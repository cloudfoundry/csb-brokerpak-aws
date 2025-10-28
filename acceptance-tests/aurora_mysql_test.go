package acceptance_tests_test

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"csbbrokerpakaws/acceptance-tests/helpers/apps"
	"csbbrokerpakaws/acceptance-tests/helpers/awscli"
	"csbbrokerpakaws/acceptance-tests/helpers/cf"
	"csbbrokerpakaws/acceptance-tests/helpers/jdbcapp"
	"csbbrokerpakaws/acceptance-tests/helpers/matchers"
	"csbbrokerpakaws/acceptance-tests/helpers/random"
	"csbbrokerpakaws/acceptance-tests/helpers/services"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("Aurora MySQL", Label("aurora-mysql"), func() {
	It("can be accessed by an app", Label("JDBC-m"), func() {
		var (
			userIn, userOut jdbcapp.AppResponseUser
			sslInfo         jdbcapp.MySQLSSLInfo
		)

		By("creating a service instance")
		params := map[string]any{
			"cluster_instances":          2,
			"instance_class":             "db.t3.medium",
			"engine_version":             "8.0.mysql_aurora.3.04.2",
			"auto_minor_version_upgrade": false,
		}

		serviceInstance := services.CreateInstance(
			"csb-aws-aurora-mysql",
			services.WithPlan("default"),
			services.WithParameters(params))
		defer serviceInstance.Delete()

		By("pushing the unstarted apps")
		manifest := jdbcapp.ManifestFor(jdbcapp.MySQL)
		appWriter := apps.Push(apps.WithApp(apps.JDBCTestAppMysql), apps.WithManifest(manifest))
		appReader := apps.Push(apps.WithApp(apps.JDBCTestAppMysql), apps.WithManifest(manifest))
		defer apps.Delete(appWriter, appReader)

		By("binding the the writer app")
		binding := serviceInstance.Bind(appWriter)

		By("starting the writer app")
		apps.Start(appWriter)

		By("checking that the app environment has a credhub reference for credentials")
		Expect(binding.Credential()).To(matchers.HaveCredHubRef)

		By("creating an entry using the writer app")
		value := random.Hexadecimal()
		appWriter.POSTf("", "?name=%s", value).ParseInto(&userIn)

		By("binding the reader app to the reader endpoint")
		serviceInstance.Bind(appReader, services.WithBindParameters(map[string]any{"reader_endpoint": true}))

		By("starting the reader app")
		apps.Start(appReader)

		By("getting the entry using the reader app")
		appReader.GETf("%d", userIn.ID).ParseInto(&userOut)
		Expect(userOut.Name).To(Equal(value), "The first app stored [%s] as the value, the second app retrieved [%s]", value, userOut.Name)

		By("verifying the DB connection utilises TLS")
		appWriter.GET("mysql-ssl").ParseInto(&sslInfo)

		Expect(strings.ToLower(sslInfo.VariableName)).To(Equal("ssl_cipher"))
		Expect(sslInfo.Value).NotTo(BeEmpty())

		By("deleting the entry using the writer app")
		appWriter.DELETEf("%d", userIn.ID)

		By("pushing and binding an app for verifying non-TLS connection attempts")
		golangApp := apps.Push(apps.WithApp(apps.MySQL))
		serviceInstance.Bind(golangApp)
		apps.Start(golangApp)

		By("verifying interactions with TLS enabled")
		key, value := "key", "value"
		golangApp.PUT(value, key)
		got := golangApp.GET(key)
		Expect(got.String()).To(Equal(value))

		By("verifying that non-TLS connections should fail")
		response := golangApp.GETResponsef("%s?tls=false", key)
		defer response.Body.Close()
		Expect(response).To(HaveHTTPStatus(http.StatusInternalServerError), "force TLS is enabled by default")
		b, err := io.ReadAll(response.Body)
		Expect(err).ToNot(HaveOccurred(), "error reading response body in TLS failure")
		Expect(string(b)).To(ContainSubstring("error connecting to database: failed to verify the connection"), "force TLS is enabled by default")
		Expect(string(b)).To(ContainSubstring("Error 1045 (28000): Access denied for user"), "mysql client cannot connect to the mysql server due to invalid TLS")
	})

	// As we introduce the 'use_managed_admin_password' feature, some users may wish to update existing DBs.
	// This is a tactical test that should exist for this changeover period and is not intended to be a forever test.
	// Due to limitations in Tofu/AWS provider/AWS the operation to switch fails first time, then succeeds on
	// a second attempt. That's not an ideal customer experience, and this test exists to ensure that what we
	// document works, and make us aware if the behavior changes.
	It("allows 'use_managed_admin_password' to be enabled", Label("managed-password"), func() {
		By("creating a service instance")
		params := map[string]any{
			"cluster_instances":          1,
			"instance_class":             "db.t3.medium",
			"engine_version":             "8.0.mysql_aurora.3.04.2",
			"auto_minor_version_upgrade": false,
		}
		serviceInstance := services.CreateInstance("csb-aws-aurora-mysql", services.WithPlan("default"), services.WithParameters(params))
		defer serviceInstance.Delete()

		By("pushing an unstarted app")
		appManifest := jdbcapp.ManifestFor(jdbcapp.MySQL)
		app := apps.Push(apps.WithApp(apps.JDBCTestAppMysql), apps.WithManifest(appManifest))
		defer apps.Delete(app)

		By("binding the app to the service instance")
		binding := serviceInstance.Bind(app)

		By("starting the app")
		apps.Start(app)

		By("creating an entry using the app")
		value := random.Hexadecimal()
		var userIn jdbcapp.AppResponseUser
		app.POSTf("", "?name=%s", value).ParseInto(&userIn)

		By("updating the service to set 'use_managed_admin_password' a first time which is expected to fail")
		const updateParams = `{"use_managed_admin_password": true}`
		session := cf.Start("update-service", serviceInstance.Name, "-c", updateParams, "--wait")
		Eventually(session).WithTimeout(time.Hour).Should(gexec.Exit(1), func() string {
			out, _ := cf.Run("service", serviceInstance.Name)
			return out
		})

		By("checking that it fails for the expected reason")
		msg, _ := cf.Run("service", serviceInstance.Name)
		Expect(msg).To(MatchRegexp(`message:\s+update failed:\s+Error:\s+Provider produced inconsistent final plan When expanding the plan for aws_secretsmanager_secret_rotation`))

		By("updating the service to set 'use_managed_admin_password' a second time")
		serviceInstance.Update(services.WithParameters(updateParams))

		By("waiting for the password rotation to be applied")
		identifier := fmt.Sprintf("csb-auroramysql-%s", serviceInstance.GUID())
		Eventually(dbClusterStatus(identifier)).WithTimeout(time.Hour).WithPolling(10 * time.Second).Should(Equal("available"))

		By("rebinding app")
		binding.Unbind()
		serviceInstance.Bind(app)
		app.Restage()

		By("getting the previously stored value")
		var userOut jdbcapp.AppResponseUser
		app.GETf("%d", userIn.ID).ParseInto(&userOut)
		Expect(userOut.Name).To(Equal(value), "App stored [%s] as the value, App retrieved [%s]", value, userOut.Name)

		By("updating the service to unset 'use_managed_admin_password'")
		serviceInstance.Update(services.WithParameters(`{"use_managed_admin_password": false}`))

		By("rebinding app")
		binding.Unbind()
		serviceInstance.Bind(app)
		app.Restage()

		By("getting the previously stored value")
		app.GETf("%d", userIn.ID).ParseInto(&userOut)
		Expect(userOut.Name).To(Equal(value), "App stored [%s] as the value, App retrieved [%s]", value, userOut.Name)
	})

	// While snapshot restore is an AWS feature rather than a CSB feature, its valuable to check that it works
	It("allows a snapshot to be restored when 'use_managed_admin_password' is enabled", Label("managed-password-snapshot-restore"), func() {
		By("creating a service instance")
		const engineVersion = "8.0.mysql_aurora.3.04.2"
		params := map[string]any{
			"use_managed_admin_password":  true,
			"rotate_admin_password_after": 4,
			"cluster_instances":           1,
			"instance_class":              "db.t3.medium",
			"engine_version":              engineVersion,
			"auto_minor_version_upgrade":  false,
		}
		serviceInstance := services.CreateInstance("csb-aws-aurora-mysql", services.WithPlan("default"), services.WithParameters(params))
		defer serviceInstance.Delete()

		By("pushing an unstarted app")
		appManifest := jdbcapp.ManifestFor(jdbcapp.MySQL)
		app := apps.Push(apps.WithApp(apps.JDBCTestAppMysql), apps.WithManifest(appManifest))
		defer apps.Delete(app)

		By("waiting for the DB cluster to be available")
		dbClusterIdentifier := fmt.Sprintf("csb-auroramysql-%s", serviceInstance.GUID())
		Eventually(dbClusterStatus(dbClusterIdentifier)).WithTimeout(time.Hour).WithPolling(10 * time.Second).Should(Equal("available"))

		By("binding the app to the service instance")
		binding := serviceInstance.Bind(app)

		By("starting the app")
		apps.Start(app)

		By("creating an entry using the app")
		value := random.Hexadecimal()
		var userIn jdbcapp.AppResponseUser
		app.POSTf("", "?name=%s", value).ParseInto(&userIn)

		// The cluster manages the storage, so we need to take a snapshot of the cluster, not the instance.
		By("taking a snapshot of the DB cluster")
		snapshotIdentifier := random.Name(random.WithPrefix("snapshot-restore-test"))
		awscli.AWS(
			"rds", "create-db-cluster-snapshot",
			"--db-cluster-snapshot-identifier", snapshotIdentifier,
			"--db-cluster-identifier", dbClusterIdentifier,
		)
		Eventually(dbClusterSnapshotStatus(snapshotIdentifier)).WithPolling(time.Minute).WithTimeout(time.Hour).Should(Equal("available"))

		deleteDBCluster(dbClusterIdentifier)

		By("waiting a bit to avoid a 'DBClusterAlreadyExists' error if we restore too soon")
		time.Sleep(time.Minute)

		// At the time of writing, there's no flag in the AWS CLI (or in the console) to enable AWS secrets manager when
		// restoring from a snapshot. Ideally we could restore with the "--manage-master-user-password" and there would
		// be no need to mess around with the settings after the restore, but that flag hasn't been added yet.
		By("restoring the DB cluster from the snapshot")
		awscli.AWS(
			"rds", "restore-db-cluster-from-snapshot",
			"--snapshot-identifier", snapshotIdentifier,
			"--db-cluster-identifier", dbClusterIdentifier,
			"--engine", "aurora-mysql",
			"--engine-version", engineVersion,
			"--db-subnet-group-name", dbSubnetGroupName(metadata.VPC, serviceInstance.GUID()),
		)

		By("waiting for the DB cluster to be available")
		Eventually(dbClusterStatus(dbClusterIdentifier)).WithTimeout(time.Hour).WithPolling(10 * time.Second).Should(Equal("available"))

		By("creating a DB instance in the restored cluster")
		dbInstanceIdentifier := fmt.Sprintf("%s-0", dbClusterIdentifier)
		awscli.AWS(
			"rds", "create-db-instance",
			"--db-instance-identifier", dbInstanceIdentifier,
			"--db-instance-class", "db.t3.medium",
			"--engine", "aurora-mysql",
			"--db-cluster-identifier", dbClusterIdentifier,
		)

		By("waiting for the DB instance to be available")
		Eventually(dbInstanceStatus(dbInstanceIdentifier)).WithTimeout(time.Hour).WithPolling(10 * time.Second).Should(Equal("available"))

		By("disabling 'use_managed_admin_password' to match the restored snapshot")
		serviceInstance.Update(services.WithParameters(`{"use_managed_admin_password": false}`))

		By("updating the service to set 'use_managed_admin_password' a first time which is expected to fail")
		const paramsJSON = `{"use_managed_admin_password": true, "rotate_admin_password_after": 4}`
		session := cf.Start("update-service", serviceInstance.Name, "-c", paramsJSON, "--wait")
		Eventually(session).WithTimeout(time.Hour).Should(gexec.Exit(1), func() string {
			out, _ := cf.Run("service", serviceInstance.Name)
			return out
		})

		By("checking that it failed for the expected reason")
		msg, _ := cf.Run("service", serviceInstance.Name)
		Expect(msg).To(MatchRegexp(`message:\s+update failed:\s+Error:\s+Provider produced inconsistent final plan When expanding the plan for aws_secretsmanager_secret_rotation`))

		By("updating the service to set 'use_managed_admin_password' a second time")
		serviceInstance.Update(services.WithParameters(paramsJSON))

		By("waiting for the password rotation to be applied")
		Eventually(dbClusterStatus(dbClusterIdentifier)).WithTimeout(time.Hour).WithPolling(10 * time.Second).Should(Equal("available"))

		By("checking that bindings created before the restore still work")
		var userOut1 jdbcapp.AppResponseUser
		app.GETf("%d", userIn.ID).ParseInto(&userOut1)
		Expect(userOut1.Name).To(Equal(value), "App stored [%s] as the value, App retrieved [%s]", value, userOut1.Name)

		By("rebinding app")
		binding.Unbind()
		serviceInstance.Bind(app)
		app.Restage()

		By("checking that new bindings can be created")
		var userOut2 jdbcapp.AppResponseUser
		app.GETf("%d", userIn.ID).ParseInto(&userOut2)
		Expect(userOut2.Name).To(Equal(value), "App stored [%s] as the value, App retrieved [%s]", value, userOut2.Name)
	})
})

func dbClusterStatus(clusterIdentifier string) func() string {
	return func() string {
		return awscli.AWSQuery("DBClusters[0].Status", "rds", "describe-db-clusters", "--db-cluster-identifier", clusterIdentifier)
	}
}

func dbClusterSnapshotStatus(id string) func() string {
	return func() string {
		status := awscli.AWSQuery("DBClusterSnapshots[0].Status", "rds", "describe-db-cluster-snapshots", "--db-cluster-snapshot-identifier", id)
		Expect(status).To(SatisfyAny(Equal("available"), Equal("creating")))
		return status
	}
}

func deleteDBCluster(dbClusterIdentifier string) {
	By("deleting the DB instances in the cluster")
	output := awscli.AWS("rds", "describe-db-clusters", "--db-cluster-identifier", dbClusterIdentifier, "--query", "DBClusters[0].DBClusterMembers[*].DBInstanceIdentifier", "--output", "text")
	dbInstanceIdentifiers := strings.Fields(string(output))

	for _, instanceID := range dbInstanceIdentifiers {
		awscli.AWS("rds", "delete-db-instance", "--db-instance-identifier", instanceID, "--skip-final-snapshot")
	}

	By("waiting for DB instances to be deleted")
	for _, instanceID := range dbInstanceIdentifiers {
		Eventually(func() *gbytes.Buffer {
			session := awscli.AWSSession("rds", "describe-db-instances", "--db-instance-identifier", instanceID, "--query", "DBInstances[0].DBInstanceStatus")
			session.Wait(5 * time.Minute)
			return session.Err
		}).WithPolling(time.Minute).WithTimeout(time.Hour).Should(gbytes.Say("not found"))
	}

	By("deleting the DB cluster")
	awscli.AWS("rds", "delete-db-cluster", "--db-cluster-identifier", dbClusterIdentifier, "--skip-final-snapshot")
	Eventually(func() *gbytes.Buffer {
		session := awscli.AWSSession("rds", "describe-db-clusters", "--db-cluster-identifier", dbClusterIdentifier, "--query", "DBClusters[0].Status")
		session.Wait(5 * time.Minute) // Should be quick, but it's occasionally slow, and we don't want to bail out when that happens
		return session.Err
	}).WithPolling(time.Minute).WithTimeout(time.Hour).Should(gbytes.Say("not found"))
}
