-- Drop foreign key columns from existing tables
ALTER TABLE menu_photos DROP COLUMN IF EXISTS user_id;
ALTER TABLE restaurant_suggestions DROP COLUMN IF EXISTS user_id;
ALTER TABLE ratings DROP COLUMN IF EXISTS user_id;
ALTER TABLE restaurants DROP COLUMN IF EXISTS user_id;

-- Drop tables in reverse order (respecting foreign key constraints)
DROP TABLE IF EXISTS audit_logs;
DROP TABLE IF EXISTS api_keys;
DROP TABLE IF EXISTS sessions;
DROP TABLE IF EXISTS role_permissions;
DROP TABLE IF EXISTS user_roles;
DROP TABLE IF EXISTS permissions;
DROP TABLE IF EXISTS roles;
DROP TABLE IF EXISTS users;
