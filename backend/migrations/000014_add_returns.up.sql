CREATE TABLE returns (
    id BIGSERIAL PRIMARY KEY,
    branch_id BIGINT NOT NULL REFERENCES branches(id) ON DELETE CASCADE,
    invoice_id BIGINT NOT NULL REFERENCES invoices(id) ON DELETE CASCADE,
    customer_id BIGINT REFERENCES customers(id) ON DELETE SET NULL,
    refund_amount DECIMAL(10, 2) NOT NULL DEFAULT 0,
    refund_method VARCHAR(20) NOT NULL CHECK (refund_method IN ('cash', 'store_credit')),
    reason TEXT,
    created_by BIGINT REFERENCES users(id),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE TABLE return_items (
    id BIGSERIAL PRIMARY KEY,
    return_id BIGINT NOT NULL REFERENCES returns(id) ON DELETE CASCADE,
    invoice_item_id BIGINT NOT NULL REFERENCES invoice_items(id) ON DELETE CASCADE,
    product_id BIGINT REFERENCES products(id) ON DELETE SET NULL,
    description VARCHAR(500),
    quantity DECIMAL(10, 2) NOT NULL,
    unit_price DECIMAL(10, 2) NOT NULL,
    total DECIMAL(10, 2) NOT NULL
);

CREATE INDEX idx_returns_invoice ON returns (invoice_id);
CREATE INDEX idx_return_items_invoice_item ON return_items (invoice_item_id);
