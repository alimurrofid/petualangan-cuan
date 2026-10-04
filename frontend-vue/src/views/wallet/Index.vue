<script setup lang="ts">
import { ref, onMounted, computed, watch } from "vue";
import { useWalletStore } from "@/stores/wallet";
import { useAuthStore } from "@/stores/auth";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogFooter, DialogDescription } from "@/components/ui/dialog";
import { useSwal } from "@/composables/useSwal";


import { emojiCategories, getEmoji, getIconComponent, walletIcons } from "@/lib/icons";
import { formatCurrency, parseCurrencyInput, formatCurrencyInput, formatCurrencyLive } from "@/lib/utils";
import { Plus, Pencil, Trash2, Save, Nfc, Search } from "lucide-vue-next";

const walletStore = useWalletStore();
const authStore = useAuthStore();
const swal = useSwal();
const wallets = computed(() => walletStore.wallets);

const isDialogOpen = ref(false);
const isIconPickerOpen = ref(false);
const isEditMode = ref(false);
const isSubmitting = ref(false);

const form = ref({
  id: 0,
  name: "",
  type: "Cash",
  icon: "",
  balance: 0,
  seq: 1,
});
const balanceDisplay = ref("");

const errors = ref({
  name: false,
  icon: false,
});

const totalOverall = computed(() => {
  return wallets.value.reduce((sum: any, w: any) => sum + (w.balance || 0), 0);
});

const totalAvailable = computed(() => {
  return wallets.value.reduce((sum: any, w: any) => sum + (w.available_balance || 0), 0);
});

watch(balanceDisplay, (val) => {
  const formatted = formatCurrencyLive(val);
  if (formatted !== val) {
    balanceDisplay.value = formatted;
    return;
  }
  const num = parseCurrencyInput(val);
  form.value.balance = num;
});

watch(() => form.value.balance, (val) => {
  const num = Number(val);
  const currentParsed = parseCurrencyInput(balanceDisplay.value);
  if (Math.abs(currentParsed - num) > 0.001) {
    balanceDisplay.value = val ? formatCurrencyInput(val) : "";
  }
});

const onBalanceBlur = () => {
  const num = parseCurrencyInput(balanceDisplay.value);
  if (num) balanceDisplay.value = formatCurrencyInput(num);
};



onMounted(() => {
  walletStore.fetchWallets();
});

const openAdd = () => {
  isEditMode.value = false;
  const maxSeq = wallets.value.length ? Math.max(...wallets.value.map((w: any) => w.seq ?? 0)) : 0;
  form.value = { id: 0, name: "", type: "Cash", icon: "", balance: 0, seq: maxSeq + 1 };
  balanceDisplay.value = "";
  errors.value = { name: false, icon: false };
  isSubmitting.value = false;
  isDialogOpen.value = true;
};

const openEdit = (wallet: any) => {
  isEditMode.value = true;
  form.value = { ...wallet, seq: wallet.seq ?? 0 };
  balanceDisplay.value = formatCurrencyInput(wallet.balance);
  errors.value = { name: false, icon: false };
  isSubmitting.value = false;
  isDialogOpen.value = true;
};

const selectIcon = (name: string) => {
  form.value.icon = name;
  errors.value.icon = false;
  isIconPickerOpen.value = false;
  iconSearch.value = '';
};

const iconSearch = ref('');

const filteredWalletIcons = computed(() => {
  if (!iconSearch.value) return walletIcons;
  const q = iconSearch.value.toLowerCase();
  return walletIcons.filter(i => i.name.toLowerCase().includes(q) || i.label.toLowerCase().includes(q));
});

const filteredEmojiCategories = computed(() => {
  if (!iconSearch.value) return emojiCategories;
  const q = iconSearch.value.toLowerCase();
  const result: Record<string, typeof emojiCategories[string]> = {};
  for (const [cat, list] of Object.entries(emojiCategories)) {
    const filtered = list.filter(e => e.name.toLowerCase().includes(q) || e.emoji.includes(q));
    if (filtered.length > 0) result[cat] = filtered;
  }
  return result;
});

const handleSave = async () => {
  isSubmitting.value = true;
  errors.value.name = !form.value.name;
  errors.value.icon = !form.value.icon;

  if (errors.value.name || errors.value.icon) {
    let msg = "Mohon lengkapi data berikut:";
    if (errors.value.name) msg += "<br>- Nama Dompet";
    if (errors.value.icon) msg += "<br>- Icon Dompet";
    await swal.fire({
      icon: 'error',
      title: 'Validasi Gagal',
      html: msg,
      confirmButtonColor: '#EF4444',
    });
    setTimeout(() => { isSubmitting.value = false; }, 300);
    return;
  }

  const payload = {
    name: form.value.name,
    type: form.value.type,
    balance: Number(form.value.balance),
    icon: form.value.icon,
    seq: Number(form.value.seq) || 0,
  };

  try {
    if (isEditMode.value) {
      await walletStore.updateWallet(form.value.id, payload);
      swal.success("Berhasil Update", "Dompet berhasil diperbarui");
    } else {
      await walletStore.createWallet(payload);
      swal.success("Berhasil Tambah", "Dompet baru berhasil dibuat");
    }
    isDialogOpen.value = false;
  } catch (error) {
    swal.error("Gagal Menyimpan", "Terjadi kesalahan saat menyimpan data");
  } finally {
    isSubmitting.value = false;
  }
};

const handleDelete = async () => {
  const confirmed = await swal.confirmDelete('Dompet');
  if (confirmed) {
    try {
      await walletStore.deleteWallet(form.value.id);
      isDialogOpen.value = false;
      swal.success("Terhapus", "Dompet berhasil dihapus");
    } catch (error) {
      swal.error("Gagal", "Gagal menghapus dompet");
    }
  }
};

const getCardGradient = (type: string) => {
  switch (type) {
    case 'Bank': return 'bg-gradient-to-br from-[#1e3a8a] to-[#3b82f6] text-white';
    case 'E-Wallet': return 'bg-gradient-to-br from-[#581c87] to-[#a855f7] text-white';
    case 'Cash': return 'bg-gradient-to-br from-[#064e3b] to-[#10b981] text-white';
    default: return 'bg-gradient-to-br from-slate-800 to-slate-600 text-white';
  }
};

</script>

<template>
  <div class="flex-1 pt-2 space-y-6" v-if="walletStore.isLoading">
    <div class="flex items-center justify-center min-h-[400px]">
      <p class="text-muted-foreground animate-pulse">Memuat data dompet...</p>
    </div>
  </div>
  <div class="flex-1 pt-2 space-y-6 text-foreground" v-else>

    <div class="flex flex-col justify-between gap-6 md:flex-row md:items-end">
      <div>
        <h2 class="text-3xl font-bold tracking-tight">Dompet Saya</h2>
        <p class="mt-1 text-muted-foreground">Total aset bersih Anda termasuk tabungan dan dana aktif.</p>
        <div class="mt-4 space-y-1">
          <div class="flex items-baseline gap-2">
            <span class="text-xs font-bold tracking-widest uppercase text-emerald-600 dark:text-emerald-400">Total
              Tersedia</span>
          </div>
          <div class="flex items-baseline gap-2">
            <span
              class="text-4xl font-extrabold text-transparent bg-clip-text bg-gradient-to-r from-emerald-600 to-teal-500"
              :class="{ 'privacy-blur': authStore.isPrivacyMode }">
              {{ formatCurrency(totalAvailable) }}
            </span>
          </div>
          <div class="flex items-center gap-2 text-sm font-medium text-muted-foreground/70">
            <span>Total Keseluruhan:</span>
            <span class="text-sm font-medium text-muted-foreground/70"
              :class="{ 'privacy-blur': authStore.isPrivacyMode }">{{ formatCurrency(totalOverall) }}</span>
          </div>
        </div>
      </div>

      <Button @click="openAdd"
        class="h-12 px-6 text-white transition-all rounded-full shadow-lg bg-gradient-to-r from-emerald-600 to-teal-500 hover:from-emerald-500 hover:to-teal-400 hover:scale-105 active:scale-95">
        <Plus class="w-5 h-5 mr-2" />
        Tambah Dompet
      </Button>
    </div>

    <div class="grid grid-cols-1 gap-6 sm:grid-cols-2 lg:grid-cols-3">

      <div v-for="item in wallets" :key="item.id" @click="openEdit(item)"
        :class="['relative h-56 rounded-3xl p-6 flex flex-col justify-between shadow-2xl cursor-pointer transition-all duration-300 hover:-translate-y-2 hover:shadow-xl group overflow-hidden', getCardGradient(item.type)]">
        <div
          class="absolute top-0 right-0 w-48 h-48 -mt-16 -mr-16 rounded-full pointer-events-none bg-white/5 blur-3xl">
        </div>
        <div
          class="absolute bottom-0 left-0 w-32 h-32 -mb-10 -ml-10 rounded-full pointer-events-none bg-black/10 blur-2xl">
        </div>

        <div class="relative z-10 flex items-start justify-between">
          <div class="flex items-center gap-3">
            <div
              class="flex items-center justify-center w-10 h-10 border rounded-full shadow-inner bg-white/20 backdrop-blur-md border-white/10">
              <component v-if="getIconComponent(item.icon)" :is="getIconComponent(item.icon)"
                class="w-5 h-5 text-white" />
              <span v-else-if="getEmoji(item.icon)" class="text-xl leading-none filter drop-shadow-sm">{{
                getEmoji(item.icon) }}</span>
              <component v-else :is="getIconComponent(null, 'Wallet')" class="w-5 h-5 text-white" />
            </div>
            <div>
              <p class="text-lg font-bold tracking-wide">{{ item.name }}</p>
              <p class="text-[10px] uppercase font-bold opacity-70 tracking-widest">{{ item.type }}</p>
            </div>
          </div>
          <Nfc class="w-8 h-8 rotate-90 opacity-40" />
        </div>

        <div class="relative z-10 pl-1 my-auto">
          <div
            class="relative flex items-center justify-center w-12 mb-4 overflow-hidden border rounded-md shadow-sm h-9 bg-gradient-to-br from-yellow-200 to-yellow-500 border-yellow-600/30 opacity-90">
            <div class="absolute inset-0 border-[0.5px] border-black/10 rounded-md"
              style="background-image: repeating-linear-gradient(45deg, transparent, transparent 2px, rgba(0,0,0,0.1) 2px, rgba(0,0,0,0.1) 4px);">
            </div>
          </div>

          <div class="space-y-1">
            <p class="text-[10px] font-bold opacity-70 uppercase tracking-widest text-emerald-100">Saldo Tersedia</p>
            <p class="font-mono text-2xl font-bold tracking-tight filter drop-shadow-sm"
              :class="{ 'privacy-blur': authStore.isPrivacyMode }">{{ formatCurrency(item.available_balance ??
                item.balance) }}</p>

            <div class="flex items-center gap-1 pt-2 mt-1 border-t border-white/10 opacity-80">
              <span class="text-[10px] uppercase font-medium">Total Saldo:</span>
              <span class="font-mono text-xs font-bold" :class="{ 'privacy-blur': authStore.isPrivacyMode }">{{
                formatCurrency(item.balance) }}</span>
            </div>
          </div>
        </div>

        <div class="relative z-10 flex items-center justify-between pl-1 font-mono text-xs tracking-widest opacity-70">
          <span class="uppercase">{{ authStore.user?.name || 'USER' }}</span>
          <span>**** ****</span>
        </div>

        <div
          class="absolute inset-0 bg-black/40 backdrop-blur-[1px] opacity-0 group-hover:opacity-100 transition-opacity flex items-center justify-center z-20">
          <span
            class="px-4 py-2 text-xs font-bold text-black transition-transform transform scale-90 bg-white rounded-full shadow-lg group-hover:scale-100">
            Edit Dompet
          </span>
        </div>
      </div>

    </div>

    <Dialog v-model:open="isDialogOpen">
      <DialogContent class="max-w-md p-0 overflow-hidden shadow-2xl bg-card border-border"
        @interact-outside="swal.handleSwalInteractOutside">
        <DialogHeader class="p-6 border-b">
          <DialogTitle>{{ isEditMode ? "Edit Dompet" : "Tambah Dompet" }}</DialogTitle>
          <DialogDescription>Simpan informasi detail dompet Anda.</DialogDescription>
        </DialogHeader>

        <div class="p-6 space-y-5 text-foreground">
          <div class="grid gap-2">
            <Label class="text-sm font-semibold opacity-70">Nama Dompet</Label>
            <Input v-model="form.name" placeholder="Misal: BCA Utama, Cash"
              :class="['h-11 bg-background shadow-sm', errors.name ? 'border-red-500 ring-1 ring-red-500' : '']"
              :disabled="isSubmitting" />
            <span v-if="errors.name" class="text-xs font-medium text-red-500">Nama dompet wajib diisi</span>
          </div>

          <div class="grid gap-2">
            <Label class="text-sm font-semibold opacity-70">Tipe Dompet</Label>
            <Select v-model="form.type" :disabled="isSubmitting">
              <SelectTrigger class="w-full h-11 bg-background border-border">
                <SelectValue placeholder="Pilih Tipe" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="Cash">💵 Uang Tunai (Cash)</SelectItem>
                <SelectItem value="Bank">🏦 Bank / Rekening</SelectItem>
                <SelectItem value="E-Wallet">📱 E-Wallet (Dana/OVO)</SelectItem>
              </SelectContent>
            </Select>
          </div>

          <div class="grid gap-2">
            <Label class="text-sm font-semibold opacity-70">Saldo Awal</Label>
            <Input type="text" inputmode="decimal" placeholder="Rp 0" v-model="balanceDisplay" @blur="onBalanceBlur"
              :class="['h-11 bg-background shadow-sm']" :disabled="isSubmitting" />
          </div>

          <div class="grid gap-2">
            <Label class="text-sm font-semibold opacity-70">Urutan</Label>
            <Input type="number" v-model.number="form.seq" placeholder="Contoh: 1"
              class="shadow-sm h-11 bg-background" :disabled="isSubmitting" />
            <span class="text-[11px] text-muted-foreground">Menentukan urutan dompet di menu dan pilihan transaksi (angka lebih kecil muncul lebih awal).</span>
          </div>

          <div class="grid gap-2 text-foreground">
            <Label class="text-sm font-semibold opacity-70">Icon / Emoji</Label>
            <button @click="isIconPickerOpen = true" type="button"
              :class="['w-full h-24 flex items-center justify-center border-dashed border-2 rounded-2xl hover:bg-accent/30 transition-all gap-4 bg-background border-border shadow-sm group', errors.icon ? 'border-red-500 bg-red-50/10' : '', isSubmitting ? 'opacity-50 cursor-not-allowed' : '']"
              :disabled="isSubmitting">
              <template v-if="!form.icon">
                <div
                  class="flex items-center justify-center w-12 h-12 transition-transform rounded-full bg-muted text-muted-foreground group-hover:scale-110">
                  <Plus :class="['h-6 w-6', errors.icon ? 'text-red-500' : '']" />
                </div>
                <span
                  :class="['text-sm font-medium italic', errors.icon ? 'text-red-500' : 'text-muted-foreground']">Pilih
                  icon...</span>
              </template>
              <template v-else>
                <div
                  :class="['h-14 w-14 rounded-2xl flex items-center justify-center text-white shadow-md transform group-hover:scale-105 transition-transform', getCardGradient(form.type)]">
                  <component v-if="getIconComponent(form.icon)" :is="getIconComponent(form.icon)" class="h-7 w-7" />
                  <span v-else-if="getEmoji(form.icon)" class="text-3xl leading-none">{{ getEmoji(form.icon) }}</span>
                  <component v-else :is="getIconComponent(null, 'Wallet')" class="h-7 w-7" />
                </div>
                <div class="text-left">
                  <p class="text-xs font-bold uppercase opacity-50">Icon Terpilih</p>
                  <p class="text-sm font-semibold">Klik untuk ganti</p>
                </div>
              </template>
            </button>
            <span v-if="errors.icon" class="text-xs font-medium text-red-500">Icon wajib dipilih</span>
          </div>
        </div>

        <DialogFooter class="flex flex-row items-center justify-between gap-2 p-6 border-t bg-muted/5">
          <Button v-if="isEditMode" variant="ghost" type="button"
            class="gap-2 px-4 text-red-500 hover:text-red-600 hover:bg-red-50" @click="handleDelete"
            :disabled="isSubmitting">
            <Trash2 class="w-4 h-4" /> Hapus
          </Button>
          <div class="flex gap-2 ml-auto">
            <Button variant="outline" type="button" @click="isDialogOpen = false"
              :disabled="isSubmitting">Batal</Button>
            <Button @click="handleSave" type="button"
              class="px-6 text-white shadow-md bg-gradient-to-r from-emerald-600 to-teal-500 hover:from-emerald-500 hover:to-teal-400"
              :disabled="isSubmitting" :loading="isSubmitting">
              <template v-if="isEditMode">
                <Pencil class="w-4 h-4 mr-2" /> Simpan
              </template>
              <template v-else>
                <Save class="w-4 h-4 mr-2" /> Buat
              </template>
            </Button>
          </div>
        </DialogFooter>
      </DialogContent>
    </Dialog>

    <Dialog v-model:open="isIconPickerOpen">
      <DialogContent
        class="flex flex-col max-w-md p-0 overflow-hidden shadow-2xl h-125 bg-card border-border text-foreground">
        <DialogHeader class="p-4 text-center border-b">
          <DialogTitle class="text-sm font-bold">Visual Dompet</DialogTitle>
        </DialogHeader>
        <Tabs default-value="icons" class="flex flex-col flex-1 overflow-hidden">
          <div class="px-6 pt-4">
            <TabsList class="grid w-full grid-cols-2 shadow-sm">
              <TabsTrigger value="icons">Icons</TabsTrigger>
              <TabsTrigger value="emojis">Emojis</TabsTrigger>
            </TabsList>
          </div>
          <div class="px-6 pt-3">
            <div class="relative">
              <Search class="absolute w-4 h-4 -translate-y-1/2 left-3 top-1/2 text-muted-foreground" />
              <Input v-model="iconSearch" placeholder="Cari icon atau emoji..."
                class="text-sm rounded-lg h-9 pl-9 bg-background" />
            </div>
          </div>
          <TabsContent value="icons" class="flex-1 p-6 mt-0 overflow-y-auto">
            <div class="grid grid-cols-4 gap-4">
              <Button v-for="item in filteredWalletIcons" :key="item.name" variant="ghost" type="button"
                class="flex flex-col h-20 gap-2 hover:bg-primary/10" @click="selectIcon(item.name)">
                <component :is="item.icon" class="w-6 h-6" />
                <span class="text-[10px] font-medium opacity-60 truncate w-full uppercase tracking-tighter">{{
                  item.label
                  }}</span>
              </Button>
            </div>
            <p v-if="filteredWalletIcons.length === 0" class="py-8 text-sm text-center text-muted-foreground">Tidak ada
              icon
              cocok.</p>
          </TabsContent>
          <TabsContent value="emojis" class="flex-1 p-6 mt-0 overflow-y-auto">
            <div v-for="(list, cat) in filteredEmojiCategories" :key="cat" class="mb-6">
              <p class="text-[10px] font-bold text-muted-foreground uppercase mb-3 text-left tracking-widest">{{ cat }}
              </p>
              <div class="grid grid-cols-4 gap-4">
                <button v-for="e in list" :key="e.name" type="button"
                  class="p-2 text-4xl transition-transform hover:bg-accent rounded-2xl active:scale-90"
                  @click="selectIcon(e.name)">{{ e.emoji }}</button>
              </div>
            </div>
            <p v-if="Object.keys(filteredEmojiCategories).length === 0"
              class="py-8 text-sm text-center text-muted-foreground">Tidak ada emoji cocok.</p>
          </TabsContent>
        </Tabs>
      </DialogContent>
    </Dialog>
  </div>
</template>
