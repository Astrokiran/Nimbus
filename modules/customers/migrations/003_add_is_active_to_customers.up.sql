-- Add is_active column to customers table if it doesn't already exist
ALTER TABLE customers ADD COLUMN IF NOT EXISTS is_active BOOLEAN DEFAULT TRUE;

-- Update existing rows where is_active might be NULL (optional, depends on prior state)
-- UPDATE customers SET is_active = TRUE WHERE is_active IS NULL; 