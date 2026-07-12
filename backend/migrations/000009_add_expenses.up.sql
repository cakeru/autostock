CREATE TABLE expenses (
    id BIGSERIAL PRIMARY KEY,
    branch_id BIGINT NOT NULL REFERENCES branches(id),
    category VARCHAR(100) NOT NULL,
    description TEXT,
    amount_usd DECIMAL(10, 2) NOT NULL DEFAULT 0,
    spent_at DATE NOT NULL DEFAULT CURRENT_DATE,
    created_by BIGINT REFERENCES users(id),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_expenses_branch_date ON expenses(branch_id, spent_at);
