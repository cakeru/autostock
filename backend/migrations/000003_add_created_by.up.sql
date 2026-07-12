BEGIN;

ALTER TABLE invoices ADD COLUMN created_by BIGINT REFERENCES users(id);
ALTER TABLE service_jobs ADD COLUMN created_by BIGINT REFERENCES users(id);

CREATE INDEX idx_invoices_created_by ON invoices(created_by);
CREATE INDEX idx_service_jobs_created_by ON service_jobs(created_by);

COMMIT;
