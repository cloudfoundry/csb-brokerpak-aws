package acceptance_tests_test

import (
	"csbbrokerpakaws/acceptance-tests/helpers/apps"
	dms2 "csbbrokerpakaws/acceptance-tests/helpers/awscli/dms"
	"csbbrokerpakaws/acceptance-tests/helpers/random"
	"csbbrokerpakaws/acceptance-tests/helpers/services"

	"github.com/mitchellh/mapstructure"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("PostgreSQL data migration", Label("postgresql-migration"), func() {
	It("can migrate data from the previous broker to the CSB", func() {
		By("creating a replication instance")
		replicationInstance := dms2.CreateReplicationInstance(metadata.VPC, metadata.Name, metadata.Region)
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
		sourceEndpoint := dms2.CreateEndpoint(dms2.CreateEndpointParams{
			EndpointType:    dms2.Source,
			EnvironmentName: metadata.Name,
			Username:        sourceReceiver.Username,
			Password:        sourceReceiver.Password,
			Server:          sourceReceiver.Server,
			DatabaseName:    sourceReceiver.DBName,
			Region:          metadata.Region,
			Engine:          "postgres",
			Port:            sourceReceiver.Port,
		})
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
		targetEndpoint := dms2.CreateEndpoint(dms2.CreateEndpointParams{
			EndpointType:    dms2.Target,
			EnvironmentName: metadata.Name,
			Username:        targetReceiver.Username,
			Password:        targetReceiver.Password,
			Server:          targetReceiver.Server,
			DatabaseName:    targetReceiver.DBName,
			Region:          metadata.Region,
			Engine:          "postgres",
			Port:            targetReceiver.Port,
		})
		defer targetEndpoint.Cleanup()

		By("running the replication task")
		dms2.RunReplicationTask(replicationInstance, sourceEndpoint, targetEndpoint, metadata.Region, schema)

		By("deleting the target service key to trigger data ownership update")
		targetKey.Delete()

		By("checking that the data is available in the target database")
		targetApp := apps.Push(apps.WithApp(apps.PostgreSQL))
		defer targetApp.Delete()
		targetServiceInstance.Bind(targetApp)
		targetApp.Start()
		got := targetApp.GET("%s/%s", schema, key).String()
		Expect(got).To(Equal(value))
	})
})
