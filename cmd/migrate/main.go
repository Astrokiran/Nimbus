package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/rs/zerolog/log"

	"nimbus-service/internal/config"
	"nimbus-service/internal/database"
	appLogger "nimbus-service/internal/logger"
)

func main() {
	// Command-line flags
	configPath := flag.String("config", "./configs", "Path to config directory")
	useAutoMigrate := flag.Bool("auto", false, "Use GORM's AutoMigrate instead of SQL migrations")
	moduleName := flag.String("module", "", "Module name for rollback (if empty, applies to all modules)")
	rollback := flag.Bool("rollback", false, "Rollback the last migration for the specified module")
	flag.Parse()

	// Load config
	cfg, err := config.LoadConfig(*configPath)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to load configuration")
		os.Exit(1)
	}

	// Set up logger
	appLogger.Init(cfg)

	// Override migrate settings
	cfg.Database.MigrateOnStart = true

	// Use command-line flag if provided, otherwise use config value
	autoFlag := flag.Lookup("auto")
	if autoFlag != nil && autoFlag.Value.String() != autoFlag.DefValue {
		cfg.Database.UseAutoMigrate = *useAutoMigrate
	}

	// Log migration approach
	if cfg.Database.UseAutoMigrate {
		fmt.Println("Using GORM's AutoMigrate for schema migration")
	} else {
		fmt.Println("Using modular migrations from modules/*/migrations directories")
	}

	// Initialize database connection
	fmt.Println("Running database migrations...")
	if err := database.Init(cfg); err != nil {
		// Check for dirty database error
		if strings.Contains(err.Error(), "Dirty database version") || strings.Contains(err.Error(), "dirty migration") {
			fmt.Println("\n===========================================================================")
			fmt.Println("ERROR: The database is in a dirty state due to a failed migration.")
			fmt.Println("To fix this, you have two options:")
			fmt.Println()
			fmt.Println("Option 1: Fix the migration issues and run the migrate command again with the module flag")
			fmt.Println("  Run: make migrate MODULE=<module_name>")
			fmt.Println()
			fmt.Println("Option 2: Use GORM AutoMigrate instead (may not be suitable for all cases)")
			fmt.Println("  Run: make auto-migrate")
			fmt.Println("===========================================================================")
		}
		log.Fatal().Err(err).Msg("Migration failed")
		os.Exit(1)
	}

	// If rollback flag is set, perform a rollback for the specified module
	if *rollback && *moduleName != "" {
		fmt.Printf("Rolling back last migration for module '%s'...\n", *moduleName)
		if err := database.RollbackModuleMigration(database.DB, *moduleName, log.Logger); err != nil {
			log.Fatal().Err(err).Str("module", *moduleName).Msg("Rollback failed")
			os.Exit(1)
		}
		fmt.Println("Rollback completed successfully!")
	} else {
		fmt.Println("Migrations completed successfully!")
	}
}
