package database

import (
	"errors"
	"fmt"
	"io/fs"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres" // Driver for postgres
	_ "github.com/golang-migrate/migrate/v4/source/file"       // Driver for file source
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log" // Keep for top-level logging (e.g., connection errors)
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"

	"nimbus-service/internal/config"
	appLogger "nimbus-service/internal/logger" // Use our application's logger package
)

var DB *gorm.DB

// GormZerologWriter adapts zerolog to GORM's logger interface
type GormZerologWriter struct {
	LogLevel gormlogger.LogLevel
}

// Printf implements gorm's logger.Writer interface
func (w GormZerologWriter) Printf(message string, data ...interface{}) {
	logMessage := fmt.Sprintf(message, data...)
	// Map GORM log levels to Zerolog levels. GORM's Info includes SQL.
	switch w.LogLevel {
	// Silent should not log anything via Printf used for SQL logging
	case gormlogger.Silent:
		return
	case gormlogger.Error:
		appLogger.Logger.Error().Msg(logMessage)
	case gormlogger.Warn:
		appLogger.Logger.Warn().Msg(logMessage)
	case gormlogger.Info:
		appLogger.Logger.Info().Msg(logMessage)
	default: // Default to info level for messages passed via Printf
		appLogger.Logger.Info().Msg(logMessage)
	}
}

// Init initializes the database connection and runs migrations.
func Init(cfg *config.Config) error {
	var err error

	gormLogLevel := getGormLogLevel(cfg.Database.LogLevel)

	// Configure GORM logger using our adapter
	newLogger := gormlogger.New(
		GormZerologWriter{LogLevel: gormLogLevel}, // Use our adapter
		gormlogger.Config{
			SlowThreshold:             time.Second,  // Slow SQL threshold
			LogLevel:                  gormLogLevel, // Log level for GORM's internal decisions
			IgnoreRecordNotFoundError: true,         // Ignore ErrRecordNotFound error for logger
			Colorful:                  false,        // Disable color
		},
	)

	// Connect to the database
	DB, err = gorm.Open(postgres.Open(cfg.Database.DSN), &gorm.Config{
		Logger: newLogger,
	})
	if err != nil {
		// Use the global log here for fatal startup errors
		log.Fatal().Err(err).Msg("Failed to connect to database")
		return err // Although Fatal will exit, return for completeness
	}

	log.Info().Msg("Database connection established")

	// Run migrations if configured
	if cfg.Database.MigrateOnStart {
		log.Info().Msg("Running database migrations...")

		// Choose migration method based on configuration
		if cfg.Database.UseAutoMigrate {
			// Use GORM's AutoMigrate (schema reflection)
			log.Info().Msg("Using GORM AutoMigrate for schema migration")
			if err := AutoMigrateSchema(); err != nil {
				log.Error().Err(err).Msg("Failed to run GORM auto migrations")
				return err // Return the error to fail the process
			}
		} else {
			// Use the new modular migrations system instead of the older SQL-based one
			log.Info().Msg("Using modular migrations system")
			migrationLogger := log.With().Str("component", "migration").Logger() // Create a specific logger for migrations
			if err := RunModularMigrations(DB, "modules", migrationLogger); err != nil {
				migrationLogger.Error().Err(err).Msg("Failed to run modular migrations")
				return err // Return the error to fail the process
			}
		}
	}

	return nil
}

// getGormLogLevel converts string level to gormlogger.LogLevel
func getGormLogLevel(level string) gormlogger.LogLevel {
	switch strings.ToLower(level) {
	case "silent":
		return gormlogger.Silent
	case "error":
		return gormlogger.Error
	case "warn":
		return gormlogger.Warn
	case "info":
		return gormlogger.Info
	default:
		log.Warn().Msgf("Invalid GORM log level '%s', defaulting to 'warn'", level)
		return gormlogger.Warn
	}
}

// RunMigrations finds and applies migrations from all modules.
// Added a logger argument for better context.
func RunMigrations(databaseDSN string, modulesPath string, migrationLogger zerolog.Logger) error {
	var migrationErr error
	var modulesWithMigrations []string

	// Ensure the DSN is in the format expected by golang-migrate
	// golang-migrate requires the `postgres` or `postgresql` scheme
	migrateDSN := databaseDSN
	if !strings.HasPrefix(migrateDSN, "postgres://") && !strings.HasPrefix(migrateDSN, "postgresql://") {
		// Attempt to parse without scheme, then add if necessary
		// This handles DSNs like "user:pass@host:port/db?sslmode=disable"
		parsedURL, parseErr := url.Parse(migrateDSN)
		if parseErr != nil || parsedURL.Scheme == "" { // If parse fails or scheme missing
			migrateDSN = "postgresql://" + migrateDSN // Add default scheme
			migrationLogger.Warn().Msgf("Database DSN missing scheme, assuming 'postgresql://'. Full DSN used for migration: %s", migrateDSN)
		} else {
			// If scheme exists but isn't postgresql, golang-migrate might fail later
			if parsedURL.Scheme != "postgres" && parsedURL.Scheme != "postgresql" {
				migrationLogger.Warn().Msgf("Database DSN scheme '%s' might not be compatible with golang-migrate postgres driver.", parsedURL.Scheme)
			}
		}

	}

	// Parse DSN again with scheme to check validity early
	_, err := url.Parse(migrateDSN)
	if err != nil {
		return fmt.Errorf("invalid database DSN for migration: %w", err)
	}

	// Walk the modules directory to find migration folders
	err = filepath.WalkDir(modulesPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		// Check if the entry is a directory named "migrations"
		if d.IsDir() && d.Name() == "migrations" {
			// Check if the directory is empty or contains migration files
			files, readErr := os.ReadDir(path)
			if readErr != nil {
				migrationLogger.Error().Err(readErr).Str("path", path).Msg("Failed to read migration directory")
				// Decide whether to skip or return error
				return nil // Skip this potentially problematic directory
			}
			hasMigrations := false
			for _, file := range files {
				if strings.HasSuffix(file.Name(), ".sql") {
					hasMigrations = true
					break
				}
			}

			if !hasMigrations {
				migrationLogger.Info().Str("path", path).Msg("Skipping empty migration directory")
				return filepath.SkipDir
			}

			moduleDir := filepath.Dir(path) // Get the parent module directory
			moduleName := filepath.Base(moduleDir)
			migrationLogger.Info().Str("module", moduleName).Str("path", path).Msg("Found migrations")
			modulesWithMigrations = append(modulesWithMigrations, path)
			return filepath.SkipDir // Don't walk further down into the migrations dir itself
		}
		return nil
	})

	if err != nil {
		return fmt.Errorf("error scanning modules for migrations: %w", err)
	}

	if len(modulesWithMigrations) == 0 {
		migrationLogger.Info().Msg("No non-empty module migration directories found.")
		return nil
	}

	// Run migrations for each found directory
	for _, migrationPath := range modulesWithMigrations {
		moduleDir := filepath.Dir(migrationPath)
		moduleName := filepath.Base(moduleDir)
		mLog := migrationLogger.With().Str("module", moduleName).Str("path", migrationPath).Logger()

		// Make path absolute for file source URL consistency
		absMigrationPath, absErr := filepath.Abs(migrationPath)
		if absErr != nil {
			migrationErr = errors.Join(migrationErr, fmt.Errorf("failed to get absolute path for %s: %w", migrationPath, absErr))
			continue
		}
		sourceURL := "file://" + filepath.ToSlash(absMigrationPath) // Ensure forward slashes for URL
		mLog.Info().Msgf("Applying migrations using source '%s'", sourceURL)

		m, err := migrate.New(sourceURL, migrateDSN)
		if err != nil {
			err = fmt.Errorf("failed to create migrate instance: %w", err)
			mLog.Error().Err(err).Msg("Migration setup failed")
			migrationErr = errors.Join(migrationErr, err)
			continue // Try next module
		}

		// Assign migrate's logger to our logger (optional, provides more detail)
		m.Log = &migrateLoggerAdapter{zerolog: mLog}

		if err := m.Up(); err != nil {
			if errors.Is(err, migrate.ErrNoChange) {
				mLog.Info().Msg("No changes to apply")
			} else {
				// Combine error context
				err = fmt.Errorf("failed to apply migrations: %w", err)
				mLog.Error().Err(err).Msg("Migration failed")
				migrationErr = errors.Join(migrationErr, err)
			}
		} else {
			mLog.Info().Msg("Successfully applied migrations")
		}

		srcErr, dbErr := m.Close()
		if srcErr != nil {
			err = fmt.Errorf("error closing migration source: %w", srcErr)
			mLog.Warn().Err(err).Msg("Migration source close error")
			migrationErr = errors.Join(migrationErr, err)
		}
		if dbErr != nil {
			err = fmt.Errorf("error closing migration database connection: %w", dbErr)
			mLog.Warn().Err(err).Msg("Migration DB close error")
			migrationErr = errors.Join(migrationErr, err)
		}
	}

	return migrationErr // Return combined errors, if any
}

// migrateLoggerAdapter adapts zerolog to migrate.Logger interface
type migrateLoggerAdapter struct {
	zerolog zerolog.Logger
}

func (l *migrateLoggerAdapter) Printf(format string, v ...interface{}) {
	l.zerolog.Info().Msgf(format, v...) // Log migrate messages as Info level
}

func (l *migrateLoggerAdapter) Verbose() bool {
	// You can tie this to your app's log level if desired
	// e.g., return appLogger.Logger.GetLevel() <= zerolog.DebugLevel
	return true // Enable verbose logging from migrate library
}
