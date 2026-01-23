<script setup lang="ts">
import { ref, onMounted, computed, reactive, watch } from "vue";
import { useDebtStore, type Debt, type CreateDebtInput, type UpdateDebtInput, type PayDebtInput } from "@/stores/debt";
import { useWalletStore } from "@/stores/wallet";
import { Card, CardHeader, CardTitle, CardContent } from "@/components/ui/card";
import { Tabs, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Button } from "@/components/ui/button";
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogDescription, DialogFooter } from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Plus, ArrowUpRight, ArrowDownLeft, Pencil, Trash2, HandCoins, CircleFadingArrowUp, Eye, Calendar } from "lucide-vue-next";
import { format } from "date-fns";
import { id } from "date-fns/locale";
import { getEmoji, getIconComponent } from "@/lib/icons";
import { formatCurrency, parseCurrencyInput, formatCurrencyInput, formatCurrencyLive } from "@/lib/utils";
import Swal from "sweetalert2";
import Detail from "./Detail.vue";

const debtStore = useDebtStore();
const walletStore = useWalletStore();

onMounted(async () => {
  await Promise.all([
    debtStore.fetchDebts(),
    walletStore.fetchWallets()
  ]);
});

// Local formatCurrency removed

// Summary Stats
const totalDebt = computed(() => {
  return debtStore.debts.reduce((sum: number, d: Debt) => sum + d.remaining, 0);
});

const totalReceivable = computed(() => {
  return debtStore.receivables.reduce((sum: number, d: Debt) => sum + d.remaining, 0);
});

// Filter & List Logic
const filterType = ref<'all' | 'debt' | 'receivable'>('all');

const allItems = computed(() => {
    return [...debtStore.debts, ...debtStore.receivables].sort((a, b) => {
        return new Date(b.created_at).getTime() - new Date(a.created_at).getTime();
    });
});

const filteredItems = computed(() => {
    if (filterType.value === 'all') return allItems.value;
    return allItems.value.filter(item => item.type === filterType.value);
});

// Dialog States
const isCreateOpen = ref(false);
const isEditMode = ref(false);
const editingId = ref<number | null>(null);
const isPayOpen = ref(false);
const isDetailOpen = ref(false);
const activeTab = ref("debt"); // 'debt' or 'receivable'
const selectedDebt = ref<Debt | null>(null);

// Display Refs for loose binding
const createAmountDisplay = ref("");
const payAmountDisplay = ref("");

// Forms
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

// Helper for Wallet Selection
const selectedCreateWalletObj = computed(() => walletStore.wallets.find(w => String(w.id) === String(createForm.wallet_id)));
const selectedPayWalletObj = computed(() => walletStore.wallets.find(w => String(w.id) === String(payForm.wallet_id)));



// Sync Create Amount
watch(createAmountDisplay, (val) => {
    const formatted = formatCurrencyLive(val);
    if(formatted !== val) { createAmountDisplay.value = formatted; return; }
    createForm.amount = parseCurrencyInput(val);
});
watch(() => createForm.amount, (val) => {
    const currentParsed = parseCurrencyInput(createAmountDisplay.value);
    if(Math.abs(currentParsed - val) > 0.001) createAmountDisplay.value = val ? formatCurrencyInput(val) : "";
});
const onCreateBlur = () => {
    const num = parseCurrencyInput(createAmountDisplay.value);
    if(num) createAmountDisplay.value = formatCurrencyInput(num);
};

// Sync Pay Amount
watch(payAmountDisplay, (val) => {
    const formatted = formatCurrencyLive(val);
    if(formatted !== val) { payAmountDisplay.value = formatted; return; }
    payForm.amount = parseCurrencyInput(val);
});
const onPayBlur = () => {
    const num = parseCurrencyInput(payAmountDisplay.value);
    if(num) payAmountDisplay.value = formatCurrencyInput(num);
};
// Pay amount is usually initialized from remaining, so we need init watcher or manual init.

// Proxy for Select
const createWalletIdProxy = computed({
  get: () => String(createForm.wallet_id),
  set: (val: string) => createForm.wallet_id = Number(val)
});

const payWalletIdProxy = computed({
  get: () => String(payForm.wallet_id),
  set: (val: string) => payForm.wallet_id = Number(val)
});

// Watch for tab change in Create Dialog to update form type
const onTabChange = (val: string | number) => {
    if (isEditMode.value) return; // Prevent changing type in edit mode
    createForm.type = val as 'debt' | 'receivable';
    activeTab.value = val as string;
};

const openCreateDialog = () => {
  isEditMode.value = false;
  editingId.value = null;
  activeTab.value = 'debt'; // Default
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
         Swal.fire({
          icon: 'success',
          title: 'Berhasil',
          text: 'Data berhasil diperbarui',
          timer: 1500,
          showConfirmButton: false
        });
    } else {
        await debtStore.createDebt(payload);
         Swal.fire({
          icon: 'success',
          title: 'Berhasil',
          text: 'Data berhasil disimpan',
          timer: 1500,
          showConfirmButton: false
        });
    }
    
    isCreateOpen.value = false;
  } catch (e: any) {
    console.error(e);
    Swal.fire({
      icon: 'error',
      title: 'Gagal',
      text: e.response?.data?.error || 'Terjadi kesalahan saat menyimpan data',
    });
  }
};

const handlePay = async () => {
  if (!selectedDebt.value) return;
  try {
     const payload = { ...payForm };
     payload.amount = Number(payload.amount);
     payload.wallet_id = Number(payload.wallet_id);

    await debtStore.payDebt(selectedDebt.value.id, payload);
    isPayOpen.value = false;
    
    Swal.fire({
      icon: 'success',
      title: 'Berhasil',
      text: 'Pembayaran berhasil diproses',
      timer: 1500,
      showConfirmButton: false
    });
  } catch (e: any) {
    console.error(e);
    Swal.fire({
      icon: 'error',
      title: 'Gagal',
      text: e.response?.data?.error || 'Terjadi kesalahan saat memproses pembayaran',
    });
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
  <div class="flex-1 space-y-6 pt-2" v-if="debtStore.isLoading">
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
      <Button @click="openCreateDialog()" class="bg-gradient-to-r from-emerald-600 to-teal-500 text-white hover:from-emerald-500 hover:to-teal-400 shadow-lg h-12 rounded-full transition-all hover:scale-105 active:scale-95 px-6">
        <Plus class="mr-2 h-5 w-5" /> Tambah Baru
      </Button>
    </div>

    <!-- Summary Cards -->
    <div class="grid gap-6 md:grid-cols-2">
      <Card class="bg-gradient-to-br from-red-50 to-orange-50 dark:from-red-950/20 dark:to-orange-950/20 border border-red-100 dark:border-red-900/30 shadow-md rounded-3xl overflow-hidden hover:shadow-lg transition-shadow">
        <CardHeader class="flex flex-row items-center justify-between space-y-0 pb-2">
          <CardTitle class="text-sm font-semibold text-muted-foreground uppercase tracking-widest">Total Utang Saya</CardTitle>
          <ArrowUpRight class="h-4 w-4 text-red-500" />
        </CardHeader>
        <CardContent>
          <div class="text-2xl font-bold text-red-500">{{ formatCurrency(totalDebt) }}</div>
          <p class="text-xs font-semibold text-muted-foreground mt-2 uppercase tracking-wide">Harus segera dibayar</p>
        </CardContent>
      </Card>
      <Card class="bg-gradient-to-br from-emerald-50 to-teal-50 dark:from-emerald-950/20 dark:to-teal-950/20 border border-emerald-100 dark:border-emerald-900/30 shadow-md rounded-3xl overflow-hidden hover:shadow-lg transition-shadow">
        <CardHeader class="flex flex-row items-center justify-between space-y-0 pb-2">
          <CardTitle class="text-sm font-semibold text-muted-foreground uppercase tracking-widest">Total Piutang Saya</CardTitle>
          <ArrowDownLeft class="h-4 w-4 text-emerald-500" />
        </CardHeader>
        <CardContent>
          <div class="text-2xl font-bold text-emerald-500">{{ formatCurrency(totalReceivable) }}</div>
          <p class="text-xs font-semibold text-muted-foreground mt-2 uppercase tracking-wide">Akan segera diterima</p>
        </CardContent>
      </Card>
    </div>

    <!-- Main Content -->
    <div class="space-y-6">
          <div class="px-1 flex flex-col sm:flex-row items-stretch sm:items-center justify-between gap-4">
             <div class="flex items-center gap-2 w-full sm:w-auto">
                <Select v-model="filterType">
                    <SelectTrigger class="w-full sm:w-[180px] h-11 rounded-xl shadow-sm bg-background">
                        <SelectValue placeholder="Filter Tipe" />
                    </SelectTrigger>
                    <SelectContent>
                        <SelectItem value="all">Semua</SelectItem>
                        <SelectItem value="debt">Utang (Payable)</SelectItem>
                        <SelectItem value="receivable">Piutang (Receivable)</SelectItem>
                    </SelectContent>
                </Select>
             </div>
          </div>

          <div class="grid gap-6 md:grid-cols-2 lg:grid-cols-3">
                <Card 
                  v-for="item in filteredItems" 
                  :key="item.id" 
                  class="group relative overflow-hidden transition-all duration-300 hover:shadow-lg hover:-translate-y-1"
                  :class="[
                      item.type === 'debt' ? 'hover:border-red-200 dark:hover:border-red-900' : 'hover:border-emerald-200 dark:hover:border-emerald-900'
                  ]"
               >
                  <CardHeader class="pb-3">
                      <div class="flex justify-between items-start">
                          <div class="space-y-1">
                              <div class="flex items-center gap-2 mb-1">
                                  <span v-if="item.type === 'debt'" class="px-2 py-0.5 rounded-md bg-red-100 dark:bg-red-500/20 text-red-700 dark:text-red-400 text-[10px] font-bold uppercase tracking-widest border border-red-200 dark:border-red-800">Utang</span>
                                  <span v-else class="px-2 py-0.5 rounded-md bg-emerald-100 dark:bg-emerald-500/20 text-emerald-700 dark:text-emerald-400 text-[10px] font-bold uppercase tracking-widest border border-emerald-200 dark:border-emerald-800">Piutang</span>
                              </div>
                              <CardTitle class="text-lg font-bold tracking-tight">{{ item.name }}</CardTitle>
                              <p class="text-xs text-muted-foreground line-clamp-1">{{ item.description || '-' }}</p>
                          </div>
                          
                          <div class="flex gap-1">
                              <Button variant="ghost" size="icon" class="h-8 w-8 text-muted-foreground hover:bg-slate-100 hover:text-blue-600 dark:hover:bg-slate-800" @click="openEditDialog(item)">
                                  <Pencil class="h-4 w-4" />
                              </Button>
                              <Button variant="ghost" size="icon" class="h-8 w-8 text-muted-foreground hover:bg-red-50 hover:text-red-600 dark:hover:bg-red-900/20" @click="handleDelete(item.id)">
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
                               <span v-if="item.is_paid" class="text-emerald-600">Lunas</span>
                               <span v-else class="text-amber-600">Belum Lunas</span>
                           </div>
                           <div class="text-2xl font-mono font-bold tracking-tight text-foreground">
                              {{ formatCurrency(item.remaining) }}
                           </div>
                           <div class="text-xs text-muted-foreground pt-2 border-t border-border/50 mt-2 flex justify-between">
                              <span>Total Awal: {{ formatCurrency(item.amount) }}</span>
                           </div>
                      </div>
                      
                      <div class="flex items-center gap-2 text-xs text-muted-foreground font-medium p-2">
                           <component v-if="item.wallet && getIconComponent(item.wallet.icon)" :is="getIconComponent(item.wallet.icon)" class="h-4 w-4" />
                           <span v-else-if="item.wallet">{{ getEmoji(item.wallet.icon) || '💼' }}</span>
                           <span v-else>💼</span>
                           <span>{{ item.wallet?.name || 'No Wallet' }}</span>
                           
                           <span class="mx-1 opacity-50">|</span>
                           
                           <Calendar class="h-3.5 w-3.5" />
                           <span>{{ item.due_date ? format(new Date(item.due_date), "d MMM yyyy", { locale: id }) : 'Tanpa Tenggat' }}</span>
                      </div>

                      <div class="grid grid-cols-[1fr,auto] gap-2 pt-2">
                           <Button 
                              v-if="!item.is_paid" 
                              @click="openPayDialog(item)" 
                              :class="item.type === 'debt' ? 'bg-red-600 hover:bg-red-700 text-white' : 'bg-emerald-600 hover:bg-emerald-700 text-white'"
                              class="w-full rounded-xl shadow-sm border-0 font-bold h-10 text-xs transition-all active:scale-95"
                            >
                              <circle-fading-arrow-up v-if="item.type === 'debt'" class="mr-2 h-4 w-4" />
                              <hand-coins v-else class="mr-2 h-4 w-4" />
                              {{ item.type === 'debt' ? 'Bayar Sekarang' : 'Terima Pembayaran' }}
                          </Button>
                          <Button v-else disabled class="w-full rounded-xl bg-muted text-muted-foreground border border-border h-10 text-xs font-bold opacity-50 cursor-not-allowed">
                              Selesai
                          </Button>

                           <Button variant="outline" class="w-full rounded-xl bg-background border-input hover:bg-accent hover:text-accent-foreground font-bold h-10 text-xs transition-all active:scale-95 px-4" @click="openDetailDialog(item)">
                              <Eye class="mr-2 h-4 w-4" /> Detail
                           </Button>
                      </div>
                  </CardContent>
               </Card>
               
               <!-- Empty State -->
               <div v-if="filteredItems.length === 0" class="col-span-full text-center py-20 text-muted-foreground border-2 border-dashed border-muted rounded-3xl bg-muted/10 h-80 flex flex-col items-center justify-center">
                   <div class="h-16 w-16 bg-muted rounded-full flex items-center justify-center mb-4">
                       <HandCoins class="h-8 w-8 opacity-40" />
                   </div>
                   <p class="font-medium text-lg">Belum ada catatan.</p>
                   <p class="text-sm opacity-70">Mulai catat utang atau piutang Anda.</p>
                   <Button @click="openCreateDialog()" variant="link" class="mt-2 text-emerald-600">Tambah Baru</Button>
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
            <TabsTrigger value="debt" :disabled="isEditMode" class="rounded-lg py-2 data-[state=active]:bg-red-600 data-[state=active]:text-white dark:data-[state=active]:bg-red-600 dark:data-[state=active]:text-white hover:bg-red-50 dark:hover:bg-red-900/10 hover:text-red-600 transition-colors disabled:opacity-50">Utang (Saya Berutang)</TabsTrigger>
            <TabsTrigger value="receivable" :disabled="isEditMode" class="rounded-lg py-2 data-[state=active]:bg-emerald-600 data-[state=active]:text-white dark:data-[state=active]:bg-emerald-600 dark:data-[state=active]:text-white hover:bg-emerald-50 dark:hover:bg-emerald-900/10 hover:text-emerald-600 transition-colors disabled:opacity-50">Piutang (Orang Berutang)</TabsTrigger>
          </TabsList>
        </Tabs>
 
        <div class="grid gap-4 py-2">
          <div class="grid gap-2">
            <Label>Nama {{ createForm.type === 'debt' ? 'Pemberi Utang' : 'Peminjam' }}</Label>
            <Input v-model="createForm.name" placeholder="Contoh: Budi" class="h-11 shadow-sm rounded-xl bg-background" />
          </div>
          <div class="grid gap-2">
            <Label>Nominal</Label>
            <Input type="text" inputmode="decimal" placeholder="Rp 0" v-model="createAmountDisplay" @blur="onCreateBlur" class="h-11 shadow-sm rounded-xl bg-background" />
            <p v-if="isEditMode" class="text-[10px] font-medium text-yellow-600 px-1">
                Perhatian: Mengubah nominal akan menyesuaikan ulang saldo dompet.
            </p>
          </div>
          <div class="grid gap-2">
            <Label>Dompet Terkait</Label>
            <Select v-model="createWalletIdProxy">
              <SelectTrigger class="w-full h-11 rounded-xl bg-background shadow-sm">
                  <div v-if="selectedCreateWalletObj" class="flex items-center gap-2">
                      <component v-if="getIconComponent(selectedCreateWalletObj.icon)" :is="getIconComponent(selectedCreateWalletObj.icon)" class="h-4 w-4 text-muted-foreground" />
                      <span v-else class="text-xs">{{ getEmoji(selectedCreateWalletObj.icon) || '💼' }}</span>
                      <span>{{ selectedCreateWalletObj.name }}</span>
                  </div>
                  <SelectValue v-else placeholder="Pilih Dompet" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem v-for="wallet in walletStore.wallets" :key="wallet.id" :value="String(wallet.id)">
                   <div class="flex items-center gap-2">
                      <component v-if="getIconComponent(wallet.icon)" :is="getIconComponent(wallet.icon)" class="h-4 w-4" />
                      <span v-else class="text-xs">{{ getEmoji(wallet.icon) || '💼' }}</span>
                      <span>{{ wallet.name }}</span>
                  </div>
                </SelectItem>
              </SelectContent>
            </Select>
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
            <Input v-model="createForm.description" placeholder="Optional" class="h-11 shadow-sm rounded-xl bg-background" />
          </div>
           <div class="grid gap-2">
            <Label>Jatuh Tempo</Label>
            <Input type="date" :model-value="createForm.due_date || ''" @update:model-value="v => createForm.due_date = v as string" class="h-11 shadow-sm rounded-xl bg-background block w-full" />
          </div>
        </div>
        <DialogFooter class="gap-2">
          <Button variant="outline" @click="isCreateOpen = false" class="rounded-xl h-10 px-6">Batal</Button>
          <Button @click="handleCreate" :class="createForm.type === 'debt' ? 'bg-red-600 hover:bg-red-700' : 'bg-gradient-to-r from-emerald-600 to-teal-500 hover:from-emerald-500 hover:to-teal-400 font-bold'" class="rounded-xl h-10 px-6 text-white shadow-md">
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
             <div class="text-xl font-bold text-foreground">{{ selectedDebt ? formatCurrency(selectedDebt.remaining) : 0 }}</div>
          </div>
          <div class="grid gap-2">
            <Label>Nominal Pembayaran</Label>
            <Input type="text" inputmode="decimal" placeholder="Rp 0" v-model="payAmountDisplay" @blur="onPayBlur" class="h-11 shadow-sm rounded-xl bg-background" />
          </div>
          <div class="grid gap-2">
            <Label>Dompet Sumber/Tujuan</Label>
             <Select v-model="payWalletIdProxy">
              <SelectTrigger class="w-full h-11 rounded-xl bg-background shadow-sm">
                  <div v-if="selectedPayWalletObj" class="flex items-center gap-2">
                      <component v-if="getIconComponent(selectedPayWalletObj.icon)" :is="getIconComponent(selectedPayWalletObj.icon)" class="h-4 w-4 text-muted-foreground" />
                      <span v-else class="text-xs">{{ getEmoji(selectedPayWalletObj.icon) || '💼' }}</span>
                      <span>{{ selectedPayWalletObj.name }}</span>
                  </div>
                  <SelectValue v-else placeholder="Pilih Dompet" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem v-for="wallet in walletStore.wallets" :key="wallet.id" :value="String(wallet.id)">
                   <div class="flex items-center gap-2">
                      <component v-if="getIconComponent(wallet.icon)" :is="getIconComponent(wallet.icon)" class="h-4 w-4" />
                      <span v-else class="text-xs">{{ getEmoji(wallet.icon) || '💼' }}</span>
                      <span>{{ wallet.name }}</span>
                  </div>
                </SelectItem>
              </SelectContent>
            </Select>
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
            <Input v-model="payForm.note" placeholder="Contoh: Cicilan ke-1" class="h-11 shadow-sm rounded-xl bg-background" />
          </div>
        </div>
        <DialogFooter class="gap-2">
          <Button variant="outline" @click="isPayOpen = false" class="rounded-xl h-10 px-6">Batal</Button>
          <Button @click="handlePay" :class="selectedDebt?.type === 'debt' ? 'bg-red-600 hover:bg-red-700' : 'bg-gradient-to-r from-emerald-600 to-teal-500 hover:from-emerald-500 hover:to-teal-400 font-bold'" class="rounded-xl h-10 px-6 text-white shadow-md">Proses</Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>

    <Detail 
        v-model:open="isDetailOpen" 
        :debt="selectedDebt"
    />
  </div>
</template>
