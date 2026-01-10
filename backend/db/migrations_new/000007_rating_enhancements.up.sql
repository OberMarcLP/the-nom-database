-- Add missing columns to ratings table
ALTER TABLE ratings
ADD COLUMN updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
ADD COLUMN helpful_count INTEGER NOT NULL DEFAULT 0,
ADD COLUMN not_helpful_count INTEGER NOT NULL DEFAULT 0;

-- Create index for rating timestamps
CREATE INDEX idx_ratings_created ON ratings(created_at DESC);

-- Update review_photos table to match code expectations
ALTER TABLE review_photos
ADD COLUMN photo_url VARCHAR(500),
ADD COLUMN display_order INTEGER NOT NULL DEFAULT 0;

-- Migrate existing data: copy filename to photo_url with /api/uploads/review_photos/ prefix
UPDATE review_photos
SET photo_url = '/api/uploads/review_photos/' || filename
WHERE photo_url IS NULL;

-- Make photo_url NOT NULL after migration
ALTER TABLE review_photos
ALTER COLUMN photo_url SET NOT NULL;

-- Create index for display order
CREATE INDEX idx_review_photos_display ON review_photos(rating_id, display_order);

-- Review Votes (helpful/not helpful votes on ratings)
CREATE TABLE review_votes (
    rating_id INTEGER NOT NULL REFERENCES ratings(id) ON DELETE CASCADE,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    vote_type VARCHAR(20) NOT NULL CHECK (vote_type IN ('helpful', 'not_helpful')),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    PRIMARY KEY (rating_id, user_id)
);

CREATE INDEX idx_review_votes_rating ON review_votes(rating_id);
CREATE INDEX idx_review_votes_user ON review_votes(user_id);

-- Create trigger to update helpful_count and not_helpful_count
CREATE OR REPLACE FUNCTION update_rating_vote_counts()
RETURNS TRIGGER AS $$
BEGIN
    UPDATE ratings
    SET
        helpful_count = (
            SELECT COUNT(*) FROM review_votes
            WHERE rating_id = NEW.rating_id AND vote_type = 'helpful'
        ),
        not_helpful_count = (
            SELECT COUNT(*) FROM review_votes
            WHERE rating_id = NEW.rating_id AND vote_type = 'not_helpful'
        )
    WHERE id = NEW.rating_id;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_update_rating_vote_counts
AFTER INSERT OR UPDATE OR DELETE ON review_votes
FOR EACH ROW
EXECUTE FUNCTION update_rating_vote_counts();

-- Create trigger to automatically update updated_at on ratings
CREATE OR REPLACE FUNCTION update_ratings_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_ratings_updated_at
BEFORE UPDATE ON ratings
FOR EACH ROW
EXECUTE FUNCTION update_ratings_updated_at();
