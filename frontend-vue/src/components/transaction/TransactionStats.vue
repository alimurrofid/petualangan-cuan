<script setup lang="ts">
import { computed } from "vue";
import { ArrowUpCircle, ArrowDownCircle, Wallet } from "lucide-vue-next";
import { formatCurrency } from "@/lib/utils";
import type { TransactionSummary } from "@/stores/transaction";

const props = defineProps<{
  summaryData: TransactionSummary[]
}>();

const totalIncome = computed(() => {
    return props.summaryData.reduce((sum, item) => sum + item.income, 0);
});

const totalExpense = computed(() => {
    return props.summaryData.reduce((sum, item) => sum + item.expense, 0);
});

const balance = computed(() => totalIncome.value - totalExpense.value);
</script>

<template>
    <div class="grid grid-cols-1 md:grid-cols-3 gap-6">
        <div class="relative overflow-hidden rounded-3xl p-6 bg-card border border-border shadow-sm">
             <div class="flex justify-between items-start">
                <div>
                    <p class="text-sm font-semibold text-muted-foreground uppercase tracking-widest">Saldo Bersih</p>
                    <p class="text-2xl font-bold text-foreground mt-1">{{ formatCurrency(balance) }}</p>
                </div>
                 <div class="h-10 w-10 rounded-full bg-muted flex items-center justify-center text-muted-foreground">
                    <Wallet class="h-6 w-6" />
                </div>
            </div>
        </div>

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
    </div>
</template>
