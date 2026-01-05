<script setup lang="ts">
import { computed } from 'vue';
import { format, parseISO, subDays } from "date-fns";
import { id } from "date-fns/locale";
import { formatCurrency } from "@/lib/utils";
import { getEmoji, getIconComponent } from "@/lib/icons";
import { useTransactionStore } from "@/stores/transaction";
import { Search, Pencil, Trash2, Paperclip } from "lucide-vue-next";
import { Button } from "@/components/ui/button";
import { useSwal } from "@/composables/useSwal";

const transactionStore = useTransactionStore();
const swal = useSwal();
const baseUrl = import.meta.env.VITE_API_BASE_URL;

const emit = defineEmits<{
    (e: 'page-change', page: number): void;
    (e: 'edit', transaction: any): void;
}>();

const handleDelete = async (t: any) => {
    const result = await swal.fire({
        title: 'Hapus Transaksi?',
        text: `Apakah Anda yakin ingin menghapus transaksi "${t.description}" senilai ${formatCurrency(t.amount)}?`,
        icon: 'warning',
        showCancelButton: true,
        confirmButtonColor: '#EF4444',
        cancelButtonColor: '#CBD5E1',
        confirmButtonText: 'Ya, Hapus',
        cancelButtonText: 'Batal'
    });

    if (result.isConfirmed) {
        try {
            await transactionStore.deleteTransaction(t.id);
            swal.toast({
                icon: 'success',
                title: 'Transaksi berhasil dihapus'
            });
        } catch (error) {
            swal.toast({
                icon: 'error',
                title: 'Gagal menghapus transaksi'
            });
        }
    }
};

const groupedTransactions = computed(() => {
    const groups: Record<string, any[]> = {};
    transactionStore.transactions.forEach(t => {
        // Use local timezone for grouping key
        const dateObj = parseISO(t.date);
        const dateKey = format(dateObj, 'yyyy-MM-dd');
        if (!groups[dateKey]) groups[dateKey] = [];
        groups[dateKey].push(t);
    });

    const sortedTypes = Object.keys(groups).sort((a, b) => b.localeCompare(a));

    return sortedTypes.map(dateStr => {
        const date = parseISO(dateStr);
        let label = format(date, 'EEEE, d MMMM yyyy', { locale: id });

        // Compare with today using local time
        const today = new Date();
        const yesterday = subDays(today, 1);

        // We compare using formatted strings to avoid time discrepancies
        if (format(date, 'yyyy-MM-dd') === format(today, 'yyyy-MM-dd')) label = "Hari Ini";
        if (format(date, 'yyyy-MM-dd') === format(yesterday, 'yyyy-MM-dd')) label = "Kemarin";

        return {
            date: dateStr,
            label,
            items: groups[dateStr]
        };
    });
});
</script>

<template>
    <div class="h-full flex flex-col">
        <div v-if="transactionStore.isLoading && transactionStore.transactions.length === 0"
            class="flex-1 flex items-center justify-center p-8">
            <span class="loading loading-spinner loading-md text-muted-foreground"></span>
        </div>

        <div v-else-if="transactionStore.transactions.length === 0"
            class="flex-1 flex flex-col items-center justify-center text-muted-foreground opacity-60 text-center p-8">
            <div class="h-12 w-12 bg-muted rounded-full flex items-center justify-center mb-2">
                <Search class="h-5 w-5" />
            </div>
            <p class="text-sm font-medium">Tidak ada transaksi</p>
        </div>

        <div v-else class="flex-1 overflow-y-auto space-y-6 pr-2 custom-scrollbar p-1">
            <div v-for="group in groupedTransactions" :key="group.date" class="space-y-3">
                <div class="flex items-center gap-2 sticky top-0 bg-background/95 backdrop-blur-sm z-10 py-1">
                    <span
                        class="text-[10px] font-bold uppercase tracking-widest text-muted-foreground bg-muted px-2 py-1 rounded-md">{{
                            group.label }}</span>
                    <div class="h-[1px] flex-1 bg-border"></div>
                </div>

                <div v-for="t in group.items" :key="t.id"
                    class="group relative flex items-center justify-between p-3 rounded-2xl hover:bg-muted/50 transition-all cursor-pointer border border-transparent hover:border-border">
                    <div class="flex items-center gap-3">
                        <div :class="['h-10 w-10 rounded-xl flex items-center justify-center text-lg shadow-sm transition-transform group-hover:scale-105',
                            t.type === 'expense' ? 'bg-red-50 text-red-500' :
                                t.type === 'income' ? 'bg-emerald-50 text-emerald-600' :
                                    'bg-blue-50 text-blue-600']">
                            <span v-if="getEmoji(t.category.icon)" class="text-xl leading-none filter drop-shadow-sm">{{
                                getEmoji(t.category.icon) }}</span>
                            <component v-else :is="getIconComponent(t.category.icon, 'Circle')" class="h-5 w-5" />
                        </div>
                        <div>
                            <p class="font-bold text-sm truncate max-w-[120px]">{{ t.description || 'No Description' }}
                            </p>
                            <div class="flex items-center gap-1">
                                <p class="text-[10px] text-muted-foreground font-medium flex items-center gap-1">
                                    {{ t.wallet.name }} • {{ t.category.name }}
                                </p>
                                <a v-if="t.attachment" :href="`${baseUrl}${t.attachment}`" target="_blank" @click.stop
                                    class="text-xs text-blue-500 hover:text-blue-700 flex items-center">
                                    <Paperclip class="w-3 h-3" />
                                </a>
                            </div>
                        </div>
                    </div>
                    <div class="text-right">
                        <span :class="['block font-bold text-sm',
                            t.type === 'income' ? 'text-emerald-600' :
                                t.type === 'expense' ? 'text-red-500' :
                                    'text-blue-600']">
                            {{ (t.type === 'income' || t.type === 'transfer_in') ? '+' : '-' }} {{
                                formatCurrency(t.amount) }}
                        </span>
                        <span class="text-[10px] text-muted-foreground">{{ format(parseISO(t.date), 'HH:mm') }}</span>
                    </div>


                    <!-- Action Buttons -->
                    <div
                        class="absolute right-2 top-2 opacity-0 group-hover:opacity-100 transition-opacity flex gap-1 bg-background/80 backdrop-blur-sm rounded-lg p-1 shadow-sm border border-border">
                        <button @click.stop="emit('edit', t)"
                            class="p-1.5 hover:bg-slate-100 rounded-md text-slate-500 hover:text-blue-600 transition-colors"
                            title="Edit">
                            <Pencil class="w-3.5 h-3.5" />
                        </button>
                        <button @click.stop="handleDelete(t)"
                            class="p-1.5 hover:bg-red-50 rounded-md text-slate-500 hover:text-red-500 transition-colors"
                            title="Hapus">
                            <Trash2 class="w-3.5 h-3.5" />
                        </button>
                    </div>
                </div>
            </div>

            <!-- Pagination Controls -->
            <div class="flex justify-center gap-4 py-4 mt-4 border-t border-border">
                <Button variant="outline" size="sm" :disabled="transactionStore.paginationMeta.page <= 1"
                    @click="$emit('page-change', transactionStore.paginationMeta.page - 1)">
                    Previous
                </Button>
                <div class="flex items-center text-xs text-muted-foreground">
                    Page {{ transactionStore.paginationMeta.page }} / {{ Math.ceil(transactionStore.paginationMeta.total
                        / transactionStore.paginationMeta.limit) || 1 }}
                </div>
                <Button variant="outline" size="sm"
                    :disabled="transactionStore.paginationMeta.page >= Math.ceil(transactionStore.paginationMeta.total / transactionStore.paginationMeta.limit)"
                    @click="$emit('page-change', transactionStore.paginationMeta.page + 1)">
                    Next
                </Button>
            </div>
        </div>
    </div>
</template>

<style scoped>
.custom-scrollbar::-webkit-scrollbar {
    width: 4px;
}

.custom-scrollbar::-webkit-scrollbar-track {
    background: transparent;
}

.custom-scrollbar::-webkit-scrollbar-thumb {
    background: hsl(var(--border));
    border-radius: 4px;
}

.custom-scrollbar::-webkit-scrollbar-thumb:hover {
    background: hsl(var(--muted-foreground));
}
</style>
