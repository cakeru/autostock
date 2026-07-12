# AutoStock vs Industry Standard — Business Logic & UI Comparison

**Date:** 2026-07-07
**Status:** ✅ All 4 "adopt now" gaps closed (payment ledger, invoice freeze, receive/adjust stock, created_by). Quote/approval, kanban board, global search, and day-close now implemented. Consciously-skipped items unchanged.
**Benchmarked against:** established auto-shop management systems (Shopmonkey, Tekmetric, AutoLeap, Shop-Ware, Mitchell 1) for workflow; Square/Loyverse-class POS for point-of-sale patterns; Odoo/inFlow-class tools for inventory practice.
**Framing:** AutoStock serves a small Cambodian tire & repair shop. The goal is not feature parity with $300/month US shop software — it's knowing which industry patterns are *structural* (skipping them creates real business risk) versus which are enterprise bloat to consciously skip.

---

## The Industry-Standard Shop Flow

Every mature shop-management system converges on the same pipeline:

```
Intake → Estimate → Customer Approval → Work Order → Parts & Labor
      → Invoice (locked) → Payment(s) → Receipt → Follow-up
```

AutoStock's pipeline today:

```
(Intake) → Service Job → Items → Complete → Invoice → Payment → Receipt ✓
```

### Stage-by-stage comparison

| Stage | Industry standard | AutoStock | Gap severity |
|---|---|---|---|
| **Intake** | Lookup by plate/phone/VIN; VIN decode; appointment scheduling | Customer + vehicle records, lookup by name only | Medium |
| **Estimate** | Quote built first; itemized parts + labor | **Absent** — work starts with no quoted price | **High** |
| **Approval** | Customer authorizes estimate (signature/SMS link); legally protective | **Absent** | **High** |
| **Work order** | Status pipeline, assigned technician, labor hours, inspection checklist (DVI) | Status pipeline ✓ (with enforced transitions — good), hours fields exist; no technician assignment | Medium |
| **Parts** | Added to WO reserve/allocate stock; availability shown | Stock shown in picker ✓; no reservation — availability checked only at invoicing | Medium |
| **Invoice** | **Immutable once issued**; corrections via void + reissue or credit note | Items can be added/removed after issue (and stock doesn't follow — known bug) | **High** |
| **Payment** | Payments are a **ledger** (multiple records: amount, method, date, who took it); deposits; partial payments accumulate | Single overwritable `paid_amount` field | **High** |
| **Receipt** | Print/email/SMS receipt | Print receipt ✓ (bilingual, dual currency) | ✓ Fixed |
| **Follow-up** | Service reminders, review requests | Telegram stub | Low (skip for now) |

---

## The Four Structural Gaps

These are the places where AutoStock's *data model*, not just its UI, diverges from standard practice — meaning they get more expensive to fix the longer they wait.

### 1. No estimate → approval stage

Industry systems never start work without a quoted, approved price. This isn't bureaucracy — it's the shop's protection against "I never agreed to that" disputes, and psychologically it converts the customer's commitment *before* labor is spent. AutoStock's flow quotes the price for the first time on the final invoice — after the tires are already mounted.

**Right-sized fix:** not a full estimate module. Reuse the existing job-items mechanism: a job in `pending` with items on it *is* an estimate — add a "Quoted / Approved" flag and a printable quote (the `PrintReceipt` pattern already exists). One column, one button, one print view.

### 2. Inventory only goes down — there is no inbound flow

Standard inventory practice has three movement types: **receiving** (purchase orders or at least a "goods in" action with quantity + cost), **sales deduction** (AutoStock has this now, correctly, via the ledger), and **adjustments** (damage, shrinkage, count corrections — with a reason code). In AutoStock, the only way stock increases is *editing the product's stock field by hand*, which bypasses the `stock_movements` ledger entirely — so the ledger that was just built can't actually explain a product's history, and `buy_price` is a single static field rather than tracked per receipt (no real margin visibility).

**Right-sized fix:** skip purchase orders and suppliers for now. Add a "Receive stock" and an "Adjust stock" action (quantity, unit cost, reason) that write ledger entries, and make direct stock edits impossible in the UI. This closes the loop with two small forms.

### 3. Payments overwrite instead of accumulate

Every accounting-adjacent system records payments as immutable rows: who paid, how much, method, when, taken by whom. AutoStock has one mutable `paid_amount` — so a deposit followed by a balance payment leaves no trace of either event, "partial" status is set by hand rather than derived from `sum(payments) vs total`, and there's no end-of-day cash reconciliation possible (a daily ritual in every real shop). This is also an embezzlement surface: mutable payment fields with no history are the classic small-business leak.

**Right-sized fix:** a `payments` table (invoice_id, amount, method, received_by, created_at), payment_status *derived* from the sum, and the existing "Mark as Paid — Cash" button writes a payment row. The dashboard gains a truthful "cash taken today" number.

### 4. Issued invoices are editable

Industry rule: draft invoices are editable; issued invoices are frozen; mistakes are fixed by void + reissue (which AutoStock already supports well). AutoStock creates every invoice directly in `issued` status yet still allows item add/remove afterward — which is both non-standard and the source of the known stock-asymmetry bug (items added post-issue never deduct stock but do get "restored" on void). Some jurisdictions outright require issued-invoice immutability.

**Right-sized fix:** remove item editing on issued invoices (the endpoints and UI). Void + recreate is the correct correction path and it already works. This *deletes* code and closes the last stock-integrity hole simultaneously.

---

## Smaller Deviations Worth Knowing About

- **No technician attribution.** Even 2-person shops eventually ask "who did this job / who sold this?" `voided_by` exists; `created_by`/`assigned_to` don't. Cheap to add now, painful to backfill.
- **No plate-number lookup.** Industry intake is plate-first (it's what the front desk can see from the counter). The data and index exist; the search UI doesn't.
- **Tax is per-invoice, not per-line.** Fine for Cambodia today (tax is off by default); would need rework for VAT-style itemized tax if the business formalizes.
- **No day-close / Z-report.** POS-class systems have an end-of-day summary (cash vs card totals, voids, per-user). The daily-revenue endpoint is 80% of the way there.
- **No appointment/scheduling concept.** Real gap in US-market software; for a walk-in tire shop, legitimately skippable.
- **No estimates of labor time vs actual** — fields exist (`estimated_hours`, `actual_hours`) but nothing reports on them. Fine dormant.

## Where AutoStock Matches or Beats the Standard

Credit where due — several choices are *better* than what imported software would give this shop:

- **Dual-currency USD/KHR as a first-class concept** — Shopmonkey/Tekmetric simply cannot do this; it's AutoStock's genuine competitive moat for its market, now visible end-to-end (settings-driven rate → invoice → bilingual receipt).
- **Stock movement ledger with reference tracking** (`invoice_issued` / `invoice_voided` + reference id) — this is textbook-correct inventory accounting, better than many small commercial POS products.
- **Enforced job status transitions** with a confirmation on the irreversible step — many commercial tools let statuses flap freely.
- **Void-with-mandatory-reason + voided_by + auto stock restore** — matches or exceeds standard practice.
- **Granular permission strings per module/action** — more precise than the admin/staff binary common in small-shop tools.
- **Bilingual (EN/KM) thermal-style receipt** — locally right in a way no benchmark product ships out of the box.
- **Deliberate simplicity.** Tekmetric onboarding takes weeks; AutoStock is learnable in an afternoon. For a 2–5 person shop this is a feature, not a deficiency — the right move is adopting the four structural patterns above, not the other 200 features.

---

## Visual UI vs Industry Patterns

Modern shop software converged on a few signature UI patterns; AutoStock's clean list+detail architecture covers the basics, and these are the deltas:

| Pattern | Industry norm | AutoStock | Verdict |
|---|---|---|---|
| **Workflow board** | Kanban of jobs by status (Shopmonkey/Tekmetric's home screen) — the shop's shared "what's in the bay" view | Jobs are a paginated table | **Adopt** — highest-value UI upgrade; even a simple 3-column pending/in-progress/completed board changes daily use |
| **Global search** | One omnibox: plate, phone, name, invoice # | Per-page name-only filters | **Adopt** (plate + phone first) |
| **Status color language** | One consistent color per status everywhere | Unified badge (invoices) ✓; jobs/dashboard now consistent | ✓ Close |
| **Action-oriented dashboard** | KPIs are clickable filters: "5 unpaid invoices → tap → filtered list" | Numbers mostly display-only (low-stock card navigates ✓) | Adopt — cheap |
| **Big-target, tablet-first ergonomics** | 44px+ controls, keypad-friendly amounts | Compact desktop-density controls | Partially open (visual-polish pass) |
| **Barcode scanning** | Standard for parts-heavy shops | None | Skip until SKU count justifies it |
| **Customer-facing screens/DVI photos** | Increasingly standard in US market | None | Skip — not this market segment |

---

## Recommended Adoption Order

**Adopt now (structural, cheap while the schema is young):** ✅ **All four implemented & verified 2026-07-07**
1. ✅ Payments ledger (`payments` table: amount, method, `received_by`, notes) + payment status derived from accumulated total, with overpayment guard and audit logging; frontend shows payment history and offers "Pay Full — Cash" / "Record Payment". Verified: backend builds, tests pass.
2. ✅ Issued invoices frozen — item add/remove now rejected with `INVOICE_FROZEN` on any non-draft invoice ("void and recreate instead"), which also closes the stock-asymmetry bug. `paid_amount` removed from the manual update DTO.
3. ✅ Receive/Adjust stock — `POST /products/:id/receive` (quantity + unit cost, updates buy_price) and `/adjust` (delta + reason), both writing `stock_movements` ledger entries; `stock_quantity` removed from the update DTO and the form field disabled when editing (initial quantity only at creation).
4. ✅ `created_by` on invoices and service jobs (migration 000003), populated on create and joined to the creator's name in reads. Bonus: auth middleware now verifies `is_active` in the DB per request, so deactivated users lose access immediately.

*Verification nits — all fixed 2026-07-07:* ✅ `RecordPayment` now runs in a transaction with `SELECT ... FOR UPDATE` on the invoice and derives the total from `SUM(payments)` inside the tx; ✅ `ReceiveStock`/`AdjustStock` are transactional; ✅ `AdjustStock` rejects over-adjustment below zero (`INSUFFICIENT_STOCK`, locked read) instead of clamping, keeping the ledger truthful; ✅ manual `payment_status` removed from `UpdateInvoiceRequest`; ✅ integration tests added for payments (partial → paid progression, overpayment, paid-invoice rejection) and receive/adjust (movements recorded, over-adjustment rejected) — all passing.

*Tiny leftovers (non-blocking):* `CreateInvoiceRequest` still accepts a `payment_status` at creation, so an invoice can be born "paid" with no payment row — drop it and let the walk-in flow create-then-record-payment (the UI already has "Pay Full — Cash"); the `INVOICE_FROZEN` guard has no dedicated test case.

**Adopt soon (workflow value):**
5. Quote/approval flag on jobs + printable quote (gap #1)
6. Jobs kanban board; global plate/phone search
7. Day-close summary (cash/card/void totals per day)

**Consciously skip (enterprise features wrong for this shop):**
appointments/scheduling, VIN decode, DVI photo inspections, supplier/PO management, SMS gateways (revisit Telegram instead — locally apt), accounting integrations, multi-branch UI (until branch #2 is real).
