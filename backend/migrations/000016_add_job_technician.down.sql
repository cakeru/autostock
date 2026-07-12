DROP INDEX IF EXISTS idx_service_jobs_assigned;
ALTER TABLE service_jobs DROP COLUMN IF EXISTS assigned_to;
