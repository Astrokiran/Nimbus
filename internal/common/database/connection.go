package database

import (
	"context"
	"fmt"
	"time"

	"github.com/astrokiran/nimbus/internal/common/database/migration"
	"github.com/go-jet/jet/v2/postgres"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

type Dialect struct {
	Int        func(value int64) postgres.IntegerExpression
	String     func(value string) postgres.StringExpression
	UUID       func(value fmt.Stringer) postgres.StringExpression
	ColumnList []postgres.Column
}

var PostgresDialect = Dialect{
	Int:        postgres.Int,
	String:     postgres.String,
	UUID:       postgres.UUID,
	ColumnList: []postgres.Column{},
}

type Database struct {
	Conn     *sqlx.DB
	Dialect  Dialect
	Migrator *migration.MigrationManager
}

type Config struct {
	DSN             string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	Migration       migration.Config
}

func NewDatabase(cfg Config) (*Database, error) {
	fmt.Println("Connecting to database...", cfg.DSN)
	ctx, cancel := context.WithTimeout(context.Background(), cfg.ConnMaxLifetime)
	defer cancel()

	db, err := sqlx.ConnectContext(ctx, "postgres", "postgres://"+cfg.DSN)
	if err != nil {
		return nil, err
	}

	// Set the maximum number of open connections to the database.
	db.SetMaxOpenConns(cfg.MaxOpenConns)
	// Set the maximum number of idle connections in the pool.
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	// Set the maximum lifetime of a connection.
	db.SetConnMaxLifetime(cfg.ConnMaxLifetime)

	database := &Database{
		Conn:    db,
		Dialect: PostgresDialect,
	}

	// Initialize migration manager if migrations are enabled
	if cfg.Migration.Enabled {
		migrator, err := migration.NewMigrationManager(db, cfg.Migration)
		if err != nil {
			return nil, fmt.Errorf("failed to initialize migration manager: %w", err)
		}
		database.Migrator = migrator

		// Run auto-migration if enabled
		if cfg.Migration.AutoMigrate {
			if err := migrator.AutoMigrate(ctx); err != nil {
				return nil, fmt.Errorf("failed to run auto-migrations: %w", err)
			}
		}
	}

	return database, nil
}

// Close gracefully shuts down the database connection and migration manager.
func (d *Database) Close() error {
	if d.Migrator != nil {
		if err := d.Migrator.Close(); err != nil {
			return fmt.Errorf("failed to close migration manager: %w", err)
		}
	}
	if d.Conn != nil {
		return d.Conn.Close()
	}
	return nil
}
