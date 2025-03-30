package migration

import (
	"context"
	"testing"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestDB(t *testing.T) *sqlx.DB {
	// Use test database connection string
	dsn := "postgres://postgres:postgres@localhost:5432/nimbus_test?sslmode=disable"
	db, err := sqlx.Connect("postgres", dsn)
	require.NoError(t, err)
	return db
}

func TestNewMigrationManager(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	cfg := Config{
		Enabled:     true,
		AutoMigrate: true,
		Path:        "./assets/migrations",
		LogLevel:    "info",
	}

	manager, err := NewMigrationManager(db, cfg)
	require.NoError(t, err)
	require.NotNil(t, manager)
	defer manager.Close()
}

func TestGetCurrentVersion(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	cfg := Config{
		Enabled:     true,
		AutoMigrate: true,
		Path:        "./assets/migrations",
		LogLevel:    "info",
	}

	manager, err := NewMigrationManager(db, cfg)
	require.NoError(t, err)
	defer manager.Close()

	version, err := manager.GetCurrentVersion()
	require.NoError(t, err)
	assert.GreaterOrEqual(t, version, uint(0))
}

func TestAutoMigrate(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	cfg := Config{
		Enabled:     true,
		AutoMigrate: true,
		Path:        "./assets/migrations",
		LogLevel:    "info",
	}

	manager, err := NewMigrationManager(db, cfg)
	require.NoError(t, err)
	defer manager.Close()

	err = manager.AutoMigrate(context.Background())
	require.NoError(t, err)
}

func TestRollback(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	cfg := Config{
		Enabled:     true,
		AutoMigrate: true,
		Path:        "./assets/migrations",
		LogLevel:    "info",
	}

	manager, err := NewMigrationManager(db, cfg)
	require.NoError(t, err)
	defer manager.Close()

	// First apply migrations
	err = manager.AutoMigrate(context.Background())
	require.NoError(t, err)

	// Then rollback
	err = manager.Rollback()
	require.NoError(t, err)
}

func TestRunMigrationsWithTransaction(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	cfg := Config{
		Enabled:     true,
		AutoMigrate: true,
		Path:        "./assets/migrations",
		LogLevel:    "info",
	}

	manager, err := NewMigrationManager(db, cfg)
	require.NoError(t, err)
	defer manager.Close()

	err = manager.RunMigrations(context.Background())
	require.NoError(t, err)
}
