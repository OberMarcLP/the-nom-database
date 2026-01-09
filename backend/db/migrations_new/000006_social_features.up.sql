-- Restaurant Lists (user-created collections)
CREATE TABLE restaurant_lists (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    is_public BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_restaurant_lists_user ON restaurant_lists(user_id);
CREATE INDEX idx_restaurant_lists_public ON restaurant_lists(is_public);

-- Restaurant List Items (many-to-many)
CREATE TABLE restaurant_list_items (
    id SERIAL PRIMARY KEY,
    list_id INTEGER NOT NULL REFERENCES restaurant_lists(id) ON DELETE CASCADE,
    restaurant_id INTEGER NOT NULL REFERENCES restaurants(id) ON DELETE CASCADE,
    notes TEXT,
    added_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    CONSTRAINT unique_list_restaurant UNIQUE (list_id, restaurant_id)
);

CREATE INDEX idx_list_items_list ON restaurant_list_items(list_id);
CREATE INDEX idx_list_items_restaurant ON restaurant_list_items(restaurant_id);

-- User Follows (social connections)
CREATE TABLE follows (
    follower_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    following_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    PRIMARY KEY (follower_id, following_id),
    CONSTRAINT no_self_follow CHECK (follower_id != following_id)
);

CREATE INDEX idx_follows_follower ON follows(follower_id);
CREATE INDEX idx_follows_following ON follows(following_id);

-- Activity Feed
CREATE TABLE activities (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    activity_type VARCHAR(50) NOT NULL CHECK (activity_type IN ('rating', 'restaurant', 'list', 'follow', 'suggestion')),
    restaurant_id INTEGER REFERENCES restaurants(id) ON DELETE CASCADE,
    rating_id INTEGER REFERENCES ratings(id) ON DELETE CASCADE,
    list_id INTEGER REFERENCES restaurant_lists(id) ON DELETE CASCADE,
    target_user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
    suggestion_id INTEGER REFERENCES restaurant_suggestions(id) ON DELETE CASCADE,
    details JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_activities_user ON activities(user_id);
CREATE INDEX idx_activities_type ON activities(activity_type);
CREATE INDEX idx_activities_created ON activities(created_at DESC);
CREATE INDEX idx_activities_restaurant ON activities(restaurant_id);

-- Review Photos (photos attached to ratings)
CREATE TABLE review_photos (
    id SERIAL PRIMARY KEY,
    rating_id INTEGER NOT NULL REFERENCES ratings(id) ON DELETE CASCADE,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    filename VARCHAR(255) NOT NULL,
    original_filename VARCHAR(255),
    caption VARCHAR(255),
    file_size INTEGER,
    mime_type VARCHAR(50),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_review_photos_rating ON review_photos(rating_id);
CREATE INDEX idx_review_photos_user ON review_photos(user_id);

-- Likes/Reactions on ratings
CREATE TABLE rating_likes (
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    rating_id INTEGER NOT NULL REFERENCES ratings(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    PRIMARY KEY (user_id, rating_id)
);

CREATE INDEX idx_rating_likes_user ON rating_likes(user_id);
CREATE INDEX idx_rating_likes_rating ON rating_likes(rating_id);

-- Bookmarks/Saved Restaurants
CREATE TABLE bookmarks (
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    restaurant_id INTEGER NOT NULL REFERENCES restaurants(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    PRIMARY KEY (user_id, restaurant_id)
);

CREATE INDEX idx_bookmarks_user ON bookmarks(user_id);
CREATE INDEX idx_bookmarks_restaurant ON bookmarks(restaurant_id);
