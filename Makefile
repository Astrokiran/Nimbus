# Go parameters
GO := go
TARGET := nimbus-service # Output binary name
MAIN_PKG := ./cmd/server

# Git versioning - Get the latest tag or commit hash
GIT_COMMIT := $(shell git rev-parse --short HEAD)
GIT_TAG := $(shell git describe --tags --abbrev=0 2>/dev/null || echo "dev")
BUILD_DATE := $(shell date -u +'%Y-%m-%dT%H:%M:%SZ')
LD_FLAGS = -ldflags="-X main.Version=$(GIT_TAG) -X main.Commit=$(GIT_COMMIT) -X main.BuildDate=$(BUILD_DATE)"

# Tools
AIR := $(shell which air || echo "air") # Assumes air is in PATH or current dir
MIGRATE := $(shell which migrate || echo "migrate") # Assumes migrate is in PATH or current dir
GOLANGCI_LINT := $(shell which golangci-lint || echo "golangci-lint")

# Default target
.DEFAULT_GOAL := help

## ===========================================================================
## Development
## ===========================================================================

.PHONY: run
run: build ## Build and run the application
	@echo "Running $(TARGET)..."
	@./$(TARGET)

.PHONY: dev
dev: ## Run the application with live reload using air
	@echo "Starting development server with live reload (air)..."
	@$(AIR) -c .air.toml

## ===========================================================================
## Build
## ===========================================================================

.PHONY: build
build: tidy ## Build the application binary
	@echo "Building binary $(TARGET) (Version: $(GIT_TAG), Commit: $(GIT_COMMIT))..."
	@$(GO) build -v $(LD_FLAGS) -o $(TARGET) $(MAIN_PKG)

## ===========================================================================
## Dependencies & Formatting
## ===========================================================================

.PHONY: tidy
tidy: ## Tidy Go module dependencies
	@echo "Tidying dependencies..."
	@$(GO) mod tidy
	@$(GO) mod vendor # Optional: vendor dependencies

.PHONY: fmt
fmt: ## Format Go source files
	@echo "Formatting code..."
	@$(GO) fmt ./...

## ===========================================================================
## Testing & Linting
## ===========================================================================

.PHONY: test
test: ## Run tests
	@echo "Running tests..."
	@$(GO) test -v -race ./...

.PHONY: vet
vet: ## Run go vet
	@echo "Running go vet..."
	@$(GO) vet ./...

.PHONY: lint
lint: ## Run golangci-lint
	@echo "Running linter (golangci-lint)..."
	@if ! command -v $(GOLANGCI_LINT) > /dev/null; then \
		echo "golangci-lint not found. Please install it: https://golangci-lint.run/usage/install/"; \
		exit 1; \
	fi
	@$(GOLANGCI_LINT) run ./...

## ===========================================================================
## Database Migrations
## ===========================================================================

.PHONY: migrate-create
migrate-create: ## Create a new migration file for a module (Usage: make migrate-create MODULE=users NAME=add_new_field)
	@if [ -z "$(MODULE)" ]; then echo "MODULE is not set"; exit 1; fi
	@if [ -z "$(NAME)" ]; then echo "NAME is not set"; exit 1; fi
	@if ! command -v $(MIGRATE) > /dev/null; then \
		echo "migrate CLI not found. Please install it."; \
		exit 1; \
	fi
	@if [ ! -d "modules/$(MODULE)/migrations" ]; then echo "Migration directory modules/$(MODULE)/migrations does not exist."; exit 1; fi
	@echo "Creating migration \"$(NAME)\" in module \"$(MODULE)\" ..."
	@$(MIGRATE) create -ext sql -dir modules/$(MODULE)/migrations -seq $(NAME)

.PHONY: migrate
migrate: ## Run migrations using built-in migrator (no need for CLI)
	@echo "Running migrations using built-in modular migrator..."
	@$(GO) run cmd/migrate/main.go

.PHONY: migrate-module
migrate-module: ## Run migrations for a specific module (Usage: make migrate-module MODULE=users)
	@if [ -z "$(MODULE)" ]; then echo "MODULE is not set"; exit 1; fi
	@echo "Running migrations for module $(MODULE)..."
	@$(GO) run cmd/migrate/main.go -module=$(MODULE)

.PHONY: migrate-rollback
migrate-rollback: ## Rollback the last migration for a specific module (Usage: make migrate-rollback MODULE=users)
	@if [ -z "$(MODULE)" ]; then echo "MODULE is not set"; exit 1; fi
	@echo "Rolling back last migration for module $(MODULE)..."
	@$(GO) run cmd/migrate/main.go -module=$(MODULE) -rollback

.PHONY: auto-migrate
auto-migrate: ## Run migrations using GORM's AutoMigrate (schema reflection)
	@echo "Running migrations using GORM's AutoMigrate..."
	@$(GO) run cmd/migrate/main.go -auto

## ===========================================================================
## Cleaning
## ===========================================================================

.PHONY: clean
clean: ## Remove build artifacts
	@echo "Cleaning up..."
	@rm -f $(TARGET)
	@rm -rf ./vendor

## ===========================================================================
## Help
## ===========================================================================

.PHONY: help
help: ## Display this help screen
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		sort | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}' 