<script setup lang="ts">
import { ref, computed, onMounted, watch } from "vue";
import { format, startOfMonth, endOfMonth, startOfWeek, endOfWeek, startOfDay, endOfDay, addMonths, addWeeks, addDays } from "date-fns";
import { id } from "date-fns/locale";

import { useTransactionStore, type CategoryBreakdown } from "@/stores/transaction";
import { useWalletStore } from "@/stores/wallet";
import { useCategoryStore } from "@/stores/category";

import { Card, CardHeader, CardTitle, CardContent } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import DateRangePicker from "@/components/DateRangePicker.vue";


import { ChevronLeft, ChevronRight, Calendar as CalendarIcon, PieChart, Download } from "lucide-vue-next";
import { getEmoji, getIconComponent } from "@/lib/icons";

const transactionStore = useTransactionStore();
const walletStore = useWalletStore();
const categoryStore = useCategoryStore();

type PeriodType = 'monthly' | 'weekly' | 'daily' | 'custom';

const filterWallet = ref<string>("all");
const filterType = ref<string>("all"); // Default to all as requested
const periodType = ref<PeriodType>('monthly');
const selectedDate = ref(new Date());
const customDateRange = ref({
  start: new Date(),
  end: new Date() 
});

const showDatePicker = ref(false);

watch(periodType, (val) => {
  showDatePicker.value = val === 'custom';
});

onMounted(async () => {
    // Fetch initial data
    await Promise.all([
        walletStore.fetchWallets(),
        categoryStore.fetchCategories()
    ]);
    fetchReportData();
});

const dateRange = computed(() => {
  const date = selectedDate.value;
  switch (periodType.value) {
    case 'monthly':
      return { start: startOfMonth(date), end: endOfMonth(date) };
    case 'weekly':
      return { start: startOfWeek(date, { weekStartsOn: 1 }), end: endOfWeek(date, { weekStartsOn: 1 }) };
    case 'daily':
      return { start: startOfDay(date), end: endOfDay(date) };
    case 'custom':
      return customDateRange.value;
    default:
      return { start: startOfMonth(date), end: endOfMonth(date) };
  }
});

const formattedDateRange = computed(() => {
  const { start, end } = dateRange.value;
  if (!start || !end) return '-';
  
  if (periodType.value === 'daily') {
    return format(start, "d MMMM yyyy", { locale: id });
  }
  return `${format(start, "d MMM", { locale: id })} - ${format(end, "d MMM yyyy", { locale: id })}`;
});

const navigateDate = (amount: number) => {
  const date = selectedDate.value;
  switch (periodType.value) {
    case 'monthly':
      selectedDate.value = addMonths(date, amount);
      break;
    case 'weekly':
      selectedDate.value = addWeeks(date, amount);
      break;
    case 'daily':
      selectedDate.value = addDays(date, amount);
      break;
  }
};

const fetchReportData = async () => {
    const { start, end } = dateRange.value;
    const startDateStr = format(start, 'yyyy-MM-dd');
    const endDateStr = format(end, 'yyyy-MM-dd');
    
    await transactionStore.fetchReport(
        startDateStr, 
        endDateStr, 
        filterWallet.value === 'all' ? undefined : Number(filterWallet.value), 
        filterType.value === 'all' ? undefined : filterType.value
    );
};

// Watch for changes in filters to re-fetch
watch([dateRange, filterWallet, filterType], () => {
    fetchReportData();
}, { deep: true });

// Use store data directly
const reportData = computed(() => transactionStore.reportData);

// Charts & Visuals
const formatCurrency = (value: number) => {
  return new Intl.NumberFormat("id-ID", { style: "currency", currency: "IDR", minimumFractionDigits: 0 }).format(value);
};

const totalAmount = computed(() => {
    return reportData.value.reduce((sum, item) => sum + item.total_amount, 0);
});

const chartSeries = computed(() => {
    return reportData.value.map(item => item.total_amount);
});

const chartOptions = computed(() => {
    return {
        chart: { type: 'donut', fontFamily: 'inherit', foreColor: '#94a3b8' },
        labels: reportData.value.map(item => item.category_name),
        colors: ['#10b981', '#ef4444', '#3b82f6', '#f59e0b', '#8b5cf6', '#ec4899', '#06b6d4', '#6366f1'], 
        plotOptions: {
            pie: {
                donut: {
                    size: '70%',
                    labels: {
                        show: true,
                        value: {
                            show: true,
                            formatter: (val: number) => formatCurrency(val)
                        },
                        total: {
                            show: true,
                            label: 'Total',
                            formatter: () => formatCurrency(totalAmount.value)
                        }
                    }
                }
            }
        },
        legend: { position: 'bottom' },
        stroke: { show: false },
        dataLabels: { enabled: false },
        tooltip: { 
            theme: 'dark',
            y: { formatter: (value: number) => formatCurrency(value) }
        }
    };
});



const getDisplayPercentage = (item: CategoryBreakdown) => {
    // If expense and has budget, use budget as base
    if (item.type === 'expense' && item.budget_limit > 0) {
        return Math.round((item.total_amount / item.budget_limit) * 100);
    }
    // Fallback to percentage of total amount displayed (contribution)
    if (totalAmount.value === 0) return 0;
    return Math.round((item.total_amount / totalAmount.value) * 100);
};

const getProgressBarWidth = (item: CategoryBreakdown) => {
    return Math.min(getDisplayPercentage(item), 100);
};

const getProgressColor = (item: CategoryBreakdown) => {
    if (item.type === 'expense') {
        const progress = getDisplayPercentage(item);
        if (progress >= 100) return 'bg-red-600';
        if (progress >= 80) return 'bg-yellow-500';
        return 'bg-emerald-500';
    }
    return '';
};

const handleExport = async () => {
    try {
        const { start, end } = dateRange.value;
        const startDateStr = format(start, 'yyyy-MM-dd HH:mm:ss');
        const endDateStr = format(end, 'yyyy-MM-dd HH:mm:ss');
        
        const blob = await transactionStore.exportReport(startDateStr, endDateStr, filterWallet.value, filterType.value);
        const url = window.URL.createObjectURL(blob);
        const a = document.createElement('a');
        a.href = url;
        a.download = `reports_petualangancuan_${format(new Date(), 'yyyy-MM-dd')}.xlsx`;
        document.body.appendChild(a);
        a.click();
        window.URL.revokeObjectURL(url);
        document.body.removeChild(a);
    } catch (e) {
        console.error("Export failed", e);
    }
};
</script>

<template>
  <div class="p-6 space-y-8 text-foreground min-h-screen bg-background">
    <div class="flex flex-col gap-2">
      <h2 class="text-3xl font-bold tracking-tight">Laporan Keuangan</h2>
      <p class="text-sm text-muted-foreground">Analisis pengeluaran dan pemasukan per kategori.</p>
    </div>

    <!-- Toolbar -->
    <div class="flex flex-col md:flex-row gap-4 items-center justify-between bg-card p-3 rounded-2xl border border-border shadow-sm z-20 relative">
        <!-- Date Nav -->
        <div class="flex items-center gap-2 bg-muted/30 p-1 rounded-xl w-full md:w-auto justify-between md:justify-start">
                <Button variant="ghost" size="icon" @click="navigateDate(-1)" class="h-8 w-8 rounded-lg hover:bg-background shadow-sm">
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
            <Button variant="ghost" size="icon" @click="navigateDate(1)" class="h-8 w-8 rounded-lg hover:bg-background shadow-sm">
                <ChevronRight class="h-4 w-4" />
            </Button>
        </div>

        <!-- Filters -->
        <div class="flex flex-col md:flex-row items-center gap-2 w-full md:w-auto">
             <Select v-model="periodType">
                <SelectTrigger class="w-full md:w-[120px] h-9 rounded-xl text-xs font-semibold">
                    <SelectValue placeholder="Periode" />
                </SelectTrigger>
                <SelectContent>
                    <SelectItem value="monthly">Bulanan</SelectItem>
                    <SelectItem value="weekly">Mingguan</SelectItem>
                    <SelectItem value="daily">Harian</SelectItem>
                    <SelectItem value="custom">Custom</SelectItem>
                </SelectContent>
            </Select>

            
            <Select v-model="filterWallet">
                <SelectTrigger class="w-full md:w-[140px] h-9 rounded-xl text-xs font-semibold">
                    <SelectValue placeholder="Semua Dompet" />
                </SelectTrigger>
                <SelectContent>
                    <SelectItem value="all">Semua Dompet</SelectItem>
                    <SelectItem v-for="w in walletStore.wallets" :key="w.id" :value="String(w.id)">{{ w.name }}</SelectItem>
                </SelectContent>
            </Select>
            
            <Select v-model="filterType">
                <SelectTrigger class="w-full md:w-[120px] h-9 rounded-xl text-xs font-semibold">
                    <SelectValue placeholder="Tipe" />
                </SelectTrigger>
                <SelectContent>
                    <SelectItem value="all">Semua Tipe</SelectItem>
                    <SelectItem value="income">Pemasukan</SelectItem>
                    <SelectItem value="expense">Pengeluaran</SelectItem>
                </SelectContent>
            </Select>

        </div>
    </div>

    <div v-if="periodType === 'custom' && showDatePicker" class="absolute left-0 right-0 top-32 z-50 flex justify-center p-0 animate-in fade-in zoom-in-95 duration-200 pointer-events-none">
        <div class="pointer-events-auto shadow-xl rounded-xl">
            <DateRangePicker 
                :startDate="customDateRange.start"
                :endDate="customDateRange.end"
                @update:range="(val) => { customDateRange.start = val.start; customDateRange.end = val.end }"
                @apply="showDatePicker = false"
            />
        </div>
    </div>

    <!-- Content -->
    <div class="grid lg:grid-cols-3 gap-6">
        <!-- Chart -->
        <Card class="bg-card border-border shadow-sm flex flex-col rounded-3xl overflow-hidden min-h-[400px]">
             <!-- We give it a fixed key to force re-render if needed, but apexchart handles reactivity usually -->
            <CardHeader class="pb-2 border-b border-border/50">
                <CardTitle class="text-base font-bold flex items-center gap-2">
                    <PieChart class="h-4 w-4" /> Distribusi Kategori
                </CardTitle>
            </CardHeader>
            <CardContent class="flex items-center justify-center p-6 bg-muted/10 h-full">
                <div v-if="reportData.length > 0" class="w-full max-w-[350px]">
                     <apexchart type="donut" width="100%" :options="chartOptions" :series="chartSeries" />
                </div>
                <div v-else class="text-center text-muted-foreground py-10">
                    <PieChart class="h-12 w-12 mx-auto mb-3 opacity-20" />
                    <p>Tidak ada data untuk periode ini.</p>
                </div>
            </CardContent>
        </Card>

        <!-- Category List -->
        <Card class="lg:col-span-2 bg-card border-border shadow-sm flex flex-col rounded-3xl overflow-hidden">
            <CardHeader class="pb-3 border-b border-border/50">
                <div class="flex items-center justify-between">
                    <CardTitle class="text-base font-bold">Rincian Kategori</CardTitle>
                    <Button variant="outline" size="sm" @click="handleExport" class="h-8 rounded-xl border-border shadow-sm hover:bg-muted/50 gap-2 text-xs" title="Export Excel">
                        <Download class="h-3.5 w-3.5 text-muted-foreground" />
                        <span class="text-xs font-semibold text-muted-foreground">Export</span>
                    </Button>
                </div>
            </CardHeader>
            <CardContent class="p-0 custom-scrollbar overflow-y-auto max-h-[500px]">
                 <div v-if="reportData.length === 0" class="p-8 text-center text-muted-foreground">
                    <p>Tidak ada transaksi.</p>
                </div>
                <div v-else class="divide-y divide-border">
                    <div v-for="(item, index) in reportData" :key="index" class="p-4 hover:bg-muted/30 transition-colors flex items-center justify-between group">
                        <div class="flex items-center gap-4 flex-1">
                             <div :class="['h-10 w-10 rounded-xl flex items-center justify-center text-lg shadow-sm', 
                                  item.type === 'expense' ? 'bg-red-50 text-red-500' : 'bg-emerald-50 text-emerald-600']">
                                  <span v-if="getEmoji(item.category_icon)" class="text-lg leading-none">{{ getEmoji(item.category_icon) }}</span>
                                  <component v-else :is="getIconComponent(item.category_icon)" class="h-5 w-5" />
                            </div>
                            <div class="flex-1 max-w-md">
                                <div class="flex justify-between items-center mb-1">
                                    <p class="font-bold text-sm">{{ item.category_name }}</p>
                                    <span v-if="item.is_over_budget" class="text-[10px] font-bold text-red-600 bg-red-100 px-2 py-0.5 rounded-full animate-pulse">OVER BUDGET</span>
                                </div>
                                <div v-if="item.type === 'expense'" class="flex items-center gap-2">
                                    <div class="h-1.5 flex-1 bg-muted rounded-full overflow-hidden">
                                        <div class="h-full rounded-full" 
                                            :class="getProgressColor(item)" 
                                            :style="{ width: `${getProgressBarWidth(item)}%` }">
                                        </div>
                                    </div>
                                    <span class="text-[10px] text-muted-foreground font-medium w-8 text-right">{{ getDisplayPercentage(item) }}%</span>
                                </div>
                            </div>
                        </div>
                        <div class="text-right ml-4">
                             <p :class="['font-bold text-sm', item.type === 'expense' ? 'text-red-500' : 'text-emerald-600']">
                                {{ formatCurrency(item.total_amount) }}
                             </p>
                             <div v-if="item.type === 'expense' && item.budget_limit > 0" class="flex flex-col items-end">
                                <p class="text-[10px] text-muted-foreground">Budget: {{ formatCurrency(item.budget_limit) }}</p>
                             </div>
                             <p v-else class="text-[10px] text-muted-foreground uppercase tracking-widest font-semibold">{{ item.type === 'expense' ? 'Pengeluaran' : '' }}</p>
                        </div>
                    </div>
                </div>
            </CardContent>
        </Card>
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
</style>
