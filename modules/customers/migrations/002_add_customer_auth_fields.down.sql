-- Remove constraints and columns from customers table
ALTER TABLE customers DROP CONSTRAINT IF EXISTS fk_customers_user;
DROP INDEX IF EXISTS idx_customers_user_id;

ALTER TABLE customers DROP COLUMN IF EXISTS user_id;
ALTER TABLE customers DROP COLUMN IF EXISTS otp_secret;
ALTER TABLE customers DROP COLUMN IF EXISTS otp_generated_at;
ALTER TABLE customers DROP COLUMN IF EXISTS otp_attempt_count;
ALTER TABLE customers DROP COLUMN IF EXISTS is_otp_blocked;
-- Only drop is_active if it was added in the corresponding UP migration
-- ALTER TABLE customers DROP COLUMN IF EXISTS is_active; 