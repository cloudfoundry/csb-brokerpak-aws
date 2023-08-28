package connector

import (
	"context"
	"database/sql"
	"fmt"
)

type ObjectReassigner struct {
	dbConnector *DBConnector
}

func NewObjectReassigner(connector *DBConnector) *ObjectReassigner {
	return &ObjectReassigner{dbConnector: connector}

}

func (o *ObjectReassigner) ManageObjectReassignment(ctx context.Context, username string) error {
	// It needs to be granted to the **user** database and not the **master**.
	// We change the database context through the connection.
	// We must use the specific database connection!! (withConnection or withTransaction instead of withDefaultDBConnection)
	// otherwise when trying to add the user to the role, we will not have permissions.
	// Remember that the user authentication information is stored in each database.
	exists, err := o.checkObjectReassignmentRoleExist(ctx)
	if err != nil {
		return err
	}

	if !exists {
		if err := o.createObjectReassignmentRole(ctx); err != nil {
			return err
		}
	}

	if err := o.addUserToObjectReassignmentRole(ctx, username); err != nil {
		return err
	}

	schemas, err := o.getSchemaNamesByUserOwner(ctx, username)
	if err != nil {
		return err
	}

	if err := o.transferOwnershipToObjectReassignmentRole(ctx, schemas); err != nil {
		return err
	}

	return nil
}

func (o *ObjectReassigner) checkObjectReassignmentRoleExist(ctx context.Context) (result bool, err error) {
	return result, o.dbConnector.withConnection(func(db *sql.DB) (err error) {
		statement := `SELECT 1 FROM sys.database_principals WHERE type_desc = 'DATABASE_ROLE' AND name = @p1`
		rows, err := db.QueryContext(ctx, statement, roleName)
		if err != nil {
			return fmt.Errorf("error querying existence of role %q: %w", roleName, err)
		}
		defer rows.Close()

		result = rows.Next()
		return nil
	})
}

func (o *ObjectReassigner) createObjectReassignmentRole(ctx context.Context) error {
	return o.dbConnector.withConnection(func(db *sql.DB) (err error) {
		statement := fmt.Sprintf("CREATE ROLE %s AUTHORIZATION dbo", roleName)
		if _, err := db.ExecContext(ctx, statement); err != nil {
			return fmt.Errorf("error creating role %q: %w", roleName, err)
		}
		return nil
	})
}

func (o *ObjectReassigner) addUserToObjectReassignmentRole(ctx context.Context, username string) error {
	return o.dbConnector.withConnection(func(db *sql.DB) (err error) {
		statement := `
DECLARE @sql nvarchar(max)
SET @sql = 'ALTER ROLE ' + QuoteName(@roleName) + ' ADD MEMBER ' + QuoteName(@username) + ' '
EXEC (@sql)
`
		if _, err := db.ExecContext(ctx, statement, sql.Named("roleName", roleName), sql.Named("username", username)); err != nil {
			return fmt.Errorf("error adding user %q to role %q: %w", username, roleName, err)
		}
		return nil
	})
}

func (o *ObjectReassigner) getSchemaNamesByUserOwner(ctx context.Context, username string) (schemaNames []string, err error) {
	return schemaNames, o.dbConnector.withConnection(func(db *sql.DB) (err error) {
		statement := `SELECT schema_name FROM information_schema.schemata WHERE schema_owner=@p1`
		rows, err := db.QueryContext(ctx, statement, username)
		if err != nil {
			return fmt.Errorf("error querying existence of schemas by username %q: %w", username, err)
		}
		defer rows.Close()

		for rows.Next() {
			var s string
			if err := rows.Scan(&s); err != nil {
				return fmt.Errorf("error scanning schema %w", err)
			}
			schemaNames = append(schemaNames, s)
		}

		return nil
	})
}

func (o *ObjectReassigner) transferOwnershipToObjectReassignmentRole(ctx context.Context, schemas []string) error {
	if len(schemas) == 0 {
		return nil
	}

	return o.dbConnector.withTransaction(func(tx *sql.Tx) (err error) {
		for _, schema := range schemas {
			statement := `
DECLARE @sql nvarchar(max)
SET @sql = 'ALTER AUTHORIZATION ON SCHEMA::' + QuoteName(@schemaName) + ' TO ' + QuoteName(@roleName) + ' '
EXEC (@sql)`
			if _, err := tx.ExecContext(ctx, statement, sql.Named("schemaName", schema), sql.Named("roleName", roleName)); err != nil {
				return fmt.Errorf("error transfering ownership of the %q schema to the %s role: %w", schema, roleName, err)
			}
		}
		return nil
	})
}
