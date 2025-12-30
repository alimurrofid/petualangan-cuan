<script setup lang="ts">
import { ref } from "vue";
import { useRouter } from "vue-router";
import { Plus, Bot, PenLine } from "lucide-vue-next";
import { DropdownMenu, DropdownMenuContent, DropdownMenuItem, DropdownMenuTrigger } from "@/components/ui/dropdown-menu";
import { Button } from "@/components/ui/button";
import ManualTransactionDialog from "@/components/ManualTransactionDialog.vue";

const router = useRouter();
const isDialogOpen = ref(false);

const handleCommandAi = () => {
    router.push("/chat");
};

const handleManualTransaction = () => {
    isDialogOpen.value = true;
};
</script>

<template>
  <div class="fixed bottom-20 right-4 z-50 md:bottom-10 md:right-10">
    <DropdownMenu>
        <DropdownMenuTrigger asChild>
            <Button size="icon" class="h-14 w-14 rounded-full shadow-2xl bg-gradient-to-r from-emerald-400 to-teal-500 hover:from-emerald-500 hover:to-teal-400 text-white transition-all duration-300 hover:scale-110 hover:shadow-emerald-500/50 border-2 border-white/20">
                <Plus class="h-8 w-8" />
            </Button>
        </DropdownMenuTrigger>
        <DropdownMenuContent align="end" class="w-56 mb-2">
            <DropdownMenuItem @click="handleManualTransaction" class="gap-2 p-3 cursor-pointer">
                <PenLine class="w-4 h-4" />
                <span>Manual transaction input</span>
            </DropdownMenuItem>
            <DropdownMenuItem @click="handleCommandAi" class="gap-2 p-3 cursor-pointer">
                <Bot class="w-4 h-4" />
                <span>Command Ai</span>
            </DropdownMenuItem>
        </DropdownMenuContent>
    </DropdownMenu>

    <ManualTransactionDialog v-model:open="isDialogOpen" @save="(data) => console.log(data)" />
  </div>
</template>
