<script setup lang="ts">
import { computed } from "vue";
import { useI18n } from "vue-i18n";
import { RouterLink } from "vue-router";
import { ToolCase, Wallet, ArrowLeftRight, FileChartColumn, Settings2 } from "lucide-vue-next";

const { t } = useI18n();

const items = computed(() => [
  { title: t("nav.dashboard"), url: "/dashboard", icon: ToolCase },
  { title: t("nav.wallets"), url: "/wallet", icon: Wallet },
  { title: t("nav.transactions"), url: "/transaction", icon: ArrowLeftRight },
  { title: t("nav.reports"), url: "/report", icon: FileChartColumn },
  { title: t("nav.settings"), url: "/setting", icon: Settings2 },
]);
</script>

<template>
  <div class="fixed bottom-0 left-0 right-0 z-40 bg-card border-t border-border md:hidden block pb-safe">
    <div class="flex justify-around items-center h-16">
      <RouterLink
        v-for="item in items"
        :key="item.url"
        :to="item.url"
        active-class="text-primary"
        class="flex flex-col items-center justify-center w-full h-full text-muted-foreground hover:text-foreground transition-colors gap-1"
      >
        <component :is="item.icon" class="w-5 h-5" />
        <span class="text-[10px] font-medium">{{ item.title }}</span>
      </RouterLink>
    </div>
  </div>
</template>

<style scoped>
.pb-safe {
  padding-bottom: env(safe-area-inset-bottom);
}
</style>
