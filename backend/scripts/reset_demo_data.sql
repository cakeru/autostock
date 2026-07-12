-- Wipes all transactional/experimental data while preserving the schema,
-- the admin login (user id 1), the main branch (id 1) and its settings.
-- Companion to seed_realistic.py, which repopulates a realistic dataset
-- through the live API so every ledger invariant is maintained.
BEGIN;

TRUNCATE
    telegram_events,
    audit_logs,
    stocktake_items, stocktakes,
    purchase_order_items, purchase_orders,
    return_items, returns,
    payments, invoice_items, invoices,
    service_job_items, service_jobs,
    stock_movements, batches,
    deposits, expenses, cash_shifts,
    vehicles, customers,
    products, suppliers
RESTART IDENTITY CASCADE;

-- Employees: keep only the admin's own profile; drop test-run leftovers.
DELETE FROM employees WHERE user_id IS DISTINCT FROM 1;
UPDATE employees SET position = 'Owner', pay_type = 'salary', base_salary = 0 WHERE user_id = 1;
SELECT setval('employees_id_seq', COALESCE((SELECT MAX(id) FROM employees), 1));

-- Users: keep only the admin account.
DELETE FROM users WHERE id <> 1;
SELECT setval('users_id_seq', COALESCE((SELECT MAX(id) FROM users), 1));

-- Settings and branches: keep branch 1 (and global rows), drop phantoms.
DELETE FROM settings WHERE branch_id IS NOT NULL AND branch_id <> 1;
DELETE FROM branches WHERE id <> 1;
UPDATE branches SET name = 'K&S Wheel-Tyre — Phnom Penh' WHERE id = 1;
SELECT setval('branches_id_seq', 1);

COMMIT;
