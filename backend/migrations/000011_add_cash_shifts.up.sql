CREATE TABLE cash_shifts (
    id BIGSERIAL PRIMARY KEY,
    branch_id BIGINT NOT NULL REFERENCES branches(id),
    opened_by BIGINT REFERENCES users(id),
    opening_float DECIMAL(10, 2) NOT NULL DEFAULT 0,
    opened_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    closed_by BIGINT REFERENCES users(id),
    closing_amount DECIMAL(10, 2),
    cash_sales DECIMAL(10, 2),
    expected_amount DECIMAL(10, 2),
    over_short DECIMAL(10, 2),
    note TEXT,
    closed_at TIMESTAMP WITH TIME ZONE,
    status VARCHAR(20) NOT NULL DEFAULT 'open' CHECK (status IN ('open', 'closed')),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- At most one open drawer per branch (single register), enforced by the DB.
CREATE UNIQUE INDEX idx_cash_shifts_one_open ON cash_shifts (branch_id) WHERE status = 'open';
CREATE INDEX idx_cash_shifts_branch ON cash_shifts (branch_id, opened_at DESC);
