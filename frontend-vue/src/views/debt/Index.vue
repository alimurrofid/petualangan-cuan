<script setup lang="ts">
import { ref, onMounted, computed, reactive } from "vue";
import { useDebtStore, type Debt, type CreateDebtInput, type UpdateDebtInput, type PayDebtInput } from "@/stores/debt";
import { useWalletStore } from "@/stores/wallet";
import { Card, CardHeader, CardTitle, CardContent } from "@/components/ui/card";
import { Tabs, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Button } from "@/components/ui/button";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogDescription, DialogFooter } from "@/components/ui/dialog";
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from "@/components/ui/tooltip";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Plus, ArrowUpRight, ArrowDownLeft, Filter, Pencil, Trash2, HandCoins, CircleFadingArrowUp, Eye } from "lucide-vue-next";
import { format } from "date-fns";
import { id } from "date-fns/locale";
import { getEmoji, getIconComponent } from "@/lib/icons";
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

const formatCurrency = (amount: number) => {
  return new Intl.NumberFormat("id-ID", {
    style: "currency",
    currency: "IDR",
    minimumFractionDigits: 0,
  }).format(amount);
};

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

// Currency Formatter for Inputs
const useCurrencyInput = (modelValue: { amount: number }) => {
  return computed({
    get: () => {
      if (!modelValue.amount) return "";
      return new Intl.NumberFormat("id-ID", { style: "currency", currency: "IDR", maximumFractionDigits: 0 }).format(modelValue.amount);
    },
    set: (val: string) => {
      const numericValue = Number(val.replace(/[^0-9]/g, ""));
      modelValue.amount = numericValue;
    }
  });
};

const formattedCreateAmount = useCurrencyInput(createForm);
const formattedPayAmount = useCurrencyInput(payForm);

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
  <div class="p-6 space-y-8 text-foreground min-h-screen bg-background">
    <div class="flex flex-col gap-2">
      <h2 class="text-3xl font-bold tracking-tight">Utang & Piutang</h2>
      <p class="text-sm text-muted-foreground">Kelola catatan utang dan piutang Anda.</p>
    </div>

    <!-- Summary Cards -->
    <div class="grid gap-4 md:grid-cols-2">
      <Card class="bg-card border-border shadow-sm">
        <CardHeader class="flex flex-row items-center justify-between space-y-0 pb-2">
          <CardTitle class="text-sm font-medium">Total Utang Saya</CardTitle>
          <ArrowUpRight class="h-4 w-4 text-red-500" />
        </CardHeader>
        <CardContent>
          <div class="text-2xl font-bold text-red-500">{{ formatCurrency(totalDebt) }}</div>
          <p class="text-xs text-muted-foreground mt-1">Harus dibayar</p>
        </CardContent>
      </Card>
      <Card class="bg-card border-border shadow-sm">
        <CardHeader class="flex flex-row items-center justify-between space-y-0 pb-2">
          <CardTitle class="text-sm font-medium">Total Piutang Saya</CardTitle>
          <ArrowDownLeft class="h-4 w-4 text-emerald-500" />
        </CardHeader>
        <CardContent>
          <div class="text-2xl font-bold text-emerald-500">{{ formatCurrency(totalReceivable) }}</div>
          <p class="text-xs text-muted-foreground mt-1">Akan diterima</p>
        </CardContent>
      </Card>
    </div>

    <!-- Main Content -->
    <Card class="bg-card border-border shadow-sm">
      <CardContent class="p-6">
          <div class="flex flex-col sm:flex-row items-stretch sm:items-center justify-between gap-4 mb-6">
             <div class="flex items-center gap-2 w-full sm:w-auto">
                <Select v-model="filterType">
                    <SelectTrigger class="w-full sm:w-[180px]">
                        <SelectValue placeholder="Filter Tipe" />
                    </SelectTrigger>
                    <SelectContent>
                        <SelectItem value="all">Semua</SelectItem>
                        <SelectItem value="debt">Utang (Payable)</SelectItem>
                        <SelectItem value="receivable">Piutang (Receivable)</SelectItem>
                    </SelectContent>
                </Select>
             </div>
            <Button @click="openCreateDialog()" class="gap-2 bg-emerald-600 hover:bg-emerald-700 text-white w-full sm:w-auto">
              <Plus class="h-4 w-4" />
              Tambah Baru
            </Button>
          </div>

            <div class="rounded-md border">
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>Tipe</TableHead>
                    <TableHead>Nama</TableHead>
                    <TableHead>Keterangan</TableHead>
                    <TableHead>Dompet</TableHead>
                    <TableHead>Jatuh Tempo</TableHead>
                    <TableHead class="text-right">Total</TableHead>
                    <TableHead class="text-right">Sisa</TableHead>
                    <TableHead class="text-center">Status</TableHead>
                    <TableHead class="text-right">Aksi</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                   <TableRow v-if="filteredItems.length === 0">
                      <TableCell colspan="9" class="text-center py-8 text-muted-foreground">Tidak ada data</TableCell>
                   </TableRow>
                   <TableRow v-for="item in filteredItems" :key="item.id">
                    <TableCell>
                        <span v-if="item.type === 'debt'" class="text-red-500 font-medium text-xs border border-red-200 bg-red-50 px-2 py-1 rounded">Utang</span>
                        <span v-else class="text-emerald-500 font-medium text-xs border border-emerald-200 bg-emerald-50 px-2 py-1 rounded">Piutang</span>
                    </TableCell>
                    <TableCell class="font-medium">{{ item.name }}</TableCell>
                    <TableCell class="text-muted-foreground text-sm truncate max-w-[200px]">{{ item.description }}</TableCell>
                    <TableCell>
                        <div v-if="item.wallet" class="flex items-center gap-2 text-sm">
                             <component v-if="getIconComponent(item.wallet.icon)" :is="getIconComponent(item.wallet.icon)" class="h-4 w-4 text-muted-foreground" />
                             <span v-else>{{ getEmoji(item.wallet.icon) || '💼' }}</span>
                             <span>{{ item.wallet.name }}</span>
                        </div>
                        <span v-else class="text-xs text-muted-foreground">-</span>
                    </TableCell>
                    <TableCell>
                      {{ item.due_date ? format(new Date(item.due_date), "d MMM yyyy", { locale: id }) : '-' }}
                    </TableCell>
                    <TableCell class="text-right">{{ formatCurrency(item.amount) }}</TableCell>
                    <TableCell class="text-right font-bold" :class="item.type === 'debt' ? 'text-red-500' : 'text-emerald-500'">
                        {{ formatCurrency(item.remaining) }}
                    </TableCell>
                    <TableCell class="text-center">
                      <span v-if="item.is_paid" class="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-green-100 text-green-800">
                        Lunas
                      </span>
                      <span v-else class="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-yellow-100 text-yellow-800">
                        Belum Lunas
                      </span>
                    </TableCell>
                    <TableCell class="text-right">
                       <div class="flex items-center justify-end gap-1">
                           <TooltipProvider>
                               <Tooltip>
                                 <TooltipTrigger as-child>
                                   <Button @click="openDetailDialog(item)" size="sm" variant="ghost" class="h-8 w-8 p-0 hover:bg-slate-100">
                                     <Eye class="h-4 w-4 text-slate-500" />
                                   </Button>
                                 </TooltipTrigger>
                                 <TooltipContent>
                                   <p>Lihat Detail & Riwayat</p>
                                 </TooltipContent>
                               </Tooltip>
                           </TooltipProvider>

                           <TooltipProvider>
                               <Tooltip>
                                 <TooltipTrigger as-child>
                                   <Button @click="openEditDialog(item)" size="sm" variant="ghost" class="h-8 w-8 p-0 hover:bg-blue-50">
                                     <Pencil class="h-4 w-4 text-blue-500" />
                                   </Button>
                                 </TooltipTrigger>
                                 <TooltipContent>
                                   <p>Edit Data</p>
                                 </TooltipContent>
                               </Tooltip>
                           </TooltipProvider>

                           <TooltipProvider v-if="!item.is_paid">
                               <Tooltip>
                                 <TooltipTrigger as-child>
                                   <Button @click="openPayDialog(item)" size="sm" variant="ghost" class="h-8 w-8 p-0 hover:bg-emerald-50">
                                     <CircleFadingArrowUp v-if="item.type === 'debt'" class="h-4 w-4 text-emerald-600" />
                                     <HandCoins v-else class="h-4 w-4 text-emerald-600" />
                                   </Button>
                                 </TooltipTrigger>
                                 <TooltipContent>
                                   <p>{{ item.type === 'debt' ? 'Bayar Utang' : 'Terima Piutang' }}</p>
                                 </TooltipContent>
                               </Tooltip>
                           </TooltipProvider>

                           <TooltipProvider>
                               <Tooltip>
                                 <TooltipTrigger as-child>
                                   <Button @click="handleDelete(item.id)" size="sm" variant="ghost" class="h-8 w-8 p-0 hover:bg-red-50">
                                     <Trash2 class="h-4 w-4 text-red-500" />
                                   </Button>
                                 </TooltipTrigger>
                                 <TooltipContent>
                                   <p>Hapus Data</p>
                                 </TooltipContent>
                               </Tooltip>
                           </TooltipProvider>
                       </div>
                    </TableCell>
                  </TableRow>
                </TableBody>
              </Table>
            </div>
      </CardContent>
    </Card>

    <!-- Create/Edit Dialog -->
    <Dialog :open="isCreateOpen" @update:open="isCreateOpen = $event">
      <DialogContent class="sm:max-w-[500px]">
        <DialogHeader>
          <DialogTitle>{{ isEditMode ? 'Edit Utang/Piutang' : 'Catat Utang/Piutang Baru' }}</DialogTitle>
          <DialogDescription>
             {{ isEditMode ? 'Perbarui detail data.' : 'Pilih tipe transaksi dan isi detailnya.' }}
          </DialogDescription>
        </DialogHeader>
        
        <Tabs v-model="activeTab" class="w-full mt-2" @update:modelValue="onTabChange">
          <TabsList class="grid w-full grid-cols-2 mb-4">
            <TabsTrigger value="debt" :disabled="isEditMode" class="data-[state=active]:bg-red-500 data-[state=active]:text-white disabled:opacity-50">Utang (Saya Berutang)</TabsTrigger>
            <TabsTrigger value="receivable" :disabled="isEditMode" class="data-[state=active]:bg-emerald-500 data-[state=active]:text-white disabled:opacity-50">Piutang (Orang Berutang)</TabsTrigger>
          </TabsList>
        </Tabs>

        <div class="grid gap-4 py-2">
          <div class="grid gap-2">
            <Label>Nama {{ createForm.type === 'debt' ? 'Pemberi Utang' : 'Peminjam' }}</Label>
            <Input v-model="createForm.name" placeholder="Contoh: Budi" />
          </div>
          <div class="grid gap-2">
            <Label>Nominal (Rp)</Label>
            <Input type="text" inputmode="numeric" pattern="[0-9]*" placeholder="Rp 0" v-model="formattedCreateAmount" />
            <p v-if="isEditMode" class="text-xs text-muted-foreground text-yellow-600">
                Perhatian: Mengubah nominal akan menyesuaikan ulang saldo dompet.
            </p>
          </div>
          <div class="grid gap-2">
            <Label>Dompet Terkait</Label>
            <Select v-model="createWalletIdProxy">
              <SelectTrigger class="w-full bg-background">
                  <div v-if="selectedCreateWalletObj" class="flex items-center gap-2">
                      <component v-if="getIconComponent(selectedCreateWalletObj.icon)" :is="getIconComponent(selectedCreateWalletObj.icon)" class="h-4 w-4" />
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
            <p v-if="!isEditMode" class="text-xs text-muted-foreground mt-1">
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
            <Input v-model="createForm.description" placeholder="Optional" />
          </div>
           <div class="grid gap-2">
            <Label>Jatuh Tempo</Label>
            <Input type="date" :model-value="createForm.due_date || ''" @update:model-value="v => createForm.due_date = v as string" />
          </div>
        </div>
        <DialogFooter>
          <Button variant="outline" @click="isCreateOpen = false">Batal</Button>
          <Button @click="handleCreate" :class="createForm.type === 'debt' ? 'bg-red-600 hover:bg-red-700' : 'bg-emerald-600 hover:bg-emerald-700'">
            {{ isEditMode ? 'Simpan Perubahan' : 'Simpan' }}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>

    <!-- Pay Dialog -->
    <Dialog :open="isPayOpen" @update:open="isPayOpen = $event">
      <DialogContent class="sm:max-w-[425px]">
        <DialogHeader>
          <DialogTitle>{{ selectedDebt?.type === 'debt' ? 'Bayar Utang' : 'Terima Pembayaran' }}</DialogTitle>
          <DialogDescription>
             Catat cicilan atau pelunasan. Saldo dompet akan otomatis disesuaikan.
          </DialogDescription>
        </DialogHeader>
        <div class="grid gap-4 py-4">
          <div class="grid gap-2">
             <Label class="tex-sm text-muted-foreground">Sisa Kewajiban</Label>
             <div class="text-xl font-bold">{{ selectedDebt ? formatCurrency(selectedDebt.remaining) : 0 }}</div>
          </div>
          <div class="grid gap-2">
            <Label>Nominal Pembayaran (Rp)</Label>
            <Input type="text" inputmode="numeric" pattern="[0-9]*" placeholder="Rp 0" v-model="formattedPayAmount" />
          </div>
          <div class="grid gap-2">
            <Label>Dompet Sumber/Tujuan</Label>
             <Select v-model="payWalletIdProxy">
              <SelectTrigger class="w-full bg-background">
                  <div v-if="selectedPayWalletObj" class="flex items-center gap-2">
                      <component v-if="getIconComponent(selectedPayWalletObj.icon)" :is="getIconComponent(selectedPayWalletObj.icon)" class="h-4 w-4" />
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
            <p class="text-xs text-muted-foreground mt-1">
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
            <Input v-model="payForm.note" placeholder="Contoh: Cicilan ke-1" />
          </div>
        </div>
        <DialogFooter>
          <Button variant="outline" @click="isPayOpen = false">Batal</Button>
          <Button @click="handlePay" :class="selectedDebt?.type === 'debt' ? 'bg-red-600 hover:bg-red-700' : 'bg-emerald-600 hover:bg-emerald-700'">Proses</Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>

    <Detail 
        v-model:open="isDetailOpen" 
        :debt="selectedDebt"
    />
  </div>
</template>
