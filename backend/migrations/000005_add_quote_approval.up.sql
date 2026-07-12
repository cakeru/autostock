ALTER TABLE service_jobs
  ADD COLUMN quote_approved_at TIMESTAMPTZ,
  ADD COLUMN quote_approved_by BIGINT REFERENCES users(id);
