BEGIN;

DROP INDEX IF EXISTS idx_invoices_created_by;
DROP INDEX IF EXISTS idx_service_jobs_created_by;

ALTER TABLE invoices DROP COLUMN IF EXISTS created_by;
ALTER TABLE service_jobs DROP COLUMN IF EXISTS created_by;

COMMIT;
