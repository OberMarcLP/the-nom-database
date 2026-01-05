-- Add user_id column to restaurant_suggestions table
ALTER TABLE restaurant_suggestions ADD COLUMN IF NOT EXISTS user_id INTEGER REFERENCES users(id) ON DELETE SET NULL;

-- Add index for user_id lookups
CREATE INDEX IF NOT EXISTS idx_suggestions_user_id ON restaurant_suggestions(user_id);
