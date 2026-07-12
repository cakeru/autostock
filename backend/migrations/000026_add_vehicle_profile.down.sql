DROP TABLE IF EXISTS vehicle_record_photos;
DROP TABLE IF EXISTS vehicle_records;
ALTER TABLE vehicles DROP CONSTRAINT IF EXISTS vehicles_branch_id_fkey;
DROP INDEX IF EXISTS idx_vehicles_branch;
ALTER TABLE vehicles DROP COLUMN IF EXISTS branch_id;
