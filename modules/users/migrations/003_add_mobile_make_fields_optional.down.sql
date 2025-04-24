-- +migrate Down
-- Before making columns NOT NULL, ensure no NULLs exist or handle them appropriately.
-- Example: UPDATE users SET name = 'Default Name' WHERE name IS NULL;
-- Example: UPDATE users SET email = 'default_' || id::text || '@example.com' WHERE email IS NULL;
-- Example: UPDATE users SET password = 'default_password' WHERE password IS NULL;
-- The lines above are commented out as the appropriate default/handling strategy depends on your application needs.

ALTER TABLE users ALTER COLUMN password SET NOT NULL;
ALTER TABLE users ALTER COLUMN email SET NOT NULL;
ALTER TABLE users ALTER COLUMN name SET NOT NULL;

ALTER TABLE users DROP COLUMN mobile; 