-- Add price_range to restaurants (1 = $, 2 = $$, 3 = $$$, 4 = $$$$)
-- The model and list filter already reference this column; it was missing in the schema.
ALTER TABLE restaurants
ADD COLUMN price_range INTEGER CHECK (price_range BETWEEN 1 AND 4);
