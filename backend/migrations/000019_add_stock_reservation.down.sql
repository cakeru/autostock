ALTER TABLE products DROP CONSTRAINT IF EXISTS chk_products_reserved_nonneg;
ALTER TABLE products DROP COLUMN IF EXISTS reserved_quantity;
