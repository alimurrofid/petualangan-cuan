# 🚀 AGENTS.md — Backend API (Go)

Area-specific guide for `backend-go/`. The global rules in the root [AGENTS.md](../AGENTS.md) still apply.

## ⚡ 1. Commands & Verification

```bash
# Download dependencies
go mod tidy

# Run the backend server (dev)
make run            # = go run cmd/api/main.go

# Database
make fresh          # reset + re-migrate
make seed           # run seeder
make fresh-seed     # reset + migrate + seed

# --- Verification (MUST pass before finishing a task) ---
make vet            # go vet ./...
make test           # go test ./... (handler & service suites)
make build          # build binary → bin/api
make check          # vet + test + build in one go

# --- Swagger docs (run after changing/adding handler annotations) ---
make swagger        # = swag init -g cmd/api/main.go -o docs
                    # requires: go install github.com/swaggo/swag/cmd/swag@v1.16.4
```

> Build artifacts (`bin/api`) are gitignored; never commit build outputs.

## 🏗️ 2. Architecture Layers (STRICT — Never Bypass)

Required flow: `Handler → Service → Repository → Entity`.

| Layer | Allowed | Forbidden |
|---|---|---|
| `cmd/api/main.go` | App setup, **registration of ALL routes** via `api.Group(...)` | Business logic |
| `internal/handler/` | Input validation, parse body/query/param, call Service, return JSON | Calling Repository directly; business logic |
| `internal/service/` | All business rules, logic validation, DB transaction orchestration, calling AI providers | Direct GORM queries |
| `internal/repository/` | Direct GORM database queries | Business logic |
| `internal/entity/` | Clean GORM model structs | Logic |
| `internal/provider/ai/` | AI provider integrations (local LLM, Whisper, external) | — |
| `pkg/middleware/`, `pkg/utils/` | Cross-layer helpers (JWT auth, logger, request id, ctx utils) | Business logic |

## 📋 3. Code Conventions (based on actual codebase patterns)

1. **File naming**: entities & services use `snake_case` per domain (`transaction.go`, `debt_payment.go`, `saving_goal_service.go`). Handlers come in pairs: `x_handler.go` + `x_handler_test.go`.
2. **JSON responses** (consistent pattern used by handlers):
   - Error: `c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})` — the HTTP status must be accurate (400/404/500 per condition).
   - Data success: `c.JSON(data)` / `c.JSON(categories)` directly.
   - Action success: `c.JSON(fiber.Map{"message": "..."})`.
   - Always check errors explicitly (`if err != nil`) — no silent ignores.
3. **Swagger**: annotations (`@Summary`, `@Router`, etc.) are written in handler comments; generated output lives in `backend-go/docs/` (`docs.go`, `swagger.json`, `swagger.yaml`). Every new endpoint must have annotations, then regenerate with `make swagger` (never hand-edit `docs/`).
4. **Multi-tenancy via context**: get the user via `pkg/utils.GetUserIDFromContext(ctx)` — the userID is set into fiber locals by the `Protected()` middleware.
5. **Logging**: `zerolog` structured JSON, always include `request_id`. Never print passwords/tokens/secrets.
6. **Mocks**: repository mocks for tests live in `internal/repository/mock/`.
7. **Go target version: 1.26** (see `go.mod`). Use modern features; no backward compatibility.

## 🔒 4. Configuration

All backend environment variables are defined in [backend-go/.env.example](.env.example) — **never read the real `.env`**. Config structures live in `internal/config/` (App, Database, AI, JWT, OAuth).
