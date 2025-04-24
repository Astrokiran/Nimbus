-- Remove is_active column from customers table if it exists
ALTER TABLE customers DROP COLUMN IF EXISTS is_active; 