<script setup lang="ts">
import { ref, watch } from 'vue';
import { Button } from "@/components/ui/button";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Input } from "@/components/ui/input";
import { Search, ChevronLeft, ChevronRight, Calendar as CalendarIcon } from "lucide-vue-next";
import DateRangePicker from "@/components/DateRangePicker.vue";
import { useWalletStore } from "@/stores/wallet";
import { useCategoryStore } from "@/stores/category";

const props = defineProps<{
    periodType: string,
    startDate: Date,
    endDate: Date,
    walletId: string,
    categoryId: string,
    searchQuery: string,
    formattedDateRange: string
}>();

const emit = defineEmits([
    'update:periodType', 
    'update:walletId', 
    'update:categoryId', 
    'update:searchQuery', 
    'navigateDate', 
    'update:dateRange',
    'export'
]);

const walletStore = useWalletStore();
const categoryStore = useCategoryStore();

const showDatePicker = ref(false);

const localSearch = ref(props.searchQuery);
const isFocused = ref(false);
let debounceTimer: any = null;

watch(() => props.searchQuery, (val) => {
    // Only sync from prop if user is not actively interacting
    if (!isFocused.value && localSearch.value !== val) {
        localSearch.value = val;
    }
});

watch(localSearch, (val) => {
    // Only emit if it's different to avoid redundant fetches
    if (val === props.searchQuery) return;

    clearTimeout(debounceTimer);
    debounceTimer = setTimeout(() => {
        emit('update:searchQuery', val);
    }, 300);
});

watch(() => props.periodType, (val) => {
    if (val === 'custom') {
        showDatePicker.value = true;
    } else {
        showDatePicker.value = false;
    }
});
</script>

<template>
    <div class="flex flex-col md:flex-row gap-4 items-center justify-between bg-card p-3 rounded-2xl border border-border shadow-sm">
        <!-- Date Nav -->
        <div class="flex items-center gap-2 bg-muted/30 p-1 rounded-xl w-full md:w-auto justify-between md:justify-start">
                <Button variant="ghost" size="icon" @click="$emit('navigateDate', -1)" class="h-8 w-8 rounded-lg hover:bg-background shadow-sm">
                <ChevronLeft class="h-4 w-4" />
            </Button>
            <div 
                class="flex-1 text-center md:px-4 text-sm font-bold flex items-center justify-center gap-2 min-w-[140px] transition-all duration-200"
                :class="{ 
                    'cursor-pointer bg-emerald-50 text-emerald-700 hover:bg-emerald-100 border border-emerald-200 rounded-lg py-1.5 shadow-sm': periodType === 'custom',
                    'py-1': periodType !== 'custom'
                }"
                @click="periodType === 'custom' ? (showDatePicker = !showDatePicker) : null"
            >
                <CalendarIcon class="h-4 w-4 opacity-50" />
                {{ formattedDateRange }}
            </div>
            <Button variant="ghost" size="icon" @click="$emit('navigateDate', 1)" class="h-8 w-8 rounded-lg hover:bg-background shadow-sm">
                <ChevronRight class="h-4 w-4" />
            </Button>
        </div>

        <!-- Filters -->
        <div class="flex flex-col md:flex-row items-center gap-2 w-full md:w-auto flex-1 justify-end">
            <Select :modelValue="periodType" @update:modelValue="(val) => $emit('update:periodType', val)">
                <SelectTrigger class="w-full md:w-[140px] h-9 rounded-xl text-xs font-semibold">
                    <SelectValue placeholder="Periode" />
                </SelectTrigger>
                <SelectContent>
                    <SelectItem value="monthly">Bulanan</SelectItem>
                    <SelectItem value="weekly">Mingguan</SelectItem>
                    <SelectItem value="daily">Harian</SelectItem>
                    <SelectItem value="custom">Custom</SelectItem>
                </SelectContent>
            </Select>

            <Select :modelValue="walletId" @update:modelValue="(val) => $emit('update:walletId', val)">
                <SelectTrigger class="w-full md:w-[140px] h-9 rounded-xl text-xs font-semibold">
                    <SelectValue placeholder="Semua Dompet" />
                </SelectTrigger>
                <SelectContent>
                    <SelectItem value="all">Semua Dompet</SelectItem>
                    <SelectItem v-for="w in walletStore.wallets" :key="w.id" :value="String(w.id)">{{ w.name }}</SelectItem>
                </SelectContent>
            </Select>

            <Select :modelValue="categoryId" @update:modelValue="(val) => $emit('update:categoryId', val)">
                <SelectTrigger class="w-full md:w-[140px] h-9 rounded-xl text-xs font-semibold">
                    <SelectValue placeholder="Semua Kategori" />
                </SelectTrigger>
                <SelectContent>
                    <SelectItem value="all">Semua Kategori</SelectItem>
                    <SelectItem v-for="c in categoryStore.categories" :key="c.id" :value="String(c.id)">{{ c.name }}</SelectItem>
                </SelectContent>
            </Select>
            
             <!-- Search Input -->
            <div class="relative w-full md:w-[180px]">
                <Search class="absolute left-2.5 top-2.5 h-3.5 w-3.5 text-muted-foreground" />
                <Input 
                    v-model="localSearch" 
                    @focus="isFocused = true"
                    @blur="isFocused = false"
                    placeholder="Cari transaksi..." 
                    class="h-9 pl-8 rounded-full bg-muted/50 border-transparent focus:bg-background transition-all text-xs" 
                />
            </div>

        </div>
    </div>

    <!-- Date Picker Overlay -->
    <div v-if="periodType === 'custom' && showDatePicker" class="absolute left-0 right-0 top-16 z-50 flex justify-center p-0 animate-in fade-in zoom-in-95 duration-200 pointer-events-none">
        <div class="pointer-events-auto">
            <DateRangePicker 
                :startDate="startDate"
                :endDate="endDate"
                @update:range="(val) => $emit('update:dateRange', val)"
                @apply="showDatePicker = false"
            />
        </div>
    </div>
</template>
