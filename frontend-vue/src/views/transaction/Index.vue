<script setup lang="ts">
import { ref, watch, onMounted, computed } from "vue";
import { format, startOfMonth, endOfMonth, startOfWeek, endOfWeek, startOfDay, endOfDay, addMonths, addWeeks, addDays } from "date-fns";
import { Download, Search, Loader2 } from "lucide-vue-next";
import { id } from "date-fns/locale";

import { useTransactionStore } from "@/stores/transaction";
import { useWalletStore } from "@/stores/wallet";
import { useCategoryStore } from "@/stores/category";

import { Card, CardHeader, CardTitle, CardContent } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Button } from "@/components/ui/button";
import TransactionStats from "@/components/transaction/TransactionStats.vue";
import TransactionChart from "@/components/transaction/TransactionChart.vue";
import TransactionFilter from "@/components/transaction/TransactionFilter.vue";
import TransactionList from "@/components/transaction/TransactionList.vue";
import ManualTransactionDialog from "@/components/ManualTransactionDialog.vue";

const transactionStore = useTransactionStore();
const walletStore = useWalletStore();
const categoryStore = useCategoryStore();

type PeriodType = 'monthly' | 'weekly' | 'daily' | 'custom';
const periodType = ref<PeriodType>('monthly');
const selectedDate = ref(new Date());
const customDateRange = ref({
  start: new Date(),
  end: new Date()
});

const filterWallet = ref<string[]>([]);
const filterCategory = ref<string[]>([]);
const searchQuery = ref("");

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
      return {
        start: startOfDay(customDateRange.value.start),
        end: endOfDay(customDateRange.value.end)
      };
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

const updateCustomDateRange = (range: { start: Date, end: Date }) => {
  customDateRange.value = range;
};

const fetchData = async (page = 1) => {
  const { start, end } = dateRange.value;
  const startDateStr = format(start, 'yyyy-MM-dd HH:mm:ss');
  const endDateStr = format(end, 'yyyy-MM-dd HH:mm:ss');
  transactionStore.setFilters({
    page,
    limit: 10,
    start_date: startDateStr,
    end_date: endDateStr,
    wallet_id: filterWallet.value,
    category_id: filterCategory.value,
    search: searchQuery.value,
  });

  await transactionStore.refreshData();
};

watch([periodType, selectedDate, customDateRange, filterWallet, filterCategory, searchQuery], () => {
  fetchData(1);
}, { deep: true });

onMounted(async () => {
  await Promise.all([
    walletStore.fetchWallets(),
    categoryStore.fetchCategories(),
    fetchData(1)
  ]);
});

const onPageChange = (page: number) => {
  fetchData(page);
};

const showDialog = ref(false);
const transactionToEdit = ref<any>(null);

const handleEdit = (t: any) => {
  transactionToEdit.value = t;
  showDialog.value = true;
};

const handleSave = () => {
  showDialog.value = false;
  transactionToEdit.value = null;
};


const handleExport = async () => {
  try {
    const blob = await transactionStore.exportTransactions();
    const url = window.URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = `transactions_petualangancuan_${format(new Date(), 'yyyy-MM-dd')}.xlsx`;
    document.body.appendChild(a);
    a.click();
    window.URL.revokeObjectURL(url);
    document.body.removeChild(a);
  } catch (e) {
    console.error("Export failed", e);
  }
};

const localSearch = ref("");
let debounceTimer: any = null;

watch(localSearch, (val) => {
  clearTimeout(debounceTimer);
  debounceTimer = setTimeout(() => {
    searchQuery.value = val;
  }, 300);
});
</script>

<template>
  <div class="flex-1 space-y-6 pt-2">
    <!-- Initial Loading State -->
    <div v-if="transactionStore.isLoading && !transactionStore.transactions.length && !searchQuery" class="flex-1 flex items-center justify-center min-h-[400px]">
        <div class="flex flex-col items-center gap-2">
           <p class="text-muted-foreground animate-pulse">Memuat data transaksi...</p>
        </div>
    </div>
    
    <div class="flex-1 space-y-6 pt-2 text-foreground" v-else>
      <div class="flex flex-col gap-2">
        <h2 class="text-3xl font-bold tracking-tight">Riwayat Transaksi</h2>
        <p class="text-sm text-muted-foreground">Analisis dan pantau arus kas Anda.</p>
      </div>

      <!-- Stats Grid -->
      <TransactionStats :summaryData="transactionStore.calendarData" />

      <!-- Main Content -->
      <div class="space-y-6 relative">
        <!-- Toolbar -->
        <TransactionFilter 
            v-model:periodType="periodType" 
            v-model:walletIds="filterWallet"
            v-model:categoryIds="filterCategory" 
            :startDate="customDateRange.start"
            :endDate="customDateRange.end" 
            :formattedDateRange="formattedDateRange" 
            @navigateDate="navigateDate"
            @update:dateRange="updateCustomDateRange" 
            @export="handleExport" 
        />

        <div class="grid lg:grid-cols-3 gap-6 md:h-[600px] overflow-hidden">
          <!-- Chart Section -->
          <Card class="lg:col-span-2 bg-card border-border shadow-sm flex flex-col rounded-3xl overflow-hidden h-full">
            <CardHeader class="pb-2 border-b border-border/50">
              <CardTitle class="text-base font-bold flex items-center gap-2">
                Grafik Pertumbuhan
              </CardTitle>
            </CardHeader>
            <CardContent class="flex-1 p-4 relative min-h-[250px] md:min-h-[300px]">
              <TransactionChart class="h-full w-full" :summaryData="transactionStore.calendarData" :periodType="periodType" />
            </CardContent>
          </Card>

          <!-- Transaction List -->
          <Card class="bg-card border-border shadow-sm flex flex-col rounded-3xl overflow-hidden h-full">
            <CardHeader class="pb-3 border-b border-border/50 space-y-3">
              <div class="flex items-center justify-between w-full">
                <h3 class="font-bold text-sm">Daftar Transaksi</h3>
                <Button
                  @click="handleExport"
                  title="Ekspor Excel"
                  variant="outline"
                  class="h-8 px-3 rounded-lg border-border shadow-sm hover:bg-muted/50 flex items-center gap-2 text-xs"
                >
                  <Download class="h-3.5 w-3.5 text-muted-foreground" />
                  <span>Ekspor</span>
                </Button>
              </div>
              
               <!-- Search moved here -->
               <div class="relative w-full">
                  <Search class="absolute left-3 top-2.5 h-4 w-4 text-muted-foreground" />
                  <Input 
                      v-model="localSearch" 
                      placeholder="Cari transaksi..." 
                      class="h-9 pl-9 rounded-xl bg-muted/30 border-muted-foreground/20 focus:bg-background transition-all text-xs w-full" 
                  />
                  <div v-if="transactionStore.isLoading" class="absolute right-3 top-2.5">
                      <Loader2 class="h-4 w-4 animate-spin text-emerald-500" />
                  </div>
              </div>
            </CardHeader>
            <CardContent class="overflow-y-auto p-0 flex-1 custom-scrollbar">
              <TransactionList @page-change="onPageChange" @edit="handleEdit" />
            </CardContent>
          </Card>
        </div>
      </div>


    <!-- Edit Dialog -->
    <ManualTransactionDialog :open="showDialog"
      @update:open="(val) => { showDialog = val; if (!val) transactionToEdit = null; }"
      :transactionToEdit="transactionToEdit" @save="handleSave" />
  </div>
  </div>
</template>

<style scoped>
</style>
