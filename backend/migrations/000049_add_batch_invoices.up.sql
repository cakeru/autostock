-- A single receive (batch) can come with several supplier invoices, each with
-- its own number, photo, and amount — e.g. a $100 purchase split into four
-- $25 invoices that are paid off one at a time. Payment is tracked per
-- invoice; batches.amount_paid stays the running total for the batch.
CREATE TABLE batch_invoices (
    id BIGSERIAL PRIMARY KEY,
    branch_id BIGINT NOT NULL REFERENCES branches(id),
    batch_id BIGINT NOT NULL REFERENCES batches(id) ON DELETE CASCADE,
    invoice_number VARCHAR(255),
    invoice_image TEXT,
    amount NUMERIC(10,2) NOT NULL DEFAULT 0 CHECK (amount >= 0),
    amount_paid NUMERIC(10,2) NOT NULL DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_batch_invoices_batch ON batch_invoices(batch_id);

-- Carry over the single-invoice batches recorded before this migration.
INSERT INTO batch_invoices (branch_id, batch_id, invoice_number, invoice_image, amount, amount_paid)
SELECT branch_id, id, invoice_number, invoice_image, quantity_received * unit_cost, amount_paid
FROM batches
WHERE invoice_number IS NOT NULL OR invoice_image IS NOT NULL;