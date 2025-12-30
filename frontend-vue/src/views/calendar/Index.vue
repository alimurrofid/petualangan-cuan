<script setup lang="ts">
import { ref, computed, onMounted } from "vue";
import { format, startOfMonth, endOfMonth, eachDayOfInterval, addMonths, isSameMonth, isSameDay, startOfWeek, endOfWeek, parseISO, isToday } from "date-fns";
import { id } from "date-fns/locale";
import { ChevronLeft, ChevronRight, Calendar as CalendarIcon } from "lucide-vue-next";
import { useTransactionStore } from "@/stores/transaction";
import { useCategoryStore } from "@/stores/category";
import { Card, CardContent, CardHeader } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Dialog, DialogContent, DialogHeader, DialogTitle } from "@/components/ui/dialog";
import { getEmoji, getIconComponent } from "@/lib/icons";

const transactionStore = useTransactionStore();
const categoryStore = useCategoryStore();
const currentMonth = ref(new Date());
const selectedDate = ref<Date | null>(null);
const isDialogOpen = ref(false);

const days = computed(() => {
  const start = startOfWeek(startOfMonth(currentMonth.value), { weekStartsOn: 1 });
  const end = endOfWeek(endOfMonth(currentMonth.value), { weekStartsOn: 1 });
  return eachDayOfInterval({ start, end });
});

const weekDays = ['Sen', 'Sel', 'Rab', 'Kam', 'Jum', 'Sab', 'Min'];

const getTransactionsForDay = (date: Date) => {
    return transactionStore.transactions.filter(t => isSameDay(parseISO(t.date), date));
};

const navigateMonth = (amount: number) => {
    currentMonth.value = addMonths(currentMonth.value, amount);
};

const goToToday = () => {
    currentMonth.value = new Date();
};

const onDayClick = (date: Date) => {
    selectedDate.value = date;
    isDialogOpen.value = true;
};

// Fetch data when month changes
// watch(currentMonth, fetchData); // No longer needed if we rely on global transactions

onMounted(async () => {
    // Ensure we have transactions loaded
    if (transactionStore.transactions.length === 0) {
        await transactionStore.fetchTransactions();
    }
    // and categories
    if (categoryStore.categories.length === 0) {
        await categoryStore.fetchCategories();
    }
});

const selectedDayTransactions = computed(() => {
    if (!selectedDate.value) return [];
    return transactionStore.transactions.filter(t => isSameDay(parseISO(t.date), selectedDate.value!));
});

const formatCurrency = (value: number) => {
  return new Intl.NumberFormat("id-ID", { style: "currency", currency: "IDR", minimumFractionDigits: 0 }).format(value);
};
</script>

<template>
  <div class="p-6 space-y-6 text-foreground min-h-screen bg-background">
      <div class="flex flex-col gap-2">
       <h2 class="text-3xl font-bold tracking-tight">Kalender Transaksi</h2>
       <p class="text-sm text-muted-foreground">Ringkasan aktivitas keuangan bulanan Anda.</p>
    </div>

    <!-- Calendar Section -->
    <Card class="bg-card border-border shadow-sm rounded-3xl overflow-hidden flex flex-col">
        <CardHeader class="p-6 border-b border-border/10 bg-gradient-to-r from-emerald-600 to-teal-500 text-primary-foreground">
                <div class="flex items-center justify-between gap-4">
                    <div class="flex items-center gap-2 w-full md:w-auto">
                        <Button variant="ghost" size="icon" @click="navigateMonth(-1)" class="h-9 w-9 rounded-xl hover:bg-white/20 text-white hover:text-white transition-colors bg-white/30">
                            <ChevronLeft class="h-6 w-6" />
                        </Button>
                        <h3 class="font-bold text-xl capitalize px-2 min-w-[150px] text-center tracking-tight text-white drop-shadow-sm select-none bg-white/30 rounded-xl py-1 px-3">
                            {{ format(currentMonth, 'MMMM yyyy', { locale: id }) }}
                        </h3>
                        <Button variant="ghost" size="icon" @click="navigateMonth(1)" class="h-9 w-9 rounded-xl hover:bg-white/20 text-white hover:text-white transition-colors bg-white/30">
                            <ChevronRight class="h-6 w-6" />
                        </Button>
                    </div>
                    <Button variant="outline" size="sm" @click="goToToday" class="hidden md:flex rounded-xl bg-white/30 border-white/40 dark:bg-white/30 dark:border-white/40 text-white hover:bg-white hover:text-teal-600 transition-colors font-semibold h-9 text-xs">Hari Ini</Button>
            </div>
        </CardHeader>
        <CardContent class="p-0">
            <!-- Week Header -->
            <div class="grid grid-cols-7 bg-emerald-600/5 border-b border-border">
                    <div v-for="day in weekDays" :key="day" class="py-3 text-center text-xs font-bold uppercase tracking-widest text-emerald-700 dark:text-emerald-400">
                    {{ day }}
                </div>
            </div>
            
            <!-- Days Grid -->
            <div class="grid grid-cols-7 auto-rows-[1fr]">
                    <div 
                    v-for="date in days" 
                    :key="date.toString()"
                    class="relative border-b border-r border-border/50 p-2 transition-all flex flex-col justify-start gap-1 min-h-[120px] group"
                    :class="[
                        !isSameMonth(date, currentMonth) ? 'bg-muted/10 text-muted-foreground/40' : 'bg-background hover:bg-muted/10 cursor-pointer',
                        isToday(date) ? 'bg-teal-200/75 dark:bg-teal-200/75' : ''
                    ]"
                    @click="onDayClick(date)"
                    >
                    <div class="flex justify-between items-start mb-1">
                        <span 
                            :class="[
                                'text-sm font-semibold h-6 w-6 flex items-center justify-center rounded-full transition-shadow',
                                isToday(date) 
                                    ? 'bg-gradient-to-br from-emerald-500 to-teal-500 text-white shadow-md' 
                                    : 'text-foreground/70 group-hover:text-emerald-600 group-hover:bg-emerald-50 dark:group-hover:bg-emerald-950/30'
                            ]"
                        >
                            {{ format(date, 'd') }}
                        </span>
                        
                        <!-- Transaction Count -->
                         <span v-if="getTransactionsForDay(date).length > 0" class="text-[10px] font-bold text-muted-foreground opacity-70">
                            {{ getTransactionsForDay(date).length }} Trx
                         </span>
                    </div>

                    <!-- Transaction Pills (Desktop & Mobile) -->
                    <div class="space-y-1 w-full">
                         <template v-for="(t, index) in getTransactionsForDay(date)" :key="t.id">
                            <div 
                                v-if="index < 3"
                                :class="[
                                    'px-1.5 py-0.5 rounded-md text-[9px] font-medium flex items-center justify-between shadow-sm border border-transparent w-full',
                                    (t.type === 'expense' || t.type === 'transfer_out')
                                        ? 'bg-red-50 text-red-700 dark:bg-red-950/30 dark:text-red-300' 
                                        : 'bg-emerald-50 text-emerald-700 dark:bg-emerald-950/30 dark:text-emerald-300'
                                ]"
                            >
                                <span class="truncate max-w-[60%]">{{ t.description || t.category?.name }}</span>
                                <span class="font-bold shrink-0">
                                    {{ t.amount >= 1000000 ? (t.amount/1000000).toFixed(1) + 'm' : (t.amount/1000).toFixed(0) + 'k' }}
                                </span>
                            </div>
                         </template>
                         
                         <div v-if="getTransactionsForDay(date).length > 3" class="text-[9px] text-center font-bold text-muted-foreground bg-muted/30 rounded-full py-0.5">
                            +{{ getTransactionsForDay(date).length - 3 }} 
                         </div>
                    </div>
                </div>
            </div>
        </CardContent>
    </Card>

    <!-- Details Dialog -->
    <Dialog v-model:open="isDialogOpen">
        <DialogContent class="max-w-md bg-card border-border sm:rounded-3xl">
            <DialogHeader class="border-b border-border pb-4 mb-2">
                <DialogTitle class="flex items-center gap-3">
                    <div class="h-10 w-10 rounded-full bg-primary/10 flex items-center justify-center text-primary">
                        <CalendarIcon class="h-5 w-5" />
                    </div>
                    <div>
                        <p class="text-sm font-medium text-muted-foreground uppercase tracking-widest">Transaksi Tanggal</p>
                        <p class="text-xl font-bold">{{ selectedDate ? format(selectedDate, 'd MMMM yyyy', { locale: id }) : '' }}</p>
                    </div>
                </DialogTitle>
            </DialogHeader>

            <div class="space-y-4 max-h-[60vh] overflow-y-auto pr-2 custom-scrollbar">
                 <div v-if="selectedDayTransactions.length === 0" class="text-center py-8 text-muted-foreground">
                    <div class="h-12 w-12 bg-muted rounded-full flex items-center justify-center mx-auto mb-3 opacity-50">
                        <CalendarIcon class="h-6 w-6" />
                    </div>
                    <p class="text-sm">Tidak ada transaksi tercatat.</p>
                 </div>

                 <div v-else class="space-y-3">
                     <div v-for="t in selectedDayTransactions" :key="t.id" class="flex items-center justify-between p-3 rounded-2xl bg-muted/20 hover:bg-muted/50 border border-transparent hover:border-border transition-all">
                         <div class="flex items-center gap-3">
                             <div :class="['h-10 w-10 rounded-xl flex items-center justify-center text-lg shadow-sm', 
                                  t.type === 'expense' ? 'bg-red-50 text-red-500' : 
                                  t.type === 'income' ? 'bg-emerald-50 text-emerald-600' : 'bg-blue-50 text-blue-600']">
                                 <component v-if="getIconComponent(t.category?.icon)" :is="getIconComponent(t.category?.icon)" class="h-5 w-5" />
                                 <span v-else-if="getEmoji(t.category?.icon)">{{ getEmoji(t.category?.icon) }}</span>
                                 <!-- Fallback if no emoji/icon logic, use first letter -->
                                 <span v-else class="font-bold">{{ t.category?.name?.[0] || '?' }}</span>
                             </div>
                             <div class="overflow-hidden">
                                 <p class="font-bold text-sm truncate max-w-[150px]">{{ t.description || 'No Description' }}</p>
                                 <p class="text-[10px] text-muted-foreground font-medium">{{ t.category.name }} • {{ t.wallet.name }}</p>
                             </div>
                         </div>
                         <div class="flex flex-col items-end">
                             <span :class="['font-bold text-sm whitespace-nowrap', 
                                  t.type === 'expense' ? 'text-red-500' : 
                                  t.type === 'income' ? 'text-emerald-600' : 'text-blue-600']">
                                 {{ t.type === 'income' || t.type === 'transfer_in' ? '+' : '-' }} {{ formatCurrency(t.amount) }}
                             </span>
                             <span class="text-xs text-muted-foreground">{{ format(parseISO(t.date), 'HH:mm') }}</span>
                         </div>
                     </div>
                 </div>
            </div>
             
            <div v-if="selectedDayTransactions.length > 0" class="pt-4 border-t border-border mt-2 bg-muted/40 -mx-6 -mb-6 p-6 flex justify-between items-center">
                 <div class="text-xs font-bold text-muted-foreground uppercase tracking-widest">Saldo Harian</div>
                  <div class="text-lg font-bold text-foreground">
                    {{ 
                        formatCurrency(
                            selectedDayTransactions.reduce((acc, curr) => {
                                if (curr.type === 'income' || curr.type === 'transfer_in') return acc + curr.amount;
                                if (curr.type === 'expense' || curr.type === 'transfer_out') return acc - curr.amount;
                                return acc;
                            }, 0)
                        ) 
                    }}
                  </div>
             </div>
        </DialogContent>
    </Dialog>

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
