<script setup lang="ts">
import { onMounted, ref, computed, watch } from "vue";
import { useWishlistStore, type WishlistItem } from "@/stores/wishlist";
import { useCategoryStore } from "@/stores/category";
import { useAuthStore } from "@/stores/auth";
import ManualTransactionDialog from "@/components/ManualTransactionDialog.vue";
import { Button } from "@/components/ui/button";
import { Card, CardHeader, CardTitle, CardContent } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogFooter } from "@/components/ui/dialog";
import { Label } from "@/components/ui/label";
import { Input } from "@/components/ui/input";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import SearchableSelect from "@/components/ui/searchable-select/SearchableSelect.vue";
import { Plus, Trash2, Pencil, ShoppingCart, CheckCircle, Clock, Flame, Zap } from "lucide-vue-next";
import { useSwal } from "@/composables/useSwal";
import { getEmoji, getIconComponent } from "@/lib/icons";
import { parseCurrencyInput, formatCurrencyInput, formatCurrencyLive } from "@/lib/utils";

const wishlistStore = useWishlistStore();
const categoryStore = useCategoryStore();
const authStore = useAuthStore();
const swal = useSwal();

onMounted(() => {
    wishlistStore.fetchItems();
    categoryStore.fetchCategories();
});

// State for Add/Edit Dialog
const isDialogOpen = ref(false);
const isEditing = ref(false);
const editingId = ref<number | null>(null);
const form = ref({
    name: "",
    category_id: "",
    estimated_price: "",
    priority: "low" as "low" | "medium" | "high"
});
const estimatedPriceDisplay = ref("");
const isSubmitting = ref(false);

// State for Buy Dialog
const isBuyDialogOpen = ref(false);
const selectedItemToBuy = ref<WishlistItem | null>(null);

const activeItems = computed(() => wishlistStore.items.filter(i => !i.is_bought));
const boughtItems = computed(() => wishlistStore.items.filter(i => i.is_bought));

const openAddDialog = () => {
    isEditing.value = false;
    editingId.value = null;
    form.value = { name: "", category_id: "", estimated_price: "", priority: "low" };
    estimatedPriceDisplay.value = "";
    isDialogOpen.value = true;
};

const openEditDialog = (item: WishlistItem) => {
    isEditing.value = true;
    editingId.value = item.id;
    form.value = {
        name: item.name,
        category_id: String(item.category_id),
        estimated_price: String(item.estimated_price),
        priority: item.priority
    };
    estimatedPriceDisplay.value = formatCurrencyInput(item.estimated_price);
    isDialogOpen.value = true;
};

const handleSave = async () => {
    if (!form.value.name || !form.value.category_id || !form.value.estimated_price) {
        swal.error("Gagal", "Mohon lengkapi semua field wajib");
        return;
    }

    isSubmitting.value = true;
    try {
        const payload = {
            name: form.value.name,
            category_id: Number(form.value.category_id),
            estimated_price: Number(form.value.estimated_price),
            priority: form.value.priority
        };

        let success = false;
        if (isEditing.value && editingId.value) {
            success = await wishlistStore.updateItem(editingId.value, payload);
        } else {
            success = await wishlistStore.createItem(payload);
        }

        if (success) {
            isDialogOpen.value = false;
        }
    } finally {
        isSubmitting.value = false;
    }
};

const handleDelete = async (id: number) => {
    await wishlistStore.deleteItem(id);
};

const handleBuy = (item: WishlistItem) => {
    selectedItemToBuy.value = item;
    isBuyDialogOpen.value = true;
};

const onTransactionSaved = () => {
    // This is called when ManualTransactionDialog successfully saves.
    // It should also trigger markAsBought via internal logic we added to dialog,
    // BUT we need to refresh wishlist list to see the update.
    wishlistStore.fetchItems();
    isBuyDialogOpen.value = false;
    selectedItemToBuy.value = null;
};

// Formatting Helper
const formatRp = (val: number) => new Intl.NumberFormat("id-ID", { style: "currency", currency: "IDR", maximumFractionDigits: 2 }).format(val);



// Sync Display -> Model
watch(estimatedPriceDisplay, (val) => {
    const formatted = formatCurrencyLive(val);
    if (formatted !== val) {
        estimatedPriceDisplay.value = formatted;
        return;
    }
    const num = parseCurrencyInput(val);
    form.value.estimated_price = num.toString();
});

const onPriceBlur = () => {
    const num = parseCurrencyInput(estimatedPriceDisplay.value);
    if(num) estimatedPriceDisplay.value = formatCurrencyInput(num);
};

const categoryOptions = computed(() => categoryStore.categories.filter(c => c.type === 'expense').map(c => ({
    value: String(c.id),
    label: c.name,
    icon: c.icon
})));

</script>

<template>
  <div class="flex-1 space-y-6 pt-2" v-if="wishlistStore.isLoading">
      <div class="flex items-center justify-center min-h-[400px]">
          <p class="text-muted-foreground animate-pulse">Memuat data wishlist...</p>
      </div>
  </div>
  <div class="flex-1 space-y-6 pt-2 text-foreground" v-else>
        <div class="flex flex-col sm:flex-row justify-between items-start sm:items-center gap-4 sm:gap-0">
            <div>
                <h2 class="text-3xl font-bold tracking-tight">Wishlist</h2>
                <p class="text-sm text-muted-foreground mt-1">Simpan dan wujudkan impian finansial Anda.</p>
            </div>
            <Button @click="openAddDialog" class="bg-gradient-to-r from-emerald-600 to-teal-500 text-white hover:from-emerald-500 hover:to-teal-400 shadow-lg h-12 rounded-full transition-all hover:scale-105 active:scale-95 px-6">
                <Plus class="mr-2 h-5 w-5" /> Tambah Keinginan
            </Button>
        </div>

        <!-- Active Wishlist -->
        <div class="space-y-4">
             <div class="flex items-center gap-2">
                <h3 class="text-base font-bold flex items-center gap-2">
                    <Clock class="h-5 w-5 text-muted-foreground" />
                    Sedang Diusahakan
                </h3>
                <span class="text-[10px] font-bold uppercase tracking-widest px-2 py-0.5 rounded-full bg-muted text-muted-foreground border border-border/50">{{ activeItems.length }} item</span>
            </div>

            <div class="grid gap-6 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-3">
                <Card 
                    v-for="item in activeItems" 
                    :key="item.id" 
                    class="group relative overflow-hidden transition-all duration-300 hover:shadow-lg hover:-translate-y-1 hover:border-emerald-200 dark:hover:border-emerald-900"
                >
                    <CardHeader class="pb-3">
                        <div class="flex justify-between items-start">
                            <div class="space-y-1.5">
                                <div class="flex flex-wrap gap-2 items-center">
                                    <Badge variant="outline" class="px-2 py-0.5 text-[10px] font-bold uppercase tracking-widest gap-1 border-border bg-muted/50 text-muted-foreground">
                                        <component v-if="item.category && getIconComponent(item.category.icon)" :is="getIconComponent(item.category.icon)" class="h-3 w-3" />
                                        <span v-else-if="item.category">{{ getEmoji(item.category.icon) || '📦' }}</span>
                                        <span>{{ item.category?.name }}</span>
                                    </Badge>
                                    
                                     <Badge v-if="item.priority === 'high'" class="bg-red-100 dark:bg-red-500/20 text-red-700 dark:text-red-400 border-red-200 dark:border-red-800 text-[10px] gap-1 px-2 py-0.5">
                                        <Flame class="w-3 h-3" /> Mendesak
                                    </Badge>
                                    <Badge v-else-if="item.priority === 'medium'" class="bg-amber-100 dark:bg-amber-500/20 text-amber-700 dark:text-amber-400 border-amber-200 dark:border-amber-800 text-[10px] gap-1 px-2 py-0.5">
                                        <Zap class="w-3 h-3" /> Butuh
                                    </Badge>
                                    <Badge v-else class="bg-blue-100 dark:bg-blue-500/20 text-blue-700 dark:text-blue-400 border-blue-200 dark:border-blue-800 text-[10px] gap-1 px-2 py-0.5">
                                        <Clock class="w-3 h-3" /> Santai
                                    </Badge>
                                </div>
                                <CardTitle class="text-xl font-bold tracking-tight">{{ item.name }}</CardTitle>
                            </div>
                            <div class="flex gap-1">
                                <Button variant="ghost" size="icon" class="h-8 w-8 text-muted-foreground hover:bg-slate-100 hover:text-blue-600 dark:hover:bg-slate-800" @click="openEditDialog(item)">
                                    <Pencil class="h-4 w-4" />
                                </Button>
                                <Button variant="ghost" size="icon" class="h-8 w-8 text-muted-foreground hover:bg-red-50 hover:text-red-600 dark:hover:bg-red-900/20" @click="handleDelete(item.id)">
                                    <Trash2 class="h-4 w-4" />
                                </Button>
                            </div>
                        </div>
                    </CardHeader>

                    <CardContent class="space-y-4 pt-0">
                        <div class="p-4 rounded-xl bg-muted/50 border border-border mt-2">
                            <p class="text-[10px] uppercase font-bold text-muted-foreground tracking-widest mb-1">Estimasi Harga</p>
                            <div class="text-2xl font-mono font-bold tracking-tight text-foreground" :class="{ 'privacy-blur': authStore.isPrivacyMode }">
                                {{ formatRp(item.estimated_price) }}
                            </div>
                        </div>

                        <Button class="w-full rounded-xl bg-gradient-to-r from-emerald-600 to-teal-500 text-white hover:from-emerald-500 hover:to-teal-400 shadow-sm border-0 font-bold h-11" @click="handleBuy(item)">
                            <ShoppingCart class="mr-2 h-4 w-4" /> Beli Sekarang
                        </Button>
                    </CardContent>
                </Card>
                
                <div v-if="activeItems.length === 0" class="col-span-full text-center py-20 text-muted-foreground border-2 border-dashed border-muted rounded-3xl bg-muted/10 h-80 flex flex-col items-center justify-center">
                    <div class="h-16 w-16 bg-muted rounded-full flex items-center justify-center mb-4">
                        <Clock class="h-8 w-8 opacity-40" />
                    </div>
                    <p class="font-medium text-lg">Belum ada keinginan aktif.</p>
                    <p class="text-sm opacity-70">Mulai catat wishlist impianmu sekarang.</p>
                    <Button @click="openAddDialog()" variant="link" class="mt-2 text-emerald-600">Tambah Baru</Button>
                </div>
            </div>
        </div>

        <div v-if="boughtItems.length > 0" class="space-y-4 pt-8 border-t border-border">
            <div class="flex items-center gap-2">
                <h2 class="text-base font-bold flex items-center gap-2 text-emerald-600 uppercase tracking-widest">
                    <CheckCircle class="h-5 w-5" /> Sudah Terwujud
                </h2>
                <span class="text-[10px] font-bold uppercase tracking-widest px-2 py-0.5 rounded-full bg-emerald-50 text-emerald-700 border border-emerald-200">{{ boughtItems.length }} item</span>
            </div>
            
            <div class="grid gap-6 md:grid-cols-2 lg:grid-cols-3">
                 <Card 
                    v-for="item in boughtItems" 
                    :key="item.id" 
                    class="group relative overflow-hidden transition-all duration-300 hover:shadow-md hover:border-emerald-200 dark:hover:border-emerald-900 opacity-75 hover:opacity-100"
                >
                    <CardHeader class="pb-3">
                        <div class="flex justify-between items-start">
                            <div class="space-y-1">
                                <CardTitle class="text-xl font-bold tracking-tight text-muted-foreground flex items-center gap-2">
                                     <div class="p-1.5 bg-emerald-100 dark:bg-emerald-500/20 rounded-md text-emerald-600 dark:text-emerald-400">
                                        <CheckCircle class="w-4 h-4" />
                                    </div>
                                    <span class="line-through decoration-muted-foreground/50">{{ item.name }}</span>
                                </CardTitle>
                                <div class="flex items-center gap-1.5 text-[10px] font-medium uppercase tracking-widest text-muted-foreground pl-1">
                                    <component v-if="item.category && getIconComponent(item.category.icon)" :is="getIconComponent(item.category.icon)" class="h-3 w-3" />
                                    <span v-else-if="item.category">{{ getEmoji(item.category.icon) || '📦' }}</span>
                                    <span>{{ item.category?.name }}</span>
                                </div>
                            </div>
                            <Button variant="ghost" size="icon" class="h-8 w-8 text-muted-foreground hover:bg-red-50 hover:text-red-600 dark:hover:bg-red-900/20" @click="handleDelete(item.id)">
                                <Trash2 class="h-4 w-4" />
                            </Button>
                        </div>
                    </CardHeader>
                    
                    <CardContent class="space-y-4 pt-0">
                         <div class="p-4 rounded-xl bg-muted/30 border border-border space-y-1 mt-2">
                             <div class="flex justify-between text-[10px] font-bold uppercase tracking-widest text-muted-foreground">
                                   <span>Estimasi Harga</span>
                                   <span class="text-emerald-600 flex items-center gap-1"><CheckCircle class="w-3 h-3" /> Terbeli</span>
                               </div>
                            <div class="text-2xl font-mono font-bold tracking-tight text-muted-foreground" :class="{ 'privacy-blur': authStore.isPrivacyMode }">
                                {{ formatRp(item.estimated_price) }}
                            </div>
                        </div>
                    </CardContent>
                </Card>
            </div>
        </div>

        <!-- Add/Edit Dialog -->
        <Dialog :open="isDialogOpen" @update:open="isDialogOpen = $event">
            <DialogContent class="sm:max-w-[425px] rounded-3xl bg-card text-foreground">
                <DialogHeader>
                    <DialogTitle>{{ isEditing ? 'Edit Keinginan' : 'Tambah Keinginan Baru' }}</DialogTitle>
                </DialogHeader>
                <div class="space-y-4 py-4">
                    <div class="space-y-2">
                        <Label class="text-xs font-bold uppercase tracking-widest text-muted-foreground">Nama Barang/Jasa</Label>
                        <Input v-model="form.name" placeholder="Misal: iPhone 15, Liburan ke Bali" class="h-11 shadow-sm rounded-xl bg-background" />
                    </div>
                    <div class="space-y-2">
                        <Label class="text-xs font-bold uppercase tracking-widest text-muted-foreground">Kategori</Label>
                        <SearchableSelect
                            v-model="form.category_id"
                            :options="categoryOptions"
                            placeholder="Pilih Kategori"
                        >
                            <template #option="{ option }">
                                <div class="flex items-center gap-2">
                                    <component v-if="getIconComponent(option.icon)" :is="getIconComponent(option.icon)" class="h-4 w-4 shrink-0" />
                                    <span v-else class="text-xs shrink-0">{{ getEmoji(option.icon) || '📦' }}</span>
                                    <span>{{ option.label }}</span>
                                </div>
                            </template>
                        </SearchableSelect>
                    </div>
                    <div class="space-y-2">
                        <Label class="text-xs font-bold uppercase tracking-widest text-muted-foreground">Estimasi Harga</Label>
                        <Input type="text" inputmode="decimal" v-model="estimatedPriceDisplay" @blur="onPriceBlur" placeholder="Rp 0" class="h-11 shadow-sm rounded-xl bg-background" />
                    </div>
                    <div class="space-y-2">
                        <Label class="text-xs font-bold uppercase tracking-widest text-muted-foreground">Prioritas</Label>
                         <Select v-model="form.priority">
                            <SelectTrigger class="w-full h-11 rounded-xl bg-background shadow-sm">
                                <template v-if="form.priority">
                                    <div class="flex items-center gap-2" v-if="form.priority === 'low'">
                                        <Clock class="h-4 w-4 text-green-500" />
                                        <span>Rendah (Santai)</span>
                                    </div>
                                    <div class="flex items-center gap-2" v-else-if="form.priority === 'medium'">
                                        <Zap class="h-4 w-4 text-yellow-500" />
                                        <span>Sedang (Butuh)</span>
                                    </div>
                                    <div class="flex items-center gap-2" v-else-if="form.priority === 'high'">
                                        <Flame class="h-4 w-4 text-red-500" />
                                        <span>Tinggi (Mendesak)</span>
                                    </div>
                                </template>
                                <SelectValue v-else placeholder="Pilih Prioritas" />
                            </SelectTrigger>
                            <SelectContent>
                                <SelectItem value="low">
                                    <div class="flex items-center gap-2">
                                        <Clock class="h-4 w-4 text-green-500" />
                                        <span>Rendah (Santai)</span>
                                    </div>
                                </SelectItem>
                                <SelectItem value="medium">
                                    <div class="flex items-center gap-2">
                                        <Zap class="h-4 w-4 text-yellow-500" />
                                        <span>Sedang (Butuh)</span>
                                    </div>
                                </SelectItem>
                                <SelectItem value="high">
                                    <div class="flex items-center gap-2">
                                        <Flame class="h-4 w-4 text-red-500" />
                                        <span>Tinggi (Mendesak)</span>
                                    </div>
                                </SelectItem>
                            </SelectContent>
                        </Select>
                    </div>
                </div>
                <DialogFooter class="gap-2">
                    <Button variant="outline" @click="isDialogOpen = false" class="rounded-xl h-10 px-6">Batal</Button>
                    <Button @click="handleSave" class="bg-gradient-to-r from-emerald-600 to-teal-500 text-white hover:from-emerald-500 hover:to-teal-400 shadow-md rounded-xl h-10 px-6 font-bold"
                        :disabled="isSubmitting" :loading="isSubmitting">
                        {{ isEditing ? 'Simpan Perubahan' : 'Simpan' }}
                    </Button>
                </DialogFooter>
            </DialogContent>
        </Dialog>


        <ManualTransactionDialog 
            v-if="selectedItemToBuy && isBuyDialogOpen"
            :open="true" 
            @update:open="(val) => {
                if (!val) {
                    isBuyDialogOpen = false;
                    wishlistStore.fetchItems();
                    selectedItemToBuy = null;
                }
            }"
            :initialData="{
                amount: selectedItemToBuy.estimated_price,
                category_id: selectedItemToBuy.category_id,
                description: selectedItemToBuy.name
            }"
            :wishlistItemId="selectedItemToBuy.id"
            @save="onTransactionSaved"
        />
    </div>
</template>
