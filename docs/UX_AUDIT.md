# AutoStock UX/UI Audit

**Status:** ✅ All items fixed. Last remaining issues (skeletons, modal a11y, invoice Notes input, visual polish) resolved.

## Verification Update — 2026-07-07

Re-reviewed after fixes; frontend builds cleanly. **All five critical screen-level issues are fixed, plus most of the High list.** The remaining gaps are the journey-level ones.

**Verified fixed:**
- ✅ Global error feedback — sonner toaster + `QueryClient` mutation-level `onError` surfacing the API's error message; success toasts added across all mutation hooks
- ✅ Invoices mobile card view
- ✅ Diagnosis / Work Performed — prefilled from the job and saved via a Save Notes button
- ✅ Invoice form — live total preview (USD + KHR with thousand separators), removable rows, zero-price rows no longer dropped, empty-submit shows a toast, exchange rate pulled from Settings
- ✅ Mark Completed — inline confirmation naming the job and warning items become locked
- ✅ Destructive actions — red `destructive` button variant; delete/void confirmations name the object; void dialog explains stock restoration
- ✅ Navigation — Users link (permission-gated), Customers in mobile nav, prefix-match active state, Back preserves history via `navigate(-1)`
- ✅ Status badges unified into one shared `InvoiceStatusBadge` with a single color mapping
- ✅ Pickers bumped to `per_page: 100` with stock counts shown; 300ms debounced search on Inventory/Customers
- ✅ Tab title "AutoStock"; fonts actually loaded; `safe-area-inset-bottom` on mobile nav

**Journey-level fixes — completed 2026-07-07 (second pass):**
1. ✅ **Job form now has customer + vehicle pickers** (`ServiceJobs.tsx`) — customer dropdown (name + phone), vehicle dropdown scoped to the chosen customer (with a helpful hint when they have none), sends `customer_id`/`vehicle_id` to the API. Old free-text field replaced by an honest "Notes (optional)". CustomerDetail gained a **New Job** button that deep-links to `/service-jobs?new=1&customer=<id>` and opens the form preselected. Jobs, service history, and invoices-from-jobs are now connected to real customers.
2. ✅ **Printable receipt** — new `components/invoice/PrintReceipt.tsx` (80mm receipt layout: shop header, invoice/date/customer/vehicle/job, itemized lines, USD + KHR totals with separators, payment status, VOIDED stamp, bilingual thank-you) rendered print-only from InvoiceDetail via a "Print Receipt" button (`window.print()`); sidebar, mobile nav, layout padding, and screen UI hidden with `print:` variants.
3. ✅ **Payment prefill + one-tap cash** — payment form now prefills status/amount/method from the invoice, and a "Mark as Paid — Cash ($total)" button settles the common case in one tap. Detail-page USD/KHR totals also got thousand separators.

**Still open:**
All items resolved. Settings page now uses per-section mutations + batch tax save with auto-clearing. Skeletons + `keepPreviousData` added. Modal a11y addressed with shared `ConfirmDialog` + `SlideOver` focus trap. Invoice Notes textarea rendered. Visual polish pass completed (spacing tiers, PageHeader, rail layout, slide-overs, sticky headers, clickable KPI cards, login upgrade).

---

**Date:** 2026-07-07
**Scope:** Full frontend review (all pages, layout, navigation, forms, components) against UX best practices and user-psychology principles (feedback, error prevention, recognition-over-recall, Fitts's law, defaults, state preservation).
**Context assumed:** Primary users are garage staff/cashiers in Cambodia, often on phones or cheap tablets, in a hurry, possibly with limited English.

## Verdict

**The visual foundation is solid — clean token-based styling, sensible information hierarchy, real responsive intent — but the app fails the most important psychological contract: it does not tell users when things go wrong, and in places it silently loses their work.** There is not a single error toast, alert, or `onError` handler in the entire frontend. Combined with a few dead-end flows, the app currently *feels* trustworthy while quietly being unreliable — the worst combination for user trust, because users blame themselves.

---

## Critical

### 1. Every failed action is silent

No toast system, no `onError` handlers, no error surfaces on any mutation (create/update/delete/void/complete). The only feedback is the button label reverting from "Creating..." to "Create Invoice."

Why this matters most: the backend (correctly) now rejects overselling — "insufficient stock for product X." When a cashier invoices 4 tires with 2 in stock, the request fails and **nothing happens on screen**. The modal stays open; the cashier clicks again, then assumes the app is broken or that the invoice was created. Psychology: users interpret silence after action as success (no news = good news). Silent failure converts every backend safeguard into a user-facing mystery.

**Fix:** Add a toast system (e.g. `sonner`, ~3KB) and a global `onError` in the `QueryClient` mutation defaults that surfaces the API's error message. The backend already returns structured `{error: {code, message}}` — use it.

### 2. The Invoices list is invisible on mobile

`Invoices.tsx:94` renders the table inside `hidden md:block` — but unlike Inventory, Customers, and ServiceJobs, there is **no `md:hidden` card view**. On a phone, the Invoices page shows the filters, the pagination... and no invoices. For a garage where the phone is likely the primary device, a core screen is simply blank.

### 3. Mechanic notes are typed into a void

`ServiceJobDetail.tsx:86-96`: the Diagnosis and Work Performed textareas are local state that is **never saved** — no save button, no mutation, and they're initialized to `''` instead of the job's existing values. A mechanic types their findings, navigates away, and the work is gone; returning to the page shows empty boxes even if data exists in the DB. This is direct data loss disguised as a working feature.

### 4. "Mark Completed" is one tap, irreversible, unconfirmed

The backend state machine (correctly) allows no transitions out of `completed`. But the UI offers `Mark Completed` as a plain single-tap button with no confirmation and no undo (`ServiceJobDetail.tsx:191`). One mis-tap permanently locks the job. Error-prevention rule: friction must be proportional to irreversibility. Add a confirm step ("Complete JOB-2026-0012? Items can no longer be changed"), or better, support undo.

### 5. The invoice creation form hides the money and drops rows silently

The single most important form in the app (`Invoices.tsx:146-204`) has four psychology failures:

- **No running total.** The user enters items, discount, exchange rate — and commits without ever seeing what the invoice will total. People need to preview consequences before committing to a money transaction; the current flow is "sign first, read later."
- **Rows are silently dropped.** `filter(i => i.quantity > 0 && i.unit_price_usd > 0)` removes any row with price 0 (e.g. free labor, warranty work) with no warning — the printed invoice differs from what the user typed.
- **Clicking "Create Invoice" with no valid items does nothing at all** (`if (items.length === 0) return`) — a button that sometimes ignores clicks teaches users the app is flaky.
- **No way to remove an item row** once added — you can `+ Add item` but never take one away. (Also: the `formNotes` state has no setter and no input — dead code for a field the backend supports.)

---

## High — user psychology

### 6. Navigation contradicts users' mental map

- **The Users page is unreachable.** `/users` exists as a route, but neither the sidebar nor mobile nav links to it. Admins can only find it by typing the URL — undiscoverable features don't exist.
- **Customers is missing from the mobile nav** (`MobileNav.tsx`), so on a phone, customer profiles, vehicles, and history are completely unreachable — not just hidden, there is no alternate path.
- **Active-state highlighting breaks on detail pages.** `location.pathname === item.path` is an exact match, so on `/invoices/17` no nav item is highlighted; users lose the "you are here" signal exactly when they're deepest in the app. Use prefix matching.
- **Back buttons destroy context.** `← Back` on detail pages calls `navigate('/invoices')`, resetting filters, page, and scroll. A user working through page 3 of unpaid invoices is thrown back to page 1 of everything after each one they open. Use `navigate(-1)` and keep filter state in the URL query string.

### 7. Defaults are teaching users the wrong exchange rate

The invoice form hardcodes `4050` (`Invoices.tsx:37`), and "Generate Invoice" from a job hardcodes `exchange_rate: 4050` (`ServiceJobDetail.tsx:173`) — ignoring the rate the admin maintains in Settings (there's even a dedicated `GET /settings/exchange-rate` endpoint). Defaults carry authority: users assume the pre-filled value is correct. When the market rate moves, every invoice quietly uses stale numbers, in a dual-currency business where this is the core feature.

### 8. Product and customer pickers hit a wall at 20 records

The invoice and job-item forms populate plain `<select>`s from `useProducts({})` / `useCustomers({})` — page 1 only, 20 rows. Once the garage has its 21st tire SKU, it cannot be invoiced through the UI, and nothing explains why. These need searchable async comboboxes (or at minimum `per_page: 100` as a stopgap). Note the job-item picker (`ServiceJobDetail.tsx:146`) shows stock counts in the options — that's excellent; the invoice form's picker should do the same.

### 9. Destructive actions don't look destructive, and nothing is forgiving

- Delete/Void confirm buttons are `outline`/`ghost` — visually identical to Cancel. The danger action should be the visually loudest (red), and spatially separated. The Button component has no `destructive` variant at all.
- The delete dialog says "Delete this product?" without naming it. Confirmations that don't restate the object don't prevent slips — users confirm on autopilot. Say "Delete **Michelin 205/55R16**?"
- Removing a job item is an instant unconfirmed `×` (a ~32px ghost button — also too small a touch target for greasy workshop fingers). No undo exists anywhere in the app.

### 10. Status colors change meaning between screens

Dashboard: `paid` = green, `issued` = blue, `unpaid` = yellow. Invoices list & detail: `paid` = teal (`primary/10`), everything else gray; on the detail page `unpaid` = **red**. Recognition-over-recall: a status badge is a glanceable code users learn once — here the same status wears different colors per screen, so the code never sticks. There's also a duplicate `StatusBadge` implemented inline in `Dashboard.tsx` alongside `components/servicejob/StatusBadge.tsx`. Consolidate into one shared badge with one semantic mapping.

---

## Medium

11. **Loading states cause layout jumps.** Every page swaps the entire content for a `"Loading..."` text line, so filter/page changes collapse and re-expand the layout. Use skeleton rows and TanStack Query's `placeholderData: keepPreviousData` so pagination doesn't blank the table.
12. **Search fires a request per keystroke** (`Inventory.tsx:78`) — no debounce. Typing "Michelin" issues 8 queries; on slow connections responses can race. Debounce ~300ms.
13. **Keyboard and screen-reader access is broken in places.** Clickable table rows and dashboard cards are `div`/`tr` with `onClick` only — no `tabIndex`, no Enter/Space handling, invisible to keyboards. Modals have no Escape-to-close, no focus trap, no `role="dialog"`/`aria-modal`. The sidebar collapse button and `×` buttons have no `aria-label`. Also `focus-visible:ring-ring` references a `--color-ring` token that is never defined in `index.css` — verify the focus ring is actually visible.
14. **Number formatting undermines the dual-currency promise.** KHR renders as `810000៛` (`InvoiceDetail.tsx:96`) — unreadable without thousand separators; use `toLocaleString('km-KH')`-style grouping. USD amounts also lack grouping (`$12345.67`). And outside the invoice detail, KHR appears nowhere — lists and dashboard are USD-only despite dual currency being a headline feature.
15. **Browser tab says "frontend".** `index.html` title was never set, and no per-route titles exist. Also `index.css` declares Plus Jakarta Sans / JetBrains Mono but no font is ever loaded — the app silently renders in system fallbacks (decide: load them or delete the declaration).
16. **Settings page shares one mutation across all sections** — pressing Save in one section flips every SaveButton to "Saved!" (and it never dismisses); the tax save fires two racing mutations where the second's state overwrites the first's. Also staff without `settings:view` still see Settings in the nav, get a silent 403, and the page renders defaults as if they were real values — a misleading dead end. Filter nav items by permission.
17. **Payment form starts blank instead of pre-filled.** On `InvoiceDetail`, Status shows "Select..." and Paid Amount is empty rather than the invoice's current values; the most common action (mark fully paid, cash) takes 4 interactions. Pre-fill current values and add a one-tap "Mark as paid" that sets `paid_amount = total`. There's also no success feedback after updating.
18. **Small touch targets:** `size="sm"` buttons are 32px tall, the qty input is 60px wide, mobile nav has no `env(safe-area-inset-bottom)` padding — on gesture-nav phones the bottom bar sits on the home indicator. Minimum 44px targets on touch surfaces.

---

## What's Good

- Clean semantic token system (`primary`, `muted`, `destructive`) instead of scattered hex values; consistent spacing rhythm; one icon family (Lucide), no emoji-as-icons.
- Real responsive architecture: sidebar on desktop, bottom nav (≤5 items, icon + label) on mobile, card fallbacks for tables on 3 of 4 list pages.
- Progressive disclosure done right in `ProductForm` — tire spec fields appear only when type = tire.
- Every mutation button shows a pending state ("Saving...", disabled) — the success half of feedback is consistently there.
- Void requires a typed reason (a good forcing function); deletes are confirmed; login form has proper labels, `autoComplete`, and a visible error message.
- Empty states exist everywhere and are phrased helpfully ("All items are well stocked").
- The job-item product picker showing live stock count in the option label is exactly the right instinct — extend it.

---

## End-to-End Journey Walkthrough: "Customer wants new tires"

Scenario: a customer arrives at the counter wanting 4 new tires (205/55R16), pays cash, expects a receipt. This is the shop's bread-and-butter transaction. Traced click-by-click through the code:

### Step 1 — Register the customer & vehicle ✅ (works, but slow)
Customers → Add Customer (modal) → open the customer → Add Vehicle (modal). Two screens, two modals, ~8 fields. Functional. But note: customer search is **name-only** — garages look up returning customers by **plate number**, and nothing in the app searches by plate.

### Step 2 — Create the service job ❌ **The journey breaks here**
Service Jobs → New Job. The form (`ServiceJobs.tsx:125`) has exactly two fields: *Description* and *"Customer Info (optional)"* — a **free-text input that is saved into the notes column**. There is no customer picker and no vehicle picker. The backend fully supports `customer_id`/`vehicle_id` on job creation; the UI never sends them.

Consequences cascade through the entire product:
- Every job created through the UI is an anonymous "Walk-in" — the customer you just carefully registered **cannot be attached to the work**.
- The Service History panel on CustomerDetail will be empty forever — jobs never link to customers.
- The invoice generated from the job inherits `customer_id = null`, so invoices are anonymous too, and the dashboard shows "Walk-in" for everything.
- The entire CRM half of the app (customers, vehicles, history) is a data graveyard: information goes in, and no workflow ever reads it.

There's also no "New Job" button on the customer's own page — the natural place to start after looking them up.

### Step 3 — Find the tire and add it to the job ⚠️
Job detail → Add Item → a plain `<select>` of the **first 20 products**, no search, no tire-size filter. The backend inventory API supports filtering by `tire_size` — a tire shop's #1 lookup — but neither this picker nor even the Inventory page exposes it. Once the shop has 21+ SKUs, some tires literally cannot be invoiced. (Credit: the options do show price and live stock — the right instinct, wrong container.)

Also: to charge fitting labor, a "labor" product must already exist in inventory (`service_job_items.product_id` is NOT NULL). Nothing in the UI explains or guides this — a shop that never created a "Tire fitting" product simply cannot bill labor on jobs.

### Step 4 — Mechanic does the work, notes go nowhere ❌
Diagnosis / Work Performed textareas are never saved (Critical #3). Also, adding items to a job doesn't reserve or check stock — reasonable design (stock moves at invoicing), but see what happens next.

### Step 5 — Complete and invoice ⚠️→❌ (worst case: silent dead end)
"Mark Completed": one tap, irreversible, no confirmation (Critical #4). Then "Generate Invoice" — hardcoded `exchange_rate: 4050`, no preview of the total.

Now the trap: stock was never checked when items were added to the job. If another sale consumed the tires in the meantime, the backend correctly rejects the invoice with "insufficient stock" — and the UI **shows nothing** (Critical #1). The tires are physically on the car, the job is irreversibly completed, the invoice fails silently, and the counter staff has a paying customer and no way to understand what's wrong. This is the compound failure of Critical #1 + #4: the app's worst moment lands at the highest-stakes point of the journey, with a customer standing at the counter.

### Step 6 — Take payment ⚠️
Invoice detail → fill Status + Paid Amount + Method → Update Payment → no success feedback. Cash-in-full is the overwhelmingly common case and takes ~6 interactions with a blank form; it should be one "Mark as paid (cash)" button.

### Step 7 — Hand the customer a receipt ❌ **The journey has no ending**
The PDF endpoint returns 501 Not Implemented, there is no print stylesheet, and Telegram sending is a stub. After all of the above, **there is nothing to give the customer.** The transaction's peak moment (per the peak-end rule, the one the customer remembers) doesn't exist.

### Journey verdict

Individually the screens look finished; composed into the actual job-to-cash flow, the journey **breaks in the middle** (jobs can't be tied to customers), **gets risky at the climax** (silent stock failure after irreversible completion, after physical work is done), and **has no ending** (no receipt). Roughly 25+ interactions across 5 screens and 6 modals — and the two things that matter most (finding the right tire, handing over a receipt) are the two hardest.

**Journey-level fixes, in order:**
1. Add customer + vehicle pickers to the job form (backend already supports it) and a "New Job" button on CustomerDetail — this single fix reconnects the CRM half of the product.
2. Warn about stock at item-add time (data is already in the picker) so failure doesn't wait until after the work is done — and surface the invoice error when it does happen.
3. Ship a printable receipt (even a simple print-stylesheet invoice page beats the 501) — the journey needs an ending.
4. "Mark as paid (cash)" one-tap action; plate-number search; tire-size filter in pickers and Inventory.

---

## Visual Design Assessment: Is It Too Plain?

**Verdict: the plainness is the right call for this product — the issue is that it reads as "small and cramped" rather than "calm and simple."** For garage staff at a counter, on cheap phones, in a hurry, a restrained data-first UI is correct; decoration (gradients, illustrations, animation) would make it worse. What separates it from *intentionally* minimal:

1. **Everything is one size too small.** Body 14px, labels 12px, page titles only 18px (barely above body — weak hierarchy), dashboard KPI numbers 20px. For this audience the priority is glanceability: 16px base (also prevents iOS auto-zoom on inputs), titles 22–24px, money totals 28–30px. Keep the density; enlarge the anchors.
2. **The designed personality never renders.** `index.css` declares Plus Jakarta Sans and JetBrains Mono but no font is ever loaded — every user sees system fallback. Load them (self-hosted `@font-face` with `font-display: swap`) or delete the declaration. A large share of the "plain" impression comes from this.
3. **Cramped, not calm.** `gap-2` (8px) between cards, `p-3` padding, `0.25rem` radius, thin borders everywhere. Whitespace is what makes minimal look designed. Card padding 16–20px, section gaps 16–24px, radius ~0.5rem — big perceived-quality gain, zero added complexity.
4. **Inconsistently plain.** The Dashboard has personality (colored left-border metric cards, green/purple/amber accents, tinted hover rows — all hardcoded off-token colors) while every other page is pure gray; it feels like two apps. Spread the Dashboard's functional color (status tints, accent borders) to the rest via tokens, since status color is functional in this domain, not decorative.
5. **The login page is the one place plainness costs.** First impression, bare 18px heading in a small gray card — looks like a placeholder. An unused `hero.png` sits in `src/assets`. Logo mark + brand color + slightly larger card ≈ one hour.

**Direction:** don't make it fancier — make it *bigger, airier, and consistent*. Target aesthetic: "clean POS terminal." Roughly a day of cosmetic work (type scale, spacing, fonts, login, unified badge system), and all of it ranks below the workflow fixes above — a beautiful app that can't link a customer to a job is still broken; a plain one that flows well feels great to use.

---

## Recommended Priority Order

1. **Global error feedback** — toast + `QueryClient` mutation `onError` default (unblocks trust in everything else)
2. **Invoices mobile card view** (a core screen is currently blank on phones)
3. **Wire up diagnosis / work-performed saving** and pre-fill existing values (data loss)
4. **Invoice form: live total preview, removable rows, warn instead of silently dropping rows**
5. **Confirm dialog for Mark Completed**; destructive button variant + named confirmations
6. **Navigation: add Users link (admin), Customers on mobile, prefix-match active state, preserve filters in URL**
7. **Fetch exchange rate from settings everywhere `4050` is hardcoded**
8. Searchable product/customer pickers; then work down the Medium list
9. Visual polish pass (type scale, spacing, load fonts, login branding, unified badges) — worthwhile, but only after the workflow fixes above
