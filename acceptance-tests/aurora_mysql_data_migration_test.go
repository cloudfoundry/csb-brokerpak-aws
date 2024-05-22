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

var _ = Describe("Aurora MySQL data migration", Label("aurora-mysql-migration"), func() {
	It("can migrate data from the previous broker to the CSB", func() {

		By("creating a legacy service instance with the previous broker to serve as source")
		legacyServiceInstance := services.CreateInstance(
			"aws-rds-aurora",
			services.WithPlan("migration"),
			services.WithBrokerName("aws-services-broker"),
		)

		defer legacyServiceInstance.Delete()

		By("pushing an app to the legacy service instance")
		app := apps.Push(apps.WithApp(apps.MySQL))
		defer app.Delete()

		By("binding an app to the legacy service instance")
		legacyBinding := legacyServiceInstance.Bind(app)
		app.Start()

		By("adding some data in the source legacy database")
		key := random.Hexadecimal()
		value := random.Hexadecimal()
		app.PUT(value, "%s?tls=true", key)

		By("creating a target service instance using the CSB")
		params := map[string]any{
			"cluster_instances":          2,
			"instance_class":             "db.t3.medium",
			"auto_minor_version_upgrade": false,
		}
		csbServiceInstance := services.CreateInstance(
			"csb-aws-aurora-mysql",
			services.WithPlan("default"),
			services.WithParameters(params),
			services.WithDefaultBroker(),
		)
		defer csbServiceInstance.Delete()

		By("creating a replication instance")
		replicationInstance := dms2.CreateReplicationInstance(metadata.VPC, metadata.Name, metadata.Region)
		defer replicationInstance.Cleanup()

		By("waiting for the replication instance to be ready")
		replicationInstance.Wait()

		By("creating a DMS source endpoint")
		sourceCreds := legacyBinding.Credential()
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
			Engine:          "aurora",
			Port:            sourceReceiver.Port,
		})
		defer sourceEndpoint.Cleanup()

		By("creating a DMS target endpoint")
		csbKey := csbServiceInstance.CreateServiceKey()
		defer csbKey.Delete()
		var targetReceiver struct {
			Username string `json:"username"`
			Password string `json:"password"`
			Server   string `json:"hostname"`
			DBName   string `json:"name"`
			Port     int    `json:"port"`
		}
		csbKey.Get(&targetReceiver)
		targetEndpoint := dms2.CreateEndpoint(dms2.CreateEndpointParams{
			EndpointType:    dms2.Target,
			EnvironmentName: metadata.Name,
			Username:        targetReceiver.Username,
			Password:        targetReceiver.Password,
			Server:          targetReceiver.Server,
			DatabaseName:    targetReceiver.DBName,
			Region:          metadata.Region,
			Engine:          "aurora",
			Port:            targetReceiver.Port,
		})
		defer targetEndpoint.Cleanup()

		By("running the replication task")
		dms2.RunReplicationTask(replicationInstance, sourceEndpoint, targetEndpoint, metadata.Region, "%")

		By("deleting the target service key to trigger data ownership update")
		csbKey.Delete()

		By("switching the app data source from the legacy MySQL to the new CSB created MySQL")
		legacyBinding.Unbind()
		csbServiceInstance.Bind(app)
		apps.Restart(app)

		By("checking that the data is available in the target CSB created database")
		Expect(app.GET(key).String()).To(Equal(value))
	})
})
