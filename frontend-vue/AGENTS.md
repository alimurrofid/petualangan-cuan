# 🌐 AGENTS.md — Frontend Web (Vue 3)

Area-specific guide for `frontend-vue/`. The global rules in the root [AGENTS.md](../AGENTS.md) still apply.

## ⚡ 1. Commands & Verification

```bash
pnpm install            # must use pnpm, not npm/yarn

pnpm run dev            # Vite dev server (HMR) → http://localhost:5173
pnpm run typecheck      # vue-tsc -b — type checking only (fast)
pnpm run lint           # ESLint --fix + report
pnpm run build          # vue-tsc -b && vite build (full production)
pnpm run preview        # preview the production build
```

> [!WARNING]
> ### Verification Loop (MUST pass before finishing a task)
> `pnpm lint && pnpm typecheck` — not `pnpm build`, unless production bundling behavior needs testing.
>
> **ESLint Baseline**: legacy errors are recorded in `eslint-suppressions.json` (created via `eslint . --suppress-all`). **New** errors still make lint fail → **must be fixed, not suppressed**. Once legacy is cleaned up, trim the baseline: `pnpm exec eslint . --prune-suppressions`.

## 🏗️ 2. Architecture & Conventions

1. **Components**: Composition API with `<script setup lang="ts">` — mandatory for all new components.
2. **State (Pinia)**: stores are isolated per domain in `src/stores/`. The real current stores:
   `auth.ts`, `category.ts`, `dashboard.ts`, `debt.ts`, `financialHealth.ts`, `saving_goal.ts`, `settings.ts`, `transaction.ts`, `wallet.ts`, `wishlist.ts`.
   All API calls go through stores; components never call axios directly.
3. **Data types**: interfaces are defined **inline in each domain's store** (e.g. `Wallet`, `CreateWalletInput` in `stores/wallet.ts`) — **there is no `src/types/` folder**. The `any` rule: **forbidden** — enforced by ESLint (`@typescript-eslint/no-explicit-any` = error).
4. **NO double property fallbacks `||`**:
   - ❌ `const name = item.title || item.name || item.description || ''`
   - ❌ `const userId = data.user_id || data.userId || data.ownerId`
   These hide wrong column names/API contract mistakes/schema mismatches. Empty data → `null` (state) / render `-` (UI).
5. **API client**: a single axios instance in `src/lib/api.ts`:
   - `baseURL` from `VITE_API_BASE_URL`, `withCredentials: true`.
   - Automatically attaches `Authorization: Bearer <token>` from localStorage.
   - 401 interceptor → auto-refresh the token via `authStore.refreshAccessToken()` (login/refresh routes are excluded; on refresh failure → redirect to `/login`).
   **Never create a second axios instance.**
6. **Styling**: Tailwind CSS v4 utility classes. shadcn-vue atomic components live in `src/components/ui/` (Button, Dialog, Card, etc.) — **use existing ones; never reimplement from scratch**. Single-word component filenames are allowed in `ui/` (the `vue/multi-word-component-names` rule is intentionally off).
7. **UI Handling States**: every data fetch must have clear visual feedback: Loading (skeleton/spinner), Empty state (render `-`), Error (toast/alert). Confirmations use `sweetalert2` (helper: `src/composables/useSwal.ts`).
8. **Views**: `src/views/<domain>/Index.vue`, `Detail.vue`, etc. The existing domains — reuse the closest one instead of inventing a new layout:
   `auth/`, `calendar/`, `category/`, `chat/`, `debt/`, `financial-health/`, `report/`, `saving_goal/`, `setting/`, `transaction/`, `wallet/`, `wishlist/`, plus a bare `views/Dashboard.vue`.
   Folder naming is mixed (snake_case and kebab-case) and `Dashboard.vue` sits outside a folder — follow the domain you touch; never rename existing files.
9. **Icons/others**: `src/lib/icons.ts`, `src/lib/utils.ts`, `src/composables/` — check these before creating new utilities.
10. **i18n**: `vue-i18n` is available; never hardcode inline duplicated strings.

## 🔒 3. Configuration

Env template: [frontend-vue/.env.example](.env.example) (`VITE_API_BASE_URL`). Never read the real `.env`.
