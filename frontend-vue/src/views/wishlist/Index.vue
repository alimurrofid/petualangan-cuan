<script setup lang="ts">
import { onMounted, ref, computed } from "vue";
import { useWishlistStore, type WishlistItem } from "@/stores/wishlist";
import { useCategoryStore } from "@/stores/category";
import ManualTransactionDialog from "@/components/ManualTransactionDialog.vue";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogFooter } from "@/components/ui/dialog";
import { Label } from "@/components/ui/label";
import { Input } from "@/components/ui/input";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Plus, Trash2, Pencil, ShoppingCart, CheckCircle, Clock, Flame, Zap } from "lucide-vue-next";
import { useSwal } from "@/composables/useSwal";
import { getEmoji, getIconComponent } from "@/lib/icons";

const wishlistStore = useWishlistStore();
const categoryStore = useCategoryStore();
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

// State for Buy Dialog
const isBuyDialogOpen = ref(false);
const selectedItemToBuy = ref<WishlistItem | null>(null);

const activeItems = computed(() => wishlistStore.items.filter(i => !i.is_bought));
const boughtItems = computed(() => wishlistStore.items.filter(i => i.is_bought));

const openAddDialog = () => {
    isEditing.value = false;
    editingId.value = null;
    form.value = { name: "", category_id: "", estimated_price: "", priority: "low" };
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
const formatRp = (val: number) => new Intl.NumberFormat("id-ID", { style: "currency", currency: "IDR", maximumFractionDigits: 0 }).format(val);

const formattedPrice = computed({
    get: () => {
        if (!form.value.estimated_price) return "";
        return formatRp(Number(form.value.estimated_price));
    },
    set: (val: string) => {
        const numericValue = Number(val.replace(/[^0-9]/g, ""));
        form.value.estimated_price = numericValue.toString();
    }
});
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
            <Button @click="openAddDialog" class="w-full sm:w-auto bg-gradient-to-r from-emerald-600 to-teal-500 text-white hover:from-emerald-500 hover:to-teal-400 shadow-md h-10 rounded-xl transition-all hover:scale-105 active:scale-95 px-4">
                <Plus class="mr-2 h-4 w-4" /> Tambah Keinginan
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
                <Card v-for="item in activeItems" :key="item.id" class="rounded-3xl border-border shadow-sm hover:shadow-md transition-all overflow-hidden bg-card flex flex-col group relative">
                    <CardHeader class="pb-3 border-b border-border/50">
                        <div class="flex justify-between items-start">
                            <div class="space-y-1">
                                <CardTitle class="text-base font-bold tracking-tight">{{ item.name }}</CardTitle>
                                <div class="flex flex-wrap gap-2 items-center text-[10px] font-medium uppercase tracking-widest">
                                    <Badge variant="outline" :class="{
                                        'border-red-500/50 text-red-600': item.priority === 'high',
                                        'border-yellow-500/50 text-yellow-600': item.priority === 'medium',
                                        'border-green-500/50 text-green-600': item.priority === 'low'
                                    }" class="capitalize px-2 py-1 rounded-md border text-[9px] font-bold tracking-widest flex items-center gap-1">
                                        <template v-if="item.priority === 'high'">
                                            <Flame class="w-3 h-3" /> Mendesak
                                        </template>
                                        <template v-else-if="item.priority === 'medium'">
                                            <Zap class="w-3 h-3" /> Butuh
                                        </template>
                                        <template v-else>
                                            <Clock class="w-3 h-3" /> Santai
                                        </template>
                                    </Badge>
                                    <div class="flex items-center gap-1.5 text-muted-foreground">
                                        <div class="p-1 bg-muted/50 rounded-md">
                                            <component v-if="item.category && getIconComponent(item.category.icon)" :is="getIconComponent(item.category.icon)" class="h-3 w-3" />
                                            <span v-else-if="item.category">{{ getEmoji(item.category.icon) || '📦' }}</span>
                                        </div>
                                        <span>{{ item.category?.name }}</span>
                                    </div>
                                </div>
                            </div>
                            <div class="flex gap-1">
                                <Button variant="ghost" size="icon" class="h-7 w-7 p-0 hover:bg-blue-50 rounded-lg text-blue-500" @click="openEditDialog(item)">
                                    <Pencil class="h-3.5 w-3.5" />
                                </Button>
                                <Button variant="ghost" size="icon" class="h-7 w-7 p-0 hover:bg-red-50 rounded-lg text-red-500" @click="handleDelete(item.id)">
                                    <Trash2 class="h-3.5 w-3.5" />
                                </Button>
                            </div>
                        </div>
                    </CardHeader>
                    <CardContent class="pt-6 flex-1 flex flex-col justify-between gap-6">
                        <div class="space-y-1">
                            <p class="text-[10px] uppercase font-bold text-muted-foreground tracking-widest">Estimasi Harga</p>
                            <div class="text-2xl font-bold text-foreground">
                                {{ formatRp(item.estimated_price) }}
                            </div>
                        </div>
                        <Button class="w-full rounded-xl bg-gradient-to-r from-emerald-600 to-teal-500 hover:from-emerald-500 hover:to-teal-400 text-white shadow-sm transition-all active:scale-95 h-9 text-xs font-bold" @click="handleBuy(item)">
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
            
            <div class="grid gap-6 md:grid-cols-2 lg:grid-cols-3 opacity-90 hover:opacity-100 transition-opacity">
                 <Card v-for="item in boughtItems" :key="item.id" class="rounded-3xl border-border bg-card hover:bg-muted/50 transition-all group overflow-hidden border">
                    <CardHeader class="pb-3 border-b border-border/50">
                        <div class="flex justify-between items-start">
                            <div class="space-y-1">
                                <CardTitle class="text-base font-bold line-through decoration-emerald-500/50 dark:decoration-emerald-400 text-muted-foreground dark:text-slate-400">{{ item.name }}</CardTitle>
                                <div class="flex items-center gap-1.5 text-[10px] font-bold uppercase tracking-widest text-muted-foreground">
                                    <component v-if="item.category && getIconComponent(item.category.icon)" :is="getIconComponent(item.category.icon)" class="h-3 w-3" />
                                    <span v-else-if="item.category">{{ getEmoji(item.category.icon) || '📦' }}</span>
                                    <span>{{ item.category?.name }}</span>
                                </div>
                            </div>
                            <Button variant="ghost" size="icon" class="h-7 w-7 p-0 hover:bg-red-50 rounded-lg transition-all text-red-500" @click="handleDelete(item.id)">
                                <Trash2 class="h-4 w-4" />
                            </Button>
                        </div>
                    </CardHeader>
                    <CardContent class="pt-4 flex justify-between items-center">
                        <div class="text-lg font-bold text-muted-foreground/60 dark:text-white">
                            {{ formatRp(item.estimated_price) }}
                        </div>
                        <Badge variant="default" class="bg-emerald-500/10 dark:bg-emerald-500/20 text-emerald-600 dark:text-emerald-400 border-emerald-200 dark:border-emerald-900 shadow-none text-[9px] font-bold uppercase tracking-widest">
                            <CheckCircle class="w-3 h-3 mr-1" /> Terbeli
                        </Badge>
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
                        <Input type="text" inputmode="numeric" pattern="[0-9]*" v-model="formattedPrice" placeholder="Rp 0" class="h-11 shadow-sm rounded-xl bg-background" />
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

        <!-- Buy Dialog (ManualTransactionDialog Wrapper) -->
        <ManualTransactionDialog 
            v-if="selectedItemToBuy"
            :open="isBuyDialogOpen" 
            @update:open="isBuyDialogOpen = $event"
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
