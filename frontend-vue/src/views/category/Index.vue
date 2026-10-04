<script setup lang="ts">
import { ref, onMounted, computed, watch } from "vue";
import { useCategoryStore, type Category } from "@/stores/category";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogFooter, DialogDescription } from "@/components/ui/dialog";
import { useSwal } from "@/composables/useSwal";

import { emojiCategories, getEmoji, getIconComponent, categoryIcons } from "@/lib/icons";
import { Plus, Pencil, Trash2, LayoutGrid, Save, TrendingUp, TrendingDown, Search } from "lucide-vue-next";
import { formatCurrencyInput, parseCurrencyInput, formatCurrencyLive } from "@/lib/utils";

interface CategoryForm {
  id: number;
  name: string;
  icon: string;
  type: "income" | "expense";
  budgetLimit?: number;
  seq?: number;
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
  seq: 1,
});
const budgetLimitDisplay = ref("");

const errors = ref({
  name: false,
  icon: false,
});



onMounted(() => {
  categoryStore.fetchCategories();
});

const filteredCategories = computed(() => {
  return categoryStore.categories.filter((c) => c.type === currentTab.value);
});

const isSubmitting = ref(false);

const openAdd = () => {
  isEditMode.value = false;
  const maxSeq = categoryStore.categories.length ? Math.max(...categoryStore.categories.map((c: any) => c.seq ?? 0)) : 0;
  form.value = { id: 0, name: "", icon: "", type: currentTab.value, budgetLimit: 0, seq: maxSeq + 1 };
  budgetLimitDisplay.value = "";
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
    budgetLimit: category.budget_limit || 0,
    seq: category.seq ?? 0,
  };
  budgetLimitDisplay.value = category.budget_limit ? formatCurrencyInput(category.budget_limit) : "";
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

const filteredCategoryIcons = computed(() => {
  if (!iconSearch.value) return categoryIcons;
  const q = iconSearch.value.toLowerCase();
  return categoryIcons.filter(i => i.name.toLowerCase().includes(q) || i.label.toLowerCase().includes(q));
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
    if (errors.value.name) msg += "<br>- Nama Kategori";
    if (errors.value.icon) msg += "<br>- Icon Kategori";
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
    icon: form.value.icon,
    budget_limit: Number(form.value.budgetLimit),
    seq: Number(form.value.seq) || 0,
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

const getGradientIcon = (type: string) => {
  return type === 'expense'
    ? 'bg-gradient-to-br from-red-50 to-red-100 text-red-600 dark:from-red-900 dark:to-red-800 dark:text-red-100'
    : 'bg-gradient-to-br from-emerald-50 to-emerald-100 text-emerald-600 dark:from-emerald-900 dark:to-emerald-800 dark:text-emerald-100';
};

// Live Format for Budget Limit
watch(budgetLimitDisplay, (val) => {
  const formatted = formatCurrencyLive(val);
  if (formatted !== val) {
    budgetLimitDisplay.value = formatted;
    return;
  }
  const num = parseCurrencyInput(val);
  form.value.budgetLimit = num;
});

const onBudgetBlur = () => {
  const num = parseCurrencyInput(budgetLimitDisplay.value);
  if (num) budgetLimitDisplay.value = formatCurrencyInput(num);
};
</script>

<template>
  <div class="flex-1 pt-2 space-y-6" v-if="categoryStore.isLoading">
    <div class="flex items-center justify-center min-h-[400px]">
      <p class="text-muted-foreground animate-pulse">Memuat data kategori...</p>
    </div>
  </div>
  <div class="flex-1 pt-2 space-y-6 text-foreground" v-else>
    <div class="flex flex-col justify-between gap-6 md:flex-row md:items-end">
      <div>
        <h2 class="text-3xl font-bold tracking-tight">Kategori</h2>
        <p class="mt-1 text-muted-foreground">Klasifikasikan transaksi Anda agar lebih terorganisir.</p>
      </div>
      <Button @click="openAdd"
        class="h-12 px-6 text-white transition-all rounded-full shadow-lg bg-gradient-to-r from-emerald-600 to-teal-500 hover:from-emerald-500 hover:to-teal-400 hover:scale-105 active:scale-95">
        <Plus class="w-5 h-5 mr-2" />
        Tambah Kategori
      </Button>
    </div>

    <Tabs v-model:model-value="currentTab" class="space-y-6">
      <TabsList class="grid w-full h-auto grid-cols-2 p-1 rounded-full bg-muted/60">
        <TabsTrigger value="income"
          class="rounded-full data-[state=active]:bg-emerald-500 data-[state=active]:text-white dark:data-[state=active]:bg-emerald-600 dark:data-[state=active]:text-white data-[state=active]:shadow-md transition-all py-2.5 font-bold hover:bg-emerald-500/10 data-[state=active]:hover:bg-emerald-600">
          <TrendingUp class="w-4 h-4 mr-2" /> Pemasukan
        </TabsTrigger>
        <TabsTrigger value="expense"
          class="rounded-full data-[state=active]:bg-red-500 data-[state=active]:text-white dark:data-[state=active]:bg-red-600 dark:data-[state=active]:text-white data-[state=active]:shadow-md transition-all py-2.5 font-bold hover:bg-red-500/10 data-[state=active]:hover:bg-red-600">
          <TrendingDown class="w-4 h-4 mr-2" /> Pengeluaran
        </TabsTrigger>
      </TabsList>

      <div
        class="grid grid-cols-1 gap-6 duration-500 sm:grid-cols-2 lg:grid-cols-4 xl:grid-cols-5 animate-in fade-in slide-in-from-bottom-4">
        <div v-for="item in filteredCategories" :key="item.id" @click="openEdit(item)"
          :class="['group relative bg-card rounded-3xl p-5 cursor-pointer transition-all duration-300 hover:-translate-y-2 hover:shadow-xl border border-border/60 hover:border-border flex flex-col gap-4', item.type === 'expense' ? 'hover:border-red-200 dark:hover:border-red-500/50' : 'hover:border-emerald-200 dark:hover:border-emerald-500/50']">
          <!-- Top Section: Type Label & Edit Hint -->
          <div class="flex items-start justify-between">
            <div
              :class="['px-2.5 py-1 rounded-lg text-[10px] uppercase font-bold tracking-widest border', item.type === 'expense' ? 'bg-red-50 text-red-600 border-red-100 dark:bg-red-900/20 dark:text-red-300 dark:border-red-900/30' : 'bg-emerald-50 text-emerald-600 border-emerald-100 dark:bg-emerald-900/20 dark:text-emerald-300 dark:border-emerald-900/30']">
              {{ item.type === 'expense' ? 'Pengeluaran' : 'Pemasukan' }}
            </div>
            <div class="transition-opacity opacity-0 group-hover:opacity-100">
              <div
                class="bg-slate-100 dark:bg-slate-800 p-1.5 rounded-full transform rotate-12 group-hover:rotate-0 transition-transform shadow-sm">
                <Pencil class="w-3.5 h-3.5 text-blue-600 dark:text-blue-400" />
              </div>
            </div>
          </div>

          <!-- Main Content: Icon & Title -->
          <div class="flex items-center gap-4">
            <!-- Icon -->
            <div
              :class="['h-14 w-14 shrink-0 rounded-2xl flex items-center justify-center border shadow-sm transition-transform group-hover:scale-110', getGradientIcon(item.type), item.type === 'expense' ? 'border-red-100 dark:border-red-800' : 'border-emerald-100 dark:border-emerald-800']">
              <component v-if="getIconComponent(item.icon)" :is="getIconComponent(item.icon)" class="h-7 w-7" />
              <span v-else-if="getEmoji(item.icon)" class="text-3xl leading-none filter drop-shadow-sm">{{
                getEmoji(item.icon) }}</span>
              <component v-else :is="getIconComponent(null, 'LayoutGrid')" class="h-7 w-7" />
            </div>

            <!-- Title & Target -->
            <div class="flex-1 min-w-0 space-y-1">
              <p
                class="text-lg font-bold leading-tight tracking-wide truncate transition-colors text-foreground group-hover:text-primary">
                {{ item.name }}</p>
              <p v-if="item.type === 'expense' && item.budget_limit"
                class="flex items-center gap-1 text-xs text-muted-foreground">
                <span class="opacity-70">Target:</span>
                <span class="font-medium">{{ new Intl.NumberFormat("id-ID", {
                  style: "currency", currency: "IDR",
                  maximumFractionDigits: 0
                }).format(item.budget_limit) }}</span>
              </p>
            </div>
          </div>
        </div>

        <div v-if="filteredCategories.length === 0"
          class="flex flex-col items-center justify-center py-20 border-2 border-dashed opacity-50 col-span-full text-muted-foreground border-border rounded-3xl bg-muted/20">
          <div class="flex items-center justify-center w-16 h-16 mb-4 rounded-full bg-muted">
            <LayoutGrid class="w-8 h-8 opacity-40" />
          </div>
          <p class="text-lg font-medium">Belum ada kategori {{ currentTab === "income" ? "pemasukan" : "pengeluaran" }}
          </p>
          <p class="text-sm">Klik tombol tambah untuk membuat baru.</p>
        </div>
      </div>
    </Tabs>

    <Dialog v-model:open="isDialogOpen">
      <DialogContent class="max-w-md p-0 overflow-hidden shadow-2xl bg-card border-border"
        @interact-outside="swal.handleSwalInteractOutside">
        <DialogHeader class="p-6 border-b">
          <DialogTitle>{{ isEditMode ? "Edit Kategori" : "Tambah Kategori" }}</DialogTitle>
          <DialogDescription>Sesuaikan nama dan visualisasi kategori.</DialogDescription>
        </DialogHeader>

        <div class="p-6 space-y-6">
          <div class="grid gap-2">
            <Label class="text-sm font-semibold opacity-70">Nama Kategori</Label>
            <Input v-model="form.name" placeholder="Misal: Belanja, Gaji"
              :class="['h-11 bg-background shadow-sm', errors.name ? 'border-red-500 ring-1 ring-red-500' : '']"
              :disabled="isSubmitting" />
            <span v-if="errors.name" class="text-xs font-medium text-red-500">Nama kategori wajib diisi</span>
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
            <Input v-model="budgetLimitDisplay" @blur="onBudgetBlur" type="text" inputmode="decimal" placeholder="Rp 0"
              class="shadow-sm h-11 bg-background" :disabled="isSubmitting" />
            <p class="text-[10px] text-muted-foreground">Isi 0 jika tidak ingin membatasi pengeluaran.</p>
          </div>

          <div class="grid gap-2">
            <Label class="text-sm font-semibold opacity-70">Urutan</Label>
            <Input v-model.number="form.seq" type="number" placeholder="Contoh: 1"
              class="shadow-sm h-11 bg-background" :disabled="isSubmitting" />
            <p class="text-[10px] text-muted-foreground">Menentukan urutan kategori di menu dan dropdown pilihan (angka lebih kecil muncul lebih awal).</p>
          </div>

          <div class="grid gap-2 text-foreground">
            <Label class="text-sm font-semibold opacity-70">Icon / Emoji</Label>
            <button @click="isIconPickerOpen = true" type="button"
              :class="['w-full h-28 flex items-center justify-center border-dashed border-2 rounded-2xl hover:bg-accent/30 transition-all gap-4 bg-background border-border shadow-sm group', errors.icon ? 'border-red-500 bg-red-50/10' : '', isSubmitting ? 'opacity-50 cursor-not-allowed' : '']"
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
                  :class="['h-16 w-16 rounded-2xl flex items-center justify-center text-4xl shadow-md transform group-hover:scale-105 transition-transform', form.type === 'expense' ? 'bg-red-50 text-red-500' : 'bg-emerald-50 text-emerald-600']">
                  <component v-if="getIconComponent(form.icon)" :is="getIconComponent(form.icon)" class="w-8 h-8" />
                  <span v-else-if="getEmoji(form.icon)" class="leading-none">{{ getEmoji(form.icon) }}</span>
                  <component v-else :is="getIconComponent(null, 'LayoutGrid')" class="w-8 h-8" />
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
              class="px-6 text-white shadow-md bg-gradient-to-r from-emerald-600 to-teal-500 hover:from-emerald-500 hover:to-teal-400 hover:bg-foreground/90"
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
          <DialogTitle class="text-sm font-bold">Visual Kategori</DialogTitle>
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
              <Button v-for="item in filteredCategoryIcons" :key="item.name" variant="ghost" type="button"
                class="flex flex-col h-20 gap-2 hover:bg-primary/10" @click="selectIcon(item.name)">
                <component :is="item.icon" class="w-6 h-6" />
                <span class="text-[9px] font-medium opacity-60 truncate w-full">{{ item.label }}</span>
              </Button>
            </div>
            <p v-if="filteredCategoryIcons.length === 0" class="py-8 text-sm text-center text-muted-foreground">Tidak
              ada
              icon cocok.</p>
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
