BEGIN;
ALTER TABLE user_auth
ALTER created_at SET DATA TYPE TIMESTAMP,
ALTER updated_at SET DATA TYPE TIMESTAMP,
ALTER OTP_created_at SET DATA TYPE TIMESTAMP,
ALTER expires_at SET DATA TYPE TIMESTAMP;
COMMIT;