<script setup lang="ts">
import { SidebarProvider } from "@/components/ui/sidebar";
import AppSidebar from "@/components/AppSidebar.vue";
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
import ModeToggle from "@/components/ModeToggle.vue";
</script>

<template>
  <SidebarProvider>
    <AppSidebar />

    <main class="w-full bg-background min-h-screen">
      <header class="flex h-14 items-center justify-end border-b bg-card px-4">
        <div class="flex items-center gap-4">
          <ModeToggle />

          <div class="h-6 w-px bg-border"></div>

          <div class="flex items-center gap-2">
            <span class="text-sm font-medium text-foreground">Heyho, Bro</span>
            <Avatar class="h-8 w-8">
              <AvatarImage src="https://github.com/shadcn.png" />
              <AvatarFallback>CN</AvatarFallback>
            </Avatar>
          </div>
        </div>
      </header>

      <div class="p-4 md:p-6">
        <RouterView />
      </div>
    </main>
  </SidebarProvider>
</template>
