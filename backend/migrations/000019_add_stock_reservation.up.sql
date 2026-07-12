-- Scheduled jobs/quotes can hold stock against a promise without touching
-- stock_quantity itself, so POS and other jobs can't sell the same units out
-- from under a job that's already committed to using them.
ALTER TABLE products ADD COLUMN reserved_quantity INT NOT NULL DEFAULT 0;
ALTER TABLE products ADD CONSTRAINT chk_products_reserved_nonneg CHECK (reserved_quantity >= 0);
