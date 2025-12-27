<script setup lang="ts">
import { RouterLink } from "vue-router";
import { ToolCase, Wallet, ArrowLeftRight, FileChartColumn, Settings } from "lucide-vue-next";

const items = [
  { title: "Dashboard", url: "/dashboard", icon: ToolCase },
  { title: "Wallet", url: "/wallet", icon: Wallet },
  { title: "Transaksi", url: "/transaction", icon: ArrowLeftRight },
  { title: "Laporan", url: "/report", icon: FileChartColumn },
  { title: "Setting", url: "/setting", icon: Settings },
];
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
