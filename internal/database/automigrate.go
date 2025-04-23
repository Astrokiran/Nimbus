package database

import (
	"errors"

	"github.com/rs/zerolog/log"

	// Import all models that need to be auto-migrated
	"nimbus-service/modules/customers/models"
)

// RegisterModels returns all database models that should be registered with GORM
func RegisterModels() []interface{} {
	return []interface{}{
		// Add all models that should be auto-migrated here
		&models.Customer{},
		// Add additional models as needed
		// &users.User{},
		// &products.Product{},
	}
}

// AutoMigrateSchema uses GORM's AutoMigrate to update the database schema
// This is an alternative to SQL migrations that uses GORM's schema reflection
func AutoMigrateSchema() error {
	if DB == nil {
		return ErrDatabaseNotInitialized
	}

	log.Info().Msg("Running GORM AutoMigrate...")

	// Get all models to migrate
	models := RegisterModels()

	// Configure migration options
	// Uncomment if you need to disable foreign key constraints
	// DB = DB.Set("gorm:table_options", "ENGINE=InnoDB").Session(&gorm.Session{
	//     DisableForeignKeyConstraintWhenMigrating: true,
	// })

	// Perform the migration
	if err := DB.AutoMigrate(models...); err != nil {
		log.Error().Err(err).Msg("GORM AutoMigrate failed")
		return err
	}

	log.Info().Msg("GORM AutoMigrate completed successfully")
	return nil
}

// ErrDatabaseNotInitialized is returned when trying to access DB before initialization
var ErrDatabaseNotInitialized = errors.New("database not initialized")
