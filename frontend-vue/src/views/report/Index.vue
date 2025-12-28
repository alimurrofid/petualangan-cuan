<script setup lang="ts">
import { ref, computed, onMounted } from "vue";
import { 
  format, 
  parseISO, 
  startOfMonth, 
  endOfMonth, 
  isWithinInterval, 
  addMonths, 
  startOfWeek, 
  endOfWeek, 
  addWeeks, 
  startOfDay, 
  endOfDay, 
  addDays,
} from "date-fns";
import { id } from "date-fns/locale";

import { Card, CardHeader, CardTitle, CardContent } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";

import { ChevronLeft, ChevronRight, Calendar as CalendarIcon, PieChart as PieChartIcon } from "lucide-vue-next";

interface WalletItem {
  id: number;
  name: string;
}

interface CategoryItem {
  id: number;
  name: string;
  icon: string;
  isEmoji: boolean;
  type: "income" | "expense";
  budgetLimit?: number;
}

interface Transaction {
  id: number;
  title: string;
  amount: number;
  date: string;
  walletId: number;
  categoryId: number;
  type: "income" | "expense";
}

type PeriodType = 'monthly' | 'weekly' | 'daily' | 'custom';
type TransactionType = 'all' | 'income' | 'expense';

const transactions = ref<Transaction[]>([]);
const wallets = ref<WalletItem[]>([]);
const categories = ref<CategoryItem[]>([]);

const filterWallet = ref<string>("all");
const filterType = ref<TransactionType>("expense");
const periodType = ref<PeriodType>('monthly');

const selectedDate = ref(new Date());
const customDateRange = ref({
  start: new Date(),
  end: new Date() 
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

const loadData = () => {
    const savedWallets = localStorage.getItem("mock_wallets");
    if (savedWallets) wallets.value = JSON.parse(savedWallets);
    else wallets.value = [{ id: 1, name: "BCA Utama" }];

    const savedCategories = localStorage.getItem("mock_categories");
    if (savedCategories) categories.value = JSON.parse(savedCategories);
    else {
        categories.value = [
            { id: 1, name: "Makanan", icon: "Utensils", isEmoji: false, type: "expense", budgetLimit: 2000000 },
            { id: 2, name: "Gaji", icon: "💰", isEmoji: true, type: "income" },
            { id: 3, name: "Transport", icon: "Car", isEmoji: false, type: "expense", budgetLimit: 1000000 },
            { id: 4, name: "Bonus", icon: "Gift", isEmoji: false, type: "income" },
            { id: 5, name: "Belanja", icon: "ShoppingBag", isEmoji: false, type: "expense", budgetLimit: 1500000 },
            { id: 6, name: "Hiburan", icon: "Gamepad2", isEmoji: false, type: "expense", budgetLimit: 500000 },
            { id: 7, name: "Tagihan", icon: "Zap", isEmoji: false, type: "expense", budgetLimit: 750000 },
        ];
    }

    const savedTransactions = localStorage.getItem("mock_transactions");
    if (savedTransactions) transactions.value = JSON.parse(savedTransactions);
};

onMounted(loadData);

const filteredTransactions = computed(() => {
  return transactions.value.filter((t) => {
    const tDate = parseISO(t.date);
    const { start, end } = dateRange.value;
    if (!start || !end) return false;

    const matchesPeriod = isWithinInterval(tDate, { start, end });
    const matchesWallet = filterWallet.value === "all" || t.walletId === Number(filterWallet.value);
    const matchesType = filterType.value === "all" || t.type === filterType.value;

    return matchesPeriod && matchesWallet && matchesType;
  });
});

const totalAmount = computed(() => {
    return filteredTransactions.value.reduce((sum, t) => sum + t.amount, 0);
});

const categoryBreakdown = computed(() => {
    const groups: Record<number, number> = {};
    filteredTransactions.value.forEach(t => {
        if (!groups[t.categoryId]) groups[t.categoryId] = 0;
        groups[t.categoryId] = (groups[t.categoryId] || 0) + t.amount;
    });

    return Object.keys(groups).map(catId => {
        const id = Number(catId);
        const cat = categories.value.find(c => c.id === id);
        const budget = cat?.budgetLimit || 0;
        const spendingPercentage = budget > 0 ? ((groups[id] || 0) / budget) * 100 : 0;
        
        return {
            name: cat?.name || 'Unknown',
            amount: groups[id] || 0,
            percentage: totalAmount.value ? ((groups[id] || 0) / totalAmount.value) * 100 : 0,
            budget,
            spendingPercentage,
            isOverbudget: budget > 0 && (groups[id] || 0) > budget
        };
    }).sort((a,b) => (b.amount || 0) - (a.amount || 0));
});

const chartSeries = computed(() => categoryBreakdown.value.map(c => c.amount));
const chartLabels = computed(() => categoryBreakdown.value.map(c => c.name));

const chartOptions = computed(() => ({
    chart: { type: 'donut', fontFamily: 'inherit', foreColor: '#94a3b8' },
    labels: chartLabels.value,
    plotOptions: {
        pie: {
            donut: {
                size: '65%',
                labels: {
                    show: true,
                    total: {
                        show: true,
                        label: 'Total',
                        formatter: () => formatCurrency(totalAmount.value, true),
                        color: '#94a3b8'
                    }
                }
            }
        }
    },
    dataLabels: { enabled: false },
    legend: { position: 'bottom' },
    colors: ['#10b981', '#3b82f6', '#f59e0b', '#ef4444', '#8b5cf6', '#ec4899', '#6366f1'],
    stroke: { show: false },
    tooltip: {
        theme: 'dark',
        y: { formatter: (val: number) => formatCurrency(val) }
    }
}));


const formatCurrency = (value: number, short = false) => {
    if (short) {
        if (value >= 1000000) return (value / 1000000).toFixed(1) + 'jt';
        if (value >= 1000) return (value / 1000).toFixed(0) + 'rb';
    }
    return new Intl.NumberFormat("id-ID", { style: "currency", currency: "IDR", minimumFractionDigits: 0 }).format(value);
};

</script>

<template>
  <div class="p-6 space-y-6 text-foreground min-h-screen bg-background">
      
    <div class="flex flex-col gap-2">
      <h2 class="text-3xl font-bold tracking-tight">Laporan Keuangan</h2>
      <p class="text-sm text-muted-foreground">Analisis pengeluaran dan pemasukan Anda secara detail.</p>
    </div>

    <div class="flex flex-col md:flex-row gap-4 items-center justify-between bg-card p-3 rounded-2xl border border-border shadow-sm">
        
        <div class="flex items-center gap-2 bg-muted/30 p-1 rounded-xl w-full md:w-auto justify-between md:justify-start">
                <Button variant="ghost" size="icon" @click="navigateDate(-1)" class="h-8 w-8 rounded-lg hover:bg-background shadow-sm">
                <ChevronLeft class="h-4 w-4" />
            </Button>
            <div class="flex-1 text-center md:px-4 text-sm font-bold flex items-center justify-center gap-2 min-w-[140px]">
                <CalendarIcon class="h-4 w-4 opacity-50" />
                {{ formattedDateRange }}
            </div>
            <Button variant="ghost" size="icon" @click="navigateDate(1)" class="h-8 w-8 rounded-lg hover:bg-background shadow-sm">
                <ChevronRight class="h-4 w-4" />
            </Button>
        </div>

        <div class="flex flex-col md:flex-row items-center gap-2 w-full md:w-auto">
            <Select v-model="periodType">
                <SelectTrigger class="w-full md:w-[120px] h-9 rounded-xl text-xs font-semibold">
                    <SelectValue />
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
                        <SelectItem v-for="w in wallets" :key="w.id" :value="String(w.id)">{{ w.name }}</SelectItem>
                    </SelectContent>
                </Select>

                <Select v-model="filterType">
                    <SelectTrigger class="flex-1 md:w-[140px] h-9 rounded-xl text-xs font-semibold">
                        <SelectValue />
                    </SelectTrigger>
                    <SelectContent>
                        <SelectItem value="all">Semua Tipe</SelectItem>
                        <SelectItem value="expense">Pengeluaran</SelectItem>
                        <SelectItem value="income">Pemasukan</SelectItem>
                    </SelectContent>
                </Select>
            </div>
        </div>
    </div>

    <div v-if="periodType === 'custom'" class="flex flex-col md:flex-row gap-4 p-4 bg-muted/40 rounded-2xl border border-dashed border-border">
        <div class="flex-1 space-y-1">
            <Label class="text-xs">Dari Tanggal</Label>
            <Input type="date" :value="format(customDateRange.start, 'yyyy-MM-dd')" @input="(e: any) => customDateRange.start = new Date(e.target.value)" class="bg-background" />
        </div>
        <div class="flex-1 space-y-1">
            <Label class="text-xs">Sampai Tanggal</Label>
            <Input type="date" :value="format(customDateRange.end, 'yyyy-MM-dd')" @input="(e: any) => customDateRange.end = new Date(e.target.value)" class="bg-background" />
        </div>
    </div>

    <div class="grid lg:grid-cols-3 gap-6">
        
        <Card class="bg-card border-border shadow-sm rounded-3xl lg:col-span-1 overflow-hidden">
            <CardHeader>
                <CardTitle class="text-base flex items-center gap-2">
                    <PieChartIcon class="h-4 w-4 text-muted-foreground" /> Breakdown Kategori
                </CardTitle>
            </CardHeader>
            <CardContent class="flex items-center justify-center min-h-[300px]">
                <div v-if="filteredTransactions.length === 0" class="text-center text-muted-foreground text-sm opacity-60">
                    Tidak ada data untuk ditampilkan.
                </div>
                <apexchart v-else type="donut" width="100%" :options="chartOptions" :series="chartSeries" />
            </CardContent>
        </Card>

        <Card class="bg-card border-border shadow-sm rounded-3xl lg:col-span-2 flex flex-col h-[500px] overflow-hidden">
             <CardHeader class="pb-2 border-b border-border/50">
                <CardTitle class="text-base flex items-center justify-between">
                     <span>Rincian Kategori</span>
                     <span class="text-sm font-bold text-muted-foreground bg-muted/50 px-3 py-1 rounded-full">Total: {{ formatCurrency(totalAmount) }}</span>
                </CardTitle>
            </CardHeader>
            <CardContent class="flex-1 overflow-y-auto p-4 custom-scrollbar space-y-4">
                 <div v-if="categoryBreakdown.length === 0" class="text-center py-20 text-muted-foreground opacity-60">
                    Belum ada data.
                </div>

                <div v-for="(cat, idx) in categoryBreakdown" :key="idx" class="flex items-center justify-between p-3 rounded-2xl border border-border/60 hover:bg-muted/30 transition-all">
                     <div class="flex items-center gap-3">
                         <div class="h-10 w-10 flex items-center justify-center font-bold text-muted-foreground bg-muted rounded-xl">
                             {{ idx + 1 }}
                         </div>
                         <div>
                             <p class="font-bold text-sm">{{ cat.name }}</p>
                             <div class="w-24 h-1.5 bg-muted rounded-full mt-1 overflow-hidden" v-if="cat.budget > 0">
                                 <div :class="['h-full rounded-full', cat.isOverbudget ? 'bg-red-500' : 'bg-primary']" :style="{ width: Math.min(cat.spendingPercentage, 100) + '%' }"></div>
                             </div>
                             <div v-else class="w-24 h-1.5 bg-muted rounded-full mt-1 overflow-hidden">
                                  <div class="h-full bg-primary rounded-full" :style="{ width: cat.percentage + '%' }"></div>
                             </div>
                         </div>
                     </div>
                     <div class="text-right">
                         <p class="font-bold text-sm">{{ formatCurrency(cat.amount) }}</p>
                         <p v-if="cat.budget > 0" :class="['text-[10px]', cat.isOverbudget ? 'text-red-500 font-bold' : 'text-muted-foreground']">
                            {{ cat.isOverbudget ? 'Overbudget' : 'Budget' }}: {{ formatCurrency(cat.budget, true) }}
                         </p>
                         <p v-else class="text-[10px] text-muted-foreground">{{ cat.percentage.toFixed(1) }}%</p>
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
