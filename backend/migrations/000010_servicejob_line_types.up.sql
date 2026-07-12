-- Let service jobs hold the same line types as invoices (labor / fee / custom),
-- not just products. product_id becomes optional; item_type classifies the line.
ALTER TABLE service_job_items ALTER COLUMN product_id DROP NOT NULL;

ALTER TABLE service_job_items
  ADD COLUMN item_type VARCHAR(50) NOT NULL DEFAULT 'product'
  CHECK (item_type IN ('product', 'labor', 'fee', 'custom'));
