<script setup lang="ts">
import { ref, computed, onMounted, watch } from "vue";
import { format, parseISO, startOfMonth, endOfMonth, isWithinInterval, eachDayOfInterval, addMonths, startOfWeek, endOfWeek, addWeeks, startOfDay, endOfDay, addDays, isSameDay, subDays as subDaysFn, eachHourOfInterval, isSameHour } from "date-fns";
import { id } from "date-fns/locale";

import { useTransactionStore } from "@/stores/transaction";
import { useWalletStore } from "@/stores/wallet";
import { useCategoryStore } from "@/stores/category";

import { Card, CardHeader, CardTitle, CardContent } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import DateRangePicker from "@/components/DateRangePicker.vue";

import * as LucideIcons from "lucide-vue-next";
import { Search, ChevronLeft, ChevronRight, Calendar as CalendarIcon, ArrowUpCircle, ArrowDownCircle, Wallet } from "lucide-vue-next";

const transactionStore = useTransactionStore();
const walletStore = useWalletStore();
const categoryStore = useCategoryStore();

type PeriodType = 'monthly' | 'weekly' | 'daily' | 'custom';

const filterWallet = ref<string>("all");
const filterCategory = ref<string>("all");
const searchQuery = ref("");

const periodType = ref<PeriodType>('monthly');
const selectedDate = ref(new Date());
const customDateRange = ref({
  start: new Date(),
  end: new Date() 
});

const showDatePicker = ref(false);

watch(periodType, (val) => {
  if (val === 'custom') {
    showDatePicker.value = true;
  } else {
    showDatePicker.value = false;
  }
});

onMounted(async () => {
    await Promise.all([
        transactionStore.fetchTransactions(),
        walletStore.fetchWallets(),
        categoryStore.fetchCategories()
    ]);
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
  }
});

const formattedDateRange = computed(() => {
  const { start, end } = dateRange.value;
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

// Custom date inputs handled by DateRangePicker component now

const filteredTransactions = computed(() => {
  return transactionStore.transactions.filter((t) => {
    const tDate = parseISO(t.date);
    const { start, end } = dateRange.value;
    
    if (!start || !end) return false;

    let matchesPeriod = false;
    if (periodType.value === 'daily') {
         matchesPeriod = isSameDay(tDate, selectedDate.value);
    } else {
         matchesPeriod = isWithinInterval(tDate, { start, end });
    }

    const matchesWallet = filterWallet.value === "all" || t.wallet_id === Number(filterWallet.value);
    const matchesCategory = filterCategory.value === "all" || t.category_id === Number(filterCategory.value);
    // Backend uses 'description' but UI display uses it as title. Search 'description'
    const matchesSearch = (t.description || "").toLowerCase().includes(searchQuery.value.toLowerCase());

    return matchesPeriod && matchesWallet && matchesCategory && matchesSearch;
  });
});

const groupedTransactions = computed(() => {
    const groups: Record<string, any[]> = {};
    filteredTransactions.value.forEach(t => {
        const dateKey = format(parseISO(t.date), 'yyyy-MM-dd');
        if (!groups[dateKey]) groups[dateKey] = [];
        groups[dateKey].push(t);
    });

    const sortedTypes = Object.keys(groups).sort((a,b) => b.localeCompare(a));

    return sortedTypes.map(dateStr => {
        const date = parseISO(dateStr);
        let label = format(date, 'EEEE, d MMMM yyyy', { locale: id });
        
        if (isSameDay(date, new Date())) label = "Hari Ini";
        if (isSameDay(date, subDaysFn(new Date(), 1))) label = "Kemarin";

        return {
            date: dateStr,
            label,
            items: groups[dateStr]
        };
    });
});

const totalIncome = computed(() => {
    return filteredTransactions.value
        .filter(t => t.type === 'income')
        .reduce((sum, t) => sum + t.amount, 0);
});

const totalExpense = computed(() => {
    return filteredTransactions.value
        .filter(t => t.type === 'expense')
        .reduce((sum, t) => sum + t.amount, 0);
});

// Helper for icons (similarly used in other views)
const getIconComponent = (name: string | undefined) => {
    if (!name) return LucideIcons.Circle;
    return (LucideIcons as any)[name] || LucideIcons.Circle;
};

const emojiCategories: Record<string, string> = {
  Em_MoneyBag: "💰", Em_DollarBill: "💵", Em_Card: "💳", Em_Bank: "🏦", Em_MoneyWing: "💸", Em_Coin: "🪙",
  Em_Pizza: "🍕", Em_Cart: "🛒", Em_Coffee: "☕", Em_Game: "🎮", Em_Airplane: "✈️", Em_Gift: "🎁",
  Em_Star: "⭐", Em_Fire: "🔥", Em_Lock: "🔒", Em_Check: "✅", Em_Idea: "💡"
};

const getEmoji = (name: string | undefined) => {
  if (!name) return null;
  if (emojiCategories[name]) return emojiCategories[name];
  if (/\p{Emoji}/u.test(name)) return name;
  return null;
};

const formatCurrency = (value: number) => {
  return new Intl.NumberFormat("id-ID", { style: "currency", currency: "IDR", minimumFractionDigits: 0 }).format(value);
};

const chartSeries = computed(() => {
  const { start, end } = dateRange.value;
  
  if (periodType.value === 'daily') {
       // For daily view, show hourly buckets
       const hours = eachHourOfInterval({ start, end });
       
       const incomeData = hours.map(hour => {
          return filteredTransactions.value
            .filter(t => t.type.toLowerCase() === 'income' && isSameHour(parseISO(t.date), hour))
            .reduce((sum, t) => sum + Number(t.amount), 0);
       });

        const expenseData = hours.map(hour => {
          return filteredTransactions.value
            .filter(t => t.type.toLowerCase() === 'expense' && isSameHour(parseISO(t.date), hour))
            .reduce((sum, t) => sum + Number(t.amount), 0);
       });

        return [
            { name: 'Pemasukan', data: incomeData },
            { name: 'Pengeluaran', data: expenseData }
        ];
  }

  // Monthly/Weekly views (daily accumulation)
  const days = eachDayOfInterval({ start, end });
  
  const incomeData = days.map(day => {
      return filteredTransactions.value
        .filter(t => t.type.toLowerCase() === 'income' && format(parseISO(t.date), 'yyyy-MM-dd') === format(day, 'yyyy-MM-dd'))
        .reduce((sum, t) => sum + Number(t.amount), 0);
  });
  
  const expenseData = days.map(day => {
      return filteredTransactions.value
        .filter(t => t.type.toLowerCase() === 'expense' && format(parseISO(t.date), 'yyyy-MM-dd') === format(day, 'yyyy-MM-dd'))
        .reduce((sum, t) => sum + Number(t.amount), 0);
  });

  return [
    { name: 'Pemasukan', data: incomeData },
    { name: 'Pengeluaran', data: expenseData }
  ];
});

const chartOptions = computed(() => {
    const { start, end } = dateRange.value;
    
    let categories: string[] = [];

    if (periodType.value === 'daily') {
         const hours = eachHourOfInterval({ start, end });
         categories = hours.map(h => format(h, 'HH:mm'));
    } else {
        const days = eachDayOfInterval({ start, end });
        categories = days.map((d) => {
            return format(d, "d MMM", { locale: id });
        });
    }

    return {
        chart: { type: 'area', height: 300, toolbar: { show: false }, fontFamily: 'inherit', zoom: { enabled: false }, foreColor: '#94a3b8' },
        dataLabels: { enabled: false },
        stroke: { curve: 'smooth', width: 2 },
        fill: { type: 'gradient', gradient: { shadeIntensity: 1, opacityFrom: 0.4, opacityTo: 0.05, stops: [0, 90, 100] } },
        xaxis: { categories: categories, axisBorder: { show: false }, axisTicks: { show: false }, labels: { style: { fontSize: '10px' }, rotate: 0, hideOverlappingLabels: true }, tooltip: { enabled: false } },
        yaxis: { labels: { style: { fontSize: '10px' }, formatter: (value: number) => { if (value >= 1000000) return (value / 1000000).toFixed(1) + 'M'; if (value >= 1000) return (value / 1000).toFixed(0) + 'k'; return value; } } },
        grid: { borderColor: '#334155', strokeDashArray: 4, yaxis: { lines: { show: true } }, xaxis: { lines: { show: false } }, padding: { top: 0, right: 0, bottom: 0, left: 10 } },
        colors: ['#10b981', '#ef4444'],
        tooltip: { theme: 'dark', x: { show: true }, y: { formatter: (value: number) => formatCurrency(value) } },
        legend: { position: 'top', horizontalAlign: 'right', offsetY: -20, itemMargin: { horizontal: 10, vertical: 0 } }
    };
});
</script>

<template>
  <div class="p-6 space-y-8 text-foreground min-h-screen bg-background">
    <div class="flex flex-col gap-2">
      <h2 class="text-3xl font-bold tracking-tight">Riwayat Transaksi</h2>
      <p class="text-sm text-muted-foreground">Analisis dan pantau arus kas Anda.</p>
    </div>

    <!-- Stats Grid -->
    <div class="grid grid-cols-1 md:grid-cols-3 gap-6">
        <div class="relative overflow-hidden rounded-3xl p-6 bg-gradient-to-br from-emerald-50 to-teal-50 dark:from-emerald-950/20 dark:to-teal-950/20 border border-emerald-100 dark:border-emerald-900/30 shadow-sm">
            <div class="flex justify-between items-start">
                <div>
                    <p class="text-sm font-semibold text-emerald-600/80 uppercase tracking-widest">Pemasukan</p>
                    <p class="text-2xl font-bold text-emerald-700 dark:text-emerald-400 mt-1">{{ formatCurrency(totalIncome) }}</p>
                </div>
                <div class="h-10 w-10 rounded-full bg-emerald-100 dark:bg-emerald-900/30 flex items-center justify-center text-emerald-600">
                    <ArrowUpCircle class="h-6 w-6" />
                </div>
            </div>
        </div>

        <div class="relative overflow-hidden rounded-3xl p-6 bg-gradient-to-br from-red-50 to-orange-50 dark:from-red-950/20 dark:to-orange-950/20 border border-red-100 dark:border-red-900/30 shadow-sm">
            <div class="flex justify-between items-start">
                <div>
                    <p class="text-sm font-semibold text-red-600/80 uppercase tracking-widest">Pengeluaran</p>
                    <p class="text-2xl font-bold text-red-700 dark:text-red-400 mt-1">{{ formatCurrency(totalExpense) }}</p>
                </div>
                <div class="h-10 w-10 rounded-full bg-red-100 dark:bg-red-900/30 flex items-center justify-center text-red-600">
                    <ArrowDownCircle class="h-6 w-6" />
                </div>
            </div>
        </div>

        <div class="relative overflow-hidden rounded-3xl p-6 bg-card border border-border shadow-sm">
             <div class="flex justify-between items-start">
                <div>
                    <p class="text-sm font-semibold text-muted-foreground uppercase tracking-widest">Saldo Bersih</p>
                    <p class="text-2xl font-bold text-foreground mt-1">{{ formatCurrency(totalIncome - totalExpense) }}</p>
                </div>
                 <div class="h-10 w-10 rounded-full bg-muted flex items-center justify-center text-muted-foreground">
                    <Wallet class="h-6 w-6" />
                </div>
            </div>
        </div>
    </div>

    <!-- Main Content -->
    <div class="space-y-6 relative">
        <!-- Toolbar -->
        <div class="flex flex-col md:flex-row gap-4 items-center justify-between bg-card p-3 rounded-2xl border border-border shadow-sm">
            
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

                <div class="flex gap-2 w-full md:w-auto">
                     <Select v-model="filterWallet">
                        <SelectTrigger class="flex-1 md:w-[140px] h-9 rounded-xl text-xs font-semibold">
                            <SelectValue placeholder="Semua Dompet" />
                        </SelectTrigger>
                        <SelectContent>
                            <SelectItem value="all">Semua Dompet</SelectItem>
                            <SelectItem v-for="w in walletStore.wallets" :key="w.id" :value="String(w.id)">{{ w.name }}</SelectItem>
                        </SelectContent>
                    </Select>

                    <Select v-model="filterCategory">
                        <SelectTrigger class="flex-1 md:w-[140px] h-9 rounded-xl text-xs font-semibold">
                            <SelectValue placeholder="Semua Kategori" />
                        </SelectTrigger>
                        <SelectContent>
                            <SelectItem value="all">Semua Kategori</SelectItem>
                             <SelectItem v-for="c in categoryStore.categories" :key="c.id" :value="String(c.id)">{{ c.name }}</SelectItem>
                        </SelectContent>
                    </Select>
                </div>
            </div>
        </div>

        <div v-if="periodType === 'custom' && showDatePicker" class="absolute left-0 right-0 top-16 z-50 flex justify-center p-0 animate-in fade-in zoom-in-95 duration-200 pointer-events-none">
            <div class="pointer-events-auto">
                <DateRangePicker 
                    :startDate="customDateRange.start"
                    :endDate="customDateRange.end"
                    @update:startDate="(val) => customDateRange.start = val"
                    @update:endDate="(val) => customDateRange.end = val"
                    @apply="showDatePicker = false"
                />
            </div>
        </div>

        <div class="grid lg:grid-cols-3 gap-6 md:h-[600px] overflow-hidden">
             <!-- Chart Section -->
             <Card class="lg:col-span-2 bg-card border-border shadow-sm flex flex-col rounded-3xl overflow-hidden h-full">
                <CardHeader class="pb-2 border-b border-border/50">
                    <CardTitle class="text-base font-bold flex items-center gap-2">
                        Grafik Pertumbuhan
                    </CardTitle>
                </CardHeader>
                <CardContent class="flex-1 p-4 relative min-h-[250px] md:min-h-[300px]">
                    <apexchart type="area" height="100%" width="100%" :options="chartOptions" :series="chartSeries" />
                </CardContent>
            </Card>

            <!-- Transaction List -->
             <Card class="bg-card border-border shadow-sm flex flex-col rounded-3xl overflow-hidden h-full">
                 <CardHeader class="pb-3 border-b border-border/50">
                    <div class="flex items-center gap-2">
                        <div class="relative flex-1">
                            <Search class="absolute left-2.5 top-2.5 h-3.5 w-3.5 text-muted-foreground" />
                            <Input v-model="searchQuery" placeholder="Cari transaksi..." class="h-9 pl-8 rounded-full bg-muted/50 border-transparent focus:bg-background transition-all text-xs" />
                        </div>
                    </div>
                </CardHeader>
                <CardContent class="overflow-y-auto p-4 space-y-6 flex-1 pr-2 custom-scrollbar">
                    
                     <div v-if="filteredTransactions.length === 0" class="flex flex-col items-center justify-center h-full text-muted-foreground opacity-60 text-center">
                        <div class="h-12 w-12 bg-muted rounded-full flex items-center justify-center mb-2">
                            <Search class="h-5 w-5" />
                        </div>
                        <p class="text-sm font-medium">Tidak ada transaksi</p>
                    </div>

                    <div v-for="group in groupedTransactions" :key="group.date" class="space-y-3">
                        <div class="flex items-center gap-2">
                            <span class="text-[10px] font-bold uppercase tracking-widest text-muted-foreground bg-muted px-2 py-1 rounded-md">{{ group.label }}</span>
                            <div class="h-[1px] flex-1 bg-border"></div>
                        </div>
                        
                        <div v-for="t in group.items" :key="t.id" class="group relative flex items-center justify-between p-3 rounded-2xl hover:bg-muted/50 transition-all cursor-default border border-transparent hover:border-border">
                             <div class="flex items-center gap-3">
                                 <div :class="['h-10 w-10 rounded-xl flex items-center justify-center text-lg shadow-sm transition-transform group-hover:scale-105', 
                                     t.type === 'expense' ? 'bg-red-50 text-red-500' : 
                                     t.type === 'income' ? 'bg-emerald-50 text-emerald-600' : 
                                     'bg-blue-50 text-blue-600']">
                                     <span v-if="getEmoji(t.category.icon)" class="text-xl leading-none filter drop-shadow-sm">{{ getEmoji(t.category.icon) }}</span>
                                     <component v-else :is="getIconComponent(t.category.icon)" class="h-5 w-5" />
                                </div>
                                <div>
                                    <p class="font-bold text-sm truncate max-w-[120px]">{{ t.description || 'No Description' }}</p>
                                    <p class="text-[10px] text-muted-foreground font-medium flex items-center gap-1">
                                        {{ t.wallet.name }} • {{ t.category.name }}
                                    </p>
                                </div>
                            </div>
                             <div class="text-right">
                                 <span :class="['block font-bold text-sm', 
                                    t.type === 'income' ? 'text-emerald-600' : 
                                    t.type === 'expense' ? 'text-red-500' : 
                                    'text-blue-600']">
                                    {{ (t.type === 'income' || t.type === 'transfer_in') ? '+' : '-' }} {{ formatCurrency(t.amount) }}
                                 </span>
                                 <span class="text-[10px] text-muted-foreground">{{ format(parseISO(t.date), 'HH:mm') }}</span>
                            </div>
                        </div>
                    </div>
                </CardContent>
            </Card>
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
