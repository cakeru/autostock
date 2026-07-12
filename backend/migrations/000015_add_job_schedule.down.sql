DROP INDEX IF EXISTS idx_service_jobs_scheduled;
ALTER TABLE service_jobs DROP COLUMN IF EXISTS scheduled_at;
