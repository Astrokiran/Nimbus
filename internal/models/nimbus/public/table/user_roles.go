//
// Code generated by go-jet DO NOT EDIT.
//
// WARNING: Changes to this file may cause incorrect behavior
// and will be lost if the code is regenerated
//

package table

import (
	"github.com/go-jet/jet/v2/postgres"
)

var UserRoles = newUserRolesTable("public", "user_roles", "")

type userRolesTable struct {
	postgres.Table

	// Columns
	UserID    postgres.ColumnString
	RoleID    postgres.ColumnInteger
	CreatedAt postgres.ColumnTimestamp

	AllColumns     postgres.ColumnList
	MutableColumns postgres.ColumnList
}

type UserRolesTable struct {
	userRolesTable

	EXCLUDED userRolesTable
}

// AS creates new UserRolesTable with assigned alias
func (a UserRolesTable) AS(alias string) *UserRolesTable {
	return newUserRolesTable(a.SchemaName(), a.TableName(), alias)
}

// Schema creates new UserRolesTable with assigned schema name
func (a UserRolesTable) FromSchema(schemaName string) *UserRolesTable {
	return newUserRolesTable(schemaName, a.TableName(), a.Alias())
}

// WithPrefix creates new UserRolesTable with assigned table prefix
func (a UserRolesTable) WithPrefix(prefix string) *UserRolesTable {
	return newUserRolesTable(a.SchemaName(), prefix+a.TableName(), a.TableName())
}

// WithSuffix creates new UserRolesTable with assigned table suffix
func (a UserRolesTable) WithSuffix(suffix string) *UserRolesTable {
	return newUserRolesTable(a.SchemaName(), a.TableName()+suffix, a.TableName())
}

func newUserRolesTable(schemaName, tableName, alias string) *UserRolesTable {
	return &UserRolesTable{
		userRolesTable: newUserRolesTableImpl(schemaName, tableName, alias),
		EXCLUDED:       newUserRolesTableImpl("", "excluded", ""),
	}
}

func newUserRolesTableImpl(schemaName, tableName, alias string) userRolesTable {
	var (
		UserIDColumn    = postgres.StringColumn("user_id")
		RoleIDColumn    = postgres.IntegerColumn("role_id")
		CreatedAtColumn = postgres.TimestampColumn("created_at")
		allColumns      = postgres.ColumnList{UserIDColumn, RoleIDColumn, CreatedAtColumn}
		mutableColumns  = postgres.ColumnList{UserIDColumn, RoleIDColumn, CreatedAtColumn}
	)

	return userRolesTable{
		Table: postgres.NewTable(schemaName, tableName, alias, allColumns...),

		//Columns
		UserID:    UserIDColumn,
		RoleID:    RoleIDColumn,
		CreatedAt: CreatedAtColumn,

		AllColumns:     allColumns,
		MutableColumns: mutableColumns,
	}
}
