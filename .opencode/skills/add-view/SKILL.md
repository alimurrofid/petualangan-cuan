---
name: Add View
description: >
  Add a new frontend page/feature to the Vue 3 app: Pinia store with inline
  types, API calls via the shared axios client, view component, router entry,
  navigation menu, and complete loading/empty/error UI states. Use when the
  user asks to add/create a frontend view, halaman baru, page baru, store
  baru, fitur baru frontend, or tampilan baru.
---

# Add View (Frontend Vue)

Adding a new page/domain in `frontend-vue/`. It is MANDATORY to read [`frontend-vue/AGENTS.md`](../../../frontend-vue/AGENTS.md) first.

## Workflow

1. **Study the existing pattern**: pick a similar domain view (e.g. `views/wallet/` + `stores/wallet.ts`) and follow its structure. The list of real stores is in `frontend-vue/AGENTS.md`.
2. **New Pinia store** in `src/stores/<domain>.ts`:
   - Type interfaces **inline in the store** — there is no `src/types/` folder; do not create a duplicate.
   - All API calls only through the axios instance from `src/lib/api.ts` (never a second `axios.create`).
   - Empty state = `null` per strict types; property fallbacks `a.x || a.y` are forbidden.
3. **View** in `src/views/<domain>/` (follow the naming style of the related domain when one exists, e.g. `saving_goal/`):
   - `<script setup lang="ts">`, Tailwind v4, atomic components from `src/components/ui/` first before writing custom ones.
   - Confirmations/dialogs via `src/composables/useSwal.ts`.
4. **All three UI states are mandatory**: Loading (skeleton/spinner), Empty (render `-`), Error (toast/alert). No fetch without feedback.
5. **Router**: entry in `src/router/index.ts`, consistent with other routes (with auth guard following the existing pattern when needed).
6. **Navigation**: add the menu/sidebar item in the layout (`src/layouts/`), following the pattern of other entries.
7. **Backend endpoint missing?** Load the `add-endpoint` skill and do it first — never invent an API contract. Take the contract from the Swagger docs or the actual handler.
8. **Verification**:
   ```bash
   cd frontend-vue && pnpm lint && pnpm typecheck
   ```
   New errors → fix them; adding entries to `eslint-suppressions.json` is forbidden.
9. Report: files created/changed (store, view, router, layout) + verification confirmation.

## Hard Rules

- `any` is forbidden (active ESLint error) and double property fallbacks `||` likewise.
- Single-word UI components (Button, Card) only exist in `src/components/ui/` — shadcn-vue convention; the `vue/multi-word-component-names` rule is intentionally off.
- New frontend config vars go into `frontend-vue/.env.example` only.
