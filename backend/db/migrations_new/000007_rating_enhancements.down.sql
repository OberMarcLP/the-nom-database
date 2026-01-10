-- Drop triggers
DROP TRIGGER IF EXISTS trigger_ratings_updated_at ON ratings;
DROP TRIGGER IF EXISTS trigger_update_rating_vote_counts ON review_votes;

-- Drop functions
DROP FUNCTION IF EXISTS update_ratings_updated_at();
DROP FUNCTION IF EXISTS update_rating_vote_counts();

-- Drop review_votes table
DROP TABLE IF EXISTS review_votes;

-- Drop indexes
DROP INDEX IF EXISTS idx_ratings_created;
DROP INDEX IF EXISTS idx_review_photos_display;

-- Revert review_photos columns
ALTER TABLE review_photos
DROP COLUMN IF EXISTS photo_url,
DROP COLUMN IF EXISTS display_order;

-- Remove columns from ratings table
ALTER TABLE ratings
DROP COLUMN IF EXISTS updated_at,
DROP COLUMN IF EXISTS helpful_count,
DROP COLUMN IF EXISTS not_helpful_count;
