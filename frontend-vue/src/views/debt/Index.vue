<script setup lang="ts">
import { ref, onMounted, computed, reactive, watch } from "vue";
import { useDebtStore, type Debt, type CreateDebtInput, type UpdateDebtInput, type PayDebtInput } from "@/stores/debt";
import { useWalletStore } from "@/stores/wallet";
import { useAuthStore } from "@/stores/auth";
import { Card, CardHeader, CardTitle, CardContent } from "@/components/ui/card";
import { Tabs, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Button } from "@/components/ui/button";
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogDescription, DialogFooter } from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import SearchableSelect from "@/components/ui/searchable-select/SearchableSelect.vue";
import MultiSelect from "@/components/ui/multi-select/MultiSelect.vue";
import { Plus, ArrowUpRight, ArrowDownLeft, Pencil, Trash2, HandCoins, CircleFadingArrowUp, Eye, Calendar, CheckCircle, X, ClipboardClock } from "lucide-vue-next";
import { getEmoji, getIconComponent } from "@/lib/icons";
import { formatCurrency, parseCurrencyInput, formatCurrencyInput, formatCurrencyLive, formatDate } from "@/lib/utils";
import Swal from "sweetalert2";
import Detail from "./Detail.vue";

const debtStore = useDebtStore();
const walletStore = useWalletStore();
const authStore = useAuthStore();

const isInitialLoading = ref(true);

onMounted(async () => {
  try {
    await Promise.all([
      debtStore.fetchDebts(),
      walletStore.fetchWallets()
    ]);
  } finally {
    isInitialLoading.value = false;
  }
});

const totalDebt = computed(() => {
  return debtStore.debts.reduce((sum: number, d: Debt) => sum + d.remaining, 0);
});

const totalReceivable = computed(() => {
  return debtStore.receivables.reduce((sum: number, d: Debt) => sum + d.remaining, 0);
});

const filterType = ref<'all' | 'debt' | 'receivable'>('all');
const filterWallet = ref<string[]>([]);

const allItems = computed(() => {
  return [...debtStore.debts, ...debtStore.receivables].sort((a, b) => {
    return new Date(b.created_at).getTime() - new Date(a.created_at).getTime();
  });
});

const filteredItems = computed(() => {
  let items = allItems.value;
  if (filterType.value !== 'all') {
    items = items.filter(item => item.type === filterType.value);
  }
  if (filterWallet.value.length > 0) {
    items = items.filter(item => filterWallet.value.includes(String(item.wallet_id)));
  }
  return items;
});

const activeItems = computed(() => filteredItems.value.filter(item => !item.is_paid));
const completedItems = computed(() => filteredItems.value.filter(item => item.is_paid));

const isCreateOpen = ref(false);
const isEditMode = ref(false);
const editingId = ref<number | null>(null);
const isPayOpen = ref(false);
const isDetailOpen = ref(false);
const activeTab = ref("debt");
const selectedDebt = ref<Debt | null>(null);
const isSubmitting = ref(false);

const createAmountDisplay = ref("");
const payAmountDisplay = ref("");

const createForm = reactive<CreateDebtInput>({
  name: "",
  amount: 0,
  wallet_id: 0,
  type: "debt",
  description: "",
  due_date: "",
});

const payForm = reactive<PayDebtInput>({
  amount: 0,
  wallet_id: 0,
  note: "",
});

const walletOptions = computed(() => walletStore.wallets.map(w => ({
  value: String(w.id),
  label: w.name,
  icon: w.icon
})));

const hasActiveFilters = computed(() => filterType.value !== 'all' || filterWallet.value.length > 0);

const resetFilters = () => {
  filterType.value = 'all';
  filterWallet.value = [];
};



watch(createAmountDisplay, (val) => {
  const formatted = formatCurrencyLive(val);
  if (formatted !== val) { createAmountDisplay.value = formatted; return; }
  createForm.amount = parseCurrencyInput(val);
});
watch(() => createForm.amount, (val) => {
  const currentParsed = parseCurrencyInput(createAmountDisplay.value);
  if (Math.abs(currentParsed - val) > 0.001) createAmountDisplay.value = val ? formatCurrencyInput(val) : "";
});
const onCreateBlur = () => {
  const num = parseCurrencyInput(createAmountDisplay.value);
  if (num) createAmountDisplay.value = formatCurrencyInput(num);
};

watch(payAmountDisplay, (val) => {
  const formatted = formatCurrencyLive(val);
  if (formatted !== val) { payAmountDisplay.value = formatted; return; }
  payForm.amount = parseCurrencyInput(val);
});
const onPayBlur = () => {
  const num = parseCurrencyInput(payAmountDisplay.value);
  if (num) payAmountDisplay.value = formatCurrencyInput(num);
};

const createWalletIdProxy = computed({
  get: () => String(createForm.wallet_id),
  set: (val: string) => createForm.wallet_id = Number(val)
});

const payWalletIdProxy = computed({
  get: () => String(payForm.wallet_id),
  set: (val: string) => payForm.wallet_id = Number(val)
});

const onTabChange = (val: string | number) => {
  if (isEditMode.value) return;
  createForm.type = val as 'debt' | 'receivable';
  activeTab.value = val as string;
};

const openCreateDialog = () => {
  isEditMode.value = false;
  editingId.value = null;
  activeTab.value = 'debt';
  createForm.type = 'debt';
  createForm.name = "";
  createForm.amount = 0;
  createForm.wallet_id = walletStore.wallets.length > 0 ? walletStore.wallets[0]!.id : 0;
  createForm.description = "";
  createForm.due_date = "";
  isCreateOpen.value = true;
};

const openEditDialog = (item: Debt) => {
  isEditMode.value = true;
  editingId.value = item.id;
  activeTab.value = item.type;

  createForm.type = item.type;
  createForm.name = item.name;
  createForm.amount = item.amount;
  createAmountDisplay.value = formatCurrencyInput(item.amount);
  createForm.wallet_id = item.wallet_id;
  createForm.description = item.description;

  if (item.due_date) {
    try {
      const date = new Date(item.due_date);
      if (date.getFullYear() > 1) {
        createForm.due_date = date.toISOString().split('T')[0];
      } else {
        createForm.due_date = "";
      }
    } catch (e) {
      createForm.due_date = "";
    }
  } else {
    createForm.due_date = "";
  }

  isCreateOpen.value = true;
};

const openPayDialog = (debt: Debt) => {
  selectedDebt.value = debt;
  payForm.amount = debt.remaining;
  payAmountDisplay.value = formatCurrencyInput(debt.remaining);
  payForm.wallet_id = debt.wallet_id || (walletStore.wallets.length > 0 ? walletStore.wallets[0]!.id : 0);
  payForm.note = "";
  isPayOpen.value = true;
};

const openDetailDialog = (debt: Debt) => {
  selectedDebt.value = debt;
  isDetailOpen.value = true;
};

const handleCreate = async () => {
  isSubmitting.value = true;
  try {
    const payload: any = { ...createForm };
    payload.amount = Number(payload.amount);
    payload.wallet_id = Number(payload.wallet_id);

    if (!payload.due_date) {
      payload.due_date = null;
    } else {
      const date = new Date(payload.due_date);
      if (!isNaN(date.getTime())) {
        payload.due_date = date.toISOString();
      } else {
        payload.due_date = null;
      }
    }

    if (isEditMode.value && editingId.value) {
      const updatePayload: UpdateDebtInput = {
        wallet_id: payload.wallet_id,
        name: payload.name,
        amount: payload.amount,
        description: payload.description,
        due_date: payload.due_date
      };
      await debtStore.updateDebt(editingId.value, updatePayload);
    } else {
      await debtStore.createDebt(payload);
    }

    isCreateOpen.value = false;
  } catch (e: any) {
    console.error(e);
  } finally {
    isSubmitting.value = false;
  }
};

const handlePay = async () => {
  if (!selectedDebt.value) return;
  isSubmitting.value = true;
  try {
    const payload = { ...payForm };
    payload.amount = Number(payload.amount);
    payload.wallet_id = Number(payload.wallet_id);

    await debtStore.payDebt(selectedDebt.value.id, payload);
    isPayOpen.value = false;

  } catch (e: any) {
    console.error(e);
  } finally {
    isSubmitting.value = false;
  }
};

const handleDelete = async (id: number) => {
  const result = await Swal.fire({
    title: 'Apakah Anda yakin?',
    text: "Data yang dihapus tidak dapat dikembalikan! Saldo wallet akan dikembalikan ke nilai sebelum transaksi ini.",
    icon: 'warning',
    showCancelButton: true,
    confirmButtonColor: '#d33',
    cancelButtonColor: '#3085d6',
    confirmButtonText: 'Ya, hapus!',
    cancelButtonText: 'Batal'
  });

  if (result.isConfirmed) {
    try {
      await debtStore.deleteDebt(id);
      Swal.fire(
        'Terhapus!',
        'Data berhasil dihapus.',
        'success'
      );
    } catch (e: any) {
      Swal.fire({
        icon: 'error',
        title: 'Gagal',
        text: e.response?.data?.error || 'Gagal menghapus data',
      });
    }
  }
};

</script>

<template>
  <div class="flex-1 space-y-6 pt-2" v-if="isInitialLoading">
    <div class="flex items-center justify-center min-h-[400px]">
      <p class="text-muted-foreground animate-pulse">Memuat data utang & piutang...</p>
    </div>
  </div>
  <div class="flex-1 space-y-6 pt-2 text-foreground" v-else>
    <div class="flex flex-col sm:flex-row justify-between items-start sm:items-center gap-4 sm:gap-0">
      <div>
        <h2 class="text-3xl font-bold tracking-tight">Utang & Piutang</h2>
        <p class="text-sm text-muted-foreground mt-1">Kelola catatan utang dan piutang Anda.</p>
      </div>
      <Button @click="openCreateDialog()"
        class="bg-gradient-to-r from-emerald-600 to-teal-500 text-white hover:from-emerald-500 hover:to-teal-400 shadow-lg h-12 rounded-full transition-all hover:scale-105 active:scale-95 px-6">
        <Plus class="mr-2 h-5 w-5" /> Tambah Baru
      </Button>
    </div>

    <!-- Summary Cards -->
    <div class="grid gap-6 md:grid-cols-2">
      <Card
        class="bg-gradient-to-br from-red-50 to-orange-50 dark:from-red-950/20 dark:to-orange-950/20 border border-red-100 dark:border-red-900/30 shadow-md rounded-3xl overflow-hidden hover:shadow-lg transition-shadow">
        <CardHeader class="flex flex-row items-center justify-between space-y-0 pb-2">
          <CardTitle class="text-sm font-semibold text-muted-foreground uppercase tracking-widest">Total Utang Saya
          </CardTitle>
          <ArrowUpRight class="h-4 w-4 text-red-500" />
        </CardHeader>
        <CardContent>
          <div class="text-2xl font-bold text-red-500" :class="{ 'privacy-blur': authStore.isPrivacyMode }">{{
            formatCurrency(totalDebt) }}</div>
          <p class="text-xs font-semibold text-muted-foreground mt-2 uppercase tracking-wide">Harus segera dibayar</p>
        </CardContent>
      </Card>
      <Card
        class="bg-gradient-to-br from-emerald-50 to-teal-50 dark:from-emerald-950/20 dark:to-teal-950/20 border border-emerald-100 dark:border-emerald-900/30 shadow-md rounded-3xl overflow-hidden hover:shadow-lg transition-shadow">
        <CardHeader class="flex flex-row items-center justify-between space-y-0 pb-2">
          <CardTitle class="text-sm font-semibold text-muted-foreground uppercase tracking-widest">Total Piutang Saya
          </CardTitle>
          <ArrowDownLeft class="h-4 w-4 text-emerald-500" />
        </CardHeader>
        <CardContent>
          <div class="text-2xl font-bold text-emerald-500" :class="{ 'privacy-blur': authStore.isPrivacyMode }">{{
            formatCurrency(totalReceivable) }}</div>
          <p class="text-xs font-semibold text-muted-foreground mt-2 uppercase tracking-wide">Akan segera diterima</p>
        </CardContent>
      </Card>
    </div>

    <!-- Main Content -->
    <div class="space-y-6">
      <div class="px-1 flex flex-col sm:flex-row items-stretch sm:items-center justify-between gap-4">
        <div class="flex flex-col sm:flex-row items-center gap-2 w-full sm:w-auto">
          <MultiSelect v-model="filterWallet" :options="walletOptions" placeholder="Semua Dompet" count-label="Dompet"
            class="w-full sm:w-[200px]" />

          <Select v-model="filterType">
            <SelectTrigger class="w-full sm:w-[180px] h-9 rounded-xl text-xs font-semibold bg-background">
              <SelectValue placeholder="Filter Tipe" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="all">Semua Tipe</SelectItem>
              <SelectItem value="debt">Utang (Payable)</SelectItem>
              <SelectItem value="receivable">Piutang (Receivable)</SelectItem>
            </SelectContent>
          </Select>

          <Button v-if="hasActiveFilters" variant="ghost"
            class="h-11 px-4 rounded-xl text-muted-foreground hover:text-foreground hover:bg-slate-100 dark:hover:bg-slate-800 gap-2"
            @click="resetFilters">
            <X class="h-4 w-4" />
            Reset
          </Button>
        </div>
      </div>

      <!-- Active Items -->
      <div>
        <h3 class="text-base font-bold flex items-center gap-2 mb-4">
          <Calendar class="h-5 w-5 text-muted-foreground" />
          Sedang Berjalan
          <span
            class="text-[10px] font-bold uppercase tracking-widest px-2 py-0.5 rounded-full bg-muted text-muted-foreground border border-border/50">{{
              activeItems.length }} item</span>
        </h3>

        <div class="grid gap-6 md:grid-cols-2 lg:grid-cols-3">
          <Card v-for="item in activeItems" :key="item.id"
            class="group relative overflow-hidden transition-all duration-300 hover:shadow-lg hover:-translate-y-1"
            :class="[
              item.type === 'debt' ? 'hover:border-red-200 dark:hover:border-red-900' : 'hover:border-emerald-200 dark:hover:border-emerald-900'
            ]">
            <CardHeader class="pb-3">
              <div class="flex justify-between items-start">
                <div class="space-y-1">
                  <div class="flex items-center gap-2 mb-1">
                    <span v-if="item.type === 'debt'"
                      class="px-2 py-0.5 rounded-md bg-red-100 dark:bg-red-500/20 text-red-700 dark:text-red-400 text-[10px] font-bold uppercase tracking-widest border border-red-200 dark:border-red-800">Utang</span>
                    <span v-else
                      class="px-2 py-0.5 rounded-md bg-emerald-100 dark:bg-emerald-500/20 text-emerald-700 dark:text-emerald-400 text-[10px] font-bold uppercase tracking-widest border border-emerald-200 dark:border-emerald-800">Piutang</span>
                  </div>
                  <CardTitle class="text-lg font-bold tracking-tight">{{ item.name }}</CardTitle>
                  <p class="text-xs text-muted-foreground line-clamp-1">{{ item.description || '-' }}</p>
                </div>

                <div class="flex gap-1">
                  <Button variant="ghost" size="icon"
                    class="h-8 w-8 text-muted-foreground hover:bg-slate-100 hover:text-blue-600 dark:hover:bg-slate-800"
                    @click="openEditDialog(item)">
                    <Pencil class="h-4 w-4" />
                  </Button>
                  <Button variant="ghost" size="icon"
                    class="h-8 w-8 text-muted-foreground hover:bg-red-50 hover:text-red-600 dark:hover:bg-red-900/20"
                    @click="handleDelete(item.id)">
                    <Trash2 class="h-4 w-4" />
                  </Button>
                </div>
              </div>
            </CardHeader>

            <CardContent class="space-y-4">
              <!-- Amount Section -->
              <div class="p-4 rounded-xl bg-muted/50 border border-border space-y-1">
                <div class="flex justify-between text-[10px] font-bold uppercase tracking-widest text-muted-foreground">
                  <span>Sisa {{ item.type === 'debt' ? 'Bayar' : 'Terima' }}</span>
                  <span class="text-amber-600">Belum Lunas</span>
                </div>
                <div class="text-2xl font-mono font-bold tracking-tight text-foreground"
                  :class="{ 'privacy-blur': authStore.isPrivacyMode }">
                  {{ formatCurrency(item.remaining) }}
                </div>
                <div class="text-xs text-muted-foreground pt-2 border-t border-border/50 mt-2 flex justify-between">
                  <span>Total Awal: <span :class="{ 'privacy-blur': authStore.isPrivacyMode }">{{
                    formatCurrency(item.amount) }}</span></span>
                </div>
              </div>

              <div class="flex flex-wrap items-center gap-x-3 gap-y-1 text-xs text-muted-foreground font-medium p-2">
                <div class="flex items-center gap-1.5">
                  <component v-if="item.wallet && getIconComponent(item.wallet.icon)"
                    :is="getIconComponent(item.wallet.icon)" class="h-4 w-4" />
                  <span v-else-if="item.wallet">{{ getEmoji(item.wallet.icon) || '💼' }}</span>
                  <span v-else>💼</span>
                  <span>{{ item.wallet?.name || 'No Wallet' }}</span>
                </div>

                <span class="opacity-30">·</span>

                <div class="flex items-center gap-1 text-[10px]">
                  <Calendar class="h-3 w-3" />
                  <span>Dibuat {{ formatDate(item.created_at) }}</span>
                </div>

                <span class="opacity-30">·</span>

                <div class="flex items-center gap-1">
                  <ClipboardClock class="h-3 w-3" />
                  <span>{{ item.due_date ? formatDate(item.due_date) : 'Tanpa Tenggat' }}</span>
                </div>
              </div>

              <div class="grid grid-cols-[1fr,auto] gap-2 pt-2">
                <Button @click="openPayDialog(item)"
                  :class="item.type === 'debt' ? 'bg-red-600 hover:bg-red-700 text-white' : 'bg-emerald-600 hover:bg-emerald-700 text-white'"
                  class="w-full rounded-xl shadow-sm border-0 font-bold h-10 text-xs transition-all active:scale-95">
                  <circle-fading-arrow-up v-if="item.type === 'debt'" class="mr-2 h-4 w-4" />
                  <hand-coins v-else class="mr-2 h-4 w-4" />
                  {{ item.type === 'debt' ? 'Bayar Sekarang' : 'Terima Pembayaran' }}
                </Button>

                <Button variant="outline"
                  class="w-full rounded-xl bg-background border-input hover:bg-accent hover:text-accent-foreground font-bold h-10 text-xs transition-all active:scale-95 px-4"
                  @click="openDetailDialog(item)">
                  <Eye class="mr-2 h-4 w-4" /> Detail
                </Button>
              </div>
            </CardContent>
          </Card>

          <!-- Empty State Active -->
          <div v-if="activeItems.length === 0"
            class="col-span-full text-center py-20 text-muted-foreground border-2 border-dashed border-muted rounded-3xl bg-muted/10 h-80 flex flex-col items-center justify-center">
            <div class="h-16 w-16 bg-muted rounded-full flex items-center justify-center mb-4">
              <HandCoins class="h-8 w-8 opacity-40" />
            </div>
            <p class="font-medium text-lg">Belum ada catatan aktif.</p>
            <p class="text-sm opacity-70">Semua utang dan piutang telah lunas!</p>
            <Button @click="openCreateDialog()" variant="link" class="mt-2 text-emerald-600">Tambah Baru</Button>
          </div>
        </div>
      </div>

      <!-- Completed Items -->
      <div v-if="completedItems.length > 0" class="pt-8 border-t border-border">
        <div class="flex items-center gap-2 mb-4">
          <h3 class="text-base font-bold flex items-center gap-2 uppercase tracking-widest text-emerald-600">
            <CircleFadingArrowUp class="h-5 w-5" />
            Riwayat Selesai
          </h3>
          <span
            class="text-[10px] font-bold uppercase tracking-widest px-2 py-0.5 rounded-full bg-emerald-50 text-emerald-700 border border-emerald-200">{{
              completedItems.length }} item</span>
        </div>

        <div class="grid gap-6 md:grid-cols-2 lg:grid-cols-3">
          <Card v-for="item in completedItems" :key="item.id"
            class="group relative overflow-hidden transition-all duration-300 hover:shadow-md hover:border-emerald-200 dark:hover:border-emerald-900 opacity-75 hover:opacity-100">
            <CardHeader class="pb-3">
              <div class="flex justify-between items-start">
                <div class="space-y-1">
                  <div class="flex items-center gap-2 mb-1">
                    <span v-if="item.type === 'debt'"
                      class="px-2 py-0.5 rounded-md bg-slate-100 dark:bg-slate-800 text-slate-600 dark:text-slate-400 text-[10px] font-bold uppercase tracking-widest border border-slate-200 dark:border-slate-700">Utang
                      Lunas</span>
                    <span v-else
                      class="px-2 py-0.5 rounded-md bg-emerald-100 dark:bg-emerald-500/20 text-emerald-700 dark:text-emerald-400 text-[10px] font-bold uppercase tracking-widest border border-emerald-200 dark:border-emerald-800">Piutang
                      Lunas</span>
                  </div>
                  <CardTitle
                    class="text-lg font-bold tracking-tight text-muted-foreground line-through decoration-muted-foreground/50">
                    {{ item.name }}</CardTitle>
                  <div
                    class="flex items-center gap-3 text-[10px] font-medium uppercase tracking-widest text-muted-foreground">
                    <div class="flex items-center gap-1">
                      <component v-if="item.wallet && getIconComponent(item.wallet.icon)"
                        :is="getIconComponent(item.wallet.icon)" class="h-3 w-3" />
                      <span v-else-if="item.wallet">{{ getEmoji(item.wallet.icon) || '💼' }}</span>
                      <span>{{ item.wallet?.name }}</span>
                    </div>
                    <span class="opacity-30">·</span>
                    <div class="flex items-center gap-1">
                      <Calendar class="w-3 h-3" />
                      <span>Dibuat {{ formatDate(item.created_at) }}</span>
                    </div>
                  </div>
                </div>
                <Button variant="ghost" size="icon"
                  class="h-8 w-8 text-muted-foreground hover:bg-red-50 hover:text-red-600 dark:hover:bg-red-900/20"
                  @click="handleDelete(item.id)">
                  <Trash2 class="h-4 w-4" />
                </Button>
              </div>
            </CardHeader>

            <CardContent class="space-y-4 pt-0">
              <div class="p-4 rounded-xl bg-muted/30 border border-border space-y-1 mt-2">
                <div class="flex justify-between text-[10px] font-bold uppercase tracking-widest text-muted-foreground">
                  <span>Total Nominal</span>
                  <span class="text-emerald-600 flex items-center gap-1">
                    <CheckCircle class="w-3 h-3" /> Selesai
                  </span>
                </div>
                <div class="text-2xl font-mono font-bold tracking-tight text-muted-foreground"
                  :class="{ 'privacy-blur': authStore.isPrivacyMode }">
                  {{ formatCurrency(item.amount) }}
                </div>
              </div>

              <Button variant="outline"
                class="w-full rounded-xl bg-background border-input hover:bg-accent hover:text-accent-foreground font-bold h-10 text-xs transition-all active:scale-95 px-4"
                @click="openDetailDialog(item)">
                <Eye class="mr-2 h-4 w-4" /> Detail
              </Button>
            </CardContent>
          </Card>
        </div>
      </div>
    </div>

    <!-- Create/Edit Dialog -->
    <Dialog :open="isCreateOpen" @update:open="isCreateOpen = $event">
      <DialogContent class="sm:max-w-[500px] rounded-3xl bg-card text-foreground">
        <DialogHeader>
          <DialogTitle>{{ isEditMode ? 'Edit Utang/Piutang' : 'Catat Utang/Piutang Baru' }}</DialogTitle>
          <DialogDescription>
            {{ isEditMode ? 'Perbarui detail data.' : 'Pilih tipe transaksi dan isi detailnya.' }}
          </DialogDescription>
        </DialogHeader>

        <Tabs v-model="activeTab" class="w-full mt-2" @update:modelValue="onTabChange">
          <TabsList class="grid w-full grid-cols-2 mb-4 h-auto p-1 bg-muted/60 rounded-xl">
            <TabsTrigger value="debt" :disabled="isEditMode"
              class="rounded-lg py-2 h-auto whitespace-normal text-xs sm:text-sm data-[state=active]:bg-red-600 data-[state=active]:text-white dark:data-[state=active]:bg-red-600 dark:data-[state=active]:text-white hover:bg-red-50 dark:hover:bg-red-900/10 hover:text-red-600 transition-colors disabled:opacity-50">
              Utang <span class="hidden sm:inline">(Saya Berutang)</span><span
                class="sm:hidden block text-[10px] opacity-80 font-normal">(Saya Berutang)</span></TabsTrigger>
            <TabsTrigger value="receivable" :disabled="isEditMode"
              class="rounded-lg py-2 h-auto whitespace-normal text-xs sm:text-sm data-[state=active]:bg-emerald-600 data-[state=active]:text-white dark:data-[state=active]:bg-emerald-600 dark:data-[state=active]:text-white hover:bg-emerald-50 dark:hover:bg-emerald-900/10 hover:text-emerald-600 transition-colors disabled:opacity-50">
              Piutang <span class="hidden sm:inline">(Orang Berutang)</span><span
                class="sm:hidden block text-[10px] opacity-80 font-normal">(Orang Berutang)</span></TabsTrigger>
          </TabsList>
        </Tabs>

        <div class="grid gap-4 py-2">
          <div class="grid gap-2">
            <Label>Nama {{ createForm.type === 'debt' ? 'Pemberi Utang' : 'Peminjam' }}</Label>
            <Input v-model="createForm.name" placeholder="Contoh: Budi"
              class="h-11 shadow-sm rounded-xl bg-background" />
          </div>
          <div class="grid gap-2">
            <Label>Nominal</Label>
            <Input type="text" inputmode="decimal" placeholder="Rp 0" v-model="createAmountDisplay" @blur="onCreateBlur"
              class="h-11 shadow-sm rounded-xl bg-background" />
            <p v-if="isEditMode" class="text-[10px] font-medium text-yellow-600 px-1">
              Perhatian: Mengubah nominal akan menyesuaikan ulang saldo dompet.
            </p>
          </div>
          <div class="grid gap-2">
            <Label>Dompet Terkait</Label>
            <SearchableSelect v-model="createWalletIdProxy" :options="walletOptions" placeholder="Pilih Dompet">
              <template #option="{ option }">
                <div class="flex items-center gap-2">
                  <component v-if="getIconComponent(option.icon)" :is="getIconComponent(option.icon)"
                    class="h-4 w-4 shrink-0" />
                  <span v-else class="text-xs shrink-0">{{ getEmoji(option.icon) || '💼' }}</span>
                  <span>{{ option.label }}</span>
                </div>
              </template>
            </SearchableSelect>
            <p v-if="!isEditMode" class="text-[10px] font-bold uppercase tracking-widest mt-1 px-1">
              <span v-if="createForm.type === 'debt'" class="text-emerald-600 flex items-center gap-1">
                <ArrowDownLeft class="h-3 w-3" /> Saldo Dompet Bertambah (Income)
              </span>
              <span v-else class="text-red-600 flex items-center gap-1">
                <ArrowUpRight class="h-3 w-3" /> Saldo Dompet Berkurang (Expense)
              </span>
            </p>
          </div>
          <div class="grid gap-2">
            <Label>Keterangan</Label>
            <Input v-model="createForm.description" placeholder="Optional"
              class="h-11 shadow-sm rounded-xl bg-background" />
          </div>
          <div class="grid gap-2">
            <Label>Jatuh Tempo</Label>
            <div class="relative">
              <input type="date" :value="createForm.due_date || ''"
                @input="(e) => createForm.due_date = (e.target as HTMLInputElement).value"
                class="peer absolute inset-0 w-full h-full opacity-0 cursor-pointer z-10"
                @click="($event.target as HTMLInputElement).showPicker()" />
              <Input type="text" readonly tabindex="-1"
                :value="createForm.due_date ? formatDate(createForm.due_date, 'dd/MM/yyyy') : ''"
                placeholder="dd/mm/yyyy"
                class="h-11 shadow-sm rounded-xl bg-background block w-full cursor-pointer peer-focus-visible:ring-2 peer-focus-visible:ring-ring peer-focus-visible:ring-offset-2 text-foreground" />
              <div
                class="absolute inset-y-0 right-0 flex items-center pr-3 pointer-events-none text-muted-foreground z-20">
                <Calendar class="w-4 h-4" />
              </div>
            </div>
          </div>
        </div>
        <DialogFooter class="gap-2">
          <Button variant="outline" @click="isCreateOpen = false" class="rounded-xl h-10 px-6">Batal</Button>
          <Button @click="handleCreate"
            :class="createForm.type === 'debt' ? 'bg-red-600 hover:bg-red-700' : 'bg-gradient-to-r from-emerald-600 to-teal-500 hover:from-emerald-500 hover:to-teal-400 font-bold'"
            class="rounded-xl h-10 px-6 text-white shadow-md" :disabled="isSubmitting" :loading="isSubmitting">
            {{ isEditMode ? 'Simpan Perubahan' : 'Simpan' }}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>

    <!-- Pay Dialog -->
    <Dialog :open="isPayOpen" @update:open="isPayOpen = $event">
      <DialogContent class="sm:max-w-[425px] rounded-3xl bg-card text-foreground">
        <DialogHeader>
          <DialogTitle>{{ selectedDebt?.type === 'debt' ? 'Bayar Utang' : 'Terima Pembayaran' }}</DialogTitle>
          <DialogDescription>
            Catat cicilan atau pelunasan. Saldo dompet akan otomatis disesuaikan.
          </DialogDescription>
        </DialogHeader>
        <div class="grid gap-4 py-4">
          <div class="grid gap-2">
            <Label>Sisa Kewajiban</Label>
            <div class="text-xl font-bold text-foreground">{{ selectedDebt ? formatCurrency(selectedDebt.remaining) : 0
            }}
            </div>
          </div>
          <div class="grid gap-2">
            <Label>Nominal Pembayaran</Label>
            <Input type="text" inputmode="decimal" placeholder="Rp 0" v-model="payAmountDisplay" @blur="onPayBlur"
              class="h-11 shadow-sm rounded-xl bg-background" />
          </div>
          <div class="grid gap-2">
            <Label>Dompet Sumber/Tujuan</Label>
            <SearchableSelect v-model="payWalletIdProxy" :options="walletOptions" placeholder="Pilih Dompet">
              <template #option="{ option }">
                <div class="flex items-center gap-2">
                  <component v-if="getIconComponent(option.icon)" :is="getIconComponent(option.icon)"
                    class="h-4 w-4 shrink-0" />
                  <span v-else class="text-xs shrink-0">{{ getEmoji(option.icon) || '💼' }}</span>
                  <span>{{ option.label }}</span>
                </div>
              </template>
            </SearchableSelect>
            <p class="text-[10px] font-bold uppercase tracking-widest mt-1 px-1">
              <span v-if="selectedDebt?.type === 'debt'" class="text-red-600 flex items-center gap-1">
                <ArrowUpRight class="h-3 w-3" /> Saldo Dompet Berkurang (Expense)
              </span>
              <span v-else class="text-emerald-600 flex items-center gap-1">
                <ArrowDownLeft class="h-3 w-3" /> Saldo Dompet Bertambah (Income)
              </span>
            </p>
          </div>
          <div class="grid gap-2">
            <Label>Catatan (Opsional)</Label>
            <Input v-model="payForm.note" placeholder="Contoh: Cicilan ke-1"
              class="h-11 shadow-sm rounded-xl bg-background" />
          </div>
        </div>
        <DialogFooter class="gap-2">
          <Button variant="outline" @click="isPayOpen = false" class="rounded-xl h-10 px-6">Batal</Button>
          <Button @click="handlePay"
            :class="selectedDebt?.type === 'debt' ? 'bg-red-600 hover:bg-red-700' : 'bg-gradient-to-r from-emerald-600 to-teal-500 hover:from-emerald-500 hover:to-teal-400 font-bold'"
            class="rounded-xl h-10 px-6 text-white shadow-md" :disabled="isSubmitting"
            :loading="isSubmitting">Proses</Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>

    <Detail v-model:open="isDetailOpen" :debt="selectedDebt" />
  </div>
</template>
