-- Add columns and constraints to guides table
ALTER TABLE guides ADD COLUMN user_id BIGINT; -- Initially allow NULL
ALTER TABLE guides ADD COLUMN otp_secret TEXT;
ALTER TABLE guides ADD COLUMN otp_generated_at TIMESTAMPTZ;
ALTER TABLE guides ADD COLUMN otp_attempt_count INTEGER DEFAULT 0;
ALTER TABLE guides ADD COLUMN is_otp_blocked BOOLEAN DEFAULT FALSE;
-- is_active should already exist

-- IMPORTANT: Add data population logic here if needed to link existing guides to users
-- before applying the NOT NULL constraint. Example:
-- UPDATE guides g SET user_id = (SELECT u.id FROM users u WHERE u.email = g.email_field) WHERE g.user_id IS NULL;

-- Add NOT NULL constraint after potential population
ALTER TABLE guides ALTER COLUMN user_id SET NOT NULL;

-- Add index and foreign key for guides.user_id (Depends on users table)
CREATE INDEX IF NOT EXISTS idx_guides_user_id ON guides(user_id);
ALTER TABLE guides ADD CONSTRAINT fk_guides_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE RESTRICT;

-- Add unique index on guides phone number (if not already present from model tag)
CREATE UNIQUE INDEX IF NOT EXISTS idx_guide_phone ON guides(area_code, phone_number); 