<script setup lang="ts">
import { ref, computed, onMounted } from "vue";
import { 
  format, 
  startOfMonth, 
  endOfMonth, 
  startOfWeek, 
  endOfWeek, 
  eachDayOfInterval, 
  addMonths, 
  subMonths, 
  isSameMonth, 
  isSameDay, 
  parseISO,
  isToday
} from "date-fns";
import { id as localeId } from "date-fns/locale";

import { Button } from "@/components/ui/button";
import { Dialog, DialogContent, DialogHeader, DialogTitle } from "@/components/ui/dialog";

import { ChevronLeft, ChevronRight, Calendar as CalendarIcon } from "lucide-vue-next";

// Interfaces
interface Transaction {
  id: number;
  title: string;
  amount: number;
  date: string;
  type: "income" | "expense";
  categoryId: number;
}

interface CategoryItem {
  id: number;
  name: string;
  icon: string;
  isEmoji: boolean;
  type: "income" | "expense";
}

// State
const currentDate = ref(new Date());
const transactions = ref<Transaction[]>([]);
const categories = ref<CategoryItem[]>([]);

// Detail Dialog
const selectedDayTransactions = ref<Transaction[]>([]);
const selectedDayDate = ref<Date | null>(null);
const isDialogOpen = ref(false);

// Load Data
const loadData = () => {
    const savedCategories = localStorage.getItem("mock_categories");
    if (savedCategories) categories.value = JSON.parse(savedCategories);
    else {
         // Fallback default
         categories.value = [
            { id: 1, name: "Makanan", icon: "Utensils", isEmoji: false, type: "expense" },
            { id: 2, name: "Gaji", icon: "💰", isEmoji: true, type: "income" },
        ];
    }

    const savedTransactions = localStorage.getItem("mock_transactions");
    if (savedTransactions) {
        transactions.value = JSON.parse(savedTransactions);
    }
};

onMounted(loadData);

// Calendar Logic
const daysOfWeek = ["Minggu", "Senin", "Selasa", "Rabu", "Kamis", "Jumat", "Sabtu"];

const calendarDays = computed(() => {
    const startMonth = startOfMonth(currentDate.value);
    const endMonth = endOfMonth(currentDate.value);
    
    const startGrid = startOfWeek(startMonth); // Default starts on Sunday
    const endGrid = endOfWeek(endMonth);

    return eachDayOfInterval({ start: startGrid, end: endGrid });
});

const currentMonthName = computed(() => {
    return format(currentDate.value, "MMMM yyyy", { locale: localeId });
});

const nextMonth = () => {
    currentDate.value = addMonths(currentDate.value, 1);
};

const prevMonth = () => {
    currentDate.value = subMonths(currentDate.value, 1);
};

const goToToday = () => {
    currentDate.value = new Date();
};

const getTransactionsForDay = (date: Date) => {
    return transactions.value.filter(t => isSameDay(parseISO(t.date), date));
};

// Formatting
const formatCurrency = (value: number) => {
  return new Intl.NumberFormat("id-ID", { style: "currency", currency: "IDR", minimumFractionDigits: 0 }).format(value);
};

const getCategory = (id: number) => categories.value.find((c) => c.id === id);

const onDayClick = (day: Date, txs: Transaction[]) => {
    if (txs.length > 0) {
        selectedDayDate.value = day;
        selectedDayTransactions.value = txs;
        isDialogOpen.value = true;
    }
};

</script>

<template>
  <div class="p-6 space-y-6 text-foreground min-h-screen bg-background">
    
    <!-- Header -->
    <div class="flex flex-col md:flex-row items-center justify-between gap-4">
        <div>
            <h2 class="text-3xl font-bold tracking-tight">Kalender</h2>
            <p class="text-muted-foreground mt-1">Lihat riwayat transaksi Anda dalam tampilan bulanan.</p>
        </div>
        
        <div class="flex items-center gap-2 bg-card border border-border p-1.5 rounded-2xl shadow-sm w-full md:w-auto justify-between md:justify-start">
            <Button variant="ghost" size="icon" @click="prevMonth" class="rounded-xl hover:bg-muted">
                <ChevronLeft class="h-5 w-5" />
            </Button>
            <div class="flex-1 text-center md:px-4 font-bold text-lg select-none min-w-[140px]">
                {{ currentMonthName }}
            </div>
             <Button variant="ghost" size="icon" @click="nextMonth" class="rounded-xl hover:bg-muted">
                <ChevronRight class="h-5 w-5" />
            </Button>
            <div class="w-[1px] h-6 bg-border mx-2 hidden md:block"></div>
            <Button variant="outline" size="sm" @click="goToToday" class="rounded-xl font-semibold hidden md:flex">Today</Button>
        </div>
    </div>

    <!-- Calendar Grid -->
    <div class="bg-card border border-border rounded-3xl shadow-sm overflow-hidden flex flex-col">
        <!-- Week Header -->
        <div class="grid grid-cols-7 bg-muted/40 border-b border-border">
            <div v-for="day in daysOfWeek" :key="day" class="py-4 text-center text-xs font-bold uppercase tracking-widest text-muted-foreground">
                {{ day }}
            </div>
        </div>

        <!-- Days -->
        <div class="grid grid-cols-7 auto-rows-[1fr]">
            <div 
                v-for="day in calendarDays" 
                :key="day.toString()"
                @click="onDayClick(day, getTransactionsForDay(day))"
                :class="[
                    'min-h-[140px] p-2 border-b border-r border-border/50 relative transition-colors',
                    !isSameMonth(day, currentDate) ? 'bg-muted/10 text-muted-foreground/40' : 'bg-background hover:bg-muted/20 cursor-pointer',
                    isToday(day) ? 'bg-primary/5' : ''
                ]"
            >
                <!-- Day Number -->
                <div class="flex justify-between items-start mb-2">
                    <span 
                        :class="[
                            'text-sm font-semibold h-7 w-7 flex items-center justify-center rounded-full',
                            isToday(day) ? 'bg-primary text-primary-foreground shadow-md' : 'text-foreground/70'
                        ]"
                    >
                        {{ format(day, 'd') }}
                    </span>
                    
                     <!-- Daily Summary (Optional) -->
                     <span v-if="getTransactionsForDay(day).length > 0" class="text-[10px] font-bold text-muted-foreground opacity-70">
                        {{ getTransactionsForDay(day).length }} Trx
                     </span>
                </div>

                <!-- Transaction List (Preview) -->
                <div class="space-y-1.5">
                    <template v-for="(t, index) in getTransactionsForDay(day)" :key="t.id">
                        <div 
                            v-if="index < 3"
                            :class="[
                                'px-2 py-1 rounded-md text-[10px] font-medium truncate flex items-center justify-between gap-1 shadow-sm border border-transparent',
                                t.type === 'expense' 
                                    ? 'bg-red-50 text-red-700 dark:bg-red-950/30 dark:text-red-300' 
                                    : 'bg-emerald-50 text-emerald-700 dark:bg-emerald-950/30 dark:text-emerald-300'
                            ]"
                        >
                            <span class="truncate">{{ t.title }}</span>
                            <span v-if="t.amount > 999999" class="font-bold shrink-0 text-[9px] opacity-80">{{ (t.amount/1000000).toFixed(1) }}M</span>
                            <span v-else class="font-bold shrink-0 text-[9px] opacity-80">{{ (t.amount/1000).toFixed(0) }}k</span>
                        </div>
                    </template>
                    
                    <div v-if="getTransactionsForDay(day).length > 3" class="text-[10px] text-center font-bold text-muted-foreground mt-1 bg-muted/50 rounded-md py-0.5">
                        + {{ getTransactionsForDay(day).length - 3 }} Lainnya
                    </div>
                </div>
            </div>
        </div>
    </div>


    <!-- Details Dialog -->
    <Dialog v-model:open="isDialogOpen">
        <DialogContent class="max-w-md bg-card border-border">
            <DialogHeader class="border-b border-border pb-4 mb-4">
                <DialogTitle class="flex items-center gap-2">
                    <CalendarIcon class="h-5 w-5 text-primary" />
                    <span>{{ selectedDayDate ? format(selectedDayDate, 'd MMMM yyyy', { locale: localeId }) : '' }}</span>
                </DialogTitle>
            </DialogHeader>

            <div class="space-y-3 max-h-[60vh] overflow-y-auto pr-2">
                 <div v-for="t in selectedDayTransactions" :key="t.id" class="flex items-center justify-between p-3 rounded-2xl border border-border hover:bg-muted/30">
                     <div class="flex items-center gap-3">
                         <div class="h-10 w-10 text-xl flex items-center justify-center bg-muted rounded-xl">
                            {{ getCategory(t.categoryId)?.isEmoji ? getCategory(t.categoryId)?.icon : '💳' }}
                         </div>
                         <div>
                             <p class="font-bold text-sm">{{ t.title }}</p>
                             <p class="text-xs text-muted-foreground">{{ getCategory(t.categoryId)?.name }}</p>
                         </div>
                     </div>
                     <span :class="['font-bold text-sm', t.type === 'expense' ? 'text-red-500' : 'text-emerald-500']">
                         {{ t.type === 'income' ? '+' : '-' }} {{ formatCurrency(t.amount) }}
                     </span>
                 </div>
            </div>
             
             <div class="pt-4 border-t border-border mt-2 bg-muted/30 p-4 rounded-xl flex justify-between items-center">
                 <div class="text-xs font-semibold text-muted-foreground uppercase tracking-widest">Total Hari Ini</div>
                  <div class="text-lg font-bold">
                    {{ 
                        formatCurrency(
                            selectedDayTransactions.reduce((acc, curr) => acc + (curr.type === 'income' ? curr.amount : -curr.amount), 0)
                        ) 
                    }}
                  </div>
             </div>
        </DialogContent>
    </Dialog>

  </div>
</template>
