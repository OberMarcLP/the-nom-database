-- Remove price_range from restaurants
ALTER TABLE restaurants
DROP COLUMN IF EXISTS price_range;
