// ESLint flat config — quality gate cepat untuk AI agent & developer.
// Lint dijalankan lewat `pnpm lint`, validasi tipe terpisah lewat `pnpm typecheck`.
import { globalIgnores } from 'eslint/config'
import pluginVue from 'eslint-plugin-vue'
import { defineConfigWithVueTs, vueTsConfigs } from '@vue/eslint-config-typescript'

export default defineConfigWithVueTs(
  {
    name: 'app/files-to-lint',
    files: ['**/*.{ts,js,vue}'],
  },
  globalIgnores(['**/dist/**', '**/node_modules/**', '**/coverage/**']),
  pluginVue.configs['flat/essential'],
  vueTsConfigs.recommended,
  {
    rules: {
      // Bertabrakan dengan konvensi shadcn-vue (Avatar.vue, Card.vue, dst.)
      // dan pola routing views (Index.vue, Detail.vue).
      'vue/multi-word-component-names': 'off',
    },
  },
)
