<script setup lang="ts">
import { SidebarProvider, SidebarInset, SidebarTrigger } from "@/components/ui/sidebar";
import AppSidebar from "@/components/AppSidebar.vue";
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
import ModeToggle from "@/components/ModeToggle.vue";
import { Separator } from "@/components/ui/separator";
import MobileBottomNav from "@/components/MobileBottomNav.vue";
import FloatingActionMenu from "@/components/FloatingActionMenu.vue";
</script>

<template>
  <SidebarProvider>
    <AppSidebar />

    <SidebarInset>
      <header class="flex h-14 shrink-0 items-center gap-2 border-b bg-card px-4 sticky top-0 z-50 transition-[width,height] ease-linear group-has-[[data-collapsible=icon]]/sidebar-wrapper:h-12">
        <div class="flex items-center gap-2 px-4">
          <SidebarTrigger class="-ml-1" />
          <Separator orientation="vertical" class="h-4 mr-2" />
        </div>

        <div class="flex-1"></div>

        <div class="flex items-center gap-4">
          <ModeToggle />

          <div class="w-px h-6 bg-border"></div>

          <div class="flex items-center gap-2">
            <span class="text-sm font-medium text-foreground">Heyho, Bro</span>
            <Avatar class="w-8 h-8">
              <AvatarImage src="https://github.com/shadcn.png" />
              <AvatarFallback>CN</AvatarFallback>
            </Avatar>
          </div>
        </div>
      </header>

      <div class="flex-1 p-4 pt-0 space-y-4 md:p-6 pb-24 md:pb-6">
        <RouterView />
      </div>
    </SidebarInset>

    <MobileBottomNav />
    <FloatingActionMenu />
  </SidebarProvider>
</template>
