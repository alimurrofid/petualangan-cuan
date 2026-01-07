<script setup lang="ts">
import { computed, onMounted } from 'vue';
import { useAuthStore } from "@/stores/auth";
import { useDashboardStore } from "@/stores/dashboard";
import { 
  Card, 
  CardContent, 
  CardHeader, 
  CardTitle,
  CardDescription
} from '@/components/ui/card';
import { Button } from "@/components/ui/button";

import { 
  ArrowUp, 
  ArrowDown, 
  Wallet,
  TrendingUp,
  Lightbulb,
  Paperclip
} from 'lucide-vue-next';
import { getEmoji, getIconComponent } from "@/lib/icons";
import { format, parseISO, isSameDay, subDays } from 'date-fns';
import { id as localeId } from "date-fns/locale";

const authStore = useAuthStore();
const dashboardStore = useDashboardStore();
const baseUrl = import.meta.env.VITE_API_BASE_URL;

onMounted(() => {
    dashboardStore.fetchDashboard();
});

const data = computed(() => dashboardStore.data);

// Charts & Visuals
const formatCurrency = (value: number) => {
  return new Intl.NumberFormat("id-ID", { style: "currency", currency: "IDR", minimumFractionDigits: 0 }).format(value);
};

// Trend Chart
const filledMonthlyTrend = computed(() => {
    if (!data.value) return [];
    
    const filled: { date: string; income: number; expense: number }[] = [];
    const today = new Date();
    
    // Generate last 6 months
    for (let i = 5; i >= 0; i--) {
        // We just need YYYY-MM keys
        const dateObj = new Date(today.getFullYear(), today.getMonth() - i, 1);
        const key = format(dateObj, 'yyyy-MM');
        
        const existing = (data.value.monthly_trend || []).find(item => item.date === key);
        if (existing) {
            filled.push(existing);
        } else {
            filled.push({ date: key, income: 0, expense: 0 });
        }
    }
    return filled;
});

const chartSeriesArea = computed(() => {
    return [
        { name: 'Pemasukan', data: filledMonthlyTrend.value.map(d => d.income) },
        { name: 'Pengeluaran', data: filledMonthlyTrend.value.map(d => d.expense) }
    ];
});

const chartOptionsArea = computed(() => ({
  chart: {
    type: 'area',
    toolbar: { show: false },
    zoom: { enabled: false },
    foreColor: '#94a3b8' 
  },
  dataLabels: { enabled: false },
  stroke: { curve: 'smooth', width: 2 },
  colors: ['#10b981', '#ef4444'],
  fill: {
    type: 'gradient',
    gradient: {
      shadeIntensity: 1,
      opacityFrom: 0.4,
      opacityTo: 0.1,
      stops: [0, 90, 100]
    }
  },
  xaxis: {
    categories: filledMonthlyTrend.value.map(d => format(parseISO(d.date + "-01"), "MMM", { locale: localeId })),
    axisBorder: { show: false },
    axisTicks: { show: false }
  },
  yaxis: { show: false },
  grid: { 
     show: true,
     borderColor: '#334155', 
     strokeDashArray: 4,
  },
  tooltip: { theme: 'dark' }
}));

// Donut Chart (Breakdown)
const chartSeriesDonut = computed(() => {
    return data.value?.expense_breakdown?.map(item => item.total_amount) || [];
});

const chartLabelsDonut = computed(() => {
    return data.value?.expense_breakdown?.map(item => item.category_name) || [];
});

const totalExpenseAmount = computed(() => {
    return data.value?.expense_breakdown?.reduce((sum, item) => sum + item.total_amount, 0) || 0;
});

const chartOptionsDonut = computed(() => ({
    chart: { type: 'donut', fontFamily: 'inherit', foreColor: '#94a3b8' },
    labels: chartLabelsDonut.value,
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
                        formatter: () => formatCurrency(totalExpenseAmount.value)
                    }
                }
            }
        }
    },
    dataLabels: { enabled: false },
    legend: { position: 'bottom' },
    colors: ['#ef4444', '#f59e0b', '#3b82f6', '#10b981', '#8b5cf6'],
    stroke: { show: false },
    tooltip: {
        theme: 'dark',
        y: { formatter: (val: number) => formatCurrency(val) }
    }
}));

// Budget Status (Top 5 Expenses)
const budgetStatus = computed(() => {
    if (!data.value || !data.value.expense_breakdown) return [];
    // Filter expense items that have budget logic or just use breakdown
    return data.value.expense_breakdown
        .filter(item => item.type === 'expense')
        // Removed slice(0, 5) to show all categories as requested
        .map(item => {
            const limit = item.budget_limit || 0;
            const spent = item.total_amount;
            let percentage = 0;
            if (limit > 0) percentage = Math.round((spent / limit) * 100);
            
            return {
                name: item.category_name,
                percentage,
                isOver: item.is_over_budget,
                limit
            };
        })
        .sort((a, b) => b.percentage - a.percentage);
});

const getWalletColorClass = (type: string) => {
    const t = type.toLowerCase();
    if (t.includes('bank')) return 'from-blue-500 to-blue-600 text-white';
    if (t.includes('wallet')) return 'from-purple-500 to-purple-600 text-white';
    // Cash or others
    return 'from-emerald-500 to-emerald-600 text-white'; 
};


// Recent Transactions Grouping
const groupedRecentTransactions = computed(() => {
    if (!data.value || !data.value.recent_transactions) return [];
    const groups: Record<string, any[]> = {};
    
    data.value.recent_transactions.forEach(t => {
        const dateKey = format(parseISO(t.date), 'yyyy-MM-dd');
        if (!groups[dateKey]) groups[dateKey] = [];
        groups[dateKey].push(t);
    });

    const sortedTypes = Object.keys(groups).sort((a,b) => b.localeCompare(a));

    return sortedTypes.map(dateStr => {
        const date = parseISO(dateStr);
        let label = format(date, 'EEEE, d MMMM yyyy', { locale: localeId });
        
        if (isSameDay(date, new Date())) label = "Hari Ini";
        if (isSameDay(date, subDays(new Date(), 1))) label = "Kemarin";

        return {
            date: dateStr,
            label,
            items: groups[dateStr]
        };
    });
});</script>

<template>
  <div class="flex-1 space-y-6 pt-2" v-if="dashboardStore.isLoading">
      <div class="flex items-center justify-center min-h-[400px]">
          <p class="text-muted-foreground animate-pulse">Memuat data dashboard...</p>
      </div>
  </div>
  
  <div class="flex-1 space-y-6 pt-2 text-foreground" v-else-if="data">
    <div class="flex items-center justify-between">
      <div>
          <h2 class="text-3xl font-bold tracking-tight">Dashboard</h2>
          <p class="text-sm text-muted-foreground mt-1">Selamat datang kembali, <span class="capitalize">{{ authStore.user?.name || 'Bro' }}</span> 👋</p>
      </div>
    </div>

    <div class="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
      <Card class="bg-gradient-to-br from-indigo-500 to-purple-600 text-white border-none shadow-md">
        <CardHeader class="flex flex-row items-center justify-between space-y-0 pb-2">
          <CardTitle class="text-sm font-medium text-white/90">Total Saldo</CardTitle>
          <Wallet class="h-4 w-4 text-white/70" />
        </CardHeader>
        <CardContent>
          <div class="text-2xl font-bold">{{ formatCurrency(data.total_balance) }}</div>
          <p class="text-xs text-indigo-100/70 mt-1">Total aset saat ini</p>
        </CardContent>
      </Card>

      <Card>
        <CardHeader class="flex flex-row items-center justify-between space-y-0 pb-2">
          <CardTitle class="text-sm font-medium">Pemasukan</CardTitle>
          <div class="h-8 w-8 rounded-full bg-emerald-100 flex items-center justify-center">
             <ArrowUp class="h-4 w-4 text-emerald-600" />
          </div>
        </CardHeader>
        <CardContent>
          <div class="text-2xl font-bold text-emerald-600">{{ formatCurrency(data.total_income_month) }}</div>
          <p class="text-xs text-muted-foreground mt-1">Bulan Ini</p>
        </CardContent>
      </Card>

      <Card>
        <CardHeader class="flex flex-row items-center justify-between space-y-0 pb-2">
          <CardTitle class="text-sm font-medium">Pengeluaran</CardTitle>
          <div class="h-8 w-8 rounded-full bg-rose-100 flex items-center justify-center">
             <ArrowDown class="h-4 w-4 text-rose-600" />
          </div>
        </CardHeader>
        <CardContent>
          <div class="text-2xl font-bold text-rose-600">{{ formatCurrency(data.total_expense_month) }}</div>
          <p class="text-xs text-muted-foreground mt-1">Bulan Ini</p>
        </CardContent>
      </Card>

      <Card class="bg-gradient-to-br from-amber-100 to-orange-50 border-amber-200">
        <CardHeader class="flex flex-row items-center justify-between space-y-0 pb-2">
          <CardTitle class="text-sm font-medium text-amber-900">AI Tip Harian</CardTitle>
          <Lightbulb class="h-4 w-4 text-amber-600" />
        </CardHeader>
        <CardContent>
          <p class="text-xs font-medium text-amber-800 leading-relaxed">
            "Hindari belanja impulsif dengan menerapkan aturan tunggu 24 jam sebelum membeli barang non-pokok."
          </p>
        </CardContent>
      </Card>
    </div>

    <div class="grid gap-4 grid-cols-1 md:grid-cols-2 lg:grid-cols-7">
      
      <Card class="col-span-1 md:col-span-2 lg:col-span-4 bg-card shadow-sm rounded-3xl">
        <CardHeader>
          <CardTitle class="flex items-center gap-2">
              <TrendingUp class="h-4 w-4 text-primary" /> Tren Keuangan
          </CardTitle>
          <CardDescription>Perbandingan pemasukan dan pengeluaran 6 bulan terakhir.</CardDescription>
        </CardHeader>
        <CardContent class="pl-0">
            <apexchart type="area" height="300" :options="chartOptionsArea" :series="chartSeriesArea" />
        </CardContent>
      </Card>

      <Card class="col-span-1 md:col-span-2 lg:col-span-3 bg-card shadow-sm rounded-3xl">
        <CardHeader>
            <CardTitle>Distribusi Pengeluaran</CardTitle>
            <CardDescription>Berdasarkan kategori bulan ini.</CardDescription>
        </CardHeader>
        <CardContent class="flex items-center justify-center flex-col">
             <div v-if="chartSeriesDonut.length === 0" class="text-center py-10 text-muted-foreground text-sm">Belum ada data pengeluaran.</div>
             <apexchart v-else type="donut" width="100%" :options="chartOptionsDonut" :series="chartSeriesDonut" />

             <div v-if="budgetStatus.length > 0" class="w-full mt-6 space-y-3">
                 <p class="text-xs font-bold uppercase text-muted-foreground tracking-widest">Status Anggaran</p>
                 <div v-for="item in budgetStatus" :key="item.name" class="flex items-center justify-between text-sm">
                     <span class="font-medium truncate max-w-[100px]">{{ item.name }}</span>
                     <div class="flex items-center gap-2 flex-1 mx-3">
                         <div class="h-1.5 flex-1 bg-muted rounded-full overflow-hidden">
                             <div :class="['h-full rounded-full', item.isOver ? 'bg-red-500' : item.percentage > 80 ? 'bg-amber-500' : 'bg-emerald-500']" :style="{ width: Math.min(item.percentage, 100) + '%' }"></div>
                         </div>
                     </div>
                     <span :class="['text-xs font-bold', item.isOver ? 'text-red-500' : 'text-muted-foreground']">
                         {{ item.percentage }}%
                     </span>
                 </div>
             </div>
        </CardContent>
      </Card>

    </div>

    <div class="grid gap-4 grid-cols-1 md:grid-cols-2 lg:grid-cols-3">
        <Card class="col-span-1 md:col-span-2 shadow-sm rounded-3xl">
            <CardHeader>
                <CardTitle>Transaksi Terakhir</CardTitle>
                <CardDescription>5 aktivitas finansial terbaru.</CardDescription>
            </CardHeader>
            <CardContent>
                <div class="space-y-4">
                    <div v-if="groupedRecentTransactions.length === 0" class="text-center py-6 text-muted-foreground text-sm">
                        Belum ada transaksi.
                    </div>
                    <div v-for="group in groupedRecentTransactions" :key="group.date" class="space-y-3">
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
                                     <span v-if="getEmoji(t.category?.icon)" class="text-lg leading-none">{{ getEmoji(t.category?.icon) }}</span>
                                     <component v-else :is="getIconComponent(t.category?.icon, 'Circle')" class="h-5 w-5" />
                                </div>
                                <div>
                                    <p class="font-bold text-sm truncate max-w-[180px] pb-0.5">{{ t.description || 'Tanpa Keterangan' }}</p>
                                    <p class="text-[10px] text-muted-foreground font-medium flex items-center gap-1">
                                        {{ t.wallet?.name }} • {{ t.category?.name }}
                                        <a v-if="t.attachment"
                                            :href="`${baseUrl.replace(/\/$/, '')}/${t.attachment.replace(/^\//, '')}`"
                                            target="_blank" @click.stop
                                            class="text-xs text-blue-500 hover:text-blue-700 flex items-center">
                                            <Paperclip class="w-3 h-3" />
                                        </a>
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
                </div>
            </CardContent>
        </Card>

        <Card class="shadow-sm rounded-3xl">
            <CardHeader>
                <CardTitle>Dompet Saya</CardTitle>
                <CardDescription>Status saldo saat ini.</CardDescription>
            </CardHeader>
            <CardContent>
                <div class="space-y-4">
                     <div v-for="w in data.wallets" :key="w.id" :class="['p-4 rounded-xl border border-transparent shadow-md flex items-center justify-between w-full bg-gradient-to-br min-h-[80px]', getWalletColorClass(w.type)]">
                         <div class="flex items-center gap-4">
                             <div class="h-10 w-10 rounded-full bg-white/20 backdrop-blur-sm flex items-center justify-center text-xl shadow-sm">
                                  <span v-if="getEmoji(w.icon)" class="text-lg leading-none">{{ getEmoji(w.icon) }}</span>
                                  <component v-else :is="getIconComponent(w.icon || 'Wallet')" class="h-5 w-5 text-white" />
                             </div>
                             <div>
                                 <p class="text-xs font-bold opacity-80 uppercase tracking-wide">{{ w.type }}</p>
                                 <p class="font-bold text-base truncate max-w-[140px]">{{ w.name }}</p>
                             </div>
                         </div>
                         <p class="font-bold text-lg whitespace-nowrap drop-shadow-sm">{{ formatCurrency(w.balance) }}</p>
                     </div>
                     <Button variant="outline" class="w-full text-xs h-8 rounded-xl" @click="$router.push('/wallet')">
                        Lihat Semua Dompet
                     </Button>
                </div>
            </CardContent>
        </Card>
    </div>
  </div>
</template>