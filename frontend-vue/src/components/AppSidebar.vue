<script setup lang="ts">
import { Sidebar, SidebarContent, SidebarGroup, SidebarGroupContent, SidebarGroupLabel, SidebarMenu, SidebarMenuButton, SidebarMenuItem, SidebarHeader } from "@/components/ui/sidebar";
import { Wallet, Bot, FolderOpen, ArrowLeftRight, ToolCase, FileChartColumn, Settings2, Calendar1, HandCoins, Heart, PiggyBank, HeartPulse } from "lucide-vue-next";

const items = [
  {
    title: "Dashboard",
    url: "/dashboard",
    icon: ToolCase,
  },
  {
    title: "Kesehatan Keuangan",
    url: "/financial-health",
    icon: HeartPulse,
  },
  {
    title: "Asisten AI",
    url: "/chat",
    icon: Bot,
  },
  {
    title: "Dompet Saya",
    url: "/wallet",
    icon: Wallet,
  },
  {
    title: "Kategori",
    url: "/category",
    icon: FolderOpen,
  },
  {
    title: "Transaksi",
    url: "/transaction",
    icon: ArrowLeftRight,
  },
  {
    title: "Kalender",
    url: "/calendar",
    icon: Calendar1,
  },
  {
    title: "Laporan",
    url: "/report",
    icon: FileChartColumn,
  },
  {
    title: "Utang & Piutang",
    url: "/debt",
    icon: HandCoins,
  },
  {
    title: "Keinginan",
    url: "/wishlist",
    icon: Heart,
  },
  {
    title: "Target Menabung",
    url: "/saving-goal",
    icon: PiggyBank,
  },
];

const settingsItems = [
  {
    title: "Pengaturan",
    url: "/setting",
    icon: Settings2,
  },
];
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
        <SidebarGroupLabel>Menu Utama</SidebarGroupLabel>
        <SidebarGroupContent>
          <SidebarMenu>
            <SidebarMenuItem v-for="item in items" :key="item.title">
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
        <SidebarGroupLabel>Lainnya</SidebarGroupLabel>
        <SidebarGroupContent>
          <SidebarMenu>
            <SidebarMenuItem v-for="item in settingsItems" :key="item.title">
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
