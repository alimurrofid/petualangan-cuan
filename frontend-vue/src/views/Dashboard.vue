<script setup lang="ts">
import { ref, computed, onMounted } from 'vue';
import { 
  Card, 
  CardContent, 
  CardHeader, 
  CardTitle,
  CardDescription
} from '@/components/ui/card';
import { 
  ArrowUp, 
  ArrowDown, 
  Wallet,
  TrendingUp,
  Lightbulb
} from 'lucide-vue-next';
import * as LucideIcons from "lucide-vue-next";
import { startOfMonth, endOfMonth, isWithinInterval, parseISO, format, isSameDay, subDays } from 'date-fns';
import { id as localeId } from "date-fns/locale";

interface Transaction {
  id: number;
  title: string;
  amount: number;
  date: string;
  type: "income" | "expense";
  categoryId: number;
  walletId: number;
}

interface WalletItem {
  id: number;
  name: string;
  type: string;
  balance: number;
  cardNumber: string;
  holderName: string;
}

interface CategoryItem {
  id: number;
  name: string;
  icon: string;
  isEmoji: boolean;
  type: "income" | "expense";
  budgetLimit?: number;
}

const transactions = ref<Transaction[]>([]);
const wallets = ref<WalletItem[]>([]);
const categories = ref<CategoryItem[]>([]);

onMounted(() => {
    const savedTransactions = localStorage.getItem("mock_transactions");
    if (savedTransactions) transactions.value = JSON.parse(savedTransactions);

    const savedWallets = localStorage.getItem("mock_wallets");
    if (savedWallets) wallets.value = JSON.parse(savedWallets);
    else wallets.value = [
        { id: 1, name: "BCA Utama", type: "Bank", balance: 15450000, cardNumber: "**** 4521", holderName: "ALIMURROFID" },
        { id: 2, name: "GoPay", type: "E-Wallet", balance: 250000, cardNumber: "0812****99", holderName: "ALIMURROFID" },
        { id: 3, name: "Uang Tunai", type: "Cash", balance: 750000, cardNumber: "-", holderName: "-" },
    ];

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
});

const thisMonthTransactions = computed(() => {
    const now = new Date();
    const start = startOfMonth(now);
    const end = endOfMonth(now);

    return transactions.value.filter(t => {
        return isWithinInterval(parseISO(t.date), { start, end });
    });
});

const totalBalance = computed(() => wallets.value.reduce((acc, w) => acc + w.balance, 0));

const totalIncome = computed(() => thisMonthTransactions.value.filter(t => t.type === 'income').reduce((acc, t) => acc + t.amount, 0));
const totalExpense = computed(() => thisMonthTransactions.value.filter(t => t.type === 'expense').reduce((acc, t) => acc + t.amount, 0));

const recentTransactions = computed(() => {
    return transactions.value
        .sort((a,b) => new Date(b.date).getTime() - new Date(a.date).getTime())
        .slice(0, 5);
});

const groupedRecentTransactions = computed(() => {
    const groups: Record<string, Transaction[]> = {};
    recentTransactions.value.forEach(t => {
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
});

const expenseBreakdown = computed(() => {
    const groups: Record<string, number> = {};
    thisMonthTransactions.value.filter(t => t.type === 'expense').forEach(t => {
        const cat = categories.value.find(c => c.id === t.categoryId);
        const name = cat?.name || 'Lainnya';
        if (!groups[name]) groups[name] = 0;
        groups[name] += t.amount;
    });
    return groups;
});


const budgetStatus = computed(() => {
    const status: { name: string; spent: number; limit: number; percentage: number; isOver: boolean }[] = [];
    
    categories.value.filter(c => c.type === 'expense' && c.budgetLimit && c.budgetLimit > 0).forEach(c => {
        const spent = thisMonthTransactions.value
            .filter(t => t.categoryId === c.id && t.type === 'expense')
            .reduce((acc, t) => acc + t.amount, 0);
        
        status.push({
            name: c.name,
            spent,
            limit: c.budgetLimit || 0,
            percentage: (spent / (c.budgetLimit || 1)) * 100,
            isOver: spent > (c.budgetLimit || 0)
        });
    });

    return status.sort((a,b) => b.percentage - a.percentage);
});

const chartSeriesDonut = computed(() => Object.values(expenseBreakdown.value));
const chartLabelsDonut = computed(() => Object.keys(expenseBreakdown.value));

const chartOptionsDonut = computed(() => ({
    chart: { type: 'donut', fontFamily: 'inherit', foreColor: '#94a3b8' },
    labels: chartLabelsDonut.value,
    plotOptions: {
        pie: {
            donut: {
                size: '65%',
                labels: {
                    show: false
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

const chartOptionsArea = computed(() => ({
  chart: {
    type: 'area',
    toolbar: { show: false },
    zoom: { enabled: false },
    foreColor: '#94a3b8' // Slate-400: Readable on both Dark/Light
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
    categories: ['Jan', 'Feb', 'Mar', 'Apr', 'May', 'Jun', 'Jul'],
    axisBorder: { show: false },
    axisTicks: { show: false }
  },
  yaxis: { show: false },
  grid: { 
     show: true,
     borderColor: '#334155', // Slate-700
     strokeDashArray: 4,
  },
  tooltip: { theme: 'dark' }
}));

const chartSeriesArea = [
    { name: 'Pemasukan', data: [120, 200, 150, 300, 250, 400, 350] }, 
    { name: 'Pengeluaran', data: [80, 100, 120, 180, 150, 200, 180] }
];

const formatCurrency = (value: number) => {
  return new Intl.NumberFormat("id-ID", { style: "currency", currency: "IDR", minimumFractionDigits: 0 }).format(value);
};

const getWalletName = (id: number) => wallets.value.find((w) => w.id === id)?.name || "Unknown Wallet";
const getCategory = (id: number) => categories.value.find((c) => c.id === id);
const getIconComponent = (name: string) => (LucideIcons as any)[name] || LucideIcons.Circle;

</script>

<template>
  <div class="flex-1 space-y-6 pt-2">
    <div class="flex items-center justify-between">
      <div>
          <h2 class="text-3xl font-bold tracking-tight">Dashboard</h2>
          <p class="text-sm text-muted-foreground mt-1">Selamat datang kembali, Alimurrofid!</p>
      </div>
    </div>

    <div class="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
      <Card class="bg-gradient-to-br from-indigo-500 to-purple-600 text-white border-none shadow-md">
        <CardHeader class="flex flex-row items-center justify-between space-y-0 pb-2">
          <CardTitle class="text-sm font-medium text-white/90">Total Saldo</CardTitle>
          <Wallet class="h-4 w-4 text-white/70" />
        </CardHeader>
        <CardContent>
          <div class="text-2xl font-bold">{{ formatCurrency(totalBalance) }}</div>
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
          <div class="text-2xl font-bold text-emerald-600">+{{ formatCurrency(totalIncome) }}</div>
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
          <div class="text-2xl font-bold text-rose-600">-{{ formatCurrency(totalExpense) }}</div>
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
            <CardDescription>Berdasarkan kategori.</CardDescription>
        </CardHeader>
        <CardContent class="flex items-center justify-center">
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
                         {{ item.percentage.toFixed(0) }}%
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
                    <div v-if="recentTransactions.length === 0" class="text-center py-6 text-muted-foreground text-sm">
                        Belum ada transaksi.
                    </div>
                    <div v-for="group in groupedRecentTransactions" :key="group.date" class="space-y-3">
                         <div class="flex items-center gap-2">
                             <span class="text-[10px] font-bold uppercase tracking-widest text-muted-foreground bg-muted px-2 py-1 rounded-md">{{ group.label }}</span>
                             <div class="h-[1px] flex-1 bg-border"></div>
                         </div>

                        <div v-for="t in group.items" :key="t.id" class="group relative flex items-center justify-between p-3 rounded-2xl hover:bg-muted/50 transition-all cursor-default border border-transparent hover:border-border">
                             <div class="flex items-center gap-3">
                                 <div :class="['h-10 w-10 rounded-xl flex items-center justify-center text-lg shadow-sm transition-transform group-hover:scale-105', t.type === 'expense' ? 'bg-red-50 text-red-500' : 'bg-emerald-50 text-emerald-600']">
                                     <template v-if="getCategory(t.categoryId)?.isEmoji">
                                        {{ getCategory(t.categoryId)?.icon }}
                                     </template>
                                     <component v-else :is="getIconComponent(getCategory(t.categoryId)?.icon || 'Circle')" class="h-5 w-5" />
                                </div>
                                <div>
                                    <p class="font-bold text-sm truncate max-w-[120px] pb-1">{{ t.title }}</p>
                                    <p class="text-[10px] text-muted-foreground font-medium flex items-center gap-1">
                                        {{ getWalletName(t.walletId) }} • {{ getCategory(t.categoryId)?.name }}
                                    </p>
                                </div>
                            </div>
                             <div class="text-right">
                                 <span :class="['block font-bold text-sm', t.type === 'income' ? 'text-emerald-600' : 'text-red-500']">
                                    {{ t.type === 'income' ? '+' : '-' }} {{ formatCurrency(t.amount) }}
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
                     <div v-for="w in wallets" :key="w.id" class="p-3 bg-gradient-to-r from-gray-50 to-gray-100 dark:from-zinc-900 dark:to-zinc-800 rounded-xl border border-border/50 flex items-center justify-between w-full">
                         <div>
                             <p class="text-xs font-bold text-muted-foreground uppercase">{{ w.type }}</p>
                             <p class="font-bold text-sm truncate max-w-[120px]">{{ w.name }}</p>
                         </div>
                         <p class="font-bold text-sm text-primary whitespace-nowrap">{{ formatCurrency(w.balance) }}</p>
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