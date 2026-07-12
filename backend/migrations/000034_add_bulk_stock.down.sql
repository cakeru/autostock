-- 000034_add_bulk_stock.down.sql

BEGIN;

ALTER TABLE products DROP COLUMN IF EXISTS is_bulk;

ALTER TABLE stocktake_items ALTER COLUMN variance     TYPE INT USING ROUND(variance);
ALTER TABLE stocktake_items ALTER COLUMN counted_qty  TYPE INT USING ROUND(counted_qty);
ALTER TABLE stocktake_items ALTER COLUMN expected_qty TYPE INT USING ROUND(expected_qty);

ALTER TABLE stock_movements ALTER COLUMN quantity_change TYPE INTEGER USING ROUND(quantity_change);

ALTER TABLE batches ALTER COLUMN quantity_remaining TYPE INTEGER USING ROUND(quantity_remaining);
ALTER TABLE batches ALTER COLUMN quantity_received  TYPE INTEGER USING ROUND(quantity_received);

ALTER TABLE products ALTER COLUMN min_stock_alert  TYPE INTEGER USING ROUND(min_stock_alert);
ALTER TABLE products ALTER COLUMN reserved_quantity TYPE INT USING ROUND(reserved_quantity);
ALTER TABLE products ALTER COLUMN stock_quantity   TYPE INTEGER USING ROUND(stock_quantity);

COMMIT;
