# AutoStock - Agent Instructions

## Project Overview

Garage management system for small auto workshops in Cambodia (tire sales + general repair). Single location MVP, multi-branch ready.

**Stack:**
- Backend: Go + Gin + PostgreSQL
- Frontend: React + TypeScript + Vite + TanStack Query + Tailwind + shadcn/ui
- Deployment: Docker Compose

**Key docs (read these first):**
- `docs/ARCHITECTURE.md` - System design, module structure
- `docs/DATABASE.md` - Complete schema
- `docs/API.md` - All endpoints
- `docs/FRONTEND.md` - Design system, component library, mobile patterns
- `docs/ROADMAP.md` - Development phases

## Architecture

**Modular monolith** with domain modules: `auth`, `inventory`, `customer`, `service`, `invoice`, `settings`, `telegram`.

Each module follows: `handler/` → `service/` → `models/` → `dto/`

**Monorepo structure:**
```
backend/  - Go application (cmd/server, internal/, pkg/)
frontend/ - React app (src/components, src/hooks, src/pages)
docs/     - All documentation
```

## Critical Conventions

**Dual currency:** All monetary values stored in USD. KHR calculated via exchange rate (stored per-invoice for historical accuracy).

**Multi-branch ready:** All tenant tables have `branch_id`. Middleware extracts from JWT. Queries filter by branch.

**Invoice numbering:** `INV-YYYY-NNNN` format (e.g., `INV-2026-0001`).

**Tire-specific fields:** Products table includes tire_size, tire_brand, tire_model, dot_code, load_index, speed_rating, tire_type (nullable for non-tire products).

**Permissions:** Array of strings in user record (e.g., `["inventory:view", "invoice:create"]`). Admin gets all. Staff gets configurable subset.

**Design:** Avoid generic AI aesthetics. Plus Jakarta Sans font, teal-blue primary color (not generic `blue-600`), `rounded` = 4px (not `rounded-lg`), dense spacing on desktop. Collapsible sidebar on desktop, bottom nav on mobile. See `docs/FRONTEND.md` for full design system.

## Development Commands

*Not yet available - code not implemented. Update this section when backend/frontend are initialized.*

Expected commands:
```bash
# Backend
cd backend && go run ./cmd/server
cd backend && go test ./...

# Frontend  
cd frontend && npm run dev
cd frontend && npm run build
cd frontend && npm run test

# Database
docker-compose up -d postgres
migrate -path ./migrations -database "$DATABASE_URL" up
```

## Behavioral Guidelines

## 1. Think Before Coding

**Don't assume. Don't hide confusion. Surface tradeoffs.**

Before implementing:
- State your assumptions explicitly. If uncertain, ask.
- If multiple interpretations exist, present them - don't pick silently.
- If a simpler approach exists, say so. Push back when warranted.
- If something is unclear, stop. Name what's confusing. Ask.

## 2. Simplicity First

**Minimum code that solves the problem. Nothing speculative.**

- No features beyond what was asked.
- No abstractions for single-use code.
- No "flexibility" or "configurability" that wasn't requested.
- No error handling for impossible scenarios.
- If you write 200 lines and it could be 50, rewrite it.

Ask yourself: "Would a senior engineer say this is overcomplicated?" If yes, simplify.

## 3. Surgical Changes

**Touch only what you must. Clean up only your own mess.**

When editing existing code:
- Don't "improve" adjacent code, comments, or formatting.
- Don't refactor things that aren't broken.
- Match existing style, even if you'd do it differently.
- If you notice unrelated dead code, mention it - don't delete it.

When your changes create orphans:
- Remove imports/variables/functions that YOUR changes made unused.
- Don't remove pre-existing dead code unless asked.

The test: Every changed line should trace directly to the user's request.

## 4. Goal-Driven Execution

**Define success criteria. Loop until verified.**

Transform tasks into verifiable goals:
- "Add validation" → "Write tests for invalid inputs, then make them pass"
- "Fix the bug" → "Write a test that reproduces it, then make it pass"
- "Refactor X" → "Ensure tests pass before and after"

For multi-step tasks, state a brief plan:
```
1. [Step] → verify: [check]
2. [Step] → verify: [check]
3. [Step] → verify: [check]
```

Strong success criteria let you loop independently. Weak criteria ("make it work") require constant clarification.

---

**These guidelines are working if:** fewer unnecessary changes in diffs, fewer rewrites due to overcomplication, and clarifying questions come before implementation rather than after mistakes.