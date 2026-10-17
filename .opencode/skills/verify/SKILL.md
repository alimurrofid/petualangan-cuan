---
name: Verify Changes
description: >
  Run this project's fast verification loop for the areas touched by current
  changes (backend make check, frontend pnpm lint + typecheck, wa-gateway go
  vet + test), fix any failures, and report pass/fail. Use after finishing a
  task, before declaring work done, or when the user says "verify", "check",
  "cek", "validasi", or "uji perubahan".
---

# Verify Changes

The mandatory verification procedure for the Petualangan Cuan repo. Detailed conventions per area live in each `AGENTS.md` (root, `backend-go/`, `frontend-vue/`, `wa-gateway/`).

## Steps

1. Determine which areas were touched by the changes:
   ```bash
   git status --short
   ```
   If there are no tracked changes or it is unclear, ask the user which area is meant. Do not verify every area when only one changed.

2. Run the verification for every touched area:

   | Area | Command | Includes |
   |---|---|---|
   | `backend-go/` | `make check` | `go vet` + `go test ./...` + `go build` |
   | `frontend-vue/` | `pnpm lint` then `pnpm typecheck` | ESLint auto-fix + report, then vue-tsc |
   | `wa-gateway/src/` | `go vet ./... && go test ./...` | static check + unit tests |

3. Evaluate failures with these rules:
   - **ESLint**: errors in NEW code → **fix them, do NOT suppress**. Adding/renaming entries in `eslint-suppressions.json` is only allowed after legacy is actually cleaned and the user approves the baseline trim (`pnpm exec eslint . --prune-suppressions`).
   - **TypeScript / go vet / build errors** → fix directly (fail-fast, no `||` fallbacks or empty try-catch).
   - **Failing tests** → follow systematic debugging: find the root cause first, never patch blindly.
   - Fixes must be real; repeat steps 2–3 until everything is green.

4. Final report to the user, concise per area:
   - ✅/❌ per verification component (`vet`, `test`, `build` / `lint`, `typecheck`).
   - Summarize what was fixed & why, if fixes were applied.

## Hard Rules

- Never skip verification with reasoning like "probably fine" — run the commands.
- If a verification prerequisite is unavailable (e.g. a DB for tests that need one), explicitly report what was not verified and why — never assume it passed.
- The root AGENTS.md end-of-task checklist still applies on top: dead code, `||` property fallbacks in frontend, `.env` untouched/never printed.
