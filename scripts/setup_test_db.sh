#!/bin/bash

# Create test database
PGPASSWORD=postgres psql -h localhost -U postgres -c "DROP DATABASE IF EXISTS nimbus_test;"
PGPASSWORD=postgres psql -h localhost -U postgres -c "CREATE DATABASE nimbus_test;"

# Run migrations on test database
DB_DSN="postgres:postgres@localhost:5432/nimbus_test?sslmode=disable" make migrations/up 