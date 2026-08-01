-- Jobs created from a saved POS cart carry the agreed discount so it survives
-- the job → invoice conversion (previously only a note was kept).
BEGIN;

ALTER TABLE service_jobs ADD COLUMN discount NUMERIC(10,2) NOT NULL DEFAULT 0;

COMMIT;
