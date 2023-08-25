package csbsqlserver_test

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/google/uuid"

	"github.com/cloudfoundry/csb-brokerpak-aws/terraform-provider-csbsqlserver/connector"
	"github.com/cloudfoundry/csb-brokerpak-aws/terraform-provider-csbsqlserver/csbsqlserver"
	"github.com/cloudfoundry/csb-brokerpak-aws/terraform-provider-csbsqlserver/testhelpers"
)

const (
	providerName = "csbsqlserver"
)

var _ = Describe("csbsqlserver_binding resource", func() {

	Context("database exists", func() {
		When("bindings are created", func() {

			It("can apply and destroy multiple bindings", func() {

				var (
					adminPassword = testhelpers.RandomPassword()
					port          = testhelpers.FreePort()
				)

				shutdownServerFn := testhelpers.StartServer(adminPassword, port)
				DeferCleanup(func() { shutdownServerFn(time.Minute) })

				resource.Test(GinkgoT(), getTestCase(adminPassword, port))
			})
		})
	})

	Context("database does not exists", func() {
		When("binding is created", func() {
			It("should create a database", func() {
				var (
					adminPassword = testhelpers.RandomPassword()
					port          = testhelpers.FreePort()
				)

				shutdownServerFn := testhelpers.StartServer(adminPassword, port, testhelpers.WithSPConfigure())
				DeferCleanup(func() { shutdownServerFn(time.Minute) })

				resource.Test(GinkgoT(), getTestCase(adminPassword, port))
			})
		})
	})
})

func getTestCase(adminPassword string, port int) resource.TestCase {
	var (
		tfStateResourceBinding1Name        = fmt.Sprintf("%s.binding1", csbsqlserver.ResourceNameKey)
		tfStateResourceBinding2Name        = fmt.Sprintf("%s.binding2", csbsqlserver.ResourceNameKey)
		bindingUser1, bindingUser2         = uuid.NewString(), uuid.NewString()
		bindingPassword1, bindingPassword2 = testhelpers.RandomPassword(), testhelpers.RandomPassword()
		databaseName                       = testhelpers.RandomDatabaseName()
		provider                           = initTestProvider()
		db                                 = testhelpers.Connect(testhelpers.AdminUser, adminPassword, databaseName, port)
	)
	return resource.TestCase{
		IsUnitTest:        true,
		ProviderFactories: getTestProviderFactories(provider),
		Steps: []resource.TestStep{{
			ResourceName: csbsqlserver.ResourceNameKey,
			Config:       testGetConfiguration(port, adminPassword, bindingUser1, bindingPassword1, bindingUser2, bindingPassword2, databaseName),
			Check: resource.ComposeTestCheckFunc(
				resource.TestCheckResourceAttr(tfStateResourceBinding1Name, "username", bindingUser1),
				resource.TestCheckResourceAttr(tfStateResourceBinding1Name, "password", bindingPassword1),
				resource.TestCheckResourceAttr(tfStateResourceBinding1Name, "roles.0", "db_accessadmin"),
				resource.TestCheckResourceAttr(tfStateResourceBinding1Name, "roles.1", "db_datareader"),
				resource.TestCheckResourceAttr(tfStateResourceBinding2Name, "username", bindingUser2),
				resource.TestCheckResourceAttr(tfStateResourceBinding2Name, "password", bindingPassword2),
				resource.TestCheckResourceAttr(tfStateResourceBinding2Name, "roles.0", "db_accessadmin"),
				resource.TestCheckResourceAttr(tfStateResourceBinding2Name, "roles.1", "db_datareader"),
				testCheckDatabaseExists(databaseName, provider),
				testCheckUserExists(db, bindingUser1),
				testCheckUserExists(db, bindingUser2),
			),
		}},
		CheckDestroy: func(state *terraform.State) error {
			for _, user := range []string{bindingUser1, bindingUser2} {
				if testhelpers.UserExists(db, user) {
					return fmt.Errorf("user unexpectedly exists: %s", user)
				}
			}
			return nil
		},
	}
}

func testCheckUserExists(db *sql.DB, username string) func(state *terraform.State) error {
	return func(state *terraform.State) error {
		if !testhelpers.UserExists(db, username) {
			return fmt.Errorf("user does not exist: %s", username)
		}
		return nil
	}
}

func testCheckDatabaseExists(databaseName string, provider *schema.Provider) func(state *terraform.State) error {
	return func(state *terraform.State) error {
		c := provider.Meta().(*connector.Connector)
		exists, err := c.CheckDatabaseExists(context.Background(), databaseName)
		if err != nil {
			return err
		}

		if !exists {
			return fmt.Errorf("database %s was not created", databaseName)
		}

		return nil
	}
}

func getTestProviderFactories(provider *schema.Provider) map[string]func() (*schema.Provider, error) {
	return map[string]func() (*schema.Provider, error){
		providerName: func() (*schema.Provider, error) {
			if provider == nil {
				return provider, errors.New("provider cannot be nil")
			}

			return provider, nil
		},
	}
}

func initTestProvider() *schema.Provider {
	testAccProvider := &schema.Provider{
		Schema: csbsqlserver.GetProviderSchema(),
		ResourcesMap: map[string]*schema.Resource{
			csbsqlserver.ResourceNameKey: csbsqlserver.BindingResource(),
		},
		ConfigureContextFunc: csbsqlserver.ProviderContextFunc,
	}
	err := testAccProvider.InternalValidate()
	Expect(err).NotTo(HaveOccurred())

	return testAccProvider
}

func testGetConfiguration(port int, adminPassword, bindingUser1, bindingPassword1, bindingUser2, bindingPassword2, databaseName string) string {
	return fmt.Sprintf(`
			provider "csbsqlserver" {
				server   = "%s"
				port     = "%d"
				database = "%s"
				username = "%s"
				password = "%s"
				encrypt  = "disable"
			}

			resource "csbsqlserver_binding" "binding1" {
				username = "%s"
				password = "%s"
				roles    = ["db_accessadmin", "db_datareader"]
			}

			resource "csbsqlserver_binding" "binding2" {
				username  = "%s"
				password  = "%s"
				roles     = ["db_accessadmin", "db_datareader"]
                depends_on = [csbsqlserver_binding.binding1]
			}`,
		testhelpers.Server,
		port,
		databaseName,
		testhelpers.AdminUser,
		adminPassword,
		bindingUser1,
		bindingPassword1,
		bindingUser2,
		bindingPassword2,
	)
}
