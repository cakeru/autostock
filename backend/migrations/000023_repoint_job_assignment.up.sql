-- Technician assignment should key off the HR identity (employees), not the
-- login account (users), since most technicians won't have a login. Remap
-- existing assignments via the 1:1 backfill from migration 000022, then
-- repoint the foreign key.
ALTER TABLE service_jobs DROP CONSTRAINT IF EXISTS service_jobs_assigned_to_fkey;

UPDATE service_jobs sj
SET assigned_to = e.id
FROM employees e
WHERE e.user_id = sj.assigned_to;

ALTER TABLE service_jobs
    ADD CONSTRAINT service_jobs_assigned_to_fkey FOREIGN KEY (assigned_to) REFERENCES employees(id) ON DELETE SET NULL;
