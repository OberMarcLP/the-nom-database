-- Drop index
DROP INDEX IF EXISTS idx_suggestions_user_id;

-- Drop user_id column
ALTER TABLE restaurant_suggestions DROP COLUMN IF EXISTS user_id;
