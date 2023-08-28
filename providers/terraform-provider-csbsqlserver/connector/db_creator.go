package connector

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/hashicorp/terraform-plugin-log/tflog"
)

type DBCreator struct {
	dbConnector *DBConnector
}

func NewDBCreator(connector *DBConnector) *DBCreator {
	return &DBCreator{dbConnector: connector}
}

func (c *DBCreator) ManageDBCreation(ctx context.Context, dbName string) error {
	exists, err := c.checkDatabaseExists(ctx, dbName)
	if err != nil {
		return err
	}

	if exists {
		return nil
	}

	// First binding => future provision phase
	// Transact-SQL Statements not allowed in a Transaction:
	// https://learn.microsoft.com/en-us/previous-versions/sql/sql-server-2008-r2/ms191544(v=sql.105)
	tflog.Debug(ctx, "creating database")
	if err := c.createDatabase(ctx, dbName); err != nil {
		return err
	}

	tflog.Debug(ctx, "setting auto close option")
	if err := c.setAutoClose(ctx, dbName); err != nil {
		return err
	}

	return nil
}

func (c *DBCreator) checkDatabaseExists(ctx context.Context, dbName string) (result bool, err error) {
	return result, c.dbConnector.withDefaultDBConnection(func(db *sql.DB) (err error) {
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

func (c *DBCreator) createDatabase(ctx context.Context, dbName string) error {
	statement := `
DECLARE @sql nvarchar(max)
SET @sql = 'CREATE DATABASE ' + QuoteName(@databaseName) + ' CONTAINMENT=PARTIAL' 
EXEC (@sql)
`
	return c.dbConnector.withDefaultDBConnection(func(db *sql.DB) (err error) {
		if _, err := db.ExecContext(ctx, statement, sql.Named("databaseName", dbName)); err != nil {
			return fmt.Errorf("error creating database %q: %w", dbName, err)
		}
		return nil
	})
}

func (c *DBCreator) setAutoClose(ctx context.Context, dbName string) error {
	statement := `
DECLARE @sql nvarchar(max)
SET @sql = 'ALTER DATABASE ' + QuoteName(@databaseName) + ' SET AUTO_CLOSE OFF' 
EXEC (@sql)
`
	return c.dbConnector.withDefaultDBConnection(func(db *sql.DB) (err error) {
		if _, err := db.ExecContext(ctx, statement, sql.Named("databaseName", dbName)); err != nil {
			return fmt.Errorf("error setting autoclose %q: %w", dbName, err)
		}
		return nil
	})
}
