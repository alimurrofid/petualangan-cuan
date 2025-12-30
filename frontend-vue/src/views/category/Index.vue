<script setup lang="ts">
import { ref, onMounted, computed } from "vue";
import { useCategoryStore, type Category } from "@/stores/category";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogFooter, DialogDescription } from "@/components/ui/dialog";
import { useSwal } from "@/composables/useSwal";

import * as LucideIcons from "lucide-vue-next";
import { Plus, Pencil, Trash2, LayoutGrid, Save, TrendingUp, TrendingDown } from "lucide-vue-next";

// Define the form structure (frontend representation)
interface CategoryForm {
  id: number;
  name: string;
  icon: string;
  type: "income" | "expense";
  budgetLimit?: number;
}

const categoryStore = useCategoryStore();
const currentTab = ref<"expense" | "income">("income");
const swal = useSwal();

const isDialogOpen = ref(false);
const isIconPickerOpen = ref(false);
const isEditMode = ref(false);

const form = ref<CategoryForm>({
  id: 0,
  name: "",
  icon: "",
  type: "expense",
  budgetLimit: 0,
});

const errors = ref({
  name: false,
  icon: false,
});

const iconOptions = [
  { name: "Utensils", icon: LucideIcons.Utensils, label: "Makan" },
  { name: "ShoppingBag", icon: LucideIcons.ShoppingBag, label: "Belanja" },
  { name: "Coffee", icon: LucideIcons.Coffee, label: "Kopi" },
  { name: "Car", icon: LucideIcons.Car, label: "Transport" },
  { name: "Zap", icon: LucideIcons.Zap, label: "Tagihan" },
  { name: "Heart", icon: LucideIcons.Heart, label: "Amal" },
  { name: "Gamepad2", icon: LucideIcons.Gamepad2, label: "Hiburan" },
  { name: "Briefcase", icon: LucideIcons.Briefcase, label: "Gaji" },
  { name: "TrendingUp", icon: LucideIcons.TrendingUp, label: "Investasi" },
  { name: "Gift", icon: LucideIcons.Gift, label: "Hadiah" },
  { name: "Stethoscope", icon: LucideIcons.Stethoscope, label: "Medis" },
  { name: "Home", icon: LucideIcons.Home, label: "Rumah" },
];

const emojiCategories = {
  Keuangan: [
    { name: "Em_MoneyBag", emoji: "💰" },
    { name: "Em_DollarBill", emoji: "💵" },
    { name: "Em_Card", emoji: "💳" },
    { name: "Em_Bank", emoji: "🏦" },
    { name: "Em_MoneyWing", emoji: "💸" },
    { name: "Em_Coin", emoji: "🪙" },
  ],
  Lifestyle: [
    { name: "Em_Pizza", emoji: "🍕" },
    { name: "Em_Cart", emoji: "🛒" },
    { name: "Em_Coffee", emoji: "☕" },
    { name: "Em_Game", emoji: "🎮" },
    { name: "Em_Airplane", emoji: "✈️" },
    { name: "Em_Gift", emoji: "🎁" },
  ],
  Simbol: [
    { name: "Em_Star", emoji: "⭐" },
    { name: "Em_Fire", emoji: "🔥" },
    { name: "Em_Lock", emoji: "🔒" },
    { name: "Em_Check", emoji: "✅" },
    { name: "Em_Idea", emoji: "💡" },
  ],
};

onMounted(() => {
  categoryStore.fetchCategories();
});

const filteredCategories = computed(() => {
  return categoryStore.categories.filter((c) => c.type === currentTab.value);
});

const isSubmitting = ref(false);

const openAdd = () => {
  isEditMode.value = false;
  form.value = { id: 0, name: "", icon: "", type: currentTab.value, budgetLimit: 0 };
  errors.value = { name: false, icon: false };
  isSubmitting.value = false;
  isDialogOpen.value = true;
};

const openEdit = (category: Category) => {
  isEditMode.value = true;
  form.value = {
    id: category.id,
    name: category.name,
    icon: category.icon || "",
    type: category.type,
    // Note: Backend JSON for budget_limit -> category.budget_limit
    budgetLimit: category.budget_limit || 0,
  };
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
      if (errors.value.name) msg += "<br>- Nama Kategori";
      if (errors.value.icon) msg += "<br>- Icon Kategori";
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
    icon: form.value.icon,
    budget_limit: Number(form.value.budgetLimit),
  };

  try {
    if (isEditMode.value) {
      await categoryStore.updateCategory(form.value.id, payload);
      swal.success("Berhasil Update", "Kategori berhasil diperbarui");
    } else {
      await categoryStore.createCategory(payload);
      swal.success("Berhasil Tambah", "Kategori baru berhasil dibuat");
    }
    isDialogOpen.value = false;
  } catch (error) {
    swal.error("Gagal Menyimpan", "Terjadi kesalahan saat menyimpan data");
  } finally {
    isSubmitting.value = false;
  }
};

const handleDelete = async () => {
  const confirmed = await swal.confirmDelete('Kategori');
  if (confirmed) {
    try {
      await categoryStore.deleteCategory(form.value.id);
      isDialogOpen.value = false;
      swal.success("Terhapus", "Kategori berhasil dihapus");
    } catch (error) {
       swal.error("Gagal", "Gagal menghapus kategori");
    }
  }
};

const getIconComponent = (name: string | undefined) => {
  if (!name) return LayoutGrid;
  return (LucideIcons as any)[name] || null;
};

const getEmoji = (name: string | undefined) => {
  if (!name) return null;
  for (const category of Object.values(emojiCategories)) {
    const found = category.find((e) => e.name === name);
    if (found) return found.emoji;
  }
  // Fallback if name itself is an emoji (legacy support)
  if (/\p{Emoji}/u.test(name)) return name;
  return null;
};

const getGradientIcon = (type: string) => {
    return type === 'expense' 
        ? 'bg-gradient-to-br from-red-50 to-red-100 text-red-600 dark:from-red-900 dark:to-red-800 dark:text-red-100' 
        : 'bg-gradient-to-br from-emerald-50 to-emerald-100 text-emerald-600 dark:from-emerald-900 dark:to-emerald-800 dark:text-emerald-100';
};

const formattedBudgetLimit = computed({
  get: () => {
    if (!form.value.budgetLimit) return "";
    return new Intl.NumberFormat("id-ID", { style: "currency", currency: "IDR", maximumFractionDigits: 0 }).format(form.value.budgetLimit);
  },
  set: (val: string) => {
    const numericValue = Number(val.replace(/[^0-9]/g, ""));
    form.value.budgetLimit = numericValue;
  }
});
</script>

<template>
  <div class="p-6 space-y-8 text-foreground min-h-screen bg-background">
    <div class="flex flex-col md:flex-row md:items-end justify-between gap-6">
      <div>
        <h2 class="text-3xl font-bold tracking-tight">Kategori</h2>
        <p class="text-muted-foreground mt-1">Klasifikasikan transaksi Anda agar lebih terorganisir.</p>
      </div>
      <Button @click="openAdd" class="bg-gradient-to-r from-emerald-600 to-teal-500 text-white hover:from-emerald-500 hover:to-teal-400 shadow-lg px-6 h-12 rounded-full transition-all hover:scale-105 active:scale-95"> 
        <Plus class="w-5 h-5 mr-2" /> 
        Tambah Kategori 
      </Button>
    </div>

    <Tabs v-model:model-value="currentTab" class="space-y-6">
        <TabsList class="grid w-full grid-cols-2 p-1 bg-muted/60 rounded-full h-auto">
            <TabsTrigger value="income" class="rounded-full data-[state=active]:bg-emerald-500 data-[state=active]:text-white dark:data-[state=active]:bg-emerald-600 dark:data-[state=active]:text-white data-[state=active]:shadow-md transition-all py-2.5 font-bold hover:bg-emerald-500/10 data-[state=active]:hover:bg-emerald-600">
                <TrendingUp class="w-4 h-4 mr-2" /> Pemasukan
            </TabsTrigger>
            <TabsTrigger value="expense" class="rounded-full data-[state=active]:bg-red-500 data-[state=active]:text-white dark:data-[state=active]:bg-red-600 dark:data-[state=active]:text-white data-[state=active]:shadow-md transition-all py-2.5 font-bold hover:bg-red-500/10 data-[state=active]:hover:bg-red-600">
                 <TrendingDown class="w-4 h-4 mr-2" /> Pengeluaran
            </TabsTrigger>
        </TabsList>

        <div class="grid gap-6 grid-cols-2 md:grid-cols-3 lg:grid-cols-4 xl:grid-cols-5 animate-in fade-in slide-in-from-bottom-4 duration-500">
            <div
                v-for="item in filteredCategories"
                :key="item.id"
                @click="openEdit(item)"
                :class="['group relative bg-card h-48 rounded-3xl p-5 flex flex-col justify-between cursor-pointer transition-all duration-300 hover:-translate-y-2 hover:shadow-xl border border-border/60 hover:border-border', item.type === 'expense' ? 'hover:border-red-200 dark:hover:border-red-500/50' : 'hover:border-emerald-200 dark:hover:border-emerald-500/50']"
            >
                <!-- Header: Icon & Edit Hint -->
                <div class="relative z-10 flex justify-between items-start">
                    <div :class="['h-14 w-14 rounded-2xl flex items-center justify-center border shadow-sm transition-transform group-hover:scale-110', getGradientIcon(item.type), item.type === 'expense' ? 'border-red-100 dark:border-red-800' : 'border-emerald-100 dark:border-emerald-800']">
                         <component v-if="getIconComponent(item.icon)" :is="getIconComponent(item.icon)" class="h-7 w-7" />
                         <span v-else-if="getEmoji(item.icon)" class="text-3xl leading-none filter drop-shadow-sm">{{ getEmoji(item.icon) }}</span>
                         <span v-else class="text-xl leading-none filter drop-shadow-sm">{{ item.icon }}</span>
                    </div>
                    <div class="opacity-0 group-hover:opacity-100 transition-opacity">
                         <div class="bg-muted p-2 rounded-full transform rotate-12 group-hover:rotate-0 transition-transform">
                            <Pencil class="w-3.5 h-3.5 text-muted-foreground" />
                         </div>
                    </div>
                </div>

                <!-- Footer: Name & Type -->
                <div class="relative z-10 mt-auto">
                    <p class="font-bold text-lg tracking-wide truncate leading-tight text-foreground group-hover:text-primary transition-colors">{{ item.name }}</p>
                    <div class="flex items-center gap-1.5 mt-2">
                         <div :class="['h-1.5 w-1.5 rounded-full', item.type === 'expense' ? 'bg-red-500' : 'bg-emerald-500']"></div>
                         <p class="text-[10px] uppercase font-bold tracking-widest text-muted-foreground">{{ item.type === 'expense' ? 'Pengeluaran' : 'Pemasukan' }}</p>
                    </div>
                     <p v-if="item.type === 'expense' && item.budget_limit" class="text-xs text-muted-foreground mt-1">
                        Target: {{ new Intl.NumberFormat("id-ID", { style: "currency", currency: "IDR", maximumFractionDigits: 0 }).format(item.budget_limit) }}
                    </p>
                </div>
            </div>

             <div v-if="filteredCategories.length === 0" class="col-span-full py-20 flex flex-col items-center justify-center text-muted-foreground border-2 border-dashed border-border rounded-3xl opacity-50 bg-muted/20">
                <div class="h-16 w-16 bg-muted rounded-full flex items-center justify-center mb-4">
                     <LayoutGrid class="w-8 h-8 opacity-40" />
                </div>
                <p class="text-lg font-medium">Belum ada kategori {{ currentTab === "income" ? "pemasukan" : "pengeluaran" }}</p>
                <p class="text-sm">Klik tombol tambah untuk membuat baru.</p>
            </div>
        </div>
    </Tabs>

    <Dialog v-model:open="isDialogOpen">
       <DialogContent class="max-w-md bg-card p-0 overflow-hidden border-border shadow-2xl" @interact-outside="swal.handleSwalInteractOutside">
        <DialogHeader class="p-6 border-b">
          <DialogTitle>{{ isEditMode ? "Edit Kategori" : "Tambah Kategori" }}</DialogTitle>
          <DialogDescription>Sesuaikan nama dan visualisasi kategori.</DialogDescription>
        </DialogHeader>

        <div class="p-6 space-y-6">
          <div class="grid gap-2">
            <Label class="text-sm font-semibold opacity-70">Nama Kategori</Label>
            <Input v-model="form.name" placeholder="Misal: Belanja, Gaji" :class="['h-11 bg-background shadow-sm', errors.name ? 'border-red-500 ring-1 ring-red-500' : '']" :disabled="isSubmitting" />
            <span v-if="errors.name" class="text-xs text-red-500 font-medium">Nama kategori wajib diisi</span>
          </div>

          <div class="grid gap-2">
            <Label class="text-sm font-semibold opacity-70">Tipe Kategori</Label>
            <Select v-model="form.type" :disabled="isSubmitting">
              <SelectTrigger class="w-full h-11 bg-background border-border">
                <SelectValue placeholder="Pilih Tipe" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="expense">📉 Pengeluaran (Expense)</SelectItem>
                <SelectItem value="income">📈 Pemasukan (Income)</SelectItem>
              </SelectContent>
            </Select>
          </div>

          <div v-if="form.type === 'expense'" class="grid gap-2">
            <Label class="text-sm font-semibold opacity-70">Target Pengeluaran (Rp)</Label>
            <Input v-model="formattedBudgetLimit" type="text" placeholder="Rp 0" class="h-11 bg-background shadow-sm" :disabled="isSubmitting" />
            <p class="text-[10px] text-muted-foreground">Isi 0 jika tidak ingin membatasi pengeluaran.</p>
          </div>

          <div class="grid gap-2 text-foreground">
            <Label class="text-sm font-semibold opacity-70">Icon / Emoji</Label>
            <button
              @click="isIconPickerOpen = true"
              type="button"
              :class="['w-full h-28 flex items-center justify-center border-dashed border-2 rounded-2xl hover:bg-accent/30 transition-all gap-4 bg-background border-border shadow-sm group', errors.icon ? 'border-red-500 bg-red-50/10' : '', isSubmitting ? 'opacity-50 cursor-not-allowed' : '']"
              :disabled="isSubmitting"
            >
              <template v-if="!form.icon">
                <div class="h-12 w-12 rounded-full bg-muted flex items-center justify-center text-muted-foreground group-hover:scale-110 transition-transform">
                  <Plus :class="['h-6 w-6', errors.icon ? 'text-red-500' : '']" />
                </div>
                <span :class="['text-sm font-medium italic', errors.icon ? 'text-red-500' : 'text-muted-foreground']">Pilih icon...</span>
              </template>
              <template v-else>
                <div :class="['h-16 w-16 rounded-2xl flex items-center justify-center text-4xl shadow-md transform group-hover:scale-105 transition-transform', form.type === 'expense' ? 'bg-red-50 text-red-500' : 'bg-emerald-50 text-emerald-600']">
                  <component v-if="getIconComponent(form.icon)" :is="getIconComponent(form.icon)" class="h-8 w-8" />
                  <span v-else-if="getEmoji(form.icon)" class="leading-none">{{ getEmoji(form.icon) }}</span>
                  <span v-else class="leading-none">{{ form.icon }}</span>
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
            <Button @click="handleSave" type="button" class="bg-gradient-to-r from-emerald-600 to-teal-500 text-white hover:from-emerald-500 hover:to-teal-400 px-6 shadow-md hover:bg-foreground/90" :disabled="isSubmitting">
              <template v-if="isEditMode"> <Pencil class="w-4 h-4 mr-2" /> Simpan </template>
              <template v-else> <Save class="w-4 h-4 mr-2" /> Buat </template>
            </Button>
          </div>
        </DialogFooter>
      </DialogContent>
    </Dialog>

    <Dialog v-model:open="isIconPickerOpen">
      <DialogContent class="max-w-md h-125 flex flex-col p-0 overflow-hidden bg-card border-border shadow-2xl text-foreground">
        <DialogHeader class="p-4 border-b text-center"><DialogTitle class="text-sm font-bold">Visual Kategori</DialogTitle></DialogHeader>
        <Tabs default-value="icons" class="flex-1 flex flex-col overflow-hidden">
          <div class="px-6 pt-4">
            <TabsList class="grid w-full grid-cols-2 shadow-sm"><TabsTrigger value="icons">Icons</TabsTrigger><TabsTrigger value="emojis">Emojis</TabsTrigger></TabsList>
          </div>
          <TabsContent value="icons" class="flex-1 overflow-y-auto p-6 mt-0">
            <div class="grid grid-cols-4 gap-4">
              <Button v-for="item in iconOptions" :key="item.name" variant="ghost" type="button" class="h-20 flex flex-col gap-2 hover:bg-primary/10" @click="selectIcon(item.name)">
                <component :is="item.icon" class="h-6 w-6" />
                <span class="text-[9px] font-medium opacity-60 truncate w-full">{{ item.label }}</span>
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
