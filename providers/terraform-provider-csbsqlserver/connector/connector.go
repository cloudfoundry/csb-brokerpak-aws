// Package connector manages the creating and deletion of service bindings to MS SQL Server
package connector

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/hashicorp/terraform-plugin-log/tflog"
)

func New(server string, port int, username, password, database, encrypt string) *Connector {
	return &Connector{
		database:    database,
		dbConnector: NewDBConnector(server, port, username, password, database, encrypt),
	}
}

type Connector struct {
	database    string
	dbConnector *DBConnector
}

// CreateBinding creates the binding user, adds roles and grants permission to execute store procedures
// It is idempotent.
func (c *Connector) CreateBinding(ctx context.Context, username, password string, roles []string) error {
	err := c.checkEngineContainmentIsEnabled(ctx)
	if err != nil {
		return err
	}

	exists, err := c.CheckDatabaseExists(ctx, c.database)
	if err != nil {
		return err
	}

	if !exists {
		// First binding => future provision phase
		// Transact-SQL Statements not allowed in a Transaction:
		// https://learn.microsoft.com/en-us/previous-versions/sql/sql-server-2008-r2/ms191544(v=sql.105)
		tflog.Debug(ctx, "creating database")
		if err := c.createDatabase(ctx, c.database); err != nil {
			return err
		}

		tflog.Debug(ctx, "setting auto close option")
		if err := c.setAutoClose(ctx, c.database); err != nil {
			return err
		}
	}

	err = c.checkDatabaseIsContained(ctx, c.database)
	if err != nil {
		return err
	}

	return c.withTransaction(func(tx *sql.Tx) error {

		if err := createUser(ctx, tx, username, password); err != nil {
			return err
		}

		if err := addRoles(ctx, tx, username, roles); err != nil {
			return err
		}

		if err := grantExec(ctx, tx, username); err != nil {
			return err
		}

		return nil
	})
}

// DeleteBinding drops the binding user. It is idempotent.
func (c *Connector) DeleteBinding(ctx context.Context, username string) error {
	return c.withTransaction(func(tx *sql.Tx) error {
		if err := dropUser(ctx, tx, username); err != nil {
			return err
		}
		if err := dropLogin(ctx, tx, username); err != nil {
			return err
		}

		return nil
	})
}

// ReadBinding checks whether the binding exists
func (c *Connector) ReadBinding(ctx context.Context, username string) (result bool, err error) {
	return result, c.withConnection(func(db *sql.DB) (err error) {
		result, err = checkUser(ctx, db, username)
		return
	})
}

func (c *Connector) CheckDatabaseExists(ctx context.Context, dbName string) (result bool, err error) {
	return result, c.withDefaultDBConnection(func(db *sql.DB) (err error) {
		statement := `SELECT 1 FROM master.dbo.sysdatabases where name=@p1`
		rows, err := db.QueryContext(ctx, statement, dbName)
		if err != nil {
			return fmt.Errorf("error querying existence of database %q: %w", dbName, err)
		}
		defer rows.Close()

		result = rows.Next()
		return nil
	})
}

func (c *Connector) createDatabase(ctx context.Context, dbName string) error {
	statement := `
DECLARE @sql nvarchar(max)
SET @sql = 'CREATE DATABASE ' + QuoteName(@databaseName) + ' CONTAINMENT=PARTIAL' 
EXEC (@sql)
`
	return c.withDefaultDBConnection(func(db *sql.DB) (err error) {
		if _, err := db.ExecContext(ctx, statement, sql.Named("databaseName", dbName)); err != nil {
			return fmt.Errorf("error creating database %q: %w", dbName, err)
		}
		return nil
	})
}

func (c *Connector) setAutoClose(ctx context.Context, dbName string) error {
	statement := `
DECLARE @sql nvarchar(max)
SET @sql = 'ALTER DATABASE ' + QuoteName(@databaseName) + ' SET AUTO_CLOSE OFF' 
EXEC (@sql)
`
	return c.withDefaultDBConnection(func(db *sql.DB) (err error) {
		if _, err := db.ExecContext(ctx, statement, sql.Named("databaseName", dbName)); err != nil {
			return fmt.Errorf("error setting autoclose %q: %w", dbName, err)
		}
		return nil
	})
}

func (c *Connector) checkEngineContainmentIsEnabled(ctx context.Context) error {
	return c.withDefaultDBConnection(func(db *sql.DB) error {
		statement := `SELECT 1 FROM sys.configurations
				WHERE [name] = N'contained database authentication'
				AND value_in_use = 1
				AND value = 1`

		var i int
		err := db.QueryRowContext(ctx, statement).Scan(&i)
		if err == sql.ErrNoRows {
			return fmt.Errorf("engine containment is not enabled %w", err)
		}

		if err != nil {
			return fmt.Errorf("error checking engine containment %w", err)
		}

		return nil
	})
}

func (c *Connector) checkDatabaseIsContained(ctx context.Context, dbName string) error {
	return c.withDefaultDBConnection(func(db *sql.DB) error {
		statement := `SELECT 1 FROM sys.databases
				WHERE [name] = @p1
				AND containment <> 0`

		var i int
		err := db.QueryRowContext(ctx, statement, dbName).Scan(&i)
		if err == sql.ErrNoRows {
			return fmt.Errorf("database %q is not contained %w", dbName, err)
		}

		if err != nil {
			return fmt.Errorf("error checking if database is contained %w", err)
		}

		return nil
	})
}

// withConnection returns a DB connection tied to the user's database connection
func (c *Connector) withConnection(callback func(*sql.DB) error) error {
	return c.dbConnector.withConnection(callback)
}

// withDefaultDBConnection returns a DB connection tied to the default database of the server
func (c *Connector) withDefaultDBConnection(callback func(*sql.DB) error) error {
	return c.dbConnector.withDefaultDBConnection(callback)
}

func (c *Connector) withTransaction(callback func(*sql.Tx) error) error {
	return c.dbConnector.withTransaction(callback)
}

type databaseActor interface {
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
}

func checkUser(ctx context.Context, db databaseActor, username string) (bool, error) {
	rows, err := db.QueryContext(ctx, `SELECT NAME FROM sys.database_principals WHERE NAME = @p1 AND TYPE = 'S'`, username)
	if err != nil {
		return false, fmt.Errorf("error querying existence of user %q: %w", username, err)
	}
	defer rows.Close()
	return rows.Next(), nil
}

func dropUser(ctx context.Context, tx *sql.Tx, username string) error {
	statement := fmt.Sprintf(`DROP USER IF EXISTS [%s]`, username)
	if _, err := tx.ExecContext(ctx, statement); err != nil {
		return fmt.Errorf("error deleting user %q: %w", username, err)
	}

	return nil
}

// dropLogin drops a legacy login. In the past the brokerpak created a login and
// then created a user from the login, but this was problematic in a failover because
// the login did not exist on the replica. See: https://www.pivotaltracker.com/story/show/179168006
func dropLogin(ctx context.Context, tx *sql.Tx, username string) error {
	// Unfortunately there's no "DROP LOGIN IF EXISTS" and checking for the existence of the login (via sys.sql_logins)
	// proved unreliable as it worked on the database fake, but not on an Azure database. So we just attempt the operation
	// and ignore any error. The issue with checking for the "right" error is that any check would be fragile if there were
	// future changes to the message, SQLState, etc... and the benefit was not considered to be worth the risk
	_, _ = tx.ExecContext(ctx, fmt.Sprintf(`DROP LOGIN [%s]`, username))

	return nil
}

func createUser(ctx context.Context, tx *sql.Tx, username, password string) error {
	ok, err := checkUser(ctx, tx, username)
	switch {
	case err != nil:
		return err
	case ok:
		return nil // already exists. In the future we could update the password with "ALTER USER".
	}

	statement := fmt.Sprintf(`CREATE USER [%s] with PASSWORD='%s'`, username, password)
	if _, err := tx.ExecContext(ctx, statement); err != nil {
		return fmt.Errorf("error creating user %q: %w", username, err)
	}

	return nil
}

func addRoles(ctx context.Context, tx *sql.Tx, username string, roles []string) error {
	for _, role := range roles {
		statement := fmt.Sprintf(`ALTER ROLE %s ADD MEMBER [%s]`, role, username)
		if _, err := tx.ExecContext(ctx, statement); err != nil {
			return fmt.Errorf("error adding user %q to role %q: %w", username, role, err)
		}
	}

	return nil
}

// grantExec grants EXECUTE permission to all the user stored procedures.
// Right out of the box, SQL Server makes it pretty easy to grant SELECT, INSERT, UPDATE, and DELETE to all user tables.
// That’s accomplished by using the built-in db_datareader (SELECT) and
// db_datawriter (INSERT, UPDATE, and DELETE) database roles in every user database.
// Any user you add to those database roles will be granted those permissions.
// But what if you want to grant EXECUTE permission to all the user stored procedures.
// Where’s the built-in database role for that? Nowhere to be found.
func grantExec(ctx context.Context, tx *sql.Tx, username string) error {
	statement := fmt.Sprintf(`GRANT EXEC TO [%s]`, username)
	if _, err := tx.ExecContext(ctx, statement); err != nil {
		return fmt.Errorf("error granting exec to user %q: %w", username, err)
	}

	return nil
}
