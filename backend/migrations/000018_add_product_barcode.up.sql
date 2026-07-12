-- Retail parts carry a printed barcode (UPC/EAN) distinct from the internal SKU;
-- the POS scanner should match either.
ALTER TABLE products ADD COLUMN barcode VARCHAR(100);

CREATE INDEX idx_products_barcode ON products (branch_id, barcode) WHERE barcode IS NOT NULL;
