<script setup lang="ts">
import { ref, computed, onMounted } from "vue";
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
import { Plus, Pencil, Trash2, Save, Nfc } from "lucide-vue-next";

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
});

const errors = ref({
  name: false,
  icon: false,
});

const totalBalance = computed(() => {
    return wallets.value.reduce((sum: any, w: any) => sum + w.balance, 0);
});



onMounted(() => {
  walletStore.fetchWallets();
});

const openAdd = () => {
  isEditMode.value = false;
  form.value = { id: 0, name: "", type: "Cash", icon: "", balance: 0 };
  errors.value = { name: false, icon: false };
  isSubmitting.value = false;
  isDialogOpen.value = true;
};

const openEdit = (wallet: any) => {
  isEditMode.value = true;
  form.value = { ...wallet };
  errors.value = { name: false, icon: false };
  isSubmitting.value = false;
  isDialogOpen.value = true;
};

const selectIcon = (name: string) => {
  form.value.icon = name;
  errors.value.icon = false;
  isIconPickerOpen.value = false;
};

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
      // Small delay to prevent ghost clicks after modal closes
      setTimeout(() => { isSubmitting.value = false; }, 300);
      return;
  }
  
  const payload = {
    name: form.value.name,
    type: form.value.type,
    balance: Number(form.value.balance),
    icon: form.value.icon,
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

const formatCurrency = (value: number) => {
  return new Intl.NumberFormat("id-ID", { style: "currency", currency: "IDR", minimumFractionDigits: 0 }).format(value);
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
  <div class="p-6 space-y-8 text-foreground min-h-screen bg-background">
    
    <div class="flex flex-col md:flex-row md:items-end justify-between gap-6">
      <div>
        <h2 class="text-3xl font-bold tracking-tight">Dompet Saya</h2>
        <p class="text-muted-foreground mt-1">Total aset bersih Anda saat ini.</p>
        <div class="mt-4 flex items-baseline gap-2">
            <span class="text-4xl font-extrabold text-transparent bg-clip-text bg-gradient-to-r from-emerald-600 to-teal-500">
                {{ formatCurrency(totalBalance) }}
            </span>
        </div>
      </div>
      
      <Button @click="openAdd" class="bg-gradient-to-r from-emerald-600 to-teal-500 text-white hover:from-emerald-500 hover:to-teal-400 shadow-lg px-6 h-12 rounded-full transition-all hover:scale-105 active:scale-95"> 
        <Plus class="w-5 h-5 mr-2" /> 
        Tambah Dompet 
      </Button>
    </div>

    <div class="grid gap-6 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-3">
        
      <div 
        v-for="item in wallets" 
        :key="item.id" 
        @click="openEdit(item)" 
        :class="['relative h-56 rounded-3xl p-6 flex flex-col justify-between shadow-2xl cursor-pointer transition-all duration-300 hover:-translate-y-2 hover:shadow-xl group overflow-hidden', getCardGradient(item.type)]"
      >
        <div class="absolute top-0 right-0 w-48 h-48 bg-white/5 rounded-full blur-3xl -mr-16 -mt-16 pointer-events-none"></div>
        <div class="absolute bottom-0 left-0 w-32 h-32 bg-black/10 rounded-full blur-2xl -ml-10 -mb-10 pointer-events-none"></div>

        <div class="relative z-10 flex justify-between items-start">
            <div class="flex items-center gap-3">
                <div class="h-10 w-10 rounded-full bg-white/20 backdrop-blur-md flex items-center justify-center border border-white/10 shadow-inner">
                     <component v-if="getIconComponent(item.icon)" :is="getIconComponent(item.icon)" class="h-5 w-5 text-white" />
                     <span v-else-if="getEmoji(item.icon)" class="text-xl leading-none filter drop-shadow-sm">{{ getEmoji(item.icon) }}</span>
                     <component v-else :is="getIconComponent(null, 'Wallet')" class="h-5 w-5 text-white" />
                </div>
                <div>
                     <p class="font-bold text-lg tracking-wide">{{ item.name }}</p>
                     <p class="text-[10px] uppercase font-bold opacity-70 tracking-widest">{{ item.type }}</p>
                </div>
            </div>
            <Nfc class="h-8 w-8 opacity-40 rotate-90" />
        </div>

        <div class="relative z-10 my-auto pl-1">
            <div class="w-12 h-9 rounded-md bg-gradient-to-br from-yellow-200 to-yellow-500 border border-yellow-600/30 shadow-sm flex items-center justify-center relative overflow-hidden mb-4 opacity-90">
                <div class="absolute inset-0 border-[0.5px] border-black/10 rounded-md" style="background-image: repeating-linear-gradient(45deg, transparent, transparent 2px, rgba(0,0,0,0.1) 2px, rgba(0,0,0,0.1) 4px);"></div>
            </div>
            
             <div class="space-y-1">
                 <p class="text-xs font-semibold opacity-70 uppercase tracking-widest">Saldo Saat Ini</p>
                 <p class="text-2xl font-mono font-bold tracking-tight">{{ formatCurrency(item.balance) }}</p>
             </div>
        </div>

        <div class="relative z-10 flex justify-between items-center opacity-70 font-mono text-xs tracking-widest pl-1">
            <span class="uppercase">{{ authStore.user?.name || 'USER' }}</span>
            <span>**** ****</span>
        </div>

        <div class="absolute inset-0 bg-black/40 backdrop-blur-[1px] opacity-0 group-hover:opacity-100 transition-opacity flex items-center justify-center z-20">
            <span class="bg-white text-black px-4 py-2 rounded-full text-xs font-bold shadow-lg transform scale-90 group-hover:scale-100 transition-transform">
                Edit Dompet
            </span>
        </div>
      </div>
    
    </div>

    <Dialog v-model:open="isDialogOpen">
       <DialogContent class="max-w-md bg-card p-0 overflow-hidden border-border shadow-2xl" @interact-outside="swal.handleSwalInteractOutside">
        <DialogHeader class="p-6 border-b">
          <DialogTitle>{{ isEditMode ? "Edit Dompet" : "Tambah Dompet" }}</DialogTitle>
          <DialogDescription>Simpan informasi detail dompet Anda.</DialogDescription>
        </DialogHeader>

        <div class="p-6 space-y-5 text-foreground">
          <div class="grid gap-2">
            <Label class="text-sm font-semibold opacity-70">Nama Dompet</Label>
            <Input v-model="form.name" placeholder="Misal: BCA Utama, Cash" :class="['h-11 bg-background shadow-sm', errors.name ? 'border-red-500 ring-1 ring-red-500' : '']" :disabled="isSubmitting" />
            <span v-if="errors.name" class="text-xs text-red-500 font-medium">Nama dompet wajib diisi</span>
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

          <div class="grid gap-2 text-foreground">
            <Label class="text-sm font-semibold opacity-70">Icon / Emoji</Label>
            <button
              @click="isIconPickerOpen = true"
              type="button"
              :class="['w-full h-24 flex items-center justify-center border-dashed border-2 rounded-2xl hover:bg-accent/30 transition-all gap-4 bg-background border-border shadow-sm group', errors.icon ? 'border-red-500 bg-red-50/10' : '', isSubmitting ? 'opacity-50 cursor-not-allowed' : '']"
              :disabled="isSubmitting"
            >
              <template v-if="!form.icon">
                <div class="h-12 w-12 rounded-full bg-muted flex items-center justify-center text-muted-foreground group-hover:scale-110 transition-transform">
                  <Plus :class="['h-6 w-6', errors.icon ? 'text-red-500' : '']" />
                </div>
                <span :class="['text-sm font-medium italic', errors.icon ? 'text-red-500' : 'text-muted-foreground']">Pilih icon...</span>
              </template>
              <template v-else>
                <div :class="['h-14 w-14 rounded-2xl flex items-center justify-center text-white shadow-md transform group-hover:scale-105 transition-transform', getCardGradient(form.type)]">
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
            <span v-if="errors.icon" class="text-xs text-red-500 font-medium">Icon wajib dipilih</span>
          </div>
        </div>

        <DialogFooter class="p-6 border-t bg-muted/5 flex flex-row items-center justify-between gap-2">
          <Button v-if="isEditMode" variant="ghost" type="button" class="text-red-500 hover:text-red-600 hover:bg-red-50 gap-2 px-4" @click="handleDelete" :disabled="isSubmitting"> <Trash2 class="w-4 h-4" /> Hapus </Button>
          <div class="flex gap-2 ml-auto">
            <Button variant="outline" type="button" @click="isDialogOpen = false" :disabled="isSubmitting">Batal</Button>
            <Button @click="handleSave" type="button" class="bg-gradient-to-r from-emerald-600 to-teal-500 text-white hover:from-emerald-500 hover:to-teal-400 px-6 shadow-md" :disabled="isSubmitting">
              <template v-if="isEditMode"> <Pencil class="w-4 h-4 mr-2" /> Simpan </template>
              <template v-else> <Save class="w-4 h-4 mr-2" /> Buat </template>
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
              <Button v-for="item in walletIcons" :key="item.name" variant="ghost" type="button" class="h-20 flex flex-col gap-2 hover:bg-primary/10" @click="selectIcon(item.name)">
                <component :is="item.icon" class="h-6 w-6" />
                <span class="text-[10px] font-medium opacity-60 truncate w-full uppercase tracking-tighter">{{ item.label }}</span>
              </Button>
            </div>
          </TabsContent>
          <TabsContent value="emojis" class="flex-1 overflow-y-auto p-6 mt-0">
            <div v-for="(list, cat) in emojiCategories" :key="cat" class="mb-6">
              <p class="text-[10px] font-bold text-muted-foreground uppercase mb-3 text-left tracking-widest">{{ cat }}</p>
              <div class="grid grid-cols-4 gap-4">
                <button v-for="e in list" :key="e.name" type="button" class="text-4xl p-2 hover:bg-accent rounded-2xl transition-transform active:scale-90" @click="selectIcon(e.name)">{{ e.emoji }}</button>
              </div>
            </div>
          </TabsContent>
        </Tabs>
      </DialogContent>
    </Dialog>
  </div>
</template>
