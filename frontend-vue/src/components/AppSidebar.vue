<script setup lang="ts">
import { computed } from "vue";
import { useI18n } from "vue-i18n";
import { Sidebar, SidebarContent, SidebarGroup, SidebarGroupContent, SidebarGroupLabel, SidebarMenu, SidebarMenuButton, SidebarMenuItem, SidebarHeader } from "@/components/ui/sidebar";
import { Wallet, Bot, FolderOpen, ArrowLeftRight, ToolCase, FileChartColumn, Settings2, Calendar1, HandCoins, Heart, PiggyBank, HeartPulse } from "lucide-vue-next";

const { t } = useI18n();

const items = computed(() => [
  {
    title: t("nav.dashboard"),
    url: "/dashboard",
    icon: ToolCase,
  },
  {
    title: t("nav.financialHealth"),
    url: "/financial-health",
    icon: HeartPulse,
  },
  {
    title: t("nav.chat"),
    url: "/chat",
    icon: Bot,
  },
  {
    title: t("nav.wallets"),
    url: "/wallet",
    icon: Wallet,
  },
  {
    title: t("nav.categories"),
    url: "/category",
    icon: FolderOpen,
  },
  {
    title: t("nav.transactions"),
    url: "/transaction",
    icon: ArrowLeftRight,
  },
  {
    title: t("nav.calendar"),
    url: "/calendar",
    icon: Calendar1,
  },
  {
    title: t("nav.reports"),
    url: "/report",
    icon: FileChartColumn,
  },
  {
    title: t("nav.debts"),
    url: "/debt",
    icon: HandCoins,
  },
  {
    title: t("nav.wishlist"),
    url: "/wishlist",
    icon: Heart,
  },
  {
    title: t("nav.savingGoals"),
    url: "/saving-goal",
    icon: PiggyBank,
  },
]);

const settingsItems = computed(() => [
  {
    title: t("nav.settings"),
    url: "/setting",
    icon: Settings2,
  },
]);
</script>

<template>
  <Sidebar>
    <SidebarHeader>
      <div class="flex items-center gap-3 px-4 py-4">
        <img src="/petualangancuan.svg" alt="Logo Petualangan Cuan" class="object-contain w-10 h-10 rounded-lg shadow-md shrink-0" />
        <div class="flex flex-col">
          <span class="text-base font-bold">Petualangan Cuan</span>
          <span class="text-xs text-muted-foreground">Kelola Keuangan Cerdas</span>
        </div>
      </div>
    </SidebarHeader>

    <SidebarContent>
      <SidebarGroup>
        <SidebarGroupLabel>{{ t("nav.mainMenu") }}</SidebarGroupLabel>
        <SidebarGroupContent>
          <SidebarMenu>
            <SidebarMenuItem v-for="item in items" :key="item.url">
              <SidebarMenuButton asChild>
                <RouterLink :to="item.url" active-class="bg-accent text-accent-foreground">
                  <component :is="item.icon" class="w-4 h-4" />
                  <span>{{ item.title }}</span>
                </RouterLink>
              </SidebarMenuButton>
            </SidebarMenuItem>
          </SidebarMenu>
        </SidebarGroupContent>
      </SidebarGroup>

      <SidebarGroup>
        <SidebarGroupLabel>{{ t("nav.preferences") }}</SidebarGroupLabel>
        <SidebarGroupContent>
          <SidebarMenu>
            <SidebarMenuItem v-for="item in settingsItems" :key="item.url">
              <SidebarMenuButton asChild>
                <RouterLink :to="item.url" active-class="bg-accent text-accent-foreground">
                  <component :is="item.icon" class="w-4 h-4" />
                  <span>{{ item.title }}</span>
                </RouterLink>
              </SidebarMenuButton>
            </SidebarMenuItem>
          </SidebarMenu>
        </SidebarGroupContent>
      </SidebarGroup>
    </SidebarContent>
  </Sidebar>
</template>
