-- 000034_add_bulk_stock.up.sql
-- Bulk stock sold by fractional units (e.g. engine oil drawn from a 208 L
-- drum, sold by the liter). The ledger has been integer-only; widen every
-- quantity column to NUMERIC(10,2) so a batch can hold 203.5 L and a sale can
-- draw 4.5 L. Arithmetic stays in SQL, so the invariant
-- SUM(batches.quantity_remaining) == products.stock_quantity holds exactly.

BEGIN;

-- Product-level stock counters.
ALTER TABLE products ALTER COLUMN stock_quantity   TYPE NUMERIC(10,2);
ALTER TABLE products ALTER COLUMN reserved_quantity TYPE NUMERIC(10,2);
ALTER TABLE products ALTER COLUMN min_stock_alert  TYPE NUMERIC(10,2);

-- Intake lots (each physical drum is one batch).
ALTER TABLE batches ALTER COLUMN quantity_received  TYPE NUMERIC(10,2);
ALTER TABLE batches ALTER COLUMN quantity_remaining TYPE NUMERIC(10,2);

-- Ledger movements.
ALTER TABLE stock_movements ALTER COLUMN quantity_change TYPE NUMERIC(10,2);

-- Physical counts.
ALTER TABLE stocktake_items ALTER COLUMN expected_qty TYPE NUMERIC(10,2);
ALTER TABLE stocktake_items ALTER COLUMN counted_qty  TYPE NUMERIC(10,2);
ALTER TABLE stocktake_items ALTER COLUMN variance     TYPE NUMERIC(10,2);

-- Marks a product as drawn from bulk (sold by fractional units). Drives the
-- barrel gauge and fractional-quantity entry in the UI; independent of
-- is_oil_product, since a sealed bottle is oil but not bulk.
ALTER TABLE products ADD COLUMN is_bulk BOOLEAN NOT NULL DEFAULT false;

COMMIT;
