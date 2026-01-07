-- Remove price_range from restaurants table
DROP INDEX IF EXISTS idx_restaurants_price_range;
ALTER TABLE restaurants DROP COLUMN IF EXISTS price_range;
