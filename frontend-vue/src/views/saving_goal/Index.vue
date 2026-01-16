<script setup lang="ts">
import { ref, onMounted, computed} from "vue";
import { useSavingGoalStore } from "@/stores/saving_goal";
import { useWalletStore } from "@/stores/wallet";
import { useCategoryStore } from "@/stores/category";
import { format } from "date-fns";
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";
import { Progress } from "@/components/ui/progress";
import { Button } from "@/components/ui/button";
import { Plus, PiggyBank, Target, Calendar } from "lucide-vue-next";
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogDescription, DialogFooter } from "@/components/ui/dialog";
import { Label } from "@/components/ui/label";
import { Input } from "@/components/ui/input";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import ManualTransactionDialog from "@/components/ManualTransactionDialog.vue";
import { getEmoji, getIconComponent } from "@/lib/icons";

const store = useSavingGoalStore();
const walletStore = useWalletStore(); // Ensure wallets are loaded for contribution
const categoryStore = useCategoryStore();
const isCreateOpen = ref(false);
const isContributeOpen = ref(false);
const selectedGoalForContribution = ref<any>(null);

// Form for New Goal
const newGoalName = ref("");
const newGoalTarget = ref("");
const newGoalDeadline = ref("");
const newGoalCategory = ref("");

// Handling Create
const handleCreate = async () => {
    if (!newGoalName.value || !newGoalTarget.value || !newGoalCategory.value) return;

    const payload = {
        name: newGoalName.value,
        target_amount: Number(newGoalTarget.value),
        category_id: newGoalCategory.value ? Number(newGoalCategory.value) : undefined,
        deadline: newGoalDeadline.value ? new Date(newGoalDeadline.value).toISOString() : null,
        icon: "PiggyBank"
    };

    const success = await store.createGoal(payload);
    if (success) {
        isCreateOpen.value = false;
        newGoalName.value = "";
        newGoalTarget.value = "";
        newGoalDeadline.value = "";
        newGoalCategory.value = "";
    }
};

// Open Contribute Dialog via ManualTransactionDialog
const openContribute = (goal: any) => {
    selectedGoalForContribution.value = goal;
    isContributeOpen.value = true;
};

// Close Handlers
const handleContributeClose = () => {
    isContributeOpen.value = false;
    selectedGoalForContribution.value = null;
    // Refresh to get updated amounts
    store.fetchGoals();
    walletStore.fetchWallets(); // Refresh wallet balances (available balance update)
};

onMounted(() => {
    store.fetchGoals();
    walletStore.fetchWallets();
    categoryStore.fetchCategories();
});

const formatCurrency = (value: number) => {
    return new Intl.NumberFormat("id-ID", { style: "currency", currency: "IDR", maximumFractionDigits: 0 }).format(value);
};

const getProgress = (current: number, target: number) => {
    if (target === 0) return 0;
    return Math.min((current / target) * 100, 100);
};

// Nominal Input Formatting
const formattedTarget = computed({
    get: () => {
        if (!newGoalTarget.value) return "";
        return new Intl.NumberFormat("id-ID", { style: "currency", currency: "IDR", maximumFractionDigits: 0 }).format(Number(newGoalTarget.value));
    },
    set: (val: string) => {
        const numericValue = Number(val.replace(/[^0-9]/g, ""));
        newGoalTarget.value = numericValue.toString();
    }
});
</script>

<template>
  <div class="flex-1 space-y-6 pt-2" v-if="store.isLoading">
      <div class="flex items-center justify-center min-h-[400px]">
          <p class="text-muted-foreground animate-pulse">Memuat data target menabung...</p>
      </div>
  </div>
  <div class="flex-1 space-y-6 pt-2 text-foreground" v-else>
        <div class="flex justify-between items-center">
            <div>
                <h1 class="text-3xl font-bold tracking-tight">Target Menabung</h1>
                <p class="text-sm text-muted-foreground mt-1">Wujudkan impianmu dengan menabung secara konsisten.</p>
            </div>
            <Button @click="isCreateOpen = true" class="bg-gradient-to-r from-emerald-600 to-teal-500 text-white hover:from-emerald-500 hover:to-teal-400 shadow-md h-10 rounded-xl transition-all hover:scale-105 active:scale-95 px-4">
                <Plus class="w-4 h-4 mr-2" /> Tambah Target
            </Button>
        </div>

        <!-- Goals Grid -->
        <div class="grid gap-6 md:grid-cols-2 lg:grid-cols-3">
            <Card v-for="goal in store.goals" :key="goal.id" class="relative overflow-hidden border-border shadow-sm hover:shadow-md transition-all rounded-3xl bg-card border group">
                <CardHeader class="pb-2 border-b border-border/50">
                    <div class="flex justify-between items-start">
                        <div class="space-y-1">
                            <CardTitle class="flex items-center gap-2 text-base font-bold">
                                <PiggyBank v-if="!goal.is_achieved" class="w-5 h-5 text-emerald-500" />
                                <Target v-else class="w-5 h-5 text-emerald-600" />
                                {{ goal.name }}
                            </CardTitle>
                            <CardDescription v-if="goal.deadline">
                                <span class="flex items-center gap-1 text-[10px] font-medium uppercase tracking-widest text-muted-foreground">
                                    <Calendar class="w-3 h-3" /> Target: {{ format(new Date(goal.deadline), 'dd MMM yyyy') }}
                                </span>
                            </CardDescription>
                        </div>
                        <div v-if="goal.is_achieved" class="px-2 py-0.5 bg-emerald-100 text-emerald-700 text-[10px] font-bold rounded-full uppercase tracking-wider">
                            Tercapai
                        </div>
                    </div>
                </CardHeader>
                <CardContent class="pt-4">
                    <div class="space-y-4">
                        <div class="flex justify-between items-end">
                            <span class="text-xs text-muted-foreground font-medium uppercase tracking-widest">Terkumpul</span>
                            <span class="font-bold text-lg text-foreground">{{ formatCurrency(goal.current_amount) }}</span>
                        </div>
                        
                        <div class="space-y-1.5">
                            <div class="flex justify-between text-[10px] font-bold uppercase tracking-widest text-muted-foreground">
                                <span>Progress {{ Math.round(getProgress(goal.current_amount, goal.target_amount)) }}%</span>
                                <span>Dari {{ formatCurrency(goal.target_amount) }}</span>
                            </div>
                            <Progress :model-value="getProgress(goal.current_amount, goal.target_amount)" class="h-1.5 bg-muted [&>div]:bg-emerald-500" />
                        </div>

                        <Button v-if="!goal.is_achieved" @click="openContribute(goal)" class="w-full rounded-xl bg-gradient-to-r from-emerald-600 to-teal-500 hover:from-emerald-500 hover:to-teal-400 text-white mt-2 shadow-sm transition-all active:scale-95 h-9 text-xs font-bold" size="sm">
                            Tabung Sekarang
                        </Button>
                        <Button v-else disabled class="w-full rounded-xl bg-emerald-50 text-emerald-700 border border-emerald-100 mt-2 h-9 text-xs font-bold" size="sm">
                            Selesai 🎉
                        </Button>
                    </div>
                </CardContent>
            </Card>

            <!-- Empty State -->
            <div v-if="store.goals.length === 0 && !store.isLoading" class="col-span-full text-center py-12">
                <div class="inline-flex items-center justify-center w-16 h-16 rounded-full bg-emerald-50 mb-4">
                    <PiggyBank class="w-8 h-8 text-emerald-600" />
                </div>
                <h3 class="text-lg font-medium">Belum ada target menabung</h3>
                <p class="text-muted-foreground mt-2 max-w-sm mx-auto">Mulai buat target impianmu sekarang, cicil sedikit demi sedikit lama-lama menjadi bukit!</p>
                <Button @click="isCreateOpen = true" class="mt-4" variant="outline">Buat Target Pertama</Button>
            </div>
        </div>

        <!-- Create Dialog -->
        <Dialog :open="isCreateOpen" @update:open="isCreateOpen = $event">
            <DialogContent class="sm:max-w-[425px] rounded-3xl bg-card text-foreground">
                <DialogHeader>
                    <DialogTitle>Buat Target Baru</DialogTitle>
                    <DialogDescription>Tentukan barang impian atau tujuan finansialmu.</DialogDescription>
                </DialogHeader>
                <div class="space-y-4 py-4">
                    <div class="space-y-2">
                        <Label class="text-xs font-bold uppercase tracking-widest text-muted-foreground">Nama Target</Label>
                        <Input v-model="newGoalName" placeholder="Misal: Beli Laptop Baru" class="h-11 shadow-sm rounded-xl bg-background" />
                    </div>
                    <div class="space-y-2">
                        <Label class="text-xs font-bold uppercase tracking-widest text-muted-foreground">Target Nominal (Rp)</Label>
                        <Input type="text" inputmode="numeric" pattern="[0-9]*" v-model="formattedTarget" placeholder="Rp 0" class="h-11 shadow-sm rounded-xl bg-background" />
                    </div>
                    <div class="space-y-2">
                        <Label class="text-xs font-bold uppercase tracking-widest text-muted-foreground">Kategori</Label>
                        <Select v-model="newGoalCategory">
                            <SelectTrigger class="w-full h-11 rounded-xl bg-background shadow-sm">
                                <SelectValue placeholder="Pilih Kategori" />
                            </SelectTrigger>
                             <SelectContent>
                                <SelectItem v-for="c in categoryStore.categories.filter(c => c.type === 'expense')" :key="c.id" :value="String(c.id)">
                                    <div class="flex items-center gap-2">
                                        <component v-if="getIconComponent(c.icon)" :is="getIconComponent(c.icon)" class="h-4 w-4" />
                                        <span v-else>{{ getEmoji(c.icon) || c.icon }}</span>
                                        <span>{{ c.name }}</span>
                                    </div>
                                </SelectItem>
                            </SelectContent>
                        </Select>
                    </div>
                    <div class="space-y-2">
                        <Label class="text-xs font-bold uppercase tracking-widest text-muted-foreground">Batas Waktu (Opsional)</Label>
                        <Input type="date" v-model="newGoalDeadline" class="h-11 rounded-xl bg-background shadow-sm" />
                    </div>
                </div>
                <DialogFooter class="gap-2">
                    <Button variant="outline" @click="isCreateOpen = false" class="rounded-xl h-10 px-6">Batal</Button>
                    <Button @click="handleCreate" class="bg-gradient-to-r from-emerald-600 to-teal-500 text-white hover:from-emerald-500 hover:to-teal-400 shadow-md rounded-xl h-10 px-6 font-bold">Buat Target</Button>
                </DialogFooter>
            </DialogContent>
        </Dialog>

        <!-- Use ManualTransactionDialog for Contribution -->
        <!-- We pass a special prop or manage state to pre-fill it -->
        <ManualTransactionDialog 
            :open="isContributeOpen" 
            @update:open="isContributeOpen = $event"
            @save="handleContributeClose"
            :savingGoalTarget="selectedGoalForContribution"
        />
    </div>
</template>
