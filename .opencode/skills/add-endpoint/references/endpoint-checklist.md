# Endpoint Checklist

Per-file checklist for adding a new domain/endpoint in `backend-go/`. Use together with SKILL.md.

## 1. Entity — `internal/entity/<domain>.go`

- [ ] One entity per file, snake_case filename per domain (`debt_payment.go`).
- [ ] Clean GORM struct: data fields only, correct GORM/SQL tags; no logic.
- [ ] Column names consistent with other entities — consumers must never compensate with double fallbacks for wrong column names.

## 2. Repository — `internal/repository/<domain>_repository.go`

- [ ] Direct GORM queries only; standard CRUD functions follow the patterns of other repositories.
- [ ] Accepts `*gorm.DB`/context the same way existing repositories do.
- [ ] If the service will be tested: add/implement the interface mock in `internal/repository/mock/`.

## 3. Service — `internal/service/<domain>_service.go`

- [ ] All business rules & logic validation live here.
- [ ] `x_service_test.go` written alongside; use the repository mock.
- [ ] User ID for data scoping comes from the fiber context (`pkg/utils.GetUserIDFromContext`), never from a client payload.

## 4. Handler — `internal/handler/<domain>_handler.go`

- [ ] Parse/validate input → call Service → return standard JSON.
- [ ] Explicit error handling; follow the response conventions:
  - error: `c.Status(http.StatusXXX).JSON(fiber.Map{"error": err.Error()})` with the correct status (400/404/500).
  - data success: `c.JSON(data)`; action success: `c.JSON(fiber.Map{"message": "..."})`.
- [ ] **Complete Swagger annotations** (`@Summary`, `@Description`, `@Tags`, `@Security`, `@Param`, `@Success`, `@Failure`, `@Router`) following existing handlers.
- [ ] `x_handler_test.go` included as well.
- [ ] Regenerate the docs afterwards with `make swagger` — never hand-edit `backend-go/docs/`.

## 5. Routes — `cmd/api/main.go`

- [ ] ALL routes are centralized here. New domain: `api.Group("/<resource>", middleware.Protected())`.
- [ ] Resource path naming (kebab/plural) matches existing groups (`/auth`, `/wallets`, `/categories`, `/transactions`, `/user`, `/debts`, `/wishlist`, `/saving-goals`, `/ai`).
- [ ] Public routes only when truly user-independent (the `/auth` group is the only example).

## 6. Miscellaneous

- [ ] Seeder: new dummy data goes into `internal/seeder` (invoked via `seeder.SeedAll`).
- [ ] New env vars are only registered in `.env.example` with explanatory comments.
- [ ] Final verification: `make swagger` (if annotations changed) then `make check` green before reporting done.
