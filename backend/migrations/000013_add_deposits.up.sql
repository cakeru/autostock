CREATE TABLE deposits (
    id BIGSERIAL PRIMARY KEY,
    branch_id BIGINT NOT NULL REFERENCES branches(id) ON DELETE CASCADE,
    customer_id BIGINT NOT NULL REFERENCES customers(id) ON DELETE CASCADE,
    amount DECIMAL(10, 2) NOT NULL CHECK (amount > 0),
    note TEXT,
    status VARCHAR(20) NOT NULL DEFAULT 'held' CHECK (status IN ('held', 'applied', 'refunded')),
    invoice_id BIGINT REFERENCES invoices(id) ON DELETE SET NULL,
    created_by BIGINT REFERENCES users(id),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    settled_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX idx_deposits_customer ON deposits (customer_id, status);
