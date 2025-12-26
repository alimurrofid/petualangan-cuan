<script setup lang="ts">
import { ref, onMounted, computed } from "vue";

// UI Components
import { Card } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogFooter, DialogDescription } from "@/components/ui/dialog";

// Icons
import * as LucideIcons from "lucide-vue-next";
import { Plus, Pencil, Trash2, LayoutGrid, Save, ArrowUpRight, ArrowDownLeft } from "lucide-vue-next";

interface CategoryItem {
  id: number;
  name: string;
  icon: string;
  isEmoji: boolean;
  type: "income" | "expense";
}

const categories = ref<CategoryItem[]>([]);
const filterType = ref<"expense" | "income">("expense");

// State untuk Popup
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
    ];
    categories.value = initial;
    localStorage.setItem("mock_categories", JSON.stringify(initial));
  }
};

onMounted(loadCategories);

const filteredCategories = computed(() => {
  return categories.value.filter((c) => c.type === filterType.value);
});

const openAdd = () => {
  isEditMode.value = false;
  form.value = { id: Date.now(), name: "", icon: "", isEmoji: false, type: filterType.value };
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
</script>

<template>
  <div class="p-6 space-y-6 text-foreground">
    <div class="flex flex-col md:flex-row md:items-center justify-between gap-4">
      <div>
        <h2 class="text-2xl font-bold tracking-tight">Kategori</h2>
        <p class="text-sm text-muted-foreground">Atur kategori pemasukan dan pengeluaran Anda.</p>
      </div>
      <Button @click="openAdd" class="bg-emerald-600 hover:bg-emerald-700 text-white shadow-sm"> <Plus class="w-4 h-4 mr-2" /> Tambah Kategori </Button>
    </div>

    <div class="flex justify-center md:justify-start">
      <div class="inline-flex p-1 bg-muted rounded-xl border border-border shadow-sm">
        <button
          @click="filterType = 'expense'"
          type="button"
          :class="['px-6 py-2 text-sm font-bold rounded-lg transition-all flex items-center gap-2', filterType === 'expense' ? 'bg-background text-red-500 shadow-sm' : 'text-muted-foreground hover:text-foreground']"
        >
          <ArrowDownLeft class="w-4 h-4" /> Pengeluaran
        </button>
        <button
          @click="filterType = 'income'"
          type="button"
          :class="['px-6 py-2 text-sm font-bold rounded-lg transition-all flex items-center gap-2', filterType === 'income' ? 'bg-background text-emerald-600 shadow-sm' : 'text-muted-foreground hover:text-foreground']"
        >
          <ArrowUpRight class="w-4 h-4" /> Pemasukan
        </button>
      </div>
    </div>

    <div class="grid gap-4 grid-cols-2 md:grid-cols-4 lg:grid-cols-6">
      <Card
        v-for="item in filteredCategories"
        :key="item.id"
        @click="openEdit(item)"
        class="cursor-pointer transition-all hover:ring-2 hover:ring-emerald-500 hover:shadow-lg active:scale-95 bg-card border-border flex flex-col items-center justify-center p-6 text-center aspect-square rounded-2xl group"
      >
        <div class="h-14 w-14 rounded-2xl bg-emerald-100 dark:bg-emerald-900/30 flex items-center justify-center text-emerald-600 mb-3 shadow-sm group-hover:scale-110 transition-transform">
          <span v-if="item.isEmoji" class="text-3xl">{{ item.icon }}</span>
          <component v-else :is="getIconComponent(item.icon)" class="h-7 w-7" />
        </div>
        <p class="text-sm font-bold truncate w-full px-2">{{ item.name }}</p>
      </Card>

      <div v-if="filteredCategories.length === 0" class="col-span-full py-20 text-center text-muted-foreground border-2 border-dashed rounded-3xl opacity-50 italic">
        Belum ada kategori {{ filterType === "income" ? "pemasukan" : "pengeluaran" }}
      </div>
    </div>

    <Dialog v-model:open="isDialogOpen">
      <DialogContent class="max-w-md bg-card p-0 overflow-hidden border-border shadow-2xl">
        <DialogHeader class="p-6 border-b">
          <DialogTitle>{{ isEditMode ? "Edit Kategori" : "Tambah Kategori" }}</DialogTitle>
          <DialogDescription>Sesuaikan nama, tipe, dan visual kategori transaksi.</DialogDescription>
        </DialogHeader>

        <div class="p-6 space-y-6">
          <div class="grid gap-2">
            <Label class="text-sm font-semibold opacity-70">Nama Kategori</Label>
            <Input v-model="form.name" placeholder="Misal: Belanja, Gaji" class="h-11 bg-background" />
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

          <div class="grid gap-2">
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
                <span class="text-xs text-muted-foreground font-medium">Pilih icon/emoji</span>
              </template>
              <template v-else>
                <div class="h-16 w-16 rounded-2xl bg-emerald-100 dark:bg-emerald-900/30 flex items-center justify-center text-emerald-600 shadow-sm">
                  <span v-if="form.isEmoji" class="text-4xl leading-none">{{ form.icon }}</span>
                  <component v-else :is="getIconComponent(form.icon)" class="h-9 w-9" />
                </div>
                <span class="text-[10px] uppercase tracking-widest font-bold opacity-40 mt-1">Klik untuk ganti</span>
              </template>
            </button>
          </div>
        </div>

        <DialogFooter class="p-6 border-t bg-muted/5 flex flex-row items-center justify-between gap-2">
          <Button v-if="isEditMode" variant="ghost" type="button" class="text-red-500 hover:text-red-600 hover:bg-red-50 gap-2" @click="handleDelete"> <Trash2 class="w-4 h-4" /> Hapus </Button>
          <div class="flex gap-2 ml-auto">
            <Button variant="outline" type="button" @click="isDialogOpen = false">Batal</Button>
            <Button @click="handleSave" type="button" class="bg-emerald-600 hover:bg-emerald-700 text-white px-8">
              <template v-if="isEditMode"> <Pencil class="w-4 h-4 mr-2" /> Edit </template>
              <template v-else> <Save class="w-4 h-4 mr-2" /> Selesai </template>
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
            <TabsList class="grid w-full grid-cols-2 shadow-sm">
              <TabsTrigger value="icons">Icons</TabsTrigger>
              <TabsTrigger value="emojis">Emojis</TabsTrigger>
            </TabsList>
          </div>
          <TabsContent value="icons" class="flex-1 overflow-y-auto p-6 mt-0">
            <div class="grid grid-cols-4 gap-4">
              <Button v-for="item in iconOptions" :key="item.name" variant="ghost" type="button" class="h-20 flex flex-col gap-2 hover:bg-emerald-50 dark:hover:bg-emerald-900/40" @click="selectIcon(item.name, false)">
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
