<script setup lang="ts">
import { ref, onMounted, computed } from "vue";
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

const props = defineProps<{
  open: boolean;
}>();

const emit = defineEmits<{
  (e: "update:open", value: boolean): void;
  (e: "save", data: any): void;
}>();

const walletStore = useWalletStore();
const categoryStore = useCategoryStore();
const transactionStore = useTransactionStore();

const activeTab = ref<"expense" | "income" | "transfer">("expense");
const date = ref(new Date().toISOString().slice(0, 10));
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


const handleSave = async () => {
    if (!amount.value) {
         alert("Mohon isi nominal");
         return;
    }

    if (activeTab.value === 'transfer') {
        if (!selectedWallet.value || !toWallet.value) {
            alert("Mohon pilih dompet asal dan tujuan");
            return;
        }
        if (selectedWallet.value === toWallet.value) {
            alert("Dompet asal dan tujuan tidak boleh sama");
            return;
        }
    } else {
        if (!selectedWallet.value || !selectedCategory.value) {
            alert("Mohon lengkapi data");
            return;
        }
    }

    try {
        // Construct date with current time
        const now = new Date();
        const [year = now.getFullYear(), month = now.getMonth() + 1, day = now.getDate()] = date.value.split('-').map(Number);
        const finalDate = new Date(year, month - 1, day, now.getHours(), now.getMinutes(), now.getSeconds());

        if (activeTab.value === 'transfer') {
             await transactionStore.transfer({
                date: finalDate.toISOString(),
                amount: Number(amount.value),
                from_wallet_id: Number(selectedWallet.value),
                to_wallet_id: Number(toWallet.value),
                description: description.value || "Transfer Antar Dompet"
            });
        } else {
            await transactionStore.createTransaction({
                type: activeTab.value,
                date: finalDate.toISOString(),
                amount: Number(amount.value),
                category_id: Number(selectedCategory.value),
                wallet_id: Number(selectedWallet.value),
                description: description.value
            });
        }
        
        // Reset form (keep date as today)
        amount.value = "";
        description.value = "";
        
        emit("save", {}); 
        emit("update:open", false);
    } catch (error) {
        alert("Failed to create transaction");
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
    <DialogContent class="max-w-md bg-card text-foreground">
      <DialogHeader>
        <DialogTitle>Tambah Transaksi</DialogTitle>
        <DialogDescription>Catat pengeluaran atau income baru.</DialogDescription>
      </DialogHeader>

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
                    <Input type="date" v-model="date" class="block w-full bg-background" />
                </div>
            </div>

            <div class="space-y-2">
                <Label>Nominal (Rp)</Label>
                <Input type="text" placeholder="Rp 0" v-model="formattedAmount" class="bg-background" />
            </div>

            <div class="space-y-2">
                <Label>{{ activeTab === 'transfer' ? 'Dari Dompet' : 'Dompet' }}</Label>
                <Select v-model="selectedWallet">
                    <SelectTrigger class="w-full bg-background">
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
            </div>

            <div v-if="activeTab === 'transfer'" class="space-y-2">
                <Label>Ke Dompet</Label>
                <Select v-model="toWallet">
                    <SelectTrigger class="w-full bg-background">
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
            </div>

            <div v-if="activeTab !== 'transfer'" class="space-y-2">
                <Label>Kategori</Label>
                 <Select v-model="selectedCategory">
                    <SelectTrigger class="w-full bg-background">
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
            </div>

            <div class="space-y-2">
                <Label>Deskripsi (Opsional)</Label>
                <Input placeholder="Misal: Makan siang, Gaji bulanan" v-model="description" class="bg-background" />
            </div>
        </div>
      </Tabs>

      <DialogFooter class="flex gap-2 justify-end mt-4">
        <Button variant="outline" @click="emit('update:open', false)">Batal</Button>
        <Button @click="handleSave" class="bg-foreground text-background hover:bg-foreground/90">Simpan</Button>
      </DialogFooter>
    </DialogContent>
  </Dialog>
</template>
