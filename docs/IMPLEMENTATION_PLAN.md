# AutoStock — Consolidated Implementation Plan

**Date:** 2026-07-07
**Status:** ✅ All phases complete. See individual doc statuses above.
**Purpose:** One ordered plan for every open item remaining across [AUDIT.md](./AUDIT.md), [UX_AUDIT.md](./UX_AUDIT.md), [INDUSTRY_COMPARISON.md](./INDUSTRY_COMPARISON.md), and [LAYOUT_PLAN.md](./LAYOUT_PLAN.md). Everything critical from those audits is already done — this is the remainder, sequenced so correctness debt clears before features.

**Status legend per phase:** each item lists → *files*, *approach*, *done-when*.

---

## Phase 1 — Correctness & Polish Debt (~half day)

Small items with known fixes; clear these before adding features on top.

### 1.1 Wire dashboard KPI links to actual filters *(the one functional gap from the layout pass)*
- **Files:** `frontend/src/pages/Invoices.tsx`, `frontend/src/pages/Inventory.tsx`
- **Approach:** `useSearchParams`; in Invoices init `paymentFilter` from `payment_status` param; in Inventory init `lowStockOnly` from `low=1`. Clear the param when the user changes the filter manually (mirror the `ServiceJobs.tsx` pattern).
- **Done when:** clicking "Today's Revenue" lands on the invoice list already filtered to paid; "Low Stock" lands on inventory with the low-stock checkbox on.

### 1.2 Fix or remove inert sticky table headers
- **Files:** all four list pages.
- **Approach:** `sticky top-0` never engages because the wrapper has `overflow-hidden`. Cheapest honest fix: drop the `sticky top-0 z-10` classes (tables are ≤20 rows/page; stickiness adds little). If keeping: wrapper gets `max-h-[70vh] overflow-y-auto` and the rounded corners move to a parent.
- **Done when:** no dead classes, or headers actually stick while scrolling.

### 1.3 Settings page state isolation
- **File:** `frontend/src/pages/Settings.tsx`
- **Approach:** one `useUpdateSetting()` instance per section (hooks are cheap); replace the double-fire tax save with a single mutation that takes `[{key, value}, ...]` (add a batch variant to the hook, or `Promise.all` inside one `mutationFn`); auto-clear the "Saved!" indicator after ~3s (`setTimeout` in `onSuccess` or use the mutation's `reset`).
- **Done when:** saving one section doesn't light up the others; tax enable + rate save atomically; "Saved!" fades.

### 1.4 Invoice form: render the Notes input
- **File:** `frontend/src/pages/Invoices.tsx` — `formNotes` state exists and is submitted; there's just no `<Textarea>` for it. Add it below Discount.

### 1.5 Backend: invoice can no longer be born "paid"
- **Files:** `backend/internal/invoice/dto/dto.go`, `service.go`, invoice handler; frontend `types/invoice.ts` + any create call passing `payment_status`.
- **Approach:** remove `PaymentStatus` from `CreateInvoiceRequest`; invoices are always created `unpaid` and become paid only via `RecordPayment` (UI already flows this way — "Pay Full — Cash" after create). `CreateFromServiceJob` passes it through too — remove there as well.
- **Done when:** grep for `payment_status` in create paths is clean; existing tests still pass.

### 1.6 Backend: add the missing `INVOICE_FROZEN` test
- **File:** `backend/internal/invoice/service/service_test.go`
- **Approach:** create invoice (issued) → `AddItem` → expect `INVOICE_FROZEN`; same for `RemoveItem`.

---

## Phase 2 — Robustness (~half day)

### 2.1 Invoice/job numbering: eliminate the `MAX+1` race *(oldest open item in the project)*
- **Files:** `backend/internal/invoice/service/service.go`, `servicejob/service/service.go`, new migration `000005`.
- **Approach (pick one):**
  - **A (preferred):** per-year Postgres sequences are awkward; instead take a transaction-scoped advisory lock: `SELECT pg_advisory_xact_lock(hashtext('invoice_number'))` inside the create transaction, then the existing `MAX+1` query via `tx`. Two lines, keeps the readable `INV-YYYY-NNNN` format, gapless.
  - **B:** a `counters(year, kind, value)` table updated with `INSERT ... ON CONFLICT ... DO UPDATE SET value = counters.value + 1 RETURNING value`.
- **Note:** `generateInvoiceNumber` currently runs on `s.pool` *before* the transaction starts — move it inside the tx for either option.
- **Done when:** a test spawning ~10 concurrent creates produces 10 distinct sequential numbers with no unique-violation errors.

### 2.2 List loading: skeletons + `keepPreviousData`
- **Files:** all list hooks (`useProducts`, `useInvoices`, `useServiceJobs`, `useCustomers`) + list pages; new `frontend/src/components/ui/Skeleton.tsx`.
- **Approach:** add `placeholderData: keepPreviousData` (TanStack v5) to list queries so pagination/filter changes keep the table rendered; replace the `"Loading..."` text with 5 skeleton rows (`animate-pulse bg-muted` bars) on first load only (`isPending && !data`).
- **Done when:** paging through Inventory never blanks the table; first load shows skeleton rows; no layout jump.

### 2.3 Shared ConfirmDialog with proper a11y
- **Files:** new `frontend/src/components/ui/ConfirmDialog.tsx`; adopt in Inventory, Customers, ServiceJobs, CustomerDetail (vehicle), InvoiceDetail (void keeps its reason field — give it the same base).
- **Approach:** one component: `role="alertdialog"`, `aria-modal`, Escape closes, focus moves to the Cancel button on open, scrim click closes, destructive button uses `variant="destructive"`, message must name the object (pass `title`/`description`). This retires ~6 hand-rolled modal blocks and closes the remaining modal-a11y audit item.
- **Done when:** every delete/void confirmation is the shared component; keyboard-only operation works end to end.

### 2.4 SlideOver focus trap (small)
- **File:** `frontend/src/components/ui/SlideOver.tsx`
- **Approach:** on Tab/Shift+Tab, cycle focus within the panel (query focusable elements; wrap at ends). ~15 lines in the existing keydown handler.

---

## Phase 3 — Workflow Features: "adopt soon" tier (~1–2 days)

### 3.1 Quote → approval on service jobs *(closes the last structural gap vs industry standard)*
- **Backend:** migration `000006`: `service_jobs.quote_approved_at TIMESTAMPTZ`, `quote_approved_by BIGINT REFERENCES users(id)`. Endpoint `POST /service-jobs/:id/approve-quote` (permission `service:update`). Optional guard: warn (not block) on `Complete` if items exist but quote was never approved.
- **Frontend:** on ServiceJobDetail — "Approve Quote" button once items exist; approved state shows a badge + timestamp; "Print Quote" button reusing the `PrintReceipt` pattern (new `PrintQuote.tsx`: same 80mm layout, titled "QUOTATION", no payment section, "valid 7 days" footer).
- **Done when:** the tire-change journey supports: add items → print quote → customer approves → work → complete → invoice.

### 3.2 Global search by plate / phone / name
- **Backend:** `GET /search?q=` → union of top-5 matches from customers (name/phone ILIKE), vehicles (plate ILIKE, joined to owner), invoices (number), jobs (number). Single endpoint, one query with `UNION ALL`, permission-filtered per entity.
- **Frontend:** search input in the `PageHeader` area or sidebar top (Cmd/K optional, not required); debounced 300ms; grouped results dropdown; Enter navigates to the top hit. Plate hits navigate to the owning customer.
- **Done when:** typing a plate number from the counter reaches the customer's page in ≤2 interactions.

### 3.3 Tire-size filter (backend already supports it)
- **Files:** `frontend/src/pages/Inventory.tsx` (add `tire_size` text filter shown when type=tire), invoice/job product pickers (client-side filter input above the `<select>`, filtering the loaded list by name *or* tire_size).
- **Done when:** a 60-SKU shop can find "205/55R16" in the picker in one keystroke sequence.

## Phase 4 — Operational Views (~1–2 days)

### 4.1 Jobs kanban board
- **Files:** `frontend/src/pages/ServiceJobs.tsx` (view toggle: board ⇄ table; board is default on ≥lg screens), new `components/servicejob/JobBoard.tsx`.
- **Approach:** three columns — Pending / In Progress / Completed (cancelled stays table-only behind a filter). Cards: job #, customer, plate, priority badge, age ("2h ago"). **No drag-and-drop in v1** — status changes go through the existing detail-page transitions; the board is a *view*. Fetch with `per_page: 100`, group client-side.
- **Done when:** the shop's "what's in the bay" question is answered by the default Jobs screen.

### 4.2 Day-close summary
- **Backend:** `GET /dashboard/day-close?date=` → payments grouped by method (from the payments ledger — this is why it exists), invoice count, voided count + total, per-user totals (`received_by`). Permission `report:view`.
- **Frontend:** "Day Close" section on Dashboard (or `/day-close` page): date picker defaulting to today, method totals, void list, print button (reuse print pattern — this is the end-of-day cash-drawer reconciliation sheet).
- **Done when:** the owner can compare the printed sheet against the cash drawer every evening.

---

## Phase 5 — Housekeeping (half day, anytime)

- **README truth pass:** remove ✅ from PDF generation and Telegram bot (or reword as "printable receipts ✓ / PDF planned"); drop "SQLC" from the stack line.
- **Delete dead sqlc scaffolding:** `backend/sqlc.yaml` + per-module `repository/queries.sql` (raw SQL in services is the de-facto standard here — make it official).
- **Contrast margin:** `--color-muted-foreground` → `hsl(220, 12%, 42%)` (from 50% lightness) for comfortable AA on white.
- **Permissions staleness decision:** `is_active` is now DB-checked per request; permissions still ride the JWT for up to 24h. Either accept (document it) or load permissions from the DB in `AuthMiddleware` (one indexed query, same table already hit — cheap). Recommended: load from DB, drop them from the token.
- **Docs:** mark completed items in the four audit docs as phases land.

---

## Explicitly Deferred (decided, not forgotten)

Appointments/scheduling, VIN decode, DVI photos, supplier/PO module, real Telegram sending, server-side PDF, dark mode, Khmer i18n, barcode scanning, searchable async comboboxes (per_page 100 + client filter is sufficient until SKU/customer counts pass ~100), multi-branch UI.

## Verification Ritual (every phase)

Backend: `go build ./... && go vet ./... && go test ./...` (integration tests need the dev Postgres up).
Frontend: `npx tsc -b && npm run build`, then eyeball 375px / 768px / 1440px.
Print paths (receipt, quote, day-close) after any layout-adjacent change.
Journey smoke test after Phase 3: customer → vehicle → job → quote → approve → complete → invoice → pay cash → print receipt.
