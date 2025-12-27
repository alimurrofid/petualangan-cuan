<script setup lang="ts">
import { SidebarProvider, SidebarInset, SidebarTrigger } from "@/components/ui/sidebar";
import AppSidebar from "@/components/AppSidebar.vue";
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
import ModeToggle from "@/components/ModeToggle.vue";
import { Separator } from "@/components/ui/separator";
import MobileBottomNav from "@/components/MobileBottomNav.vue";
import FloatingActionMenu from "@/components/FloatingActionMenu.vue";
import { 
  DropdownMenu, 
  DropdownMenuContent, 
  DropdownMenuItem, 
  DropdownMenuTrigger,
  DropdownMenuLabel,
  DropdownMenuSeparator
} from "@/components/ui/dropdown-menu";
import { User, CreditCard, LifeBuoy, Lock, LogOut } from "lucide-vue-next";
import { useRouter } from "vue-router";

const router = useRouter();
const handleLogout = () => {
    router.push('/login');
};
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

          <DropdownMenu>
            <DropdownMenuTrigger asChild>
                <div class="flex items-center gap-2 cursor-pointer hover:opacity-80 transition-opacity">
                    <span class="text-sm font-medium text-foreground">Heyho, Bro</span>
                    <Avatar class="w-8 h-8 border border-border">
                        <AvatarImage src="https://github.com/shadcn.png" />
                        <AvatarFallback>CN</AvatarFallback>
                    </Avatar>
                </div>
            </DropdownMenuTrigger>
            <DropdownMenuContent align="end" class="w-56">
                <DropdownMenuLabel>My Account</DropdownMenuLabel>
                <DropdownMenuSeparator />
                <DropdownMenuItem @click="router.push('/setting?tab=profile')">
                    <User class="mr-2 h-4 w-4" />
                    <span>Profile</span>
                </DropdownMenuItem>
                <DropdownMenuItem>
                    <CreditCard class="mr-2 h-4 w-4" />
                    <span>Payment History</span>
                </DropdownMenuItem>
                <DropdownMenuItem>
                    <LifeBuoy class="mr-2 h-4 w-4" />
                    <span>Help Center</span>
                </DropdownMenuItem>
                <DropdownMenuItem @click="router.push('/setting?tab=password')">
                    <Lock class="mr-2 h-4 w-4" />
                    <span>Change Password</span>
                </DropdownMenuItem>
                <DropdownMenuSeparator />
                <DropdownMenuItem @click="handleLogout" class="text-red-500 focus:text-red-500">
                    <LogOut class="mr-2 h-4 w-4" />
                    <span>Logout</span>
                </DropdownMenuItem>
            </DropdownMenuContent>
          </DropdownMenu>
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
