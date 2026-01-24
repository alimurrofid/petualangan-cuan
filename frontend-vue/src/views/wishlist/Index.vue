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
                
                <div v-if="activeItems.length === 0" class="col-span-full text-center py-20 text-muted-foreground border-2 border-dashed border-muted rounded-3xl bg-muted/10">
                    <Clock class="h-16 w-16 mx-auto mb-4 opacity-20" />
                    <p class="font-medium text-lg">Belum ada keinginan yang dicatat.</p>
                    <p class="text-sm opacity-70">Tekan tombol tambah untuk mulai mencatat.</p>
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
                 <div 
                    v-for="item in boughtItems" 
                    :key="item.id" 
                    class="relative rounded-3xl p-6 flex flex-col justify-between shadow-xl group overflow-hidden text-white bg-gradient-to-br from-emerald-600 to-teal-600 opacity-90 hover:opacity-100 transition-all duration-300 hover:-translate-y-2"
                >
                    <!-- Ornaments -->
                    <div class="absolute top-0 right-0 w-48 h-48 bg-white/10 rounded-full blur-3xl -mr-16 -mt-16 pointer-events-none"></div>
                    <div class="absolute bottom-0 left-0 w-32 h-32 bg-black/10 rounded-full blur-2xl -ml-10 -mb-10 pointer-events-none"></div>

                    <div class="relative z-10">
                        <div class="flex justify-between items-start mb-6">
                            <div class="space-y-1">
                                <h3 class="text-xl font-bold tracking-tight line-through decoration-white/50 text-white/50">{{ item.name }}</h3>
                                <div class="flex items-center gap-1.5 text-[10px] font-medium uppercase tracking-widest text-white/50 pl-1">
                                    <component v-if="item.category && getIconComponent(item.category.icon)" :is="getIconComponent(item.category.icon)" class="h-3 w-3" />
                                    <span v-else-if="item.category">{{ getEmoji(item.category.icon) || '📦' }}</span>
                                    <span>{{ item.category?.name }}</span>
                                </div>
                            </div>
                            <Button variant="ghost" size="icon" class="h-8 w-8 p-0 hover:bg-white/20 rounded-full text-white/70 hover:text-white" @click="handleDelete(item.id)">
                                <Trash2 class="h-4 w-4" />
                            </Button>
                        </div>
                    </div>
                    
                    <div class="relative z-10 flex justify-between items-center mt-auto p-3 bg-black/10 backdrop-blur-sm rounded-2xl border border-white/5">
                         <div class="text-xl font-mono font-bold tracking-tight text-white/90" :class="{ 'privacy-blur': authStore.isPrivacyMode }">
                            {{ formatRp(item.estimated_price) }}
                        </div>
                        <span class="flex items-center gap-1 bg-white text-emerald-600 px-3 py-1 rounded-full text-[10px] font-bold uppercase tracking-widest shadow-sm">
                            <CheckCircle class="w-3 h-3" /> Terbeli
                        </span>
                    </div>
                </div>
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
                        <Select v-model="form.category_id">
                            <SelectTrigger class="w-full h-11 rounded-xl bg-background shadow-sm">
                                <template v-if="form.category_id">
                                    <div class="flex items-center gap-2" v-if="categoryStore.categories.find(c => String(c.id) === form.category_id)">
                                        <component v-if="getIconComponent(categoryStore.categories.find(c => String(c.id) === form.category_id)?.icon || '')" 
                                            :is="getIconComponent(categoryStore.categories.find(c => String(c.id) === form.category_id)?.icon || '')" 
                                            class="h-4 w-4" 
                                        />
                                        <span v-else>{{ getEmoji(categoryStore.categories.find(c => String(c.id) === form.category_id)?.icon || '') || '📦' }}</span>
                                        <span>{{ categoryStore.categories.find(c => String(c.id) === form.category_id)?.name }}</span>
                                    </div>
                                    <span v-else>Pilih Kategori</span>
                                </template>
                                <SelectValue v-else placeholder="Pilih Kategori" />
                            </SelectTrigger>
                            <SelectContent>
                                <SelectItem v-for="cat in categoryStore.categories.filter(c => c.type === 'expense')" :key="cat.id" :value="String(cat.id)">
                                    <div class="flex items-center gap-2">
                                        <component v-if="getIconComponent(cat.icon)" :is="getIconComponent(cat.icon)" class="h-4 w-4" />
                                        <span v-else>{{ getEmoji(cat.icon) || '📦' }}</span>
                                        <span>{{ cat.name }}</span>
                                    </div>
                                </SelectItem>
                            </SelectContent>
                        </Select>
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
                    <Button @click="handleSave" class="bg-gradient-to-r from-emerald-600 to-teal-500 text-white hover:from-emerald-500 hover:to-teal-400 shadow-md rounded-xl h-10 px-6 font-bold">
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
                description: `Pembelian Wishlist: ${selectedItemToBuy.name}`
            }"
            :wishlistItemId="selectedItemToBuy.id"
            @save="onTransactionSaved"
        />
    </div>
</template>
