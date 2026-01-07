-- Add price_range to restaurants table
ALTER TABLE restaurants ADD COLUMN IF NOT EXISTS price_range INTEGER CHECK (price_range >= 1 AND price_range <= 4);

-- Add index for filtering
CREATE INDEX IF NOT EXISTS idx_restaurants_price_range ON restaurants(price_range);

-- Add comment
COMMENT ON COLUMN restaurants.price_range IS 'Price range: 1 = $, 2 = $$, 3 = $$$, 4 = $$$$';
