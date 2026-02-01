<script setup lang="ts">
import { ref, onMounted, watch, computed } from "vue";
import { useSavingGoalStore } from "@/stores/saving_goal";
import { useWalletStore } from "@/stores/wallet";
import { useCategoryStore } from "@/stores/category";
import { useAuthStore } from "@/stores/auth";
import { format } from "date-fns";
import { Button } from "@/components/ui/button";
import { Plus, PiggyBank, Target, Calendar, Pencil, Trash2, Eye, CheckCircle } from "lucide-vue-next";
import { Card, CardHeader, CardTitle, CardContent } from "@/components/ui/card";
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogDescription, DialogFooter } from "@/components/ui/dialog";
import { Label } from "@/components/ui/label";
import { Input } from "@/components/ui/input";
import SearchableSelect from "@/components/ui/searchable-select/SearchableSelect.vue";
import ManualTransactionDialog from "@/components/ManualTransactionDialog.vue";
import Detail from "./Detail.vue";
import { getEmoji, getIconComponent } from "@/lib/icons";
import { formatCurrency, parseCurrencyInput, formatCurrencyInput, formatCurrencyLive } from "@/lib/utils";
import { useSwal } from "@/composables/useSwal";

const store = useSavingGoalStore();
const walletStore = useWalletStore(); // Ensure wallets are loaded for contribution
const categoryStore = useCategoryStore();
const authStore = useAuthStore();
const swal = useSwal();
const isCreateOpen = ref(false);
const isContributeOpen = ref(false);
const isDetailOpen = ref(false);
const isEditing = ref(false);
const editingId = ref<number | null>(null);
const selectedGoalForContribution = ref<any>(null);
const selectedGoal = ref<any>(null);
const isSubmitting = ref(false);

// Form for New Goal
const newGoalName = ref("");
const newGoalTarget = ref("");
const newGoalTargetDisplay = ref("");
const newGoalDeadline = ref("");
const newGoalCategory = ref("");

// Handling Create
const handleCreate = async () => {
    if (!newGoalName.value || !newGoalTarget.value || !newGoalCategory.value) return;

    isSubmitting.value = true;
    try {
        const payload = {
            name: newGoalName.value,
            target_amount: Number(newGoalTarget.value),
            category_id: newGoalCategory.value ? Number(newGoalCategory.value) : undefined,
            deadline: newGoalDeadline.value ? new Date(newGoalDeadline.value).toISOString() : null,
            icon: "PiggyBank"
        };

        let success = false;
        if (isEditing.value && editingId.value) {
            success = await store.updateGoal(editingId.value, payload);
            if (success) swal.success("Berhasil", "Target menabung berhasil diperbarui");
        } else {
            success = await store.createGoal(payload);
            if (success) swal.success("Berhasil", "Target menabung berhasil dibuat");
        }

        if (success) {
            isCreateOpen.value = false;
            newGoalName.value = "";
            newGoalTarget.value = "";
            newGoalTargetDisplay.value = "";
            newGoalDeadline.value = "";
            newGoalCategory.value = "";
        }
    } finally {
        isSubmitting.value = false;
    }
};

// Open Contribute Dialog via ManualTransactionDialog
const openContribute = (goal: any) => {
    selectedGoalForContribution.value = goal;
    isContributeOpen.value = true;
};

const openCreateDialog = () => {
    isEditing.value = false;
    editingId.value = null;
    newGoalName.value = "";
    newGoalTarget.value = "";
    newGoalTargetDisplay.value = "";
    newGoalDeadline.value = "";
    newGoalCategory.value = "";
    isCreateOpen.value = true;
};

const openEditDialog = (goal: any) => {
    isEditing.value = true;
    editingId.value = goal.id;
    newGoalName.value = goal.name;
    newGoalTarget.value = String(goal.target_amount);
    newGoalTargetDisplay.value = formatCurrencyInput(goal.target_amount);
    newGoalCategory.value = goal.category_id ? String(goal.category_id) : "";
    if (goal.deadline) {
        try {
             const date = new Date(goal.deadline);
             newGoalDeadline.value = date.toISOString().split('T')[0] ?? "";
        } catch(e) {
             newGoalDeadline.value = "";
        }
    } else {
        newGoalDeadline.value = "";
    }
    isCreateOpen.value = true;
};

const openDetailDialog = (goal: any) => {
    selectedGoal.value = goal;
    isDetailOpen.value = true;
};

const handleDelete = async (id: number) => {
    const confirmed = await swal.confirmDelete("Target Menabung");
    if (confirmed) {
        const success = await store.deleteGoal(id);
        if (success) swal.success("Terhapus", "Target menabung berhasil dihapus");
    }
};

const handleFinish = async (goal: any) => {
    const result = await swal.fire({
        title: "Selesaikan Target?",
        text: "Dana yang terkumpul akan kembali tersedia sebagai saldo aktif di dompet Anda. Target akan ditandai sukses!",
        icon: "question",
        showCancelButton: true,
        confirmButtonText: "Ya, Cairkan!",
        cancelButtonText: "Batal",
        confirmButtonColor: "#10b981", // Emerald-500
        cancelButtonColor: "#64748b", // Slate-500
    });

    if (result.isConfirmed) {
        const success = await store.finishGoal(goal.id);
        if (success) {
            swal.success("Selamat! 🎉", "Target tercapai dan dana berhasil dicairkan ke dompet.");
             walletStore.fetchWallets(); // Real-time wallet update
        }
    }
};

// Close Handlers
const handleContributeClose = () => {
    isContributeOpen.value = false;
    selectedGoalForContribution.value = null;
    // Refresh to get updated amounts
    store.fetchGoals();
    walletStore.fetchWallets(); // Refresh wallet balances (available balance update)
};

const isInitialLoading = ref(true);

onMounted(async () => {
    try {
        await Promise.all([
            store.fetchGoals(),
            walletStore.fetchWallets(),
            categoryStore.fetchCategories()
        ]);
    } finally {
        isInitialLoading.value = false;
    }
});

// Local formatCurrency removed, using imported one

const getProgress = (current: number, target: number) => {
    if (target === 0) return 0;
    return Math.min((current / target) * 100, 100);
};

// Nominal Input Formatting


// Sync Display -> Model
watch(newGoalTargetDisplay, (val) => {
    const formatted = formatCurrencyLive(val);
    if (formatted !== val) {
        newGoalTargetDisplay.value = formatted;
        return;
    }
    const num = parseCurrencyInput(val);
    newGoalTarget.value = num.toString();
});

const onTargetBlur = () => {
    const num = parseCurrencyInput(newGoalTargetDisplay.value);
    if(num) newGoalTargetDisplay.value = formatCurrencyInput(num);
};

const activeGoals = computed(() => store.goals.filter(g => !g.is_finished));
const finishedGoals = computed(() => store.goals.filter(g => g.is_finished));

const categoryOptions = computed(() => categoryStore.categories.filter(c => c.type === 'expense').map(c => ({
    value: String(c.id),
    label: c.name,
    icon: c.icon
})));

</script>

<template>
  <div class="flex-1 space-y-6 pt-2" v-if="isInitialLoading">
      <div class="flex items-center justify-center min-h-[400px]">
          <p class="text-muted-foreground animate-pulse">Memuat data target menabung...</p>
      </div>
  </div>
  <div class="flex-1 space-y-6 pt-2 text-foreground" v-else>
        <div class="flex flex-col sm:flex-row justify-between items-start sm:items-center gap-4 sm:gap-0">
            <div>
                <h1 class="text-3xl font-bold tracking-tight">Target Menabung</h1>
                <p class="text-sm text-muted-foreground mt-1">Wujudkan impianmu dengan menabung secara konsisten.</p>
            </div>
            <Button @click="openCreateDialog()" class="bg-gradient-to-r from-emerald-600 to-teal-500 text-white hover:from-emerald-500 hover:to-teal-400 shadow-lg h-12 rounded-full transition-all hover:scale-105 active:scale-95 px-6">
                <Plus class="w-5 h-5 mr-2" /> Tambah Target
            </Button>
        </div>

        <!-- Active Goals -->
        <div class="space-y-4">
             <div class="flex items-center gap-2">
                <h3 class="text-base font-bold flex items-center gap-2">
                    <Target class="h-5 w-5 text-muted-foreground" />
                    Sedang Dikejar
                </h3>
                <span class="text-[10px] font-bold uppercase tracking-widest px-2 py-0.5 rounded-full bg-muted text-muted-foreground border border-border/50">{{ activeGoals.length }} item</span>
            </div>

            <div class="grid gap-6 md:grid-cols-2 lg:grid-cols-3">
                <Card 
                    v-for="goal in activeGoals" 
                    :key="goal.id" 
                    class="group relative overflow-hidden transition-all duration-300 hover:shadow-lg hover:-translate-y-1 hover:border-emerald-200 dark:hover:border-emerald-900"
                >
                    <CardHeader class="pb-2">
                        <div class="flex justify-between items-start">
                            <div class="space-y-1">
                                <CardTitle class="flex items-center gap-2 text-xl font-bold tracking-tight">
                                    <div class="p-2 bg-emerald-100 dark:bg-emerald-500/20 rounded-lg text-emerald-600 dark:text-emerald-400">
                                        <PiggyBank v-if="!goal.is_achieved" class="w-5 h-5" />
                                        <Target v-else class="w-5 h-5" />
                                    </div>
                                    {{ goal.name }}
                                </CardTitle>
                                <div v-if="goal.deadline" class="flex items-center gap-1 text-[10px] font-medium uppercase tracking-widest text-muted-foreground pl-1">
                                    <Calendar class="w-3 h-3" /> Target: {{ format(new Date(goal.deadline), 'dd MMM yyyy') }}
                                </div>
                            </div>
                            
                            <div v-if="goal.is_achieved" class="px-3 py-1 bg-emerald-100 dark:bg-emerald-500/20 text-emerald-600 dark:text-emerald-400 text-[10px] font-bold rounded-full uppercase tracking-wider border border-emerald-200 dark:border-emerald-800">
                                Tercapai
                            </div>
                            <div v-else class="flex gap-1">
                                 <Button variant="ghost" size="icon" class="h-8 w-8 text-muted-foreground hover:bg-slate-100 hover:text-blue-600 dark:hover:bg-slate-800" @click="openEditDialog(goal)">
                                    <Pencil class="h-4 w-4" />
                                </Button>
                                <Button variant="ghost" size="icon" class="h-8 w-8 text-muted-foreground hover:bg-red-50 hover:text-red-600 dark:hover:bg-red-900/20" @click="handleDelete(goal.id)">
                                    <Trash2 class="h-4 w-4" />
                                </Button>
                            </div>
                        </div>
                    </CardHeader>

                    <CardContent class="space-y-4 pt-2">
                        <div class="space-y-2">
                            <div class="flex justify-between items-end">
                                <span class="text-xs text-muted-foreground font-medium uppercase tracking-widest">Terkumpul</span>
                                <span class="font-mono text-2xl font-bold tracking-tight text-foreground" :class="{ 'privacy-blur': authStore.isPrivacyMode }">{{ formatCurrency(goal.current_amount) }}</span>
                            </div>
                            
                            <div class="space-y-1.5">
                                <div class="flex justify-between text-[10px] font-bold uppercase tracking-widest text-muted-foreground">
                                    <span>Progress {{ Math.round(getProgress(goal.current_amount, goal.target_amount)) }}%</span>
                                    <span>Dari <span :class="{ 'privacy-blur': authStore.isPrivacyMode }">{{ formatCurrency(goal.target_amount) }}</span></span>
                                </div>
                                <!-- Standard Progress Bar -->
                                <div class="h-2 w-full bg-secondary rounded-full overflow-hidden">
                                    <div class="h-full bg-emerald-500 rounded-full transition-all duration-1000 ease-out" 
                                         :style="{ width: getProgress(goal.current_amount, goal.target_amount) + '%' }"
                                    ></div>
                                </div>
                            </div>
                        </div>

                        <div class="grid grid-cols-[1fr,auto] gap-2 pt-2">
                            <!-- Button Actions Logic -->

                            <Button v-if="!goal.is_achieved" @click="openContribute(goal)" class="w-full rounded-xl bg-gradient-to-r from-emerald-600 to-teal-500 text-white hover:from-emerald-500 hover:to-teal-400 shadow-sm border-0 font-bold h-10 transition-all active:scale-95 text-xs" size="sm">
                                <Plus class="w-4 h-4 mr-2" /> Tabung
                            </Button>
                            
                            <Button v-else-if="goal.is_achieved && !goal.is_finished" @click="handleFinish(goal)" class="w-full rounded-xl bg-gradient-to-r from-blue-600 to-indigo-500 text-white hover:from-blue-500 hover:to-indigo-400 shadow-sm border-0 font-bold h-10 transition-all active:scale-95 text-xs" size="sm">
                                <CheckCircle class="w-4 h-4 mr-2" /> Selesaikan & Cairkan
                            </Button>

                            <Button v-else disabled class="w-full rounded-xl bg-muted text-muted-foreground border border-border h-10 text-xs font-bold ring-1 ring-inset ring-black/5" size="sm">
                                 Selesai 🎉
                            </Button>

                             <Button variant="outline" class="w-full rounded-xl bg-background border-input hover:bg-accent hover:text-accent-foreground font-bold h-10 text-xs transition-all active:scale-95 px-4" @click="openDetailDialog(goal)">
                                 <Eye class="mr-2 h-4 w-4" /> Detail
                             </Button>
                        </div>
                    </CardContent>
                </Card>

                <!-- Empty State -->
                <div v-if="activeGoals.length === 0" class="col-span-full text-center py-20 text-muted-foreground border-2 border-dashed border-muted rounded-3xl bg-muted/10 h-80 flex flex-col items-center justify-center">
                    <div class="h-16 w-16 bg-muted rounded-full flex items-center justify-center mb-4">
                        <PiggyBank class="h-8 w-8 opacity-40" />
                    </div>
                    <p class="font-medium text-lg">Belum ada target aktif.</p>
                    <p class="text-sm opacity-70">Buat target baru untuk mulai menabung.</p>
                    <Button @click="openCreateDialog()" variant="link" class="mt-2 text-emerald-600">Tambah Baru</Button>
                </div>
            </div>
        </div>

        <!-- Finished Section -->
        <div v-if="finishedGoals.length > 0" class="space-y-4 pt-8 border-t border-border">
             <div class="flex items-center gap-2">
                <h3 class="text-base font-bold flex items-center gap-2 uppercase tracking-widest text-emerald-600">
                    <CheckCircle class="h-5 w-5" />
                    Tercapai
                </h3>
                <span class="text-[10px] font-bold uppercase tracking-widest px-2 py-0.5 rounded-full bg-emerald-50 text-emerald-700 border border-emerald-200">{{ finishedGoals.length }} item</span>
            </div>

            <div class="grid gap-6 md:grid-cols-2 lg:grid-cols-3">
                 <Card 
                    v-for="goal in finishedGoals" 
                    :key="goal.id" 
                    class="group relative overflow-hidden transition-all duration-300 hover:shadow-md hover:border-emerald-200 dark:hover:border-emerald-900 opacity-75 hover:opacity-100"
                >
                    <CardHeader class="pb-2">
                        <div class="flex justify-between items-start">
                            <div class="space-y-1">
                                <CardTitle class="text-xl font-bold tracking-tight text-muted-foreground flex items-center gap-2">
                                     <div class="p-1.5 bg-emerald-100 dark:bg-emerald-500/20 rounded-md text-emerald-600 dark:text-emerald-400">
                                        <CheckCircle class="w-4 h-4" />
                                    </div>
                                    <span class="line-through decoration-muted-foreground/50">{{ goal.name }}</span>
                                </CardTitle>
                                <div class="flex items-center gap-1 text-[10px] font-medium uppercase tracking-widest text-muted-foreground pl-1">
                                   <Calendar class="w-3 h-3" /> Selesai
                                </div>
                            </div>
                            <Button variant="ghost" size="icon" class="h-8 w-8 text-muted-foreground hover:bg-red-50 hover:text-red-600 dark:hover:bg-red-900/20" @click="handleDelete(goal.id)">
                                <Trash2 class="h-4 w-4" />
                            </Button>
                        </div>
                    </CardHeader>
                    
                    <CardContent class="space-y-4 pt-2">
                        <div class="p-4 rounded-xl bg-muted/30 border border-border space-y-1">
                           <div class="flex justify-between text-[10px] font-bold uppercase tracking-widest text-muted-foreground">
                               <span>Target Tercapai</span>
                               <span class="text-emerald-600 font-bold">100%</span>
                           </div>
                           <div class="text-2xl font-mono font-bold tracking-tight text-muted-foreground" :class="{ 'privacy-blur': authStore.isPrivacyMode }">
                               {{ formatCurrency(goal.target_amount) }}
                           </div>
                        </div>

                         <Button variant="outline" class="w-full rounded-xl bg-background border-input hover:bg-accent hover:text-accent-foreground font-bold h-10 text-xs transition-all active:scale-95 px-4" @click="openDetailDialog(goal)">
                             <Eye class="mr-2 h-4 w-4" /> Detail
                         </Button>
                    </CardContent>
                </Card>
            </div>
        </div>

        <!-- Create Dialog -->
        <Dialog :open="isCreateOpen" @update:open="isCreateOpen = $event">
            <DialogContent class="sm:max-w-[425px] rounded-3xl bg-card text-foreground">
                <DialogHeader>
                    <DialogTitle>{{ isEditing ? "Edit Target" : "Buat Target Baru" }}</DialogTitle>
                    <DialogDescription>Tentukan barang impian atau tujuan finansialmu.</DialogDescription>
                </DialogHeader>
                <div class="space-y-4 py-4">
                    <div class="space-y-2">
                        <Label class="text-xs font-bold uppercase tracking-widest text-muted-foreground">Nama Target</Label>
                        <Input v-model="newGoalName" placeholder="Misal: Beli Laptop Baru" class="h-11 shadow-sm rounded-xl bg-background" />
                    </div>
                    <div class="space-y-2">
                        <Label class="text-xs font-bold uppercase tracking-widest text-muted-foreground">Target Nominal</Label>
                        <Input type="text" inputmode="decimal" v-model="newGoalTargetDisplay" @blur="onTargetBlur" placeholder="Rp 0" class="h-11 shadow-sm rounded-xl bg-background" />
                    </div>
                    <div class="space-y-2">
                        <Label class="text-xs font-bold uppercase tracking-widest text-muted-foreground">Kategori</Label>
                        <SearchableSelect
                            v-model="newGoalCategory"
                            :options="categoryOptions"
                            placeholder="Pilih Kategori"
                        >
                            <template #option="{ option }">
                                <div class="flex items-center gap-2">
                                    <component v-if="getIconComponent(option.icon)" :is="getIconComponent(option.icon)" class="h-4 w-4 shrink-0" />
                                    <span v-else class="text-xs shrink-0">{{ getEmoji(option.icon) || option.icon }}</span>
                                    <span>{{ option.label }}</span>
                                </div>
                            </template>
                        </SearchableSelect>
                    </div>
                    <div class="space-y-2">
                        <Label class="text-xs font-bold uppercase tracking-widest text-muted-foreground">Batas Waktu (Opsional)</Label>
                        <Input type="date" v-model="newGoalDeadline" class="h-11 rounded-xl bg-background shadow-sm" />
                    </div>
                </div>
                <DialogFooter class="gap-2">
                    <Button variant="outline" @click="isCreateOpen = false" class="rounded-xl h-10 px-6">Batal</Button>
                    <Button @click="handleCreate" class="bg-gradient-to-r from-emerald-600 to-teal-500 text-white hover:from-emerald-500 hover:to-teal-400 shadow-md rounded-xl h-10 px-6 font-bold"
                        :disabled="isSubmitting" :loading="isSubmitting">
                        {{ isEditing ? "Simpan Perubahan" : "Buat Target" }}
                    </Button>
                </DialogFooter>
            </DialogContent>
        </Dialog>


        <ManualTransactionDialog 
            v-if="isContributeOpen"
            :open="true" 
            @update:open="(val) => {
                if (!val) {
                    isContributeOpen = false;
                    selectedGoalForContribution = null;
                    store.fetchGoals();
                    walletStore.fetchWallets();
                }
            }"
            @save="handleContributeClose"
            :savingGoalTarget="selectedGoalForContribution"
        />

        <Detail 
            v-model:open="isDetailOpen" 
            :goal="selectedGoal"
        />
    </div>
</template>
