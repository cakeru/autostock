-- Each intake batch (a stock receive) can carry the supplier invoice it came
-- with: a reference number and a photo of the invoice. This lets the shop
-- group receives by invoice and pay specific invoices instead of lumping
-- everything together.
ALTER TABLE batches ADD COLUMN invoice_number VARCHAR(255);
ALTER TABLE batches ADD COLUMN invoice_image TEXT;