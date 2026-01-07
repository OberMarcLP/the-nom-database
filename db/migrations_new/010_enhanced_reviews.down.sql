-- Drop trigger and function for vote counts
DROP TRIGGER IF EXISTS update_rating_votes_count ON review_votes;
DROP FUNCTION IF EXISTS update_rating_vote_counts();

-- Drop vote counts from ratings
ALTER TABLE ratings DROP COLUMN IF EXISTS helpful_count;
ALTER TABLE ratings DROP COLUMN IF EXISTS not_helpful_count;

-- Drop review_votes table
DROP TABLE IF EXISTS review_votes;

-- Drop review_photos table
DROP TABLE IF EXISTS review_photos;

-- Drop trigger for ratings updated_at
DROP TRIGGER IF EXISTS update_ratings_updated_at ON ratings;

-- Drop updated_at from ratings
ALTER TABLE ratings DROP COLUMN IF EXISTS updated_at;
