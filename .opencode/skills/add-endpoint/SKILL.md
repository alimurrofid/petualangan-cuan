---
name: Add Endpoint
description: >
  Add a new API endpoint or full CRUD domain to the Go backend following the
  repo's strict Handler → Service → Repository layering: entity, repository
  (+ mock), service (+ tests), handler (+ tests + Swagger annotations), route
  registration in cmd/api/main.go, then run make check. Use when the user asks
  to add/create a backend endpoint, CRUD baru, fitur baru backend, route baru,
  API baru, or endpoint baru.
---

# Add Endpoint (Backend API)

Adding an endpoint/new CRUD in `backend-go/`. It is MANDATORY to read [`backend-go/AGENTS.md`](../../../backend-go/AGENTS.md) first — the layering rules there must never be violated.

## Workflow

1. **Study the existing pattern before writing**: pick the most similar domain (e.g. `wishlist`, `saving_goal`) and read its file pair (entity, repository + mock, service + test, handler + test). Follow that pattern exactly — never invent a new structure.
2. Create/modify all layers per `<skill-dir>/references/endpoint-checklist.md`, in order: entity → repository (+ mock when needed) → service (+ `x_service_test.go`) → handler (+ `x_handler_test.go` + Swagger annotations) → route registration in `cmd/api/main.go`.
3. If the endpoint needs dummy data: extend `internal/seeder` (invoked via `make seed` / `make fresh-seed`).
4. If new environment variables are needed: register them in `backend-go/.env.example` with comments. **Never read/touch the real `.env`.**
5. Run verification:
   ```bash
   cd backend-go && make check
   ```
   Loop until green; never report done before `make check` passes.
6. Report: summarize the files created/changed per layer and the final endpoint (method + path + route group).

## Hard Rules

- Handlers never touch GORM/repository directly; services never write queries; repositories never contain business logic.
- Responses follow the actual codebase conventions (see `backend-go/AGENTS.md` §3).
- Swagger annotations are mandatory for every new endpoint — docs in `backend-go/docs/` are generated.
- No property fallbacks, no error swallowing.
