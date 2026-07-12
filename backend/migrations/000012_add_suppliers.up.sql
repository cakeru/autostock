CREATE TABLE suppliers (
    id BIGSERIAL PRIMARY KEY,
    branch_id BIGINT NOT NULL REFERENCES branches(id) ON DELETE CASCADE,
    name VARCHAR(200) NOT NULL,
    phone VARCHAR(50),
    email VARCHAR(200),
    address TEXT,
    notes TEXT,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_suppliers_branch ON suppliers (branch_id) WHERE is_active;

-- Link purchases (batches) to a supplier and track what's been paid, so we can
-- show accounts payable. supplier (free text) stays for legacy rows.
ALTER TABLE batches
  ADD COLUMN supplier_id BIGINT REFERENCES suppliers(id) ON DELETE SET NULL,
  ADD COLUMN amount_paid DECIMAL(10, 2) NOT NULL DEFAULT 0;

CREATE INDEX idx_batches_supplier ON batches (supplier_id);
