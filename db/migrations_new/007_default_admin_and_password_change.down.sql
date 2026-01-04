-- Remove default admin user
DELETE FROM users WHERE email = 'admin@nomdb.local' AND username = 'admin';

-- Remove password_must_change column
ALTER TABLE users DROP COLUMN IF EXISTS password_must_change;
