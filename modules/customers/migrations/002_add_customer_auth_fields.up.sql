-- Add columns and constraints to customers table
ALTER TABLE customers ADD COLUMN user_id BIGINT; -- Initially allow NULL
ALTER TABLE customers ADD COLUMN otp_secret TEXT;
ALTER TABLE customers ADD COLUMN otp_generated_at TIMESTAMPTZ;
ALTER TABLE customers ADD COLUMN otp_attempt_count INTEGER DEFAULT 0;
ALTER TABLE customers ADD COLUMN is_otp_blocked BOOLEAN DEFAULT FALSE;
-- is_active might already exist, add only if needed and update existing NULLs
ALTER TABLE customers ADD COLUMN is_active BOOLEAN DEFAULT TRUE;
-- UPDATE customers SET is_active = TRUE WHERE is_active IS NULL;

-- IMPORTANT: Add data population logic here if needed to link existing customers to users
-- before applying the NOT NULL constraint. Example:
-- UPDATE customers c SET user_id = (SELECT u.id FROM users u WHERE u.email = c.email_or_some_linking_field) WHERE c.user_id IS NULL;

-- Add NOT NULL constraint after potential population
ALTER TABLE customers ALTER COLUMN user_id SET NOT NULL;

-- Add index and foreign key for customers.user_id (Depends on users table)
CREATE INDEX IF NOT EXISTS idx_customers_user_id ON customers(user_id);
ALTER TABLE customers ADD CONSTRAINT fk_customers_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE RESTRICT; 