DROP INDEX IF EXISTS idx_restaurants_search;
ALTER TABLE restaurants DROP COLUMN IF EXISTS search_vector;
