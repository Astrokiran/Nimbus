-- +migrate Down
DROP INDEX IF EXISTS idx_users_deleted_at;
-- DROP INDEX IF EXISTS idx_users_email; -- Drop if you created it above
DROP TABLE IF EXISTS users; 