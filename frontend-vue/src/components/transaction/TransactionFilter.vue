<script setup lang="ts">
import { ref, watch } from 'vue';
import { Button } from "@/components/ui/button";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import MultiSelect from "@/components/ui/multi-select/MultiSelect.vue";
import { ChevronLeft, ChevronRight, Calendar as CalendarIcon, X } from "lucide-vue-next";
import DateRangePicker from "@/components/DateRangePicker.vue";
import { useWalletStore } from "@/stores/wallet";
import { useCategoryStore } from "@/stores/category";

const props = defineProps<{
    periodType: string,
    startDate: Date,
    endDate: Date,
    walletIds: string[],
    categoryIds: string[],
    formattedDateRange: string
}>();

const emit = defineEmits([
    'update:periodType', 
    'update:walletIds', 
    'update:categoryIds', 
    'navigateDate',  
    'update:dateRange',
    'export'
]);

const walletStore = useWalletStore();
const categoryStore = useCategoryStore();

const showDatePicker = ref(false);



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


            <MultiSelect 
                :modelValue="walletIds" 
                :options="walletStore.wallets.map(w => ({ value: String(w.id), label: w.name, icon: w.icon }))"
                placeholder="Semua Dompet"
                countLabel="Dompet"
                @update:modelValue="(val) => $emit('update:walletIds', val)"
                class="w-full md:w-[200px]"
            />

            <MultiSelect 
                :modelValue="categoryIds" 
                :options="categoryStore.categories.map(c => ({ value: String(c.id), label: c.name, icon: c.icon }))"
                placeholder="Semua Kategori"
                countLabel="Kategori"
                @update:modelValue="(val) => $emit('update:categoryIds', val)"
                 class="w-full md:w-[200px]"
            />
            


            <Button 
                v-if="walletIds.length > 0 || categoryIds.length > 0"
                variant="ghost" 
                size="sm"
                @click="$emit('update:walletIds', []); $emit('update:categoryIds', [])"
                class="h-9 px-3 text-xs text-muted-foreground hover:text-foreground gap-1"
            >
                <X class="h-3.5 w-3.5" />
                Reset
            </Button>
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
