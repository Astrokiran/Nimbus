-- +migrate Up
ALTER TABLE users ADD COLUMN mobile VARCHAR(255) UNIQUE; -- Initially add as nullable
UPDATE users SET mobile = 'DEFAULT_MOBILE_' || id::text WHERE mobile IS NULL; -- Add a temporary unique value for existing rows. Adjust 'DEFAULT_MOBILE_' as needed.
ALTER TABLE users ALTER COLUMN mobile SET NOT NULL; -- Now enforce NOT NULL

ALTER TABLE users ALTER COLUMN name DROP NOT NULL;
ALTER TABLE users ALTER COLUMN email DROP NOT NULL;
ALTER TABLE users ALTER COLUMN password DROP NOT NULL; 