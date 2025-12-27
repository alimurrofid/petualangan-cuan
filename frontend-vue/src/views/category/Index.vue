<script setup lang="ts">
import { ref, onMounted, computed } from "vue";

import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogFooter, DialogDescription } from "@/components/ui/dialog";

import * as LucideIcons from "lucide-vue-next";
import { Plus, Pencil, Trash2, LayoutGrid, Save, TrendingUp, TrendingDown } from "lucide-vue-next";

interface CategoryItem {
  id: number;
  name: string;
  icon: string;
  isEmoji: boolean;
  type: "income" | "expense";
}

const categories = ref<CategoryItem[]>([]);
const currentTab = ref<"expense" | "income">("income");

const isDialogOpen = ref(false);
const isIconPickerOpen = ref(false);
const isEditMode = ref(false);
const form = ref<CategoryItem>({
  id: 0,
  name: "",
  icon: "",
  isEmoji: false,
  type: "expense",
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
  Keuangan: ["💰", "💵", "💳", "🏦", "💸", "🪙"],
  Lifestyle: ["🍕", "🛒", "☕", "🎮", "✈️", "🎁"],
  Simbol: ["⭐", "🔥", "🔒", "✅", "💡"],
};

const loadCategories = () => {
  const saved = localStorage.getItem("mock_categories");
  if (saved) {
    categories.value = JSON.parse(saved);
  } else {
    const initial: CategoryItem[] = [
      { id: 1, name: "Makanan", icon: "Utensils", isEmoji: false, type: "expense" },
      { id: 2, name: "Gaji", icon: "💰", isEmoji: true, type: "income" },
      { id: 3, name: "Transport", icon: "Car", isEmoji: false, type: "expense" },
      { id: 4, name: "Bonus", icon: "Gift", isEmoji: false, type: "income" },
    ];
    categories.value = initial;
    localStorage.setItem("mock_categories", JSON.stringify(initial));
  }
};

onMounted(loadCategories);

const filteredCategories = computed(() => {
  return categories.value.filter((c) => c.type === currentTab.value);
});

const openAdd = () => {
  isEditMode.value = false;
  form.value = { id: Date.now(), name: "", icon: "", isEmoji: false, type: currentTab.value };
  isDialogOpen.value = true;
};

const openEdit = (category: CategoryItem) => {
  isEditMode.value = true;
  form.value = { ...category };
  isDialogOpen.value = true;
};

const selectIcon = (name: string, isEmoji: boolean) => {
  form.value.icon = name;
  form.value.isEmoji = isEmoji;
  isIconPickerOpen.value = false;
};

const handleSave = () => {
  if (!form.value.name || !form.value.icon) return alert("Lengkapi data kategori");
  if (isEditMode.value) {
    categories.value = categories.value.map((c) => (c.id === form.value.id ? { ...form.value } : c));
  } else {
    categories.value.push({ ...form.value });
  }
  localStorage.setItem("mock_categories", JSON.stringify(categories.value));
  isDialogOpen.value = false;
};

const handleDelete = () => {
  if (confirm("Hapus kategori ini?")) {
    categories.value = categories.value.filter((c) => c.id !== form.value.id);
    localStorage.setItem("mock_categories", JSON.stringify(categories.value));
    isDialogOpen.value = false;
  }
};

const getIconComponent = (name: string) => {
  return (LucideIcons as any)[name] || LayoutGrid;
};

const getGradientIcon = (type: string) => {
    return type === 'expense' 
        ? 'bg-gradient-to-br from-red-50 to-red-100 text-red-600 dark:from-red-900 dark:to-red-800 dark:text-red-100' 
        : 'bg-gradient-to-br from-emerald-50 to-emerald-100 text-emerald-600 dark:from-emerald-900 dark:to-emerald-800 dark:text-emerald-100';
};
</script>

<template>
  <div class="p-6 space-y-8 text-foreground min-h-screen bg-background">
    <div class="flex flex-col md:flex-row md:items-end justify-between gap-6">
      <div>
        <h2 class="text-3xl font-bold tracking-tight">Kategori</h2>
        <p class="text-muted-foreground mt-1">Klasifikasikan transaksi Anda agar lebih terorganisir.</p>
      </div>
      <Button @click="openAdd" class="bg-foreground text-background hover:bg-foreground/90 shadow-lg px-6 h-12 rounded-full transition-all hover:scale-105 active:scale-95"> 
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
                         <span v-if="item.isEmoji" class="text-3xl leading-none filter drop-shadow-sm">{{ item.icon }}</span>
                         <component v-else :is="getIconComponent(item.icon)" class="h-7 w-7" />
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
      <DialogContent class="max-w-md bg-card p-0 overflow-hidden border-border shadow-2xl">
        <DialogHeader class="p-6 border-b">
          <DialogTitle>{{ isEditMode ? "Edit Kategori" : "Tambah Kategori" }}</DialogTitle>
          <DialogDescription>Sesuaikan nama dan visualisasi kategori.</DialogDescription>
        </DialogHeader>

        <div class="p-6 space-y-6">
          <div class="grid gap-2">
            <Label class="text-sm font-semibold opacity-70">Nama Kategori</Label>
            <Input v-model="form.name" placeholder="Misal: Belanja, Gaji" class="h-11 bg-background shadow-sm" />
          </div>

          <div class="grid gap-2">
            <Label class="text-sm font-semibold opacity-70">Tipe Kategori</Label>
            <Select v-model="form.type">
              <SelectTrigger class="h-11 bg-background border-border">
                <SelectValue placeholder="Pilih Tipe" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="expense">📉 Pengeluaran (Expense)</SelectItem>
                <SelectItem value="income">📈 Pemasukan (Income)</SelectItem>
              </SelectContent>
            </Select>
          </div>

          <div class="grid gap-2 text-foreground">
            <Label class="text-sm font-semibold opacity-70">Icon / Emoji</Label>
            <button
              @click="isIconPickerOpen = true"
              type="button"
              class="w-full h-28 flex items-center justify-center border-dashed border-2 rounded-2xl hover:bg-accent/30 transition-all gap-4 bg-background border-border shadow-sm group"
            >
              <template v-if="!form.icon">
                <div class="h-12 w-12 rounded-full bg-muted flex items-center justify-center text-muted-foreground group-hover:scale-110 transition-transform">
                  <Plus class="h-6 w-6" />
                </div>
                <span class="text-sm text-muted-foreground font-medium italic">Pilih icon...</span>
              </template>
              <template v-else>
                <div :class="['h-16 w-16 rounded-2xl flex items-center justify-center text-4xl shadow-md transform group-hover:scale-105 transition-transform', form.type === 'expense' ? 'bg-red-50 text-red-500' : 'bg-emerald-50 text-emerald-600']">
                  <span v-if="form.isEmoji" class="leading-none">{{ form.icon }}</span>
                  <component v-else :is="getIconComponent(form.icon)" class="h-8 w-8" />
                </div>
                 <div class="text-left">
                    <p class="text-xs font-bold uppercase opacity-50">Icon Terpilih</p>
                    <p class="text-sm font-semibold">Klik untuk ganti</p>
                </div>
              </template>
            </button>
          </div>
        </div>

        <DialogFooter class="p-6 border-t bg-muted/5 flex flex-row items-center justify-between gap-2">
          <Button v-if="isEditMode" variant="ghost" type="button" class="text-red-500 hover:text-red-600 hover:bg-red-50 gap-2 px-4" @click="handleDelete"> <Trash2 class="w-4 h-4" /> Hapus </Button>
          <div class="flex gap-2 ml-auto">
            <Button variant="outline" type="button" @click="isDialogOpen = false">Batal</Button>
            <Button @click="handleSave" type="button" class="bg-foreground text-background px-6 shadow-md hover:bg-foreground/90">
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
              <Button v-for="item in iconOptions" :key="item.name" variant="ghost" type="button" class="h-20 flex flex-col gap-2 hover:bg-primary/10" @click="selectIcon(item.name, false)">
                <component :is="item.icon" class="h-6 w-6" />
                <span class="text-[9px] font-medium opacity-60 truncate w-full">{{ item.label }}</span>
              </Button>
            </div>
          </TabsContent>
          <TabsContent value="emojis" class="flex-1 overflow-y-auto p-6 mt-0">
            <div v-for="(list, cat) in emojiCategories" :key="cat" class="mb-6">
              <p class="text-[10px] font-bold text-muted-foreground uppercase mb-3 text-left tracking-widest">{{ cat }}</p>
              <div class="grid grid-cols-4 gap-4">
                <button v-for="e in list" :key="e" type="button" class="text-4xl p-2 hover:bg-accent rounded-2xl transition-transform active:scale-90" @click="selectIcon(e, true)">{{ e }}</button>
              </div>
            </div>
          </TabsContent>
        </Tabs>
      </DialogContent>
    </Dialog>
  </div>
</template>
