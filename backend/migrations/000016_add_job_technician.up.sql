-- Assign a service job to a technician (mechanic) so work can be routed and
-- per-tech productivity measured.
ALTER TABLE service_jobs ADD COLUMN assigned_to BIGINT REFERENCES users(id) ON DELETE SET NULL;

CREATE INDEX idx_service_jobs_assigned ON service_jobs (branch_id, assigned_to) WHERE assigned_to IS NOT NULL;
