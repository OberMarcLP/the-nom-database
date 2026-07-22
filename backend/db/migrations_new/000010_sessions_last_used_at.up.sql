-- The refresh handler updates sessions.last_used_at on every token refresh,
-- but the column was never created (the v2.0.0 auth overhaul added the code
-- without a migration). Nullable: NULL means "never refreshed".
ALTER TABLE sessions ADD COLUMN IF NOT EXISTS last_used_at TIMESTAMP WITH TIME ZONE;
