-- Drop triggers and functions
DROP TRIGGER IF EXISTS trigger_update_restaurant_lists_updated_at ON restaurant_lists;
DROP FUNCTION IF EXISTS update_restaurant_lists_updated_at();

-- Drop indexes
DROP INDEX IF EXISTS idx_list_restaurants_restaurant_id;
DROP INDEX IF EXISTS idx_list_restaurants_list_id;
DROP INDEX IF EXISTS idx_restaurant_lists_user_id;

-- Drop tables
DROP TABLE IF EXISTS list_restaurants;
DROP TABLE IF EXISTS restaurant_lists;
