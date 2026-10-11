<script setup lang="ts">
import { Button } from "@/components/ui/button";
import { Moon, Sun } from "lucide-vue-next";
import { isDark, toggleTheme } from "@/composables/useTheme";
</script>

<template>
  <Button
    variant="ghost"
    size="icon"
    @click="toggleTheme()"
    class="rounded-full w-9 h-9 text-foreground hover:bg-accent transition-colors"
    aria-label="Toggle theme"
  >
    <Sun v-if="isDark" class="h-5 w-5 text-yellow-500 transition-all duration-300" />
    <Moon v-else class="h-5 w-5 text-slate-700 dark:text-slate-200 transition-all duration-300" />
  </Button>
</template>
