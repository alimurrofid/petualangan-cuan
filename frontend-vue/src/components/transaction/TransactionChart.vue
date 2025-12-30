<script setup lang="ts">
import { computed } from "vue";
import type { TransactionSummary } from "@/stores/transaction";
import { formatCurrency } from "@/lib/utils";
import { format, parseISO, startOfDay, addHours, isSameHour } from "date-fns";
import { id } from "date-fns/locale";

const props = defineProps<{
  summaryData: TransactionSummary[],
  periodType: string
}>();

const processedData = computed(() => {
    // If not daily, or no data, just return as is
    if (props.periodType !== 'daily') {
        return props.summaryData;
    }

    // Generate 24 hours slots
    const firstItem = props.summaryData[0];
    const baseDate = firstItem ? parseISO(firstItem.date) : new Date();
    const dayStart = startOfDay(baseDate);
    const fullDay: TransactionSummary[] = [];

    for (let i = 0; i < 24; i++) {
        const currentHour = addHours(dayStart, i);
        
        // Find if we have data for this hour
        const match = props.summaryData.find(d => {
            const dDate = parseISO(d.date);
            return isSameHour(dDate, currentHour);
        });

        if (match) {
            fullDay.push(match);
        } else {
            fullDay.push({
                date: format(currentHour, "yyyy-MM-dd HH:mm:ss"),
                income: 0,
                expense: 0
            });
        }
    }
    return fullDay;
});

const chartSeries = computed(() => {
  const data = processedData.value;
  const incomeData = data.map(d => d.income);
  const expenseData = data.map(d => d.expense);

  return [
    { name: 'Pemasukan', data: incomeData },
    { name: 'Pengeluaran', data: expenseData }
  ];
});

const chartOptions = computed(() => {
    const categories = processedData.value.map(d => {
        const date = parseISO(d.date);
        if (props.periodType === 'daily') {
            return format(date, "HH:mm", { locale: id });
        }
        return format(date, "d MMM", { locale: id });
    });

    return {
        chart: { type: 'area', height: 300, toolbar: { show: false }, fontFamily: 'inherit', zoom: { enabled: false }, foreColor: '#94a3b8' },
        dataLabels: { enabled: false },
        stroke: { curve: 'smooth', width: 2 },
        fill: { type: 'gradient', gradient: { shadeIntensity: 1, opacityFrom: 0.4, opacityTo: 0.05, stops: [0, 90, 100] } },
        xaxis: { 
            categories: categories, 
            axisBorder: { show: false }, 
            axisTicks: { show: false }, 
            labels: { 
                show: true,
                style: { fontSize: '9px' }, 
                rotate: -45, 
                rotateAlways: false,
                hideOverlappingLabels: true 
            }, 
            tooltip: { enabled: false } 
        },
        yaxis: { labels: { style: { fontSize: '10px' }, formatter: (value: number) => { if (value >= 1000000) return (value / 1000000).toFixed(1) + 'M'; if (value >= 1000) return (value / 1000).toFixed(0) + 'k'; return value; } } },
        grid: { borderColor: '#334155', strokeDashArray: 4, yaxis: { lines: { show: true } }, xaxis: { lines: { show: false } }, padding: { top: 0, right: 0, bottom: 0, left: 10 } },
        colors: ['#10b981', '#ef4444'],
        tooltip: { theme: 'dark', x: { show: true }, y: { formatter: (value: number) => formatCurrency(value) } },
        legend: { position: 'top', horizontalAlign: 'right', offsetY: -20, itemMargin: { horizontal: 10, vertical: 0 } }
    };
});
</script>

<template>
    <div class="h-full w-full">
         <apexchart type="area" height="100%" width="100%" :options="chartOptions" :series="chartSeries" />
    </div>
</template>
