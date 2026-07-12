-- Physical inventory counts: a draft sheet of expected-vs-counted quantities
-- that, once finalized, becomes real stock adjustments (with a full batch
-- ledger trail, same as a manual adjustment).
CREATE TABLE stocktakes (
    id BIGSERIAL PRIMARY KEY,
    branch_id BIGINT NOT NULL REFERENCES branches(id),
    status VARCHAR(20) NOT NULL DEFAULT 'draft' CHECK (status IN ('draft', 'completed', 'cancelled')),
    notes TEXT,
    created_by BIGINT REFERENCES users(id),
    completed_by BIGINT REFERENCES users(id),
    completed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE stocktake_items (
    id BIGSERIAL PRIMARY KEY,
    stocktake_id BIGINT NOT NULL REFERENCES stocktakes(id) ON DELETE CASCADE,
    product_id BIGINT NOT NULL REFERENCES products(id),
    expected_qty INT NOT NULL,
    counted_qty INT,
    variance INT,
    counted_by BIGINT REFERENCES users(id),
    counted_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (stocktake_id, product_id)
);

CREATE INDEX idx_stocktake_items_stocktake ON stocktake_items (stocktake_id);
CREATE INDEX idx_stocktakes_branch ON stocktakes (branch_id, status);
