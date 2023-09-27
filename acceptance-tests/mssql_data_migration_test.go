package acceptance_tests_test

import (
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/onsi/gomega/gexec"

	"csbbrokerpakaws/acceptance-tests/helpers/apps"
	"csbbrokerpakaws/acceptance-tests/helpers/dms"
	"csbbrokerpakaws/acceptance-tests/helpers/random"
	"csbbrokerpakaws/acceptance-tests/helpers/services"

	"github.com/mitchellh/mapstructure"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

const (
	envMasterUsername = "MasterUsername"
)

// The legacy_data_assister binary must be installed. See script in CI project and instructions on the Wiki
var _ = Describe("MSSQL data migration", Label("mssql-migration"), func() {
	var masterUsername string

	It("can migrate data from the previous broker to the CSB", func() {
		By("reading legacy broker RDS data")
		masterUsername = os.Getenv(envMasterUsername)
		Expect(masterUsername).NotTo(BeEmpty(), "The MasterUsername environment variable is mandatory")

		By("creating a replication instance")
		replicationInstance := dms.CreateReplicationInstance(metadata.VPC, metadata.Name, metadata.Region)
		defer replicationInstance.Cleanup()

		By("creating a legacy service instance with the previous broker to serve as source")
		legacyServiceInstance := services.CreateInstance(
			"aws-rds-sqlserver",
			services.WithPlan("premium"),
			services.WithBrokerName("aws-services-broker"),
			services.WithParameters(map[string]any{
				"CreateDbInstance": map[string]any{
					"EngineVersion": "14.00.3401.7.v1",
					"MultiAZ":       false,
				},
			}),
		)
		defer legacyServiceInstance.Delete()

		By("pushing and binding an app to the legacy service instance")
		golangAppOne := apps.Push(apps.WithApp(apps.MSSQL))
		defer apps.Delete(golangAppOne)

		By("binding the apps to the legacy service instance")
		legacyBinding := legacyServiceInstance.Bind(golangAppOne)

		By("starting the app")
		apps.Start(golangAppOne)
		By("creating a schema using the app")
		schema := random.Name(random.WithMaxLength(10))
		golangAppOne.PUT("", "%s?tls=disable", schema)

		By("setting a key-value using the app")
		key := random.Hexadecimal()
		value := random.Hexadecimal()
		golangAppOne.PUT(value, "%s/%s?tls=disable", schema, key)

		By("reading value from legacy service instance")
		got := golangAppOne.GET("%s/%s?tls=disable", schema, key).String()
		Expect(got).To(Equal(value))

		By("waiting for the replication instance to be ready")
		replicationInstance.Wait()

		legacyServiceInstanceUUID := legacyServiceInstance.GUID()
		sourceAdminPassword := getPassword(
			"/bin/bash", "-c",
			fmt.Sprintf(
				`/usr/local/bin/legacy_data_assister --instance-uuid="%s"`,
				legacyServiceInstanceUUID,
			),
		)

		Expect(sourceAdminPassword).NotTo(BeEmpty())

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
		sourceEndpoint := dms.CreateEndpoint(dms.CreateEndpointParams{
			EndpointType:    dms.Source,
			EnvironmentName: metadata.Name,
			Username:        masterUsername,
			Password:        sourceAdminPassword,
			Server:          sourceReceiver.Server,
			DatabaseName:    sourceReceiver.DBName,
			Region:          metadata.Region,
			Engine:          "sqlserver",
			Port:            sourceReceiver.Port,
		})
		defer sourceEndpoint.Cleanup()

		By("creating a target service instance using the CSB")
		csbServiceInstance := services.CreateInstance(
			"csb-aws-mssql",
			services.WithPlan("default"),
			services.WithParameters(
				map[string]any{
					"multi_az":                false,
					"backup_retention_period": 0,
					"require_ssl":             false,
				},
			),
		)
		defer csbServiceInstance.Delete()

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
		targetEndpoint := dms.CreateEndpoint(dms.CreateEndpointParams{
			EndpointType:    dms.Target,
			EnvironmentName: metadata.Name,
			Username:        targetReceiver.Username,
			Password:        targetReceiver.Password,
			Server:          targetReceiver.Server,
			DatabaseName:    targetReceiver.DBName,
			Region:          metadata.Region,
			Engine:          "sqlserver",
			Port:            targetReceiver.Port,
		})
		defer targetEndpoint.Cleanup()

		By("running the replication task")
		dms.RunReplicationTask(replicationInstance, sourceEndpoint, targetEndpoint, metadata.Region, schema)

		By("switching the app data source from the legacy MSSQL to the new CSB created MSSQL")
		legacyBinding.Unbind()
		csbServiceInstance.Bind(golangAppOne)
		apps.Restart(golangAppOne)

		By("checking that the data is available in the target CSB created database")

		Expect(golangAppOne.GET("%s/%s?tls=disable", schema, key).String()).To(Equal(value))
	})
})

func getPassword(cmd string, args ...string) string {
	command := exec.Command(cmd, args...)
	session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
	Expect(err).NotTo(HaveOccurred())
	Eventually(session, time.Minute).Should(gexec.Exit(0))
	Expect(len(session.Err.Contents())).To(BeNumerically("==", 0), fmt.Sprintf("unexpected error: %s", session.Err.Contents()))

	out := string(session.Out.Contents())
	return out
}
