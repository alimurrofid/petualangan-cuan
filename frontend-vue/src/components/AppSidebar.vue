<script setup lang="ts">
import { Sidebar, SidebarContent, SidebarGroup, SidebarGroupContent, SidebarGroupLabel, SidebarMenu, SidebarMenuButton, SidebarMenuItem, SidebarHeader } from "@/components/ui/sidebar";
import { Wallet, Settings, Bot, FolderOpen, ArrowLeftRight, TrendingUp, ToolCase, CalendarCheck, FileChartColumn } from "lucide-vue-next";

const items = [
  {
    title: "Dashboard",
    url: "/dashboard",
    icon: ToolCase,
  },
  {
    title: "AI Assistant",
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
    icon: CalendarCheck,
  },
  {
    title: "Laporan",
    url: "/report",
    icon: FileChartColumn,
  },
];

const settingsItems = [
  {
    title: "Pengaturan",
    url: "/setting",
    icon: Settings,
  },
];
</script>

<template>
  <Sidebar>
    <SidebarHeader>
      <div class="flex items-center gap-3 px-4 py-4">
        <div class="flex items-center justify-center w-10 h-10 font-bold text-white rounded-lg shadow-lg bg-linear-to-br from-emerald-500 to-teal-600">
          <TrendingUp class="w-5 h-5" />
        </div>
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
