-- 000007_add_batches.up.sql
-- Intake batch (lot) tracking + attribution of stock movements to a user.

BEGIN;

CREATE TABLE batches (
    id BIGSERIAL PRIMARY KEY,
    branch_id BIGINT NOT NULL REFERENCES branches(id) ON DELETE CASCADE,
    product_id BIGINT NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    supplier VARCHAR(200),
    dot_code VARCHAR(50),
    unit_cost DECIMAL(10, 2) NOT NULL DEFAULT 0,
    quantity_received INTEGER NOT NULL,
    quantity_remaining INTEGER NOT NULL,
    notes TEXT,
    received_by BIGINT REFERENCES users(id),
    received_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_batches_product_id ON batches(product_id);
CREATE INDEX idx_batches_branch_id ON batches(branch_id);
-- FIFO consumption order
CREATE INDEX idx_batches_fifo ON batches(product_id, received_at, id);

ALTER TABLE stock_movements ADD COLUMN recorded_by BIGINT REFERENCES users(id);
ALTER TABLE stock_movements ADD COLUMN batch_id BIGINT REFERENCES batches(id);
CREATE INDEX idx_stock_movements_batch_id ON stock_movements(batch_id);

-- Backfill an opening batch for every product that currently holds stock, so
-- SUM(batches.quantity_remaining) == products.stock_quantity from now on and
-- FIFO allocation has something to draw from for pre-existing inventory.
INSERT INTO batches (branch_id, product_id, supplier, dot_code, unit_cost,
                     quantity_received, quantity_remaining, notes, received_at)
SELECT branch_id, id, 'Opening balance', NULLIF(dot_code, ''), buy_price,
       stock_quantity, stock_quantity, 'Backfilled opening stock', COALESCE(created_at, NOW())
FROM products
WHERE is_active = true AND stock_quantity > 0;

COMMIT;
