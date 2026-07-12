-- Service jobs can be booked for a future date/time (appointments), not just
-- created as immediate walk-in work.
ALTER TABLE service_jobs ADD COLUMN scheduled_at TIMESTAMP WITH TIME ZONE;

CREATE INDEX idx_service_jobs_scheduled ON service_jobs (branch_id, scheduled_at) WHERE scheduled_at IS NOT NULL;
