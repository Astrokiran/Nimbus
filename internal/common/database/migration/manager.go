package migration

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/jmoiron/sqlx"
)

// MigrationManager handles database migrations
type MigrationManager struct {
	db      *sqlx.DB
	migrate *migrate.Migrate
	path    string
}

// Config holds migration configuration
type Config struct {
	Enabled     bool
	AutoMigrate bool
	Path        string
	LogLevel    string
}

// NewMigrationManager creates a new migration manager instance
func NewMigrationManager(db *sqlx.DB, cfg Config) (*MigrationManager, error) {
	if !cfg.Enabled {
		return nil, fmt.Errorf("migrations are disabled")
	}

	driver, err := postgres.WithInstance(db.DB, &postgres.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to create migration driver: %w", err)
	}

	fs := os.DirFS(cfg.Path)
	source, err := iofs.New(fs, ".")
	if err != nil {
		return nil, fmt.Errorf("failed to create migration source: %w", err)
	}

	m, err := migrate.NewWithInstance(
		"iofs", source,
		"postgres", driver,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create migration instance: %w", err)
	}

	return &MigrationManager{
		db:      db,
		migrate: m,
		path:    cfg.Path,
	}, nil
}

// AutoMigrate automatically runs pending migrations
func (m *MigrationManager) AutoMigrate(ctx context.Context) error {
	if err := m.migrate.Up(); err != nil {
		if err == migrate.ErrNoChange {
			log.Println("No pending migrations to apply")
			return nil
		}
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	log.Println("Successfully applied all pending migrations")
	return nil
}

// GetCurrentVersion returns the current database version
func (m *MigrationManager) GetCurrentVersion() (uint, error) {
	version, dirty, err := m.migrate.Version()
	if err != nil {
		return 0, fmt.Errorf("failed to get current version: %w", err)
	}

	if dirty {
		return version, fmt.Errorf("database is in a dirty state at version %d", version)
	}

	return version, nil
}

// GetLatestVersion returns the latest available migration version
func (m *MigrationManager) GetLatestVersion() (uint, error) {
	version, dirty, err := m.migrate.Version()
	if err != nil {
		return 0, fmt.Errorf("failed to get latest version: %w", err)
	}
	if dirty {
		return version, fmt.Errorf("database is in a dirty state at version %d", version)
	}
	return version, nil
}

// RunMigrations executes pending migrations with transaction support
func (m *MigrationManager) RunMigrations(ctx context.Context) error {
	tx, err := m.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	if err := m.migrate.Up(); err != nil {
		if err == migrate.ErrNoChange {
			log.Println("No pending migrations to apply")
			return tx.Commit()
		}
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	log.Println("Successfully applied all pending migrations")
	return nil
}

// Rollback rolls back the last applied migration
func (m *MigrationManager) Rollback() error {
	if err := m.migrate.Steps(-1); err != nil {
		return fmt.Errorf("failed to rollback migration: %w", err)
	}
	log.Println("Successfully rolled back last migration")
	return nil
}

// Close closes the migration manager and releases resources
func (m *MigrationManager) Close() error {
	sourceErr, driverErr := m.migrate.Close()
	if sourceErr != nil {
		return fmt.Errorf("failed to close migration source: %w", sourceErr)
	}
	if driverErr != nil {
		return fmt.Errorf("failed to close migration driver: %w", driverErr)
	}
	return nil
}
