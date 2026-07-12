# Layout Improvement Plan

**Status:** ✅ All 4 phases implemented. Gap (KPI query params) fixed. Nits (sticky headers, motion-reduce, focus trap) resolved.

## Verification — 2026-07-07

Implementation reviewed; `tsc` and production build pass. **All four phases are implemented as specified**, with one functional gap and three cosmetic nits:

- ✅ Phase 1: `max-w-6xl` container (print overrides intact), spacing tiers applied (`space-y-6` sections, `gap-4` grids, `p-4` cards), radius `0.5rem`, `PageHeader` (2xl title, breadcrumb, back, actions) adopted on all 10 pages.
- ✅ Phase 2: InvoiceDetail & ServiceJobDetail restructured to `lg:grid-cols-3` main-column + rail; invoice total elevated to `text-2xl font-bold`; print-receipt path verified intact.
- ✅ Phase 3: `SlideOver` built with Escape-to-close, focus-on-open, `role="dialog"`/`aria-modal`/`aria-label`, close button, animated scrim — used for ProductForm and New Invoice; confirmations correctly stayed as modals.
- ✅ Phase 4: sticky-header classes + icon action buttons on tables, KPI values `text-3xl`, all four KPI cards clickable, Login upgraded (max-w-md, brand title, gradient).

**Gap to fix:** Dashboard KPI cards navigate to `/invoices?payment_status=paid` and `/inventory?low=1`, but neither page reads those query params — the click lands on an *unfiltered* list. Wire `useSearchParams` into Invoices (init `paymentFilter` from `payment_status`) and Inventory (init `lowStockOnly` from `low=1`), or the KPI links silently lie.

**Nits (non-blocking):** `sticky top-0` on table headers is inert — the ancestor has `overflow-hidden`, so it never sticks (either give the table wrapper a fixed height + `overflow-y-auto`, or drop the classes); `SlideOver` ignores `prefers-reduced-motion` (add `motion-reduce:transition-none`); focus isn't trapped inside the SlideOver (Tab can reach the background — acceptable for now, plan only required focus-on-open + Escape).

---

**Date:** 2026-07-07
**Goal:** Move the layout from "developer default" to "clean POS terminal" — one coherent cosmetic pass, zero logic changes.
**Scope:** frontend only. No API, hook, or data changes. Estimated effort: ~1 day.
**Context:** see the "Visual Design Assessment" section in [UX_AUDIT.md](./UX_AUDIT.md) for the diagnosis. The five layout problems this fixes: fluid unbounded width, flat one-value spacing, equal-weight boxes, weak page headers, modal overuse.

---

## Phase 1 — Foundations (touch every page at once)

### 1.1 Content container

**File:** `src/components/layout/MainLayout.tsx`

Cap and center the content area:

```tsx
<main className="md:ml-[240px] print:ml-0 print:p-0">
  <div className="max-w-6xl mx-auto p-4 md:p-6 pb-20 md:pb-8">
    <Outlet />
  </div>
</main>
```

- `max-w-6xl` (1152px) — right for data tables; do NOT go narrower or Inventory's 6-column table cramps.
- Keep `print:*` overrides intact (receipt printing depends on them).

### 1.2 Spacing tiers

Define three tiers and apply them mechanically — this is a find-and-replace pass, not a judgment call per page:

| Tier | Use for | Class |
|---|---|---|
| **Section** | Between major page blocks (header → filters → table → pagination; card grids) | `space-y-6` / `gap-4` |
| **Group** | Inside a card: label → content, form field groups | `space-y-3` / `gap-3` |
| **Tight** | Label + input pairs, badge rows, icon + text | `space-y-1.5` / `gap-2` |

Concrete replacements (all pages):
- Page root `space-y-4` → `space-y-6`
- Card-grid `gap-2` → `gap-4`
- Card padding `p-3` → `p-4` (dashboard KPI cards may stay `p-4` with `gap-4`)
- Form `space-y-2` → `space-y-3`; keep `space-y-1` → `space-y-1.5` for label+input
- `index.css`: bump `--radius` from `0.25rem` to `0.5rem`

### 1.3 PageHeader component (new)

**File:** `src/components/layout/PageHeader.tsx`

One consistent header band for every page:

```tsx
interface PageHeaderProps {
  title: string
  backTo?: string | number        // renders ← back button (navigate(-1) when number)
  breadcrumb?: string             // e.g. "Invoices" shown above title on detail pages
  badges?: ReactNode              // status badges, rendered after title
  actions?: ReactNode             // right-aligned buttons
  subtitle?: string               // e.g. "Welcome, Sokha" on Dashboard
}
```

Layout: single row, `flex items-center gap-3`, title `text-2xl font-semibold` (up from `text-lg` — this is the typography fix riding along), actions pushed right with `ml-auto`, `border-b pb-4` beneath. On detail pages the breadcrumb renders as `text-xs text-muted-foreground` above the title (`Invoices / INV-2026-0012`), replacing the current inline "← Back" + title + badges jumble; badges sit next to the title and wrap *below* it on narrow screens (`flex-wrap` on an inner group, not the whole row).

**Adopt in:** Dashboard, Inventory, Customers, CustomerDetail, ServiceJobs, ServiceJobDetail, Invoices, InvoiceDetail, Settings, Users (10 pages).

---

## Phase 2 — Detail pages: main column + rail

Restructure the two heavy detail pages from equal-weight stacked cards into a dominant column with a supporting rail. Grid: `lg:grid-cols-3 gap-4` — content spans 2, rail spans 1. Below `lg`, rail stacks *after* the main column (mobile priority: the bill first).

### 2.1 InvoiceDetail

- **Main column (`lg:col-span-2`):** Items card (the bill — make totals the visual peak: Total USD `text-2xl font-bold`, KHR line under it), then Payments card (history + Pay Full / Record Payment).
- **Rail (`lg:col-span-1`):** Customer card, Notes card, Actions card (Void — destructive, spatially isolated at the rail bottom).
- Print receipt path unchanged.

### 2.2 ServiceJobDetail

- **Main column:** Items card + total, then Diagnosis / Work Performed (side-by-side pair collapses to stacked inside the column), Save Notes button.
- **Rail:** Customer/vehicle card, Description card, action buttons (Generate Invoice / View Invoice / Mark Completed — the completion confirm strip stays inline where the button is).

### 2.3 CustomerDetail

Already close: keep two-column top (Contact | Notes), Vehicles and Service History full-width below — just apply the new spacing tiers and PageHeader.

---

## Phase 3 — Slide-over panel for big forms

**File:** `src/components/ui/SlideOver.tsx` (new)

Right-side panel replacing the centered modal *only* for the two big forms:

```tsx
interface SlideOverProps {
  open: boolean
  onClose: () => void
  title: string
  children: ReactNode
}
```

- Fixed right, `w-full sm:max-w-md` (28rem), full height, `overflow-y-auto`, scrim `bg-black/50` behind, panel `bg-card border-l shadow-xl`.
- Slide-in via `transition-transform duration-200` (translate-x-full → 0); respect `prefers-reduced-motion` (no transform, instant).
- Accessibility while we're here (closes an open audit item for these two flows): `role="dialog"` `aria-modal="true"` `aria-label={title}`, close on Escape, close button top-right, focus the panel on open.

**Migrate:** ProductForm (Inventory add/edit) and the New Invoice form (Invoices). **Keep as modals:** delete/void confirmations, job create form, vehicle/customer forms, Add Item — they're small and centered-confirm is the right pattern for them.

## Phase 4 — Table & dashboard polish (small, bounded)

- **Tables (Inventory, Invoices, Customers, ServiceJobs):** header row `sticky top-0` inside the rounded container; row height `py-2.5`; actions column becomes icon buttons (`Pencil`, `Trash2` from lucide, `size="icon"` ghost, with `aria-label` + `title`) instead of text ghost buttons — reclaims ~120px of width.
- **Dashboard:** KPI value `text-2xl` → `text-3xl`; make all four KPI cards clickable filters (Revenue → `/invoices?payment_status=paid`, Jobs → `/service-jobs`, Customers → `/customers`, Low Stock → `/inventory?low=1` — Inventory already reads state for this) with `hover:shadow-md transition-shadow` affordance; chart height 200 → 220.
- **Login:** center card `max-w-sm` → `max-w-md`, add app name `text-2xl font-bold text-primary` + one-line tagline; optional: brand-tint the background (`bg-gradient-to-br from-primary/5 to-background`). Keep it restrained.

---

## Order of work & verification

1. Phase 1 (container, spacing, PageHeader) — one commit, visually diff every page.
2. Phase 2 (detail rails) — one commit.
3. Phase 3 (slide-overs) — one commit.
4. Phase 4 (tables/dashboard/login) — one commit.

After each phase: `npx tsc -b && npm run build`, then eyeball at 375px, 768px, 1440px widths. After Phase 2, print a receipt (verify `print:` classes survived the restructure). After Phase 3, tab-navigate the slide-over (Escape closes, focus lands inside).

**Explicitly out of scope:** color/token changes beyond radius, font loading changes (done), jobs kanban board (tracked in INDUSTRY_COMPARISON.md "adopt soon"), dark mode.
