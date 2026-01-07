-- Add updated_at to ratings for tracking edits
ALTER TABLE ratings ADD COLUMN IF NOT EXISTS updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP;

-- Add trigger for ratings updated_at
CREATE TRIGGER update_ratings_updated_at BEFORE UPDATE ON ratings
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Create review_photos table for photos attached to reviews
CREATE TABLE IF NOT EXISTS review_photos (
    id SERIAL PRIMARY KEY,
    rating_id INTEGER NOT NULL REFERENCES ratings(id) ON DELETE CASCADE,
    photo_url TEXT NOT NULL,
    caption TEXT,
    display_order INTEGER DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_review_photos_rating_id ON review_photos(rating_id);

-- Create review_votes table for helpfulness voting
CREATE TABLE IF NOT EXISTS review_votes (
    id SERIAL PRIMARY KEY,
    rating_id INTEGER NOT NULL REFERENCES ratings(id) ON DELETE CASCADE,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    vote_type VARCHAR(15) NOT NULL CHECK (vote_type IN ('helpful', 'not_helpful')),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(rating_id, user_id)
);

CREATE INDEX idx_review_votes_rating_id ON review_votes(rating_id);
CREATE INDEX idx_review_votes_user_id ON review_votes(user_id);

-- Add helpful/not helpful counts to ratings (denormalized for performance)
ALTER TABLE ratings ADD COLUMN IF NOT EXISTS helpful_count INTEGER DEFAULT 0;
ALTER TABLE ratings ADD COLUMN IF NOT EXISTS not_helpful_count INTEGER DEFAULT 0;

-- Create function to update vote counts
CREATE OR REPLACE FUNCTION update_rating_vote_counts()
RETURNS TRIGGER AS $$
BEGIN
    IF TG_OP = 'INSERT' THEN
        IF NEW.vote_type = 'helpful' THEN
            UPDATE ratings SET helpful_count = helpful_count + 1 WHERE id = NEW.rating_id;
        ELSE
            UPDATE ratings SET not_helpful_count = not_helpful_count + 1 WHERE id = NEW.rating_id;
        END IF;
    ELSIF TG_OP = 'UPDATE' THEN
        IF OLD.vote_type = 'helpful' AND NEW.vote_type = 'not_helpful' THEN
            UPDATE ratings SET helpful_count = helpful_count - 1, not_helpful_count = not_helpful_count + 1 WHERE id = NEW.rating_id;
        ELSIF OLD.vote_type = 'not_helpful' AND NEW.vote_type = 'helpful' THEN
            UPDATE ratings SET helpful_count = helpful_count + 1, not_helpful_count = not_helpful_count - 1 WHERE id = NEW.rating_id;
        END IF;
    ELSIF TG_OP = 'DELETE' THEN
        IF OLD.vote_type = 'helpful' THEN
            UPDATE ratings SET helpful_count = helpful_count - 1 WHERE id = OLD.rating_id;
        ELSE
            UPDATE ratings SET not_helpful_count = not_helpful_count - 1 WHERE id = OLD.rating_id;
        END IF;
        RETURN OLD;
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Create trigger for vote count updates
CREATE TRIGGER update_rating_votes_count
    AFTER INSERT OR UPDATE OR DELETE ON review_votes
    FOR EACH ROW EXECUTE FUNCTION update_rating_vote_counts();

-- Initialize vote counts for existing ratings
UPDATE ratings r
SET helpful_count = (
    SELECT COUNT(*) FROM review_votes
    WHERE rating_id = r.id AND vote_type = 'helpful'
),
not_helpful_count = (
    SELECT COUNT(*) FROM review_votes
    WHERE rating_id = r.id AND vote_type = 'not_helpful'
);
