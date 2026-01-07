-- Create restaurant_lists table
CREATE TABLE IF NOT EXISTS restaurant_lists (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    is_public BOOLEAN DEFAULT false,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create list_restaurants junction table
CREATE TABLE IF NOT EXISTS list_restaurants (
    id SERIAL PRIMARY KEY,
    list_id INTEGER NOT NULL REFERENCES restaurant_lists(id) ON DELETE CASCADE,
    restaurant_id INTEGER NOT NULL REFERENCES restaurants(id) ON DELETE CASCADE,
    notes TEXT,
    added_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(list_id, restaurant_id)
);

-- Create indexes for better performance
CREATE INDEX idx_restaurant_lists_user_id ON restaurant_lists(user_id);
CREATE INDEX idx_list_restaurants_list_id ON list_restaurants(list_id);
CREATE INDEX idx_list_restaurants_restaurant_id ON list_restaurants(restaurant_id);

-- Create updated_at trigger for restaurant_lists
CREATE OR REPLACE FUNCTION update_restaurant_lists_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_update_restaurant_lists_updated_at
    BEFORE UPDATE ON restaurant_lists
    FOR EACH ROW
    EXECUTE FUNCTION update_restaurant_lists_updated_at();

-- Grant permissions to roles
GRANT SELECT, INSERT, UPDATE, DELETE ON restaurant_lists TO authenticated;
GRANT SELECT, INSERT, UPDATE, DELETE ON list_restaurants TO authenticated;
GRANT USAGE, SELECT ON SEQUENCE restaurant_lists_id_seq TO authenticated;
GRANT USAGE, SELECT ON SEQUENCE list_restaurants_id_seq TO authenticated;
