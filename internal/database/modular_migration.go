package database

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/rs/zerolog"
	"gorm.io/gorm"
)

// ModuleMigration represents a single migration file for a module
type ModuleMigration struct {
	Module      string
	Version     int
	UpFile      string
	DownFile    string
	UpContent   string
	DownContent string
}

// ModuleMigrationRecord represents a record in the module_migrations table
type ModuleMigrationRecord struct {
	Module  string `gorm:"primary_key"`
	Version int    `gorm:"primary_key"`
	Dirty   bool
}

// RunModularMigrations finds and applies migrations from all modules using a per-module versioning system
func RunModularMigrations(db *gorm.DB, modulesPath string, logger zerolog.Logger) error {
	logger.Info().Msg("Running modular migrations...")

	// Ensure migration table exists
	err := createModuleMigrationsTable(db)
	if err != nil {
		return fmt.Errorf("failed to create module_migrations table: %w", err)
	}

	// Find all modules with migrations
	modulesDirs, err := findModulesWithMigrations(modulesPath, logger)
	if err != nil {
		return err
	}

	if len(modulesDirs) == 0 {
		logger.Info().Msg("No modules with migrations found")
		return nil
	}

	// Apply migrations for each module
	for _, moduleDir := range modulesDirs {
		moduleErr := applyModuleMigrations(db, moduleDir, logger)
		if moduleErr != nil {
			logger.Error().Err(moduleErr).Str("module", filepath.Base(filepath.Dir(moduleDir))).Msg("Failed to apply migrations for module")
			return moduleErr
		}
	}

	return nil
}

// createModuleMigrationsTable ensures the module_migrations table exists
func createModuleMigrationsTable(db *gorm.DB) error {
	// Check if table exists
	if db.Migrator().HasTable(&ModuleMigrationRecord{}) {
		return nil
	}

	// Create table if it doesn't exist
	return db.AutoMigrate(&ModuleMigrationRecord{})
}

// findModulesWithMigrations scans the filesystem for modules with migrations
func findModulesWithMigrations(modulesPath string, logger zerolog.Logger) ([]string, error) {
	var modulesDirs []string

	err := filepath.WalkDir(modulesPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Check if the entry is a directory named "migrations"
		if d.IsDir() && d.Name() == "migrations" {
			// Check if the directory contains migration files
			files, readErr := os.ReadDir(path)
			if readErr != nil {
				logger.Error().Err(readErr).Str("path", path).Msg("Failed to read migration directory")
				return nil // Skip this directory
			}

			// Check if directory has migration files
			hasMigrations := false
			for _, file := range files {
				if strings.HasSuffix(file.Name(), ".sql") {
					hasMigrations = true
					break
				}
			}

			if !hasMigrations {
				logger.Info().Str("path", path).Msg("Skipping empty migration directory")
				return filepath.SkipDir
			}

			moduleDir := filepath.Dir(path) // Get the parent module directory
			moduleName := filepath.Base(moduleDir)
			logger.Info().Str("module", moduleName).Str("path", path).Msg("Found migrations")
			modulesDirs = append(modulesDirs, path)
			return filepath.SkipDir // Don't walk further down
		}
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("error scanning modules for migrations: %w", err)
	}

	return modulesDirs, nil
}

// applyModuleMigrations applies migrations for a specific module
func applyModuleMigrations(db *gorm.DB, migrationPath string, logger zerolog.Logger) error {
	// Get module name
	moduleDir := filepath.Dir(migrationPath)
	moduleName := filepath.Base(moduleDir)
	moduleLogger := logger.With().Str("module", moduleName).Logger()

	// Load migrations for this module
	migrations, err := loadModuleMigrations(migrationPath, moduleName, moduleLogger)
	if err != nil {
		return err
	}

	if len(migrations) == 0 {
		moduleLogger.Info().Msg("No migrations found for module")
		return nil
	}

	// Sort migrations by version
	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].Version < migrations[j].Version
	})

	// Get current migration version for this module
	var currentVersion int
	var dirty bool
	err = getCurrentMigrationVersion(db, moduleName, &currentVersion, &dirty)
	if err != nil {
		return err
	}

	// If there's a dirty migration, fail
	if dirty {
		return fmt.Errorf("dirty migration detected for module %s at version %d", moduleName, currentVersion)
	}

	// Apply all migrations above the current version
	for _, migration := range migrations {
		if migration.Version <= currentVersion {
			moduleLogger.Debug().Int("version", migration.Version).Msg("Skipping already applied migration")
			continue
		}

		moduleLogger.Info().Int("version", migration.Version).Msg("Applying migration")

		// Begin transaction
		tx := db.Begin()
		if tx.Error != nil {
			return fmt.Errorf("failed to begin transaction: %w", tx.Error)
		}

		// Mark migration as dirty
		if err := markMigrationDirty(tx, moduleName, migration.Version, true); err != nil {
			tx.Rollback()
			return err
		}

		// Apply migration
		result := tx.Exec(migration.UpContent)
		if result.Error != nil {
			tx.Rollback()
			moduleLogger.Error().Err(result.Error).Int("version", migration.Version).Msg("Migration failed")
			return fmt.Errorf("failed to apply migration %d for module %s: %w", migration.Version, moduleName, result.Error)
		}

		// Mark migration as applied and not dirty
		if err := markMigrationDirty(tx, moduleName, migration.Version, false); err != nil {
			tx.Rollback()
			return err
		}

		// Commit transaction
		if err := tx.Commit().Error; err != nil {
			return fmt.Errorf("failed to commit transaction: %w", err)
		}

		moduleLogger.Info().Int("version", migration.Version).Msg("Successfully applied migration")
		currentVersion = migration.Version
	}

	return nil
}

// loadModuleMigrations loads all migration files for a module
func loadModuleMigrations(migrationPath, moduleName string, logger zerolog.Logger) ([]ModuleMigration, error) {
	var migrations []ModuleMigration
	upFiles := make(map[int]string)
	downFiles := make(map[int]string)

	// Read directory
	files, err := os.ReadDir(migrationPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read migration directory: %w", err)
	}

	// Sort files by name
	fileNames := make([]string, 0, len(files))
	for _, file := range files {
		if !file.IsDir() {
			fileNames = append(fileNames, file.Name())
		}
	}
	sort.Strings(fileNames)

	// Process files
	for _, fileName := range fileNames {
		if !strings.HasSuffix(fileName, ".sql") {
			continue
		}

		// Parse version from filename (expecting format like 001_name.up.sql or 001_name.down.sql)
		parts := strings.Split(fileName, "_")
		if len(parts) < 2 {
			logger.Warn().Str("file", fileName).Msg("Skipping file with invalid name format")
			continue
		}

		versionStr := parts[0]
		version, err := strconv.Atoi(versionStr)
		if err != nil {
			logger.Warn().Str("file", fileName).Msg("Skipping file with invalid version")
			continue
		}

		// Check if it's an up or down migration
		if strings.HasSuffix(fileName, ".up.sql") {
			upFiles[version] = filepath.Join(migrationPath, fileName)
		} else if strings.HasSuffix(fileName, ".down.sql") {
			downFiles[version] = filepath.Join(migrationPath, fileName)
		}
	}

	// Create migration objects
	for version, upFile := range upFiles {
		downFile, hasDown := downFiles[version]
		if !hasDown {
			logger.Warn().Int("version", version).Msg("Missing down migration file")
			downFile = ""
		}

		// Read file contents
		upContent, err := os.ReadFile(upFile)
		if err != nil {
			return nil, fmt.Errorf("failed to read up migration file: %w", err)
		}

		var downContent []byte
		if downFile != "" {
			downContent, err = os.ReadFile(downFile)
			if err != nil {
				return nil, fmt.Errorf("failed to read down migration file: %w", err)
			}
		}

		migrations = append(migrations, ModuleMigration{
			Module:      moduleName,
			Version:     version,
			UpFile:      upFile,
			DownFile:    downFile,
			UpContent:   string(upContent),
			DownContent: string(downContent),
		})
	}

	return migrations, nil
}

// getCurrentMigrationVersion gets the current migration version for a module
func getCurrentMigrationVersion(db *gorm.DB, moduleName string, version *int, dirty *bool) error {
	var record ModuleMigrationRecord
	result := db.Where("module = ?", moduleName).Order("version DESC").First(&record)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			// No migrations applied yet
			*version = 0
			*dirty = false
			return nil
		}
		return fmt.Errorf("failed to get current migration version: %w", result.Error)
	}

	*version = record.Version
	*dirty = record.Dirty
	return nil
}

// markMigrationDirty marks a migration as applied and sets the dirty flag
func markMigrationDirty(tx *gorm.DB, moduleName string, version int, dirty bool) error {
	// Check if record exists
	var count int64
	result := tx.Model(&ModuleMigrationRecord{}).Where("module = ? AND version = ?", moduleName, version).Count(&count)
	if result.Error != nil {
		return fmt.Errorf("failed to check for existing migration record: %w", result.Error)
	}

	// Update or insert record
	if count > 0 {
		result = tx.Model(&ModuleMigrationRecord{}).Where("module = ? AND version = ?", moduleName, version).Update("dirty", dirty)
	} else {
		result = tx.Create(&ModuleMigrationRecord{
			Module:  moduleName,
			Version: version,
			Dirty:   dirty,
		})
	}

	if result.Error != nil {
		return fmt.Errorf("failed to mark migration as %s: %w", map[bool]string{true: "dirty", false: "clean"}[dirty], result.Error)
	}

	return nil
}

// RollbackModuleMigration rolls back the last migration for a specific module
func RollbackModuleMigration(db *gorm.DB, moduleName string, logger zerolog.Logger) error {
	moduleLogger := logger.With().Str("module", moduleName).Logger()
	moduleLogger.Info().Msg("Rolling back the last migration")

	// Get the latest migration record
	var record ModuleMigrationRecord
	result := db.Where("module = ?", moduleName).Order("version DESC").First(&record)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			moduleLogger.Info().Msg("No migrations to roll back")
			return nil
		}
		return fmt.Errorf("failed to get latest migration record: %w", result.Error)
	}

	// Check if migration is dirty
	if record.Dirty {
		return fmt.Errorf("cannot roll back dirty migration (version %d)", record.Version)
	}

	// Find the down migration file
	migrationPath := filepath.Join("modules", moduleName, "migrations")
	files, err := os.ReadDir(migrationPath)
	if err != nil {
		return fmt.Errorf("failed to read migration directory: %w", err)
	}

	var downFile string
	versionStr := fmt.Sprintf("%03d", record.Version)
	for _, file := range files {
		if strings.HasPrefix(file.Name(), versionStr) && strings.HasSuffix(file.Name(), ".down.sql") {
			downFile = filepath.Join(migrationPath, file.Name())
			break
		}
	}

	if downFile == "" {
		return fmt.Errorf("down migration file not found for version %d", record.Version)
	}

	// Read down migration content
	downContent, err := os.ReadFile(downFile)
	if err != nil {
		return fmt.Errorf("failed to read down migration file: %w", err)
	}

	// Begin transaction
	tx := db.Begin()
	if tx.Error != nil {
		return fmt.Errorf("failed to begin transaction: %w", tx.Error)
	}

	// Mark migration as dirty
	if err := markMigrationDirty(tx, moduleName, record.Version, true); err != nil {
		tx.Rollback()
		return err
	}

	// Apply down migration
	result = tx.Exec(string(downContent))
	if result.Error != nil {
		tx.Rollback()
		moduleLogger.Error().Err(result.Error).Int("version", record.Version).Msg("Rollback failed")
		return fmt.Errorf("failed to apply down migration for version %d: %w", record.Version, result.Error)
	}

	// Delete migration record
	result = tx.Delete(&ModuleMigrationRecord{}, "module = ? AND version = ?", moduleName, record.Version)
	if result.Error != nil {
		tx.Rollback()
		return fmt.Errorf("failed to delete migration record: %w", result.Error)
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	moduleLogger.Info().Int("version", record.Version).Msg("Successfully rolled back migration")
	return nil
}
