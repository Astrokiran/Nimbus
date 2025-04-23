# Nimbus Service

A Go backend service built with Chi, GORM, Zerolog, and Viper, following a modular structure.

## Prerequisites

*   Go (version 1.18+ recommended)
*   PostgreSQL database
*   `make`
*   `git`
*   [Air](https://github.com/cosmtrek/air) (for live reload during development)
*   [golang-migrate CLI](https://github.com/golang-migrate/migrate/tree/master/cmd/migrate) (for manual migration management)
*   [golangci-lint](https://golangci-lint.run/usage/install/) (for linting)

## Setup

1.  **Clone the repository:**
    ```bash
    git clone <your-repo-url>
    cd nimbus-service
    ```

2.  **Configure Environment:**
    *   Copy `configs/config.yaml` to `configs/config.local.yaml` (or set environment variables).
    *   Update `configs/config.local.yaml` (or environment variables prefixed with `APP_`) with your database DSN, server port, etc.
        *   Example environment variable: `export APP_DATABASE_DSN="postgresql://user:pass@host:port/db?sslmode=disable"`
    *   **For Makefile migration commands (`migrate-up`, `migrate-down`, `migrate-create`)**, you *must* export the database DSN as an environment variable:
        ```bash
        export DB_DSN="postgresql://user:pass@host:port/db?sslmode=disable"
        ```

3.  **Install Dependencies:**
    ```bash
    make tidy
    ```

## Running the Service

*   **Development (with live reload):**
    ```bash
    # Make sure DB_DSN is exported if needed for initial auto-migration
    make dev
    ```
    The service will run based on `configs/config.local.yaml` (if it exists) or `configs/config.yaml`, watching for file changes.

*   **Production / Simple Run:**
    ```bash
    # Build the binary
    make build

    # Run the binary (ensure config is accessible or use env vars)
    ./nimbus-service
    ```

## Makefile Targets

Run `make` or `make help` to see available commands:

*   `make build`: Build the application binary.
*   `make run`: Build and run the application.
*   `make dev`: Run with live reload (uses Air).
*   `make tidy`: Install/update Go module dependencies.
*   `make fmt`: Format Go code.
*   `make test`: Run unit tests.
*   `make vet`: Run `go vet`.
*   `make lint`: Run `golangci-lint`.
*   `make clean`: Remove build artifacts.
*   `make migrate-create MODULE=<module> NAME=<n>`: Create a new SQL migration file for the specified module.
*   `make migrate-up`: Apply all pending migrations across all modules (requires `DB_DSN` env var).
*   `make migrate-down MODULE=<module>`: Roll back the last applied migration for the specified module (requires `DB_DSN` env var).
*   `make migrate`: Run migrations using built-in migrator (no need for external CLI tool or `DB_DSN` env var).
*   `make auto-migrate`: Run migrations using GORM's AutoMigrate feature (schema reflection).

## Project Structure

*   `cmd/server/main.go`: Application entry point.
*   `configs/`: Configuration files.
*   `internal/`: Core application logic (config, DB, logger, middleware).
*   `modules/`: Feature modules (e.g., users, health), each potentially containing handlers, routes, models, and migrations.
*   `Makefile`: Build, run, test, and utility commands.
*   `.air.toml`: Configuration for live reload.
*   `go.mod`, `go.sum`: Go module files.

## Versioning

The application version is automatically set during the build process (`make build`) using the latest Git tag and commit hash. This version is exposed via the `/health` endpoint.

## Database Migrations

The service supports two approaches to database migrations:

1. **SQL-based Migrations** (Default)
   * Uses the `golang-migrate` library to execute `.sql` migration files
   * Migration files are located in `modules/<module>/migrations/`
   * Better for complex schema changes and maintaining full control over SQL
   * Supports version tracking and rollbacks

2. **GORM AutoMigrate**
   * Uses GORM's `AutoMigrate()` feature to automatically generate schema changes
   * Models are registered in `internal/database/automigrate.go`
   * Simpler approach that doesn't require writing SQL
   * Limited in handling complex migrations like column drops or data transformations

To select which approach to use:
* In `configs/config.yaml`: Set `database.use_auto_migrate: true` to use GORM AutoMigrate
* Command line: Use `make migrate` for SQL migrations or `make auto-migrate` for GORM AutoMigrate
* When using the migrate tool directly: `go run cmd/migrate/main.go` or `go run cmd/migrate/main.go -auto`

### Handling Migration Failures

If a SQL migration fails, you may see an error like:
```
Dirty database version 1. Fix and force version.
```

This happens when a migration fails partway through execution. To fix this:

1. **Using the Makefile helper:**
   ```bash
   # Set your DB_DSN environment variable
   export DB_DSN="postgresql://user:pass@host:port/db?sslmode=disable"
   
   # Fix the dirty state (replace 1 with your version number)
   make fix-dirty-migration VERSION=1
   ```

2. **Using psql directly:**
   ```bash
   # Connect to your database
   psql -U your_user -d your_database
   
   # Run this SQL command (replace 1 with your version number)
   UPDATE schema_migrations SET dirty = false WHERE version = 1;
   ```

3. **Using the golang-migrate CLI:**
   ```bash
   # Install the CLI if you haven't already
   go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
   
   # Force the version (replace 1 with your version number)
   migrate -path modules/customers/migrations -database "YOUR_DB_CONNECTION" force 1
   ```

After fixing the dirty state, you should:
1. Fix the actual SQL problem in your migration file
2. Run migrations again: `make migrate` or `make auto-migrate` 