-- Closes the reorder -> PO -> receive loop: a formal purchase order against a
-- supplier, received line-by-line into real intake batches (same ledger the
-- rest of inventory already uses).
CREATE TABLE purchase_orders (
    id BIGSERIAL PRIMARY KEY,
    branch_id BIGINT NOT NULL REFERENCES branches(id),
    supplier_id BIGINT NOT NULL REFERENCES suppliers(id),
    po_number VARCHAR(20) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'draft' CHECK (status IN ('draft', 'ordered', 'partial', 'received', 'cancelled')),
    notes TEXT,
    created_by BIGINT REFERENCES users(id),
    ordered_at TIMESTAMPTZ,
    received_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (branch_id, po_number)
);

CREATE TABLE purchase_order_items (
    id BIGSERIAL PRIMARY KEY,
    purchase_order_id BIGINT NOT NULL REFERENCES purchase_orders(id) ON DELETE CASCADE,
    product_id BIGINT NOT NULL REFERENCES products(id),
    quantity_ordered INT NOT NULL CHECK (quantity_ordered > 0),
    quantity_received INT NOT NULL DEFAULT 0,
    unit_cost NUMERIC(10,2) NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_po_items_po ON purchase_order_items (purchase_order_id);
CREATE INDEX idx_po_branch_status ON purchase_orders (branch_id, status);
