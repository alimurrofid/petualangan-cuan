<script setup lang="ts">
import { ref, onMounted, computed, watch } from "vue";
import { format, parseISO } from "date-fns";
import { useWalletStore } from "@/stores/wallet";
import { useCategoryStore } from "@/stores/category";
import { useTransactionStore } from "@/stores/transaction";
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogFooter, DialogDescription } from "@/components/ui/dialog";
import { Tabs, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Label } from "@/components/ui/label";
import { Input } from "@/components/ui/input";
import { Button } from "@/components/ui/button";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import * as LucideIcons from "lucide-vue-next";
import { Wallet } from "lucide-vue-next";
import { useSwal } from "@/composables/useSwal";

const props = defineProps<{
  open: boolean;
  transactionToEdit?: any | null;
}>();

const emit = defineEmits<{
  (e: "update:open", value: boolean): void;
  (e: "save", data: any): void;
}>();

const walletStore = useWalletStore();
const categoryStore = useCategoryStore();
const transactionStore = useTransactionStore();
const swal = useSwal();

const activeTab = ref<"expense" | "income" | "transfer">("expense");
const date = ref(format(new Date(), "yyyy-MM-dd"));
const amount = ref("");
const selectedWallet = ref("");
const toWallet = ref(""); // New field for transfer
const selectedCategory = ref("");
const description = ref("");

// Initial fetch
onMounted(() => {
    walletStore.fetchWallets();
    categoryStore.fetchCategories();
});

// Watch for edit mode
watch(() => props.transactionToEdit, (newVal) => {
    if (newVal) {
        activeTab.value = newVal.type === 'transfer_in' || newVal.type === 'transfer_out' ? 'transfer' : newVal.type;
        // If transfer, we treat as simple Expense/Income for now as per plan constraints, 
        // OR better: if it's transfer_out, show as expense; transfer_in as income.
        // But if user wants to edit "Transfer", we need to know other leg.
        // Current plan: Edit as single transaction.
        if (newVal.type === 'transfer_out' || newVal.type === 'transfer_in') {
            activeTab.value = 'transfer';
            // Fetch related transaction to get the other wallet
            // Note: transactionStore needs fetchTransaction action.
            if (newVal.related_transaction_id) {
                transactionStore.fetchTransaction(newVal.related_transaction_id).then(related => {
                    if (related) {
                         if (newVal.type === 'transfer_out') {
                             toWallet.value = String(related.wallet_id);
                         } else {
                             toWallet.value = String(newVal.wallet_id); // Current is IN (to)
                             selectedWallet.value = String(related.wallet_id); // Related is OUT (from)
                         }
                    }
                });
            }
        }

        // Standard setup
        if (newVal.type === 'transfer_out') {
             selectedWallet.value = String(newVal.wallet_id);
        } else if (newVal.type === 'transfer_in') {
             if (activeTab.value === 'transfer') {
                  toWallet.value = String(newVal.wallet_id);
             } else {
                  selectedWallet.value = String(newVal.wallet_id);
             }
        } 

        // Fix date format
        date.value = format(parseISO(newVal.date), 'yyyy-MM-dd');
        amount.value = newVal.amount.toString();
        selectedWallet.value = String(newVal.wallet_id);
        selectedCategory.value = String(newVal.category_id);
        description.value = newVal.description;
    } else {
        // Reset defaults
        activeTab.value = "expense";
        date.value = format(new Date(), "yyyy-MM-dd");
        amount.value = "";
        selectedWallet.value = "";
        toWallet.value = "";
        selectedCategory.value = "";
        description.value = "";
    }
});

// Filter categories based on active tab
const filteredCategories = computed(() => {
    return categoryStore.categories.filter(c => c.type === activeTab.value);
});

// Helpers to get selected objects for display
const selectedWalletObj = computed(() => walletStore.wallets.find(w => String(w.id) === selectedWallet.value));
const toWalletObj = computed(() => walletStore.wallets.find(w => String(w.id) === toWallet.value));
const selectedCategoryObj = computed(() => categoryStore.categories.find(c => String(c.id) === selectedCategory.value));

// Utils for icon rendering
const getIconComponent = (name: string | undefined) => {
  if (!name) return Wallet;
  return (LucideIcons as any)[name] || null;
};

// Emoji map for rendering 
// (Duplicate from other files, in real app should be shared constant/composable)
const emojiCategories: Record<string, string> = {
  Em_MoneyBag: "💰", Em_DollarBill: "💵", Em_Card: "💳", Em_Bank: "🏦", Em_MoneyWing: "💸", Em_Coin: "🪙",
  Em_Pizza: "🍕", Em_Cart: "🛒", Em_Coffee: "☕", Em_Game: "🎮", Em_Airplane: "✈️", Em_Gift: "🎁",
  Em_Star: "⭐", Em_Fire: "🔥", Em_Lock: "🔒", Em_Check: "✅", Em_Idea: "💡"
};
const getEmoji = (name: string | undefined) => {
  if (!name) return null;
  if (emojiCategories[name]) return emojiCategories[name];
  if (/\p{Emoji}/u.test(name)) return name;
  return null;
};


const isSubmitting = ref(false);

const errors = ref({
  date: false,
  amount: false,
  wallet: false,
  toWallet: false,
  category: false,
});

const handleSave = async () => {
    isSubmitting.value = true;
    
    // Reset errors
    Object.keys(errors.value).forEach(k => (errors.value as any)[k] = false);

    // Validate fields
    errors.value.date = !date.value;
    errors.value.amount = !amount.value;

    if (activeTab.value === 'transfer') {
        errors.value.wallet = !selectedWallet.value;
        errors.value.toWallet = !toWallet.value;
    } else {
        errors.value.wallet = !selectedWallet.value;
        errors.value.category = !selectedCategory.value;
    }

    // Check for validation errors
    const hasError = Object.values(errors.value).some(v => v);
    if (hasError) {
         let msg = "Mohon lengkapi data berikut:";
         if (errors.value.date) msg += "<br>- Tanggal";
         if (errors.value.amount) msg += "<br>- Nominal";
         if (errors.value.wallet) msg += "<br>- " + (activeTab.value === 'transfer' ? 'Dompet Asal' : 'Dompet');
         if (errors.value.toWallet) msg += "<br>- Dompet Tujuan";
         if (errors.value.category) msg += "<br>- Kategori";

         await swal.fire({
            icon: 'error',
            title: 'Validasi Gagal',
            html: msg,
            confirmButtonColor: '#EF4444', 
         });
         setTimeout(() => { isSubmitting.value = false; }, 300);
         return;
    }

    // Logical validation for transfer
    if (activeTab.value === 'transfer' && selectedWallet.value === toWallet.value) {
         await swal.error("Gagal", "Dompet asal dan tujuan tidak boleh sama");
         setTimeout(() => { isSubmitting.value = false; }, 300);
         return;
    }

    try {
        // Construct date with current time
        const now = new Date();
        const [year = now.getFullYear(), month = now.getMonth() + 1, day = now.getDate()] = date.value.split('-').map(Number);
        const finalDate = new Date(year, month - 1, day, now.getHours(), now.getMinutes(), now.getSeconds());

        if (props.transactionToEdit) {
             let finalType: string = activeTab.value;
             let finalWalletId = Number(selectedWallet.value);

             if (activeTab.value === 'transfer') {
                  if (props.transactionToEdit.type === 'transfer_in') {
                      finalType = 'transfer_in';
                      // For transfer_in (Income), the wallet is the Destination (toWallet)
                      finalWalletId = Number(toWallet.value);
                  } else if (props.transactionToEdit.type === 'transfer_out') {
                      finalType = 'transfer_out';
                      // For transfer_out (Expense), the wallet is the Source (selectedWallet)
                      finalWalletId = Number(selectedWallet.value);
                  } else {
                      // Fallback: If editing a non-transfer but tab is transfer, assume transfer_out (Source)
                      finalType = 'transfer_out';
                  }
             }

             await transactionStore.updateTransaction(props.transactionToEdit.id, {
                wallet_id: finalWalletId,
                category_id: Number(selectedCategory.value),
                amount: Number(amount.value),
                type: finalType,
                description: description.value,
                date: format(finalDate, "yyyy-MM-dd'T'HH:mm:ssXXX"),
            });
            swal.toast({ icon: 'success', title: 'Transaksi berhasil diperbarui' });
        } else if (activeTab.value === 'transfer') {
             await transactionStore.transfer({
                date: format(finalDate, "yyyy-MM-dd'T'HH:mm:ssXXX"),
                amount: Number(amount.value),
                from_wallet_id: Number(selectedWallet.value),
                to_wallet_id: Number(toWallet.value),
                description: description.value || "Transfer Antar Dompet"
            });
            swal.success("Berhasil", "Transfer berhasil dilakukan");
        } else {
            await transactionStore.createTransaction({
                type: activeTab.value,
                date: format(finalDate, "yyyy-MM-dd'T'HH:mm:ssXXX"),
                amount: Number(amount.value),
                category_id: Number(selectedCategory.value),
                wallet_id: Number(selectedWallet.value),
                description: description.value
            });
            swal.toast({ icon: 'success', title: 'Transaksi berhasil disimpan' });
        }
        
        // Reset form (keep date as today)
        amount.value = "";
        description.value = "";
        
        emit("save", {}); 
        emit("update:open", false);
    } catch (error) {
        swal.error("Gagal", props.transactionToEdit ? "Gagal memperbarui transaksi" : "Gagal melakukan transaksi");
    } finally {
        isSubmitting.value = false;
    }
};

// Currency Formatting Logic
const formattedAmount = computed({
  get: () => {
    if (!amount.value) return "";
    return new Intl.NumberFormat("id-ID", { style: "currency", currency: "IDR", maximumFractionDigits: 0 }).format(Number(amount.value));
  },
  set: (val: string) => {
    const numericValue = Number(val.replace(/[^0-9]/g, ""));
    amount.value = numericValue.toString();
  }
});
</script>

<template>
  <Dialog :open="open" @update:open="emit('update:open', $event)">
    <DialogContent class="max-w-md bg-card text-foreground" @interact-outside="swal.handleSwalInteractOutside">
      <DialogHeader>
        <DialogTitle>{{ transactionToEdit ? 'Edit Transaksi' : 'Tambah Transaksi' }}</DialogTitle>
        <DialogDescription>Catat pengeluaran atau income baru.</DialogDescription>
      </DialogHeader>
      
      <div v-if="transactionToEdit && (transactionToEdit.type === 'transfer_in' || transactionToEdit.type === 'transfer_out')" class="bg-blue-50 p-3 rounded-lg text-xs text-blue-700 mb-4 border border-blue-200">
        <strong>Info:</strong> Transaksi ini terhubung dengan transaksi transfer lainnya. Perubahan nominal, tanggal, atau deskripsi akan otomatis diterapkan pada pasangannya.
      </div>

      <Tabs v-model="activeTab" class="w-full">
        <TabsList class="grid w-full grid-cols-3 mb-4 h-auto p-1 bg-muted/60 rounded-xl">
          <TabsTrigger value="expense" class="rounded-lg py-2 data-[state=active]:bg-red-500 data-[state=active]:text-white">Pengeluaran</TabsTrigger>
          <TabsTrigger value="income" class="rounded-lg py-2 data-[state=active]:bg-emerald-500 data-[state=active]:text-white">Pemasukan</TabsTrigger>
          <TabsTrigger value="transfer" class="rounded-lg py-2 data-[state=active]:bg-blue-500 data-[state=active]:text-white">Transfer</TabsTrigger>
        </TabsList>

        <div class="space-y-4 py-2">
            <div class="space-y-2">
                <Label>Tanggal</Label>
                <div class="relative">
                    <Input type="date" v-model="date" :class="['block w-full bg-background', errors.date ? 'border-red-500 ring-1 ring-red-500' : '']" :disabled="isSubmitting" />
                </div>
                <span v-if="errors.date" class="text-xs text-red-500 font-medium">Tanggal wajib diisi</span>
            </div>

            <div class="space-y-2">
                <Label>Nominal (Rp)</Label>
                <Input type="text" placeholder="Rp 0" v-model="formattedAmount" :class="['bg-background', errors.amount ? 'border-red-500 ring-1 ring-red-500' : '']" :disabled="isSubmitting" />
                <span v-if="errors.amount" class="text-xs text-red-500 font-medium">Nominal wajib diisi</span>
            </div>

            <div class="space-y-2">
                <Label>{{ activeTab === 'transfer' ? 'Dari Dompet' : 'Dompet' }}</Label>
                <Select v-model="selectedWallet" :disabled="isSubmitting">
                    <SelectTrigger :class="['w-full bg-background', errors.wallet ? 'border-red-500 ring-1 ring-red-500' : '']">
                         <div v-if="selectedWalletObj" class="flex items-center gap-2">
                             <component v-if="getIconComponent(selectedWalletObj.icon)" :is="getIconComponent(selectedWalletObj.icon)" class="h-4 w-4" />
                             <span v-else class="text-xs">{{ getEmoji(selectedWalletObj.icon) || '💼' }}</span>
                             <span>{{ selectedWalletObj.name }}</span>
                         </div>
                        <SelectValue v-else placeholder="Pilih Dompet" />
                    </SelectTrigger>
                    <SelectContent>
                        <SelectItem v-for="w in walletStore.wallets" :key="w.id" :value="String(w.id)">
                            <div class="flex items-center gap-2">
                                <component v-if="getIconComponent(w.icon)" :is="getIconComponent(w.icon)" class="h-4 w-4" />
                                <span v-else class="text-xs">{{ getEmoji(w.icon) || '💼' }}</span>
                                <span>{{ w.name }}</span>
                            </div>
                        </SelectItem>
                    </SelectContent>
                </Select>
                <span v-if="errors.wallet" class="text-xs text-red-500 font-medium">{{ activeTab === 'transfer' ? 'Dompet asal' : 'Dompet' }} wajib dipilih</span>
            </div>

            <div v-if="activeTab === 'transfer'" class="space-y-2">
                <Label>Ke Dompet</Label>
                <Select v-model="toWallet" :disabled="isSubmitting">
                    <SelectTrigger :class="['w-full bg-background', errors.toWallet ? 'border-red-500 ring-1 ring-red-500' : '']">
                         <div v-if="toWalletObj" class="flex items-center gap-2">
                             <component v-if="getIconComponent(toWalletObj.icon)" :is="getIconComponent(toWalletObj.icon)" class="h-4 w-4" />
                             <span v-else class="text-xs">{{ getEmoji(toWalletObj.icon) || '💼' }}</span>
                             <span>{{ toWalletObj.name }}</span>
                         </div>
                        <SelectValue v-else placeholder="Pilih Dompet Tujuan" />
                    </SelectTrigger>
                    <SelectContent>
                        <SelectItem v-for="w in walletStore.wallets" :key="w.id" :value="String(w.id)">
                            <div class="flex items-center gap-2">
                                <component v-if="getIconComponent(w.icon)" :is="getIconComponent(w.icon)" class="h-4 w-4" />
                                <span v-else class="text-xs">{{ getEmoji(w.icon) || '💼' }}</span>
                                <span>{{ w.name }}</span>
                            </div>
                        </SelectItem>
                    </SelectContent>
                </Select>
                <span v-if="errors.toWallet" class="text-xs text-red-500 font-medium">Dompet tujuan wajib dipilih</span>
            </div>

            <div v-if="activeTab !== 'transfer'" class="space-y-2">
                <Label>Kategori</Label>
                 <Select v-model="selectedCategory" :disabled="isSubmitting">
                    <SelectTrigger :class="['w-full bg-background', errors.category ? 'border-red-500 ring-1 ring-red-500' : '']">
                        <div v-if="selectedCategoryObj" class="flex items-center gap-2">
                             <component v-if="getIconComponent(selectedCategoryObj.icon)" :is="getIconComponent(selectedCategoryObj.icon)" class="h-4 w-4" />
                             <span v-else>{{ getEmoji(selectedCategoryObj.icon) || selectedCategoryObj.icon }}</span>
                             <span>{{ selectedCategoryObj.name }}</span>
                        </div>
                        <SelectValue v-else placeholder="Pilih Kategori" />
                    </SelectTrigger>
                    <SelectContent>
                         <SelectItem v-for="c in filteredCategories" :key="c.id" :value="String(c.id)">
                            <div class="flex items-center gap-2">
                                 <component v-if="getIconComponent(c.icon)" :is="getIconComponent(c.icon)" class="h-4 w-4" />
                                 <span v-else>{{ getEmoji(c.icon) || c.icon }}</span>
                                 <span>{{ c.name }}</span>
                            </div>
                        </SelectItem>
                    </SelectContent>
                </Select>
                <span v-if="errors.category" class="text-xs text-red-500 font-medium">Kategori wajib dipilih</span>
            </div>

            <div class="space-y-2">
                <Label>Deskripsi (Opsional)</Label>
                <Input placeholder="Misal: Makan siang, Gaji bulanan" v-model="description" class="bg-background" :disabled="isSubmitting" />
            </div>
        </div>
      </Tabs>

      <DialogFooter class="flex gap-2 justify-end mt-4">
        <Button variant="outline" @click="emit('update:open', false)" :disabled="isSubmitting">Batal</Button>
        <Button @click="handleSave" class="bg-gradient-to-r from-emerald-600 to-teal-500 text-white hover:from-emerald-500 hover:to-teal-400" :disabled="isSubmitting">
            {{ transactionToEdit ? 'Simpan Perubahan' : 'Simpan' }}
        </Button>
      </DialogFooter>
    </DialogContent>
  </Dialog>
</template>
