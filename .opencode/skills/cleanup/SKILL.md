---
name: Cleanup
description: >
  Hunt down and remove dead code, unused components/exports/routes/stores,
  sloppy property fallbacks (`a.x || a.y`), and other no-legacy-baggage
  violations in an area, then run the verification loop to prove nothing
  broke. Use when the user asks to bersih-bersih, cleanup, hapus kode mati,
  remove dead code, or after replacing/reworking an old implementation.
---

# Cleanup (Dead Code Sweep)

Enforcing the **no legacy baggage** principle from the root AGENTS.md. Untracked old code is a hallucination source for future agents.

## Workflow

1. **Pick the area**: ask/confirm the area (`backend-go`, `frontend-vue`, `wa-gateway`) and the scope (a specific folder vs the whole area). Read that area's AGENTS.md.
2. **Inventory candidates with evidence**, never by guessing:
   - Files/flags/exports/symbols with no references (grep the whole repo, including Dockerfiles/compose files that may use them at build/run time).
   - Vue components never imported anywhere; unused routes/stores.
   - Handlers/services/entities not registered in any route/schema.
   - Chained property fallbacks `||` (frontend), empty try-catch blocks, unused legacy wrappers/shims.
3. **Delete when confidently unused**: every deletion must come with evidence of "no references". When uncertain (possibly used via reflection/dynamic code), consult the user first — never delete silently.
4. **Verify** with the `verify` skill (or run its commands directly). Loop until green.
5. **ESLint baseline** (frontend only): when the cleanup removes legacy errors recorded in `eslint-suppressions.json`, propose the baseline trim to the user:
   ```bash
   pnpm exec eslint . --prune-suppressions
   ```
   Never run it without confirmation (this file is shared by all devs/CI).
6. Report: list what was deleted + evidence of "no references" + verification results.

## Hard Rules

- Never write shims/adapters over deleted things to be "safe" — delete completely, leaving no remnants.
- Never touch/print real `.env` files while inspecting configuration.
- Files possibly used by Docker/compose → also check `docker-compose*.yml` at the root before deleting.
