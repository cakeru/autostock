ALTER TABLE service_jobs DROP CONSTRAINT IF EXISTS service_jobs_assigned_to_fkey;

UPDATE service_jobs sj
SET assigned_to = e.user_id
FROM employees e
WHERE e.id = sj.assigned_to;

ALTER TABLE service_jobs
    ADD CONSTRAINT service_jobs_assigned_to_fkey FOREIGN KEY (assigned_to) REFERENCES users(id) ON DELETE SET NULL;
