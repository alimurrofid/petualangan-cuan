<script setup lang="ts">
import { ref, onMounted } from "vue";

// UI Components
import { Card, CardHeader, CardTitle, CardContent } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogFooter, DialogDescription } from "@/components/ui/dialog";

// Icons
import * as LucideIcons from "lucide-vue-next";
import { Plus, Pencil, Trash2, Save, Wallet } from "lucide-vue-next";

interface WalletItem {
  id: number;
  name: string;
  type: string;
  icon: string;
  isEmoji: boolean;
  balance: number;
}

const wallets = ref<WalletItem[]>([]);

// State untuk Popup
const isDialogOpen = ref(false);
const isIconPickerOpen = ref(false);
const isEditMode = ref(false);

const form = ref<WalletItem>({
  id: 0,
  name: "",
  type: "Cash",
  icon: "",
  isEmoji: false,
  balance: 0,
});

const iconOptions = [
  { name: "Wallet", icon: LucideIcons.Wallet, label: "Umum" },
  { name: "Landmark", icon: LucideIcons.Landmark, label: "Bank" },
  { name: "CreditCard", icon: LucideIcons.CreditCard, label: "Kartu" },
  { name: "Banknote", icon: LucideIcons.Banknote, label: "Tunai" },
  { name: "PiggyBank", icon: LucideIcons.PiggyBank, label: "Tabungan" },
  { name: "Coins", icon: LucideIcons.Coins, label: "Investasi" },
  { name: "Zap", icon: LucideIcons.Zap, label: "Tagihan" },
  { name: "ShoppingBag", icon: LucideIcons.ShoppingBag, label: "Belanja" },
];

const emojiList = ["💰", "💵", "💳", "🏦", "💸", "🪙", "💹", "💎", "🏠", "🚗"];

const loadWallets = () => {
  const saved = localStorage.getItem("mock_wallets");
  if (saved) {
    wallets.value = JSON.parse(saved);
  } else {
    const initial: WalletItem[] = [{ id: 1, name: "BCA Utama", type: "Bank", icon: "Landmark", isEmoji: false, balance: 0 }];
    wallets.value = initial;
    localStorage.setItem("mock_wallets", JSON.stringify(initial));
  }
};

onMounted(loadWallets);

const openAdd = () => {
  isEditMode.value = false;
  form.value = { id: Date.now(), name: "", type: "Cash", icon: "", isEmoji: false, balance: 0 };
  isDialogOpen.value = true;
};

const openEdit = (wallet: WalletItem) => {
  isEditMode.value = true;
  form.value = { ...wallet };
  isDialogOpen.value = true;
};

const selectIcon = (name: string, isEmoji: boolean) => {
  form.value.icon = name;
  form.value.isEmoji = isEmoji;
  isIconPickerOpen.value = false;
};

const handleSave = () => {
  if (!form.value.name || !form.value.icon) return alert("Lengkapi data dompet");
  if (isEditMode.value) {
    wallets.value = wallets.value.map((w) => (w.id === form.value.id ? { ...form.value } : w));
  } else {
    wallets.value.push({ ...form.value });
  }
  localStorage.setItem("mock_wallets", JSON.stringify(wallets.value));
  isDialogOpen.value = false;
};

const handleDelete = () => {
  if (confirm("Hapus dompet ini?")) {
    wallets.value = wallets.value.filter((w) => w.id !== form.value.id);
    localStorage.setItem("mock_wallets", JSON.stringify(wallets.value));
    isDialogOpen.value = false;
  }
};

const getIconComponent = (name: string) => {
  return (LucideIcons as any)[name] || Wallet;
};

const formatCurrency = (value: number) => {
  return new Intl.NumberFormat("id-ID", { style: "currency", currency: "IDR", minimumFractionDigits: 0 }).format(value);
};
</script>

<template>
  <div class="p-6 space-y-6 text-foreground">
    <div class="flex items-center justify-between">
      <div>
        <h2 class="text-2xl font-bold tracking-tight">Dompet Saya</h2>
        <p class="text-sm text-muted-foreground">Kelola sumber dana dan rekening Anda.</p>
      </div>
      <Button @click="openAdd" class="bg-emerald-600 hover:bg-emerald-700 text-white shadow-sm px-6"> <Plus class="w-4 h-4 mr-2" /> Tambah Dompet </Button>
    </div>

    <div class="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
      <Card v-for="item in wallets" :key="item.id" @click="openEdit(item)" class="cursor-pointer transition-all hover:ring-2 hover:ring-emerald-500 hover:shadow-lg active:scale-95 bg-card border-border overflow-hidden rounded-2xl group">
        <CardHeader class="flex flex-row items-center justify-between pb-2 space-y-0">
          <CardTitle class="text-sm font-bold">{{ item.name }}</CardTitle>
          <div class="h-10 w-10 rounded-xl bg-emerald-100 dark:bg-emerald-900/30 flex items-center justify-center text-emerald-600 shadow-sm group-hover:scale-110 transition-transform">
            <span v-if="item.isEmoji" class="text-xl">{{ item.icon }}</span>
            <component v-else :is="getIconComponent(item.icon)" class="h-5 w-5" />
          </div>
        </CardHeader>
        <CardContent>
          <div class="text-2xl font-bold">{{ formatCurrency(item.balance) }}</div>
          <p class="text-[10px] uppercase font-bold tracking-widest text-muted-foreground mt-1">{{ item.type }}</p>
        </CardContent>
      </Card>
    </div>

    <Dialog v-model:open="isDialogOpen">
      <DialogContent class="max-w-md bg-card p-0 overflow-hidden border-border shadow-2xl">
        <DialogHeader class="p-6 border-b">
          <DialogTitle>{{ isEditMode ? "Edit Dompet" : "Tambah Dompet" }}</DialogTitle>
          <DialogDescription>Masukkan informasi detail dompet Anda.</DialogDescription>
        </DialogHeader>

        <div class="p-6 space-y-6 text-foreground">
          <div class="grid gap-2">
            <Label class="text-sm font-semibold opacity-70">Nama Dompet</Label>
            <Input v-model="form.name" placeholder="Misal: BCA Utama, Cash" class="h-11 bg-background shadow-sm" />
          </div>

          <div class="grid gap-2">
            <Label class="text-sm font-semibold opacity-70">Tipe Dompet</Label>
            <Select v-model="form.type">
              <SelectTrigger class="h-11 bg-background border-border">
                <SelectValue placeholder="Pilih Tipe" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="Cash">💵 Uang Tunai (Cash)</SelectItem>
                <SelectItem value="Bank">🏦 Bank / Rekening</SelectItem>
                <SelectItem value="E-Wallet">📱 E-Wallet (Dana/OVO)</SelectItem>
              </SelectContent>
            </Select>
          </div>

          <div class="grid gap-2 text-foreground">
            <Label class="text-sm font-semibold opacity-70">Icon / Emoji</Label>
            <button
              @click="isIconPickerOpen = true"
              type="button"
              class="w-full h-32 flex flex-col items-center justify-center border-dashed border-2 rounded-2xl hover:bg-accent/30 transition-all gap-2 bg-background border-border shadow-sm"
            >
              <template v-if="!form.icon">
                <div class="h-10 w-10 rounded-full bg-muted flex items-center justify-center text-muted-foreground">
                  <Plus class="h-5 w-5" />
                </div>
                <span class="text-sm text-muted-foreground font-medium italic">Pilih icon/emoji</span>
              </template>
              <template v-else>
                <div class="h-16 w-16 rounded-2xl bg-emerald-100 dark:bg-emerald-900/30 flex items-center justify-center text-emerald-600 shadow-sm">
                  <span v-if="form.isEmoji" class="text-5xl leading-none">{{ form.icon }}</span>
                  <component v-else :is="getIconComponent(form.icon)" class="h-10 w-10" />
                </div>
                <span class="text-[10px] uppercase tracking-widest font-bold opacity-40 mt-1">Klik untuk ganti</span>
              </template>
            </button>
          </div>
        </div>

        <DialogFooter class="p-6 border-t bg-muted/5 flex flex-row items-center justify-between gap-2">
          <Button v-if="isEditMode" variant="ghost" type="button" class="text-red-500 hover:text-red-600 hover:bg-red-50 gap-2 px-4" @click="handleDelete"> <Trash2 class="w-4 h-4" /> Hapus </Button>
          <div class="flex gap-2 ml-auto">
            <Button variant="outline" type="button" @click="isDialogOpen = false">Batal</Button>
            <Button @click="handleSave" type="button" class="bg-emerald-600 hover:bg-emerald-700 text-white px-8 shadow-md">
              <template v-if="isEditMode"> <Pencil class="w-4 h-4 mr-2" /> Edit </template>
              <template v-else> <Save class="w-4 h-4 mr-2" /> Selesai </template>
            </Button>
          </div>
        </DialogFooter>
      </DialogContent>
    </Dialog>

    <Dialog v-model:open="isIconPickerOpen">
      <DialogContent class="max-w-md h-125 flex flex-col p-0 overflow-hidden bg-card border-border shadow-2xl text-foreground">
        <DialogHeader class="p-4 border-b text-center"><DialogTitle class="text-sm font-bold">Visual Dompet</DialogTitle></DialogHeader>
        <Tabs default-value="icons" class="flex-1 flex flex-col overflow-hidden">
          <div class="px-6 pt-4">
            <TabsList class="grid w-full grid-cols-2 shadow-sm"><TabsTrigger value="icons">Icons</TabsTrigger><TabsTrigger value="emojis">Emojis</TabsTrigger></TabsList>
          </div>
          <TabsContent value="icons" class="flex-1 overflow-y-auto p-6 mt-0">
            <div class="grid grid-cols-4 gap-4">
              <Button v-for="item in iconOptions" :key="item.name" variant="ghost" type="button" class="h-20 flex flex-col gap-2 hover:bg-emerald-50 dark:hover:bg-emerald-900/40" @click="selectIcon(item.name, false)">
                <component :is="item.icon" class="h-6 w-6" />
                <span class="text-[10px] font-medium opacity-60 truncate w-full uppercase tracking-tighter">{{ item.label }}</span>
              </Button>
            </div>
          </TabsContent>
          <TabsContent value="emojis" class="flex-1 overflow-y-auto p-6 mt-0 text-center">
            <div class="grid grid-cols-4 gap-6">
              <button v-for="e in emojiList" :key="e" type="button" class="text-4xl p-2 hover:bg-accent rounded-2xl transition-transform active:scale-90" @click="selectIcon(e, true)">{{ e }}</button>
            </div>
          </TabsContent>
        </Tabs>
      </DialogContent>
    </Dialog>
  </div>
</template>
