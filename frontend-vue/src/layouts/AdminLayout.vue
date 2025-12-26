<script setup lang="ts">
import { SidebarProvider } from "@/components/ui/sidebar"
import AppSidebar from "@/components/AppSidebar.vue"
import { Avatar, AvatarFallback, AvatarImage } from '@/components/ui/avatar'
</script>

<template>
  <SidebarProvider>
    <AppSidebar />
    
    <main class="w-full bg-muted/20 min-h-screen">
      <header class="flex h-14 items-center justify-end border-b bg-background px-4">
        <div class="flex items-center gap-2">
          <span class="text-sm font-medium">Heyho, Bro</span>
          <Avatar class="h-8 w-8">
            <AvatarImage src="https://github.com/shadcn.png" />
            <AvatarFallback>CN</AvatarFallback>
          </Avatar>
        </div>
      </header>

      <div class="p-4 md:p-6">
        <RouterView />
      </div>
    </main>
  </SidebarProvider>
</template>