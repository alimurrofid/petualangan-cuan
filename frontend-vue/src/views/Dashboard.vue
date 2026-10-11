<script setup lang="ts">
import { computed, onMounted, ref } from 'vue';
import { isDark } from '@/composables/useTheme';
import { useAuthStore } from "@/stores/auth";
import { useDashboardStore } from "@/stores/dashboard";
import { useFinancialHealthStore } from "@/stores/financialHealth";
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
  Paperclip,
  HeartPulse
} from 'lucide-vue-next';
import { getEmoji, getIconComponent, getBankIcon, getBankLogo } from "@/lib/icons";
import { formatCompactNumber } from "@/lib/utils";
import { format, parseISO, isSameDay, subDays } from 'date-fns';
import { id as localeId } from "date-fns/locale";

const authStore = useAuthStore();
const dashboardStore = useDashboardStore();
const healthStore = useFinancialHealthStore();
const baseUrl = import.meta.env.VITE_API_BASE_URL;

const data = computed(() => dashboardStore.data);
const isMounted = ref(false);

onMounted(() => {
    dashboardStore.fetchDashboard();
    healthStore.fetchFinancialHealth();
    setTimeout(() => {
        isMounted.value = true;
    }, 100);
});

const formatCurrency = (value: number) => {
  return new Intl.NumberFormat("id-ID", { style: "currency", currency: "IDR", minimumFractionDigits: 0 }).format(value);
};

const filledMonthlyTrend = computed(() => {
    if (!data.value) return [];
    
    const filled: { date: string; income: number; expense: number }[] = [];
    const today = new Date();
    
    for (let i = 5; i >= 0; i--) {
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
    height: '100%',
    toolbar: { show: false },
    zoom: { enabled: false },
    foreColor: isDark.value ? '#94a3b8' : '#64748b' 
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
  yaxis: { 
    show: false,
    labels: {
        formatter: (value: number) => formatCurrency(value)
    }
  },
  grid: { 
     show: true,
     borderColor: isDark.value ? '#334155' : '#e2e8f0', 
     strokeDashArray: 4,
  },
  tooltip: { 
    theme: isDark.value ? 'dark' : 'light',
    y: {
        formatter: (value: number) => formatCurrency(value)
    }
  }
}));

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
    chart: { type: 'donut', fontFamily: 'inherit', foreColor: isDark.value ? '#94a3b8' : '#64748b' },
    labels: chartLabelsDonut.value,
    plotOptions: {
        pie: {
            donut: {
                size: '70%',
                labels: {
                    show: true,
                    value: {
                        show: true,
                        formatter: (val: number) => formatCompactNumber(val)
                    },
                    total: {
                        show: true,
                        label: 'Total',
                        formatter: () => formatCompactNumber(totalExpenseAmount.value)
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
        theme: isDark.value ? 'dark' : 'light',
        y: { formatter: (val: number) => formatCurrency(val) }
    }
}));

const budgetStatus = computed(() => {
    if (!data.value || !data.value.expense_breakdown) return [];
    return data.value.expense_breakdown
        .filter(item => item.type === 'expense')
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

const getWalletColorClass = (w: any) => {
    const bank = getBankIcon(w?.icon);
    if (bank) {
        return `bg-gradient-to-br ${bank.gradient} ${bank.textColor}`;
    }
    const t = (w?.type || '').toLowerCase();
    if (t.includes('bank')) return 'bg-gradient-to-br from-blue-600 via-blue-700 to-indigo-900 text-white';
    if (t.includes('wallet')) return 'bg-gradient-to-br from-purple-600 via-purple-700 to-indigo-900 text-white';
    if (t.includes('cash')) return 'bg-gradient-to-br from-emerald-600 via-teal-700 to-emerald-900 text-white';
    return 'bg-gradient-to-br from-slate-700 via-slate-800 to-slate-900 text-white'; 
};

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

    <!-- Banner Sisa Aman Per Hari (The Daily Pace Metric) -->
    <div v-if="data.pace_metric" class="relative overflow-hidden rounded-3xl p-5 md:p-6 shadow-sm border transition-all"
      :class="[
        data.pace_metric.status === 'OVERPACE' 
          ? 'bg-gradient-to-r from-red-50 via-rose-50/60 to-background dark:from-red-950/40 dark:via-rose-900/20 dark:to-background border-red-200 dark:border-red-500/30' 
          : data.pace_metric.status === 'WARNING'
          ? 'bg-gradient-to-r from-amber-50 via-yellow-50/60 to-background dark:from-amber-950/40 dark:via-yellow-900/20 dark:to-background border-amber-200 dark:border-amber-500/30'
          : 'bg-gradient-to-r from-emerald-50 via-teal-50/60 to-background dark:from-emerald-950/40 dark:via-teal-900/20 dark:to-background border-emerald-200 dark:border-emerald-500/30'
      ]">
      <div class="flex flex-col md:flex-row md:items-center justify-between gap-4">
        <div class="space-y-1.5 flex-1 min-w-0">
          <div class="flex flex-wrap items-center gap-2">
            <span class="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-bold tracking-wider uppercase shrink-0"
              :class="[
                data.pace_metric.status === 'OVERPACE' ? 'bg-red-100 text-red-700 dark:bg-red-500/20 dark:text-red-400 border border-red-200 dark:border-red-500/30' :
                data.pace_metric.status === 'WARNING' ? 'bg-amber-100 text-amber-700 dark:bg-amber-500/20 dark:text-amber-400 border border-amber-200 dark:border-amber-500/30' :
                'bg-emerald-100 text-emerald-700 dark:bg-emerald-500/20 dark:text-emerald-400 border border-emerald-200 dark:border-emerald-500/30'
              ]">
              <span class="w-1.5 h-1.5 rounded-full mr-1.5 animate-pulse"
                :class="data.pace_metric.status === 'OVERPACE' ? 'bg-red-500 dark:bg-red-400' : data.pace_metric.status === 'WARNING' ? 'bg-amber-500 dark:bg-amber-400' : 'bg-emerald-500 dark:bg-emerald-400'"></span>
              {{ data.pace_metric.status === 'OVERPACE' ? 'Melebihi Kuota' : data.pace_metric.status === 'WARNING' ? 'Mendekati Batas' : 'Pace Aman' }}
            </span>
            <span class="text-xs text-muted-foreground whitespace-nowrap">• Sisa {{ data.pace_metric.days_remaining }} hari menuju gajian</span>
          </div>
          <p class="text-xs md:text-sm text-muted-foreground leading-relaxed">Sisa kuota belanja harian yang aman agar anggaran bulananmu terjaga:</p>
          <div class="flex items-baseline flex-wrap gap-1.5 sm:gap-2 pt-1">
            <span class="text-2xl sm:text-3xl md:text-4xl font-extrabold tracking-tight font-mono break-all sm:break-normal"
              :class="[
                data.pace_metric.status === 'OVERPACE' ? 'text-red-600 dark:text-red-400' :
                data.pace_metric.status === 'WARNING' ? 'text-amber-600 dark:text-amber-400' :
                'text-emerald-600 dark:text-emerald-400'
              ]">
              {{ formatCurrency(data.pace_metric.daily_safe_to_spend) }}
            </span>
            <span class="text-xs sm:text-sm font-medium text-muted-foreground">/ hari</span>
          </div>
        </div>

        <div class="grid grid-cols-2 sm:grid-cols-3 md:flex md:flex-col items-start md:items-end gap-2.5 sm:gap-4 md:gap-1 text-xs border-t md:border-t-0 md:border-l border-border/50 pt-3 md:pt-0 md:pl-6 w-full md:w-auto shrink-0">
          <div class="flex flex-col sm:flex-row sm:items-center gap-0.5 sm:gap-1.5">
            <span class="text-muted-foreground text-[11px] sm:text-xs">Belanja Hari Ini:</span>
            <span class="font-bold text-foreground font-mono">{{ formatCurrency(data.pace_metric.today_expense) }}</span>
          </div>
          <div class="flex flex-col sm:flex-row sm:items-center gap-0.5 sm:gap-1.5">
            <span class="text-muted-foreground text-[11px] sm:text-xs">Sisa Anggaran:</span>
            <span class="font-bold font-mono" :class="data.pace_metric.remaining_budget < 0 ? 'text-red-600 dark:text-red-400' : 'text-foreground'">
              {{ formatCurrency(data.pace_metric.remaining_budget) }}
            </span>
          </div>
          <div class="flex flex-col sm:flex-row sm:items-center gap-0.5 sm:gap-1.5 col-span-2 sm:col-span-1">
            <span class="text-muted-foreground text-[11px] sm:text-xs">Total Budget:</span>
            <span class="font-medium text-muted-foreground font-mono">{{ formatCurrency(data.pace_metric.total_budget) }}</span>
          </div>
        </div>
      </div>
    </div>

    <div class="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
      <Card class="bg-gradient-to-br from-indigo-500 to-purple-600 text-white border-none shadow-md">
        <CardHeader class="flex flex-row items-center justify-between space-y-0 pb-2">
          <CardTitle class="text-sm font-medium text-white/90">Saldo Tersedia</CardTitle>
          <Wallet class="h-4 w-4 text-white/70" />
        </CardHeader>
        <CardContent>
          <div class="text-2xl font-bold" :class="{ 'privacy-blur': authStore.isPrivacyMode }">{{ formatCurrency(data.total_available_balance) }}</div>
          <div class="flex items-center gap-1 mt-1 text-xs text-indigo-100/70">
              <span class="opacity-70">Saldo Keseluruhan:</span>
              <span class="font-bold" :class="{ 'privacy-blur': authStore.isPrivacyMode }">{{ formatCurrency(data.total_balance) }}</span>
          </div>
        </CardContent>
      </Card>

      <Card class="bg-gradient-to-br from-emerald-50 to-teal-50 dark:from-emerald-950/20 dark:to-teal-950/20 border-emerald-100 dark:border-emerald-900/30">
        <CardHeader class="flex flex-row items-center justify-between space-y-0 pb-2">
          <CardTitle class="text-sm font-medium">Pemasukan</CardTitle>
          <div class="h-8 w-8 rounded-full bg-emerald-100 dark:bg-emerald-950/60 flex items-center justify-center">
             <ArrowUp class="h-4 w-4 text-emerald-600 dark:text-emerald-400" />
          </div>
        </CardHeader>
        <CardContent>
          <div class="text-2xl font-bold text-emerald-600 dark:text-emerald-400" :class="{ 'privacy-blur': authStore.isPrivacyMode }">{{ formatCurrency(data.total_income_month) }}</div>
          <p class="text-xs text-muted-foreground mt-1">Bulan Ini</p>
        </CardContent>
      </Card>

      <Card class="bg-gradient-to-br from-red-50 to-orange-50 dark:from-red-950/20 dark:to-orange-950/20 border-red-100 dark:border-red-900/30">
        <CardHeader class="flex flex-row items-center justify-between space-y-0 pb-2">
          <CardTitle class="text-sm font-medium">Pengeluaran</CardTitle>
          <div class="h-8 w-8 rounded-full bg-rose-100 dark:bg-rose-950/60 flex items-center justify-center">
             <ArrowDown class="h-4 w-4 text-rose-600 dark:text-rose-400" />
          </div>
        </CardHeader>
        <CardContent>
          <div class="text-2xl font-bold text-rose-600 dark:text-rose-400" :class="{ 'privacy-blur': authStore.isPrivacyMode }">{{ formatCurrency(data.total_expense_month) }}</div>
          <p class="text-xs text-muted-foreground mt-1">Bulan Ini</p>
        </CardContent>
      </Card>

      <Card class="bg-gradient-to-br from-blue-50 to-indigo-50 border-blue-200 dark:from-blue-950/20 dark:to-indigo-950/20 dark:border-blue-900/30 cursor-pointer hover:shadow-md transition-all group" @click="$router.push('/financial-health')">
        <CardHeader class="flex flex-row items-center justify-between space-y-0 pb-2">
          <CardTitle class="text-sm font-medium text-blue-900 dark:text-blue-100">Kesehatan Keuangan</CardTitle>
          <div class="h-8 w-8 rounded-full bg-blue-100 dark:bg-blue-950/60 flex items-center justify-center">
            <HeartPulse class="h-4 w-4 text-blue-600 dark:text-blue-400 group-hover:scale-110 transition-transform" />
          </div>
        </CardHeader>
        <CardContent>
          <div v-if="healthStore.isLoading" class="h-8 w-24 bg-gray-200 dark:bg-gray-700 animate-pulse rounded"></div>
          <div v-else-if="healthStore.data" class="flex flex-col">
              <div class="flex items-baseline gap-2">
                 <div class="text-2xl font-bold" :class="{
                    'text-emerald-600': healthStore.data.overall_status === 'Sehat',
                    'text-amber-500': healthStore.data.overall_status === 'Waspada',
                    'text-red-500': healthStore.data.overall_status === 'Bahaya'
                 }">
                    {{ healthStore.data.overall_status }}
                 </div>
                 <span class="text-xs text-muted-foreground">Skor {{ healthStore.data.overall_score }}</span>
              </div>
             <p class="text-xs text-muted-foreground mt-1 truncate">{{ healthStore.data.ratios.filter((r: any) => r.status !== 'Sehat').length === 0 ? 'Semua indikator aman.' : `${healthStore.data.ratios.filter((r: any) => r.status !== 'Sehat').length} indikator perlu perhatian` }}</p>
          </div>
          <div v-else class="text-sm text-muted-foreground">Belum ada analisis.</div>
        </CardContent>
      </Card>
    </div>

    <div class="grid gap-4 grid-cols-1 md:grid-cols-2 lg:grid-cols-7">
      
      <Card class="col-span-1 md:col-span-2 lg:col-span-4 bg-card shadow-sm rounded-3xl flex flex-col">
        <CardHeader>
          <CardTitle class="flex items-center gap-2">
              <TrendingUp class="h-4 w-4 text-primary" /> Tren Keuangan
          </CardTitle>
          <CardDescription>Perbandingan pemasukan dan pengeluaran 6 bulan terakhir.</CardDescription>
        </CardHeader>
        <CardContent class="pl-0 flex-1 min-h-[300px] relative">
            <div v-if="isMounted" class="absolute inset-0 w-full h-full">
                <apexchart type="area" height="100%" width="100%" :options="chartOptionsArea" :series="chartSeriesArea" />
            </div>
        </CardContent>
      </Card>

      <Card class="col-span-1 md:col-span-2 lg:col-span-3 bg-card shadow-sm rounded-3xl">
        <CardHeader>
            <CardTitle>Distribusi Pengeluaran</CardTitle>
            <CardDescription>Berdasarkan kategori bulan ini.</CardDescription>
        </CardHeader>
        <CardContent class="flex items-center justify-center flex-col">
             <div v-if="chartSeriesDonut.length === 0" class="text-center py-10 text-muted-foreground text-sm">Belum ada data pengeluaran.</div>
             <apexchart v-else type="donut" width="100%" :options="chartOptionsDonut" :series="chartSeriesDonut" :class="{ 'privacy-blur-charts': authStore.isPrivacyMode }" />

             <div v-if="budgetStatus.length > 0" class="w-full mt-6 space-y-3">
                 <p class="text-xs font-bold uppercase text-muted-foreground tracking-widest">Status Anggaran</p>
                 <div v-for="item in budgetStatus.slice(0, 3)" :key="item.name" class="flex items-center justify-between text-sm">
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
             <Button variant="outline" class="w-full mt-4 text-xs h-9 rounded-xl" @click="$router.push('/report')">
                Lihat Semua Laporan
             </Button>
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
                             <div class="flex items-center gap-3 flex-1 min-w-0 mr-2">
                                 <div :class="['h-10 w-10 shrink-0 rounded-xl flex items-center justify-center text-lg shadow-sm transition-transform group-hover:scale-105', 
                                     t.type === 'expense' ? 'bg-red-50 text-red-600 dark:bg-red-950/40 dark:text-red-400' : 
                                     t.type === 'income' ? 'bg-emerald-50 text-emerald-600 dark:bg-emerald-950/40 dark:text-emerald-400' : 
                                     'bg-blue-50 text-blue-600 dark:bg-blue-950/40 dark:text-blue-400']">
                                     <span v-if="getEmoji(t.category?.icon)" class="text-lg leading-none">{{ getEmoji(t.category?.icon) }}</span>
                                     <component v-else :is="getIconComponent(t.category?.icon, 'Circle')" class="h-5 w-5" />
                                </div>
                                <div class="min-w-0">
                                    <p class="font-bold text-sm truncate pb-0.5">{{ t.description || 'Tanpa Keterangan' }}</p>
                                    <p class="text-[10px] text-muted-foreground font-medium flex items-center gap-1 truncate">
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
                             <div class="text-right shrink-0">
                                 <span :class="['block font-bold text-sm', 
                                    t.type === 'income' ? 'text-emerald-600 dark:text-emerald-400' : 
                                    t.type === 'expense' ? 'text-red-600 dark:text-red-400' : 
                                    'text-blue-600 dark:text-blue-400', { 'privacy-blur': authStore.isPrivacyMode }]">
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
                <div class="grid grid-cols-1 gap-4">
                     <div v-for="w in data.wallets.slice(0, 3)" :key="w.id" :class="['p-4 rounded-2xl border-none shadow-md flex items-center justify-between w-full min-h-[96px] transition-all hover:scale-[1.01] text-white', getWalletColorClass(w)]">
                         <div class="flex items-center gap-3 flex-1 min-w-0 mr-3">
                             <div class="h-11 w-11 shrink-0 rounded-2xl bg-white/95 dark:bg-white/90 backdrop-blur-sm p-1.5 flex items-center justify-center text-xl shadow-xs border border-white/30">
                                  <img v-if="getBankLogo(w.icon)" :src="getBankLogo(w.icon)" class="w-full h-full object-contain filter drop-shadow-xs" :alt="w.name" />
                                  <span v-else-if="getEmoji(w.icon)" class="text-lg leading-none">{{ getEmoji(w.icon) }}</span>
                                  <component v-else :is="getIconComponent(w.icon, 'Wallet')" class="h-5 w-5 text-slate-800" />
                             </div>
                             <div class="min-w-0">
                                 <p class="text-[10px] font-bold opacity-75 uppercase tracking-wide truncate">{{ w.type }}</p>
                                 <p class="font-bold text-sm truncate leading-tight">{{ w.name }}</p>
                             </div>
                         </div>
                         <div class="text-right shrink-0">
                             <p class="text-[10px] font-medium opacity-75 uppercase tracking-widest whitespace-nowrap">Saldo Tersedia</p>
                             <p class="font-bold text-base whitespace-nowrap drop-shadow-xs" :class="{ 'privacy-blur': authStore.isPrivacyMode }">{{ formatCurrency(w.available_balance ?? w.balance) }}</p>
                             <div class="border-t border-current/15 mt-1 pt-1 opacity-85">
                               <p class="text-[10px] font-medium opacity-75 uppercase tracking-widest whitespace-nowrap">Total Saldo</p>
                               <p class="font-bold text-xs whitespace-nowrap" :class="{ 'privacy-blur': authStore.isPrivacyMode }">{{ formatCurrency(w.balance) }}</p>
                             </div>
                         </div>
                     </div>
                </div>
                 <Button variant="outline" class="w-full mt-4 text-xs h-9 rounded-xl" @click="$router.push('/wallet')">
                    Lihat Semua Dompet
                 </Button>
            </CardContent>
        </Card>
    </div>
  </div>
</template>