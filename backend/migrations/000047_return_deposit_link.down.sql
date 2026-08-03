DROP INDEX IF EXISTS idx_deposits_return_id;
ALTER TABLE deposits DROP COLUMN IF EXISTS return_id;
