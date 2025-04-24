-- Remove constraints and columns from guides table
ALTER TABLE guides DROP CONSTRAINT IF EXISTS fk_guides_user;
DROP INDEX IF EXISTS idx_guides_user_id;
DROP INDEX IF EXISTS idx_guide_phone; -- Drop unique index added in up migration

ALTER TABLE guides DROP COLUMN IF EXISTS user_id;
ALTER TABLE guides DROP COLUMN IF EXISTS otp_secret;
ALTER TABLE guides DROP COLUMN IF EXISTS otp_generated_at;
ALTER TABLE guides DROP COLUMN IF EXISTS otp_attempt_count;
ALTER TABLE guides DROP COLUMN IF EXISTS is_otp_blocked; 