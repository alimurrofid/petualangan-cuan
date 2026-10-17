# 🤖 AGENTS.md — Petualangan Cuan (Global Guide)

Global guide for AI agents & developers. Each area has its own detailed guide — **it is mandatory to read the area guide before working in that area**:

| Area | Guide File | Contents |
|---|---|---|
| 🚀 Backend API (Go) | [`backend-go/AGENTS.md`](backend-go/AGENTS.md) | Layering, commands, responses, tests, verification |
| 🌐 Frontend Web (Vue 3) | [`frontend-vue/AGENTS.md`](frontend-vue/AGENTS.md) | Pinia, strict typing, shadcn-vue, lint baseline |
| 🚪 WA Gateway (Go) | [`wa-gateway/AGENTS.md`](wa-gateway/AGENTS.md) | DDD, JID vs Device ID, verification |

Rules in area files **tighten** this file (they may be stricter, never looser). Full architecture documentation lives in [README.md](README.md).

---

## 🛡️ 1. Security & Secrets Management (CRITICAL)

> [!CAUTION]
> ### PRIVACY POLICY & .ENV FILES
> 1. **STRICTLY FORBIDDEN TO READ / ACCESS / PRINT REAL `.env` FILES** (`.env`, `.env.prod`, `.env.staging`, or any other local secret files).
> 2. **USE TEMPLATE FILES ONLY**: Environment variable references may **ONLY** point to `.env.example` files (root & each project).
> 3. **NEVER WRITE REAL CREDENTIALS/SECRETS** into source code, markdown files, commit messages, or log output.
> 4. New features requiring config variables → register them in `.env.example` with explanatory comments.

---

## 🧹 2. Global Engineering Principles

> [!IMPORTANT]
> ### No Legacy Baggage — NO BACKWARD COMPATIBILITY NEEDED
> - **Delete dead code / leftover old implementations**: functions, entities, interfaces, components, styling, files, or routes that were replaced or are unused → **MUST BE DELETED IMMEDIATELY**.
> - **Never write retro adapters / legacy shims** for compatibility with old, unused implementations.
> - Keeping leftover old code makes future AI agents hallucinate and import the wrong modules.

> [!WARNING]
> ### FAIL-FAST — Never Mask Errors
> - Logic errors, schema mismatches, or failed requests: **let them fail loudly and explicitly**, then handle with proper error handling. No empty try-catch blocks, no fake default values, and no **double property fallbacks** (`a.field1 || a.field2`) that hide wrong column/API contract names (detailed rule in `frontend-vue/AGENTS.md`).
> - **EMPTY VALUE STANDARD**: Data layer (API/state/DB/logic) → **`null`** per strict types. UI/display layer → render **`-`** (dash).

---

## 🗺️ 3. Tech Stack

| Component | Technology |
|---|---|
| Backend API | Go 1.26, Fiber v2, GORM, Zerolog, Swagger |
| Frontend Web | Vue 3 (`<script setup>`), TypeScript, Vite, Tailwind v4, shadcn-vue, Pinia, pnpm |
| WA Gateway | Go, Whatsmeow, SQLite |
| AI Subsystem | Llama.cpp (local GGUF), Faster-Whisper, External AI Strategy (Gemini/OpenRouter) |
| Database | PostgreSQL 15 |
| Observability | Promtail, Loki, Grafana, Uptime Kuma; structured JSON logs (`X-Request-Id`) |
| Infra | Docker & Docker Compose (local, staging, prod) |

---

## ⚡ 4. Docker Commands (Local)

```bash
docker compose up -d --build      # Start the entire local stack
docker compose logs -f backend    # Logs for a specific service
docker compose down               # Stop the stack
```

Native development commands per area: see that area's AGENTS.md.

---

## 🧰 5. Installed Project Skills (`.opencode/skills/`)

| Skill | Purpose |
|---|---|
| `verify` | Verification loop for touched areas (make check / pnpm lint & typecheck / go vet & test), must pass before done |
| `add-endpoint` | SOP for adding a backend endpoint/CRUD (entity → repo → service → handler → route → `make check`) |
| `add-view` | SOP for adding a frontend page (store → view → router → menu → `pnpm lint && typecheck`) |
| `cleanup` | Evidence-based sweep of dead code / property fallbacks, then verification |

Skills are auto-discovered from their `description` when the task matches, or loaded manually via `@<skill-id>` (e.g. `@verify`). Skill rules tighten area AGENTS.md files — **keep both sides in sync** whenever project conventions change.

---

## ✅ 6. SOP & Pre-Completion Checklist

1. [ ] **Leftover dead code?** Delete unused functions/variables/imports/files.
2. [ ] **Sloppy property fallbacks (`||`)?** Empty data → `null` in state/data, `-` in display. (Frontend: `pnpm lint` also catches fallback chains via ESLint.)
3. [ ] **`.env` leak?** Real `.env` files untouched; references only to `.env.example`.
4. [ ] **Fast verification passing?** (GitHub Actions runs the same gate on push/PR to `main`/`staging` via `.github/workflows/ci.yml`.)
   - Backend: `cd backend-go && make check` (vet + test + build).
   - Frontend: `cd frontend-vue && pnpm lint && pnpm typecheck`.
   - WA Gateway: `cd wa-gateway/src && go vet ./... && go test ./...`.
5. [ ] **Architecture intact?** No layer bypass (see area guides).
