ALTER TABLE service_job_items DROP COLUMN IF EXISTS item_type;
-- Note: product_id is left nullable on rollback to avoid failing on existing
-- non-product rows.
