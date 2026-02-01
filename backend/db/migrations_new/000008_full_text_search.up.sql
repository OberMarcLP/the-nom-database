-- Full-text search for restaurants (paginated and search endpoints)
-- Uses 'simple' config to avoid stemming and match exact substrings better for names.

ALTER TABLE restaurants
ADD COLUMN IF NOT EXISTS search_vector tsvector
GENERATED ALWAYS AS (
  to_tsvector('simple', name || ' ' || COALESCE(description, ''))
) STORED;

CREATE INDEX IF NOT EXISTS idx_restaurants_search ON restaurants USING GIN(search_vector);
