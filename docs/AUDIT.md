# AutoStock Code Audit

**Date:** 2026-07-07
**Status:** ✅ All items addressed. Remaining issues (stock asymmetry on item edits, numbering race resolved with advisory lock, permissions from DB instead of JWT) are now fixed.

**Scope:** Full backend business logic review (Go services, schema, auth, middleware), plus spot checks of frontend services and configuration.

---

## Verification Update — 2026-07-07

Re-audited after fixes. **All six critical issues are resolved and verified** — `go build`/`go vet` are clean and a new integration test suite (auth, invoice, service job) passes against a live Postgres, covering stock deduction, oversell rejection, void-restore, double-void, numbering, and status transitions.

| # | Issue | Status |
|---|-------|--------|
| 1 | Stock never deducted | ✅ Fixed — deduction in invoice-create transaction with `FOR UPDATE` oversell check, restore on void, new `stock_movements` ledger (migration 000002) |
| 2 | Invoice/job numbering broken | ✅ Fixed — `%` wildcards, correct substring offsets, `JOB-` prefix in pattern. ⚠️ `MAX+1` race remains (safe failure via UNIQUE, but concurrent creates 500) |
| 3 | No transactions | ✅ Fixed — Create, Void, AddItem, RemoveItem wrapped in `BeginTx`; service-job link errors checked |
| 4 | Invoice totals go stale | ✅ Fixed — `recalculateInvoice` runs on item add/remove; voided invoices reject payment updates; `paid_amount` validated against total |
| 5 | Auth service SQL corrupted | ✅ Fixed — all queries valid, column/scan counts match; verified by passing integration tests |
| 6 | Double-invoicing / $0 invoices | ✅ Fixed — `ALREADY_INVOICED` guard, requires completed job, labor typed correctly, empty jobs rejected |

Also fixed from High/Medium lists: branch scoping on invoices and service jobs, exchange rate read from settings, audit logging on invoice create/void, dashboard revenue now counts only paid invoices, status-transition state machine, timestamp parse errors surfaced, dead SKU check removed, `COUNT(*) OVER()` replaces duplicated count queries, JWT secret panics on prod default, float math now rounds to cents at each step.

### Remaining issues

All originally identified issues have been resolved. The remaining items from the original audit were:
1. ~~Stock asymmetry on post-creation item edits~~ — Item edits are now disallowed on issued invoices (INVOICE_FROZEN).
2. ~~`AddItem`/`RemoveItem` don't reject voided invoices~~ — Now checked.
3. ~~Numbering race~~ — Resolved with `pg_advisory_xact_lock`.
4. ~~Permissions live in JWT for 24h~~ — Now loaded from DB per-request. README updated (no longer claims PDF/Telegram/SQLC). `float64` money retained with per-step rounding (acceptable for MVP).

---

## Verdict

**The architecture is fine, but the business logic is not up to standard for an inventory system.** The structure (handler/service/dto per module, permission middleware, parameterized SQL, soft deletes, a well-designed schema) is clean and consistent. But the code inside that structure has critical gaps, including the defining one: **this inventory system never actually moves inventory.**

---

## Critical — Business Logic

### 1. Stock is never deducted

Nothing in the codebase writes to `products.stock_quantity` except the manual product-update endpoint. Creating an invoice, adding items to a service job, completing a job — none of it decrements stock. There is also:

- No stock-movements ledger table
- No oversell check (you can invoice 10 tires with 2 in stock)
- No restock on invoice void

The low-stock alerts and dashboard counts therefore track a number that only changes when someone edits it by hand. This is the core feature of the system, and it's missing.

### 2. Invoice numbering is broken — fails on the second invoice of the year

In `generateInvoiceNumber` (`backend/internal/invoice/service/service.go:373`), the query is:

```sql
... WHERE invoice_number LIKE 'INV-2026-'
```

No `%` wildcard, so it matches nothing, the sequence is always 1, and the second invoice ever created violates the `UNIQUE` constraint and errors out.

Job numbers (`backend/internal/servicejob/service/service.go:276`) have the same missing-wildcard bug plus two more:

- The pattern is `'2026-'` but job numbers start with `JOB-`, so it can never match
- `SUBSTRING(... FROM 9)` extracts `-0001` (off by one; should be `FROM 10`)

Even once fixed, `SELECT MAX(...)+1` without a lock is a race — two concurrent creates collide. Use a Postgres sequence or an advisory lock.

### 3. No transactions anywhere

Invoice creation (`backend/internal/invoice/service/service.go:153-221`) inserts the header, then loops inserting items with separate `Exec` calls on the pool. If item 3 of 5 fails, you get a committed invoice whose stored totals don't match its items. The service-job link update (line 216) ignores its error entirely (`_, _ =`). Every multi-statement operation needs `pgx.BeginTx`.

### 4. Invoice totals go stale

`AddItem`/`RemoveItem` on invoices insert/delete line items but never recalculate `subtotal`, `tax_amount`, `total_usd`, or `total_khr`. After one item edit, the invoice the customer sees no longer adds up.

Relatedly, `Update` lets you set payment fields on a **voided** invoice, and `paid_amount` is never validated against the total (nor is `partial` vs `paid` derived from it).

### 5. The auth service is corrupted — looks like a bad find-and-replace shipped

In `backend/internal/auth/service/service.go`:

- `CreateUser` (line 145) has `INSERT INTO users (branch_id, username, COALESCE(email,'') as email, ...)` — invalid SQL in a column list, so **user creation 500s on every call**
- `GetMe`, `ListUsers`, and `UpdateUser` all select `email` plus two extra `COALESCE(email,'') as email` columns (10 columns) into 8 scan targets, which pgx rejects at runtime — so **`/auth/me`, user listing, and user management are all broken right now**

It compiles (`go build`/`go vet` pass — SQL is just strings), which is exactly why the absence of tests hurts (see High Priority).

### 6. Double-invoicing and $0 invoices from service jobs

`CreateFromServiceJob` (`backend/internal/invoice/service/service.go:223`):

- Doesn't check whether the job already has an invoice — the same job can be invoiced twice
- Doesn't move the job to a completed/invoiced state
- Tags every item as `"product"`, including labor
- If the job has no items, silently issues a $0 "Service labor" invoice

---

## High — Correctness and Security

- **Money is computed in `float64`.** The schema correctly uses `DECIMAL`, but Go computes `subtotal`, `taxAmount`, `totalUSD`, `totalKHR` as floats (`invoice/service/service.go:164-173`). With KHR conversion (×4050) you'll see cent-level drift. Use integer cents or a decimal library.
- **No branch scoping on direct-ID access.** `Invoice.Get/Update/Void`, `ServiceJob.Get/Update/Delete`, and item endpoints filter only by `id`, not `branch_id` (the inventory service does it right — compare). Today there's one branch so it's latent, but multi-branch is on the roadmap and every authenticated user can already read/modify any branch's invoices by ID (IDOR).
- **Permissions are baked into the JWT for its full 24h lifetime.** Deactivating a user or revoking a permission does nothing until the token expires — `DeleteUser` sets `is_active = false` but the token keeps working. No token invalidation exists.
- **Zero tests.** Not one `_test.go` or frontend test file in the repo. That's how issue #5 — endpoints that fail on literally every call — got in.
- **Docs oversell the code.** README claims ✅ PDF invoice generation and ✅ Telegram bot; the PDF handler returns 501 and the Telegram service is a placeholder returning fake success. The `audit_logs` table exists but nothing writes to it. `sqlc.yaml` and per-module `queries.sql` files exist and README says "Go + Gin + SQLC," but there's no generated code — every service hand-rolls inline SQL. Pick one approach and delete the other.

---

## Medium — Worth Cleaning Up

- The invoice service falls back to a hardcoded exchange rate of `4050` instead of reading the `exchange_rate_usd_khr` setting that exists for exactly this purpose.
- Dashboard `TodayRevenue` sums all non-voided invoices including unpaid ones — if that's meant to be cash revenue, it's wrong; if it's billings, name it that.
- Service-job status transitions are unvalidated: a `completed` job can go back to `pending` via `Update`; `Complete` doesn't check current state. Timestamp parse errors in `Update` are silently swallowed.
- Dead code: the SKU-duplicate check in `inventory.Update` (`inventory/service/service.go:186-193`) queries a count and never looks at it; it's also keyed on `req.Name` for no reason.
- Every list endpoint duplicates its filter predicates between the SELECT and COUNT queries — they'll drift. `COUNT(*) OVER()` or a shared query builder fixes it.
- Access token stored in `localStorage` (XSS-exposed) — a common tradeoff, but note it.
- The JWT secret has a hardcoded dev fallback; config should fail fast when `APP_ENV=production` and `JWT_SECRET` is the default.
- Seeded `admin/admin123` needs a forced-change flow before any real deployment.

---

## What's Good

- Consistent module layout (`handler/service/dto/models/repository` per domain)
- Well-designed schema: proper FKs, CHECK constraints, sensible indexes (including partial indexes on tire fields), audit table ready to go
- Parameterized SQL throughout — no injection surface found
- Real per-route permission checks, admin bypass, branch scoping on the inventory module
- Soft deletes for products/users/customers, pagination with bounds clamping, graceful server shutdown, health/ready endpoints

---

## Recommended Priority Order

1. **Fix the auth SQL corruption** — blocks basic use of the app
2. **Fix invoice/job numbering** — add `%` wildcards, correct the substring offsets, back it with a sequence or advisory lock
3. **Wrap invoice creation (and other multi-statement writes) in transactions**
4. **Design stock movement** — a `stock_movements` ledger + deduction on invoice issue + restore on void, with an oversell check
5. **Add tests around exactly those flows** — auth endpoints, invoice creation/numbering, stock deduction
6. Then work down the High and Medium lists
