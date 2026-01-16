<script setup lang="ts">
import { onMounted, ref, computed } from "vue";
import { useWishlistStore, type WishlistItem } from "@/stores/wishlist";
import { useCategoryStore } from "@/stores/category";
import ManualTransactionDialog from "@/components/ManualTransactionDialog.vue";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogFooter } from "@/components/ui/dialog";
import { Label } from "@/components/ui/label";
import { Input } from "@/components/ui/input";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Plus, Trash2, Pencil, ShoppingCart, CheckCircle, Clock } from "lucide-vue-next";
import { useSwal } from "@/composables/useSwal";

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
    <div class="p-6 space-y-8 text-foreground min-h-screen bg-background">
        <div class="flex flex-col md:flex-row md:items-center justify-between gap-4">
            <div>
                <h2 class="text-3xl font-bold tracking-tight">Wishlist</h2>
                <p class="text-muted-foreground mt-1">Simpan dan wujudkan impian finansial Anda.</p>
            </div>
            <Button @click="openAddDialog" class="bg-gradient-to-r from-emerald-600 to-teal-500 text-white hover:from-emerald-500 hover:to-teal-400 shadow-lg px-6 h-12 rounded-full transition-all hover:scale-105 active:scale-95">
                <Plus class="mr-2 h-5 w-5" /> Tambah Keinginan
            </Button>
        </div>

        <!-- Active Wishlist -->
        <div class="space-y-4">
             <div class="flex items-center gap-2">
                <h3 class="text-lg font-bold flex items-center gap-2">
                    <Clock class="h-5 w-5 text-indigo-500" />
                    Sedang Diusahakan
                </h3>
                <span class="text-xs font-medium px-2 py-0.5 rounded-full bg-muted text-muted-foreground">{{ activeItems.length }} item</span>
            </div>

            <div class="grid gap-6 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-3">
                <Card v-for="item in activeItems" :key="item.id" class="rounded-3xl border-border shadow-sm hover:shadow-md transition-all overflow-hidden bg-card flex flex-col">
                    <CardHeader class="pb-3 border-b border-border/50 bg-muted/20">
                        <div class="flex justify-between items-start">
                            <div class="space-y-1">
                                <CardTitle class="text-lg font-bold tracking-tight">{{ item.name }}</CardTitle>
                                <div class="flex flex-wrap gap-2 items-center text-xs">
                                    <Badge variant="outline" :class="{
                                        'border-red-500/50 text-red-600 bg-red-500/10': item.priority === 'high',
                                        'border-yellow-500/50 text-yellow-600 bg-yellow-500/10': item.priority === 'medium',
                                        'border-green-500/50 text-green-600 bg-green-500/10': item.priority === 'low'
                                    }" class="capitalize px-2 py-0.5 rounded-md border text-[10px] font-semibold tracking-wider">
                                        {{ item.priority === 'high' ? '🔥 Mendesak' : item.priority === 'medium' ? '⚡ Butuh' : '🌱 Santai' }}
                                    </Badge>
                                    <span class="text-muted-foreground flex items-center gap-1">
                                        {{ item.category?.name }}
                                    </span>
                                </div>
                            </div>
                            <div class="flex gap-1">
                                <Button variant="ghost" size="icon" class="h-8 w-8 p-0 hover:bg-blue-50 rounded-full" @click="openEditDialog(item)">
                                    <Pencil class="h-4 w-4 text-blue-500" />
                                </Button>
                                <Button variant="ghost" size="icon" class="h-8 w-8 p-0 hover:bg-red-50 rounded-full" @click="handleDelete(item.id)">
                                    <Trash2 class="h-4 w-4 text-red-500" />
                                </Button>
                            </div>
                        </div>
                    </CardHeader>
                    <CardContent class="pt-6 flex-1 flex flex-col justify-between gap-6">
                        <div class="space-y-1">
                            <p class="text-xs uppercase font-bold text-muted-foreground tracking-widest">Estimasi Harga</p>
                            <div class="text-3xl font-extrabold text-transparent bg-clip-text bg-gradient-to-br from-foreground to-muted-foreground">
                                {{ formatRp(item.estimated_price) }}
                            </div>
                        </div>
                        <Button class="w-full rounded-xl bg-indigo-600 hover:bg-indigo-500 text-white shadow-md hover:shadow-lg transition-all" @click="handleBuy(item)">
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
                <h2 class="text-xl font-bold flex items-center gap-2 text-emerald-600">
                    <CheckCircle class="h-6 w-6" /> Sudah Terwujud
                </h2>
                <span class="text-xs font-medium px-2 py-0.5 rounded-full bg-emerald-100 text-emerald-700 border border-emerald-200">{{ boughtItems.length }} item</span>
            </div>
            
            <div class="grid gap-6 md:grid-cols-2 lg:grid-cols-3 opacity-80 hover:opacity-100 transition-opacity">
                 <Card v-for="item in boughtItems" :key="item.id" class="rounded-3xl border-border bg-muted/30 hover:bg-card transition-colors group">
                    <CardHeader class="pb-2">
                        <div class="flex justify-between items-start">
                            <div class="space-y-1">
                                <CardTitle class="text-lg line-through decoration-emerald-500/50 text-muted-foreground">{{ item.name }}</CardTitle>
                                <CardDescription>{{ item.category?.name }}</CardDescription>
                            </div>
                            <Button variant="ghost" size="icon" class="h-8 w-8 p-0 hover:bg-red-50 rounded-full transition-all" @click="handleDelete(item.id)">
                                <Trash2 class="h-4 w-4 text-red-500" />
                            </Button>
                        </div>
                    </CardHeader>
                    <CardContent>
                        <div class="text-xl font-bold text-muted-foreground/70">
                            {{ formatRp(item.estimated_price) }}
                        </div>
                        <Badge variant="default" class="mt-3 bg-emerald-500/10 text-emerald-600 border-emerald-200 hover:bg-emerald-500/20">
                            <CheckCircle class="w-3 h-3 mr-1" /> Terbeli
                        </Badge>
                    </CardContent>
                </Card>
            </div>
        </div>

        <!-- Add/Edit Dialog -->
        <Dialog :open="isDialogOpen" @update:open="isDialogOpen = $event">
            <DialogContent class="sm:max-w-[425px]">
                <DialogHeader>
                    <DialogTitle>{{ isEditing ? 'Edit Keinginan' : 'Tambah Keinginan Baru' }}</DialogTitle>
                </DialogHeader>
                <div class="space-y-4 py-4">
                    <div class="space-y-2">
                        <Label>Nama Barang/Jasa</Label>
                        <Input v-model="form.name" placeholder="Misal: iPhone 15, Liburan ke Bali" />
                    </div>
                    <div class="space-y-2">
                        <Label>Kategori</Label>
                        <Select v-model="form.category_id">
                            <SelectTrigger class="w-full">
                                <SelectValue placeholder="Pilih Kategori" />
                            </SelectTrigger>
                            <SelectContent>
                                <SelectItem v-for="cat in categoryStore.categories.filter(c => c.type === 'expense')" :key="cat.id" :value="String(cat.id)">
                                    {{ cat.name }}
                                </SelectItem>
                            </SelectContent>
                        </Select>
                    </div>
                    <div class="space-y-2">
                        <Label>Estimasi Harga</Label>
                        <Input type="text" inputmode="numeric" pattern="[0-9]*" v-model="formattedPrice" placeholder="Rp 0" />
                    </div>
                    <div class="space-y-2">
                        <Label>Prioritas</Label>
                         <Select v-model="form.priority">
                            <SelectTrigger class="w-full">
                                <SelectValue placeholder="Pilih Prioritas" />
                            </SelectTrigger>
                            <SelectContent>
                                <SelectItem value="low">🌱 Rendah (Santai)</SelectItem>
                                <SelectItem value="medium">⚡ Sedang (Butuh)</SelectItem>
                                <SelectItem value="high">🔥 Tinggi (Mendesak)</SelectItem>
                            </SelectContent>
                        </Select>
                    </div>
                </div>
                <DialogFooter>
                    <Button variant="outline" @click="isDialogOpen = false">Batal</Button>
                    <Button @click="handleSave" class="bg-gradient-to-r from-emerald-600 to-teal-500 text-white hover:from-emerald-500 hover:to-teal-400 shadow-md">
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
