package acceptance_tests_test

import (
	"csbbrokerpakaws/acceptance-tests/helpers/apps"
	"csbbrokerpakaws/acceptance-tests/helpers/dms"
	"csbbrokerpakaws/acceptance-tests/helpers/random"
	"csbbrokerpakaws/acceptance-tests/helpers/services"

	"github.com/mitchellh/mapstructure"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("PostgreSQL data migration", Label("postgresql-migration"), func() {
	It("can migrate data from the previous broker to the CSB", func() {
		By("creating a replication instance")
		replicationInstance := dms.CreateReplicationInstance(metadata.VPC, metadata.Name, metadata.Region)
		defer replicationInstance.Cleanup()

		By("creating a source service instance with the previous broker")
		sourceServiceInstance := services.CreateInstance(
			"aws-rds-postgres",
			services.WithPlan("basic"),
			services.WithBrokerName("aws-services-broker"),
			services.WithParameters(map[string]any{
				"CreateDbInstance": map[string]string{
					"EngineVersion": "12",
				},
			}),
		)
		defer sourceServiceInstance.Delete()

		By("creating a target service instance using the CSB")
		targetServiceInstance := services.CreateInstance("csb-aws-postgresql", services.WithPlan("default"), services.WithDefaultBroker())
		defer targetServiceInstance.Delete()

		By("binding an app to the source service instance")
		sourceApp := apps.Push(apps.WithApp(apps.PostgreSQL))
		defer sourceApp.Delete()
		sourceInstanceBinding := sourceServiceInstance.Bind(sourceApp)
		sourceApp.Start()

		By("creating a schema and adding some data in the source database")
		schema := random.Name(random.WithMaxLength(8))
		sourceApp.PUT("", schema)

		key := random.Hexadecimal()
		value := random.Hexadecimal()
		sourceApp.PUT(value, "%s/%s", schema, key)

		By("waiting for the replication instance to be ready")
		replicationInstance.Wait()

		By("creating a DMS source endpoint")
		sourceCreds := sourceInstanceBinding.Credential()
		var sourceReceiver struct {
			DBName   string `mapstructure:"database"`
			Server   string `mapstructure:"hostname"`
			Username string `mapstructure:"username"`
			Password string `mapstructure:"password"`
			Port     int    `mapstructure:"port"`
		}
		Expect(mapstructure.Decode(sourceCreds, &sourceReceiver)).NotTo(HaveOccurred())
		sourceEndpoint := dms.CreateEndpoint(dms.Source, sourceReceiver.Username, sourceReceiver.Password, sourceReceiver.Server, sourceReceiver.DBName, metadata.Region, "postgres", sourceReceiver.Port)
		defer sourceEndpoint.Cleanup()

		By("creating a DMS target endpoint")
		targetKey := targetServiceInstance.CreateServiceKey()
		defer targetKey.Delete()
		var targetReceiver struct {
			Username string `json:"username"`
			Password string `json:"password"`
			Server   string `json:"hostname"`
			DBName   string `json:"name"`
			Port     int    `json:"port"`
		}
		targetKey.Get(&targetReceiver)
		targetEndpoint := dms.CreateEndpoint(dms.Target, targetReceiver.Username, targetReceiver.Password, targetReceiver.Server, targetReceiver.DBName, metadata.Region, "postgres", targetReceiver.Port)
		defer targetEndpoint.Cleanup()

		By("running the replication task")
		dms.RunReplicationTask(replicationInstance, sourceEndpoint, targetEndpoint, metadata.Region, schema)

		By("deleting the target service key to trigger data ownership update")
		targetKey.Delete()

		By("checking that the data is available in the target database")
		targetApp := apps.Push(apps.WithApp(apps.PostgreSQL))
		defer targetApp.Delete()
		targetServiceInstance.Bind(targetApp)
		targetApp.Start()
		got := targetApp.GET("%s/%s", schema, key)
		Expect(got).To(Equal(value))
	})
})
