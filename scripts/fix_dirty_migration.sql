-- Script to fix a dirty migration state
-- Usage: Replace VERSION with the actual version number from the error message
-- Run with: psql -U your_user -d your_database -f fix_dirty_migration.sql

-- First check the current status of migrations
SELECT * FROM schema_migrations;

-- Clear the dirty flag for a specific version 
-- (replace VERSION with the actual version number from your error message)
UPDATE schema_migrations SET dirty = false WHERE version = VERSION;

-- Verify the changes
SELECT * FROM schema_migrations; 