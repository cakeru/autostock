DROP INDEX IF EXISTS idx_vehicles_share_token;
ALTER TABLE vehicles DROP COLUMN IF EXISTS share_token;
