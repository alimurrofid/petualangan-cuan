<script setup lang="ts">
import { computed, onMounted } from 'vue';
import { useFinancialHealthStore } from '@/stores/financialHealth';
import { Card, CardContent, CardHeader } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { RotateCcw, TrendingUp, Scale, HeartPulse, PiggyBank, ShieldCheck, AlertTriangle, AlertOctagon } from 'lucide-vue-next';

const store = useFinancialHealthStore();

onMounted(() => {
    store.fetchFinancialHealth();
});

const series = computed(() => [store.data?.overall_score || 0]);

const chartOptions = computed(() => ({
    chart: {
        type: 'radialBar',
        offsetY: -10,
        sparkline: {
            enabled: true
        },
        animations: {
            enabled: true,
            easing: 'easeinout',
            speed: 800,
        }
    },
    plotOptions: {
        radialBar: {
            startAngle: -90,
            endAngle: 90,
            track: {
                background: "#f1f5f9",
                strokeWidth: '97%',
                margin: 5,
                dropShadow: {
                    enabled: true,
                    top: 2,
                    left: 0,
                    blur: 3,
                    opacity: 0.15
                }
            },
            hollow: {
                margin: 15,
                size: "65%"
            },
            dataLabels: {
                name: {
                    show: false
                },
                value: {
                    offsetY: -30,
                    fontSize: '40px',
                    fontWeight: 700,
                    color: getColor(store.data?.overall_score || 0),
                    formatter: function (val: number) {
                        return val + "%";
                    }
                }
            }
        }
    },
    fill: {
        type: 'gradient',
        gradient: {
            shade: 'light',
            shadeIntensity: 0.4,
            inverseColors: false,
            opacityFrom: 1,
            opacityTo: 1,
            stops: [0, 50, 53, 91]
        },
    },
    stroke: {
        dashArray: 0
    },
    colors: [getColor(store.data?.overall_score || 0)]
}));

function getColor(score: number) {
    if (score >= 80) return '#10b981'; // emerald-500
    if (score >= 40) return '#f59e0b'; // amber-500
    return '#ef4444'; // red-500
}

function getStatusColorClass(status: string) {
    switch (status) {
        case 'Sehat': return 'bg-emerald-100 text-emerald-700 dark:bg-emerald-500/20 dark:text-emerald-400 border-emerald-200 dark:border-emerald-800';
        case 'Waspada': return 'bg-amber-100 text-amber-700 dark:bg-amber-500/20 dark:text-amber-400 border-amber-200 dark:border-amber-800';
        case 'Bahaya': return 'bg-red-100 text-red-700 dark:bg-red-500/20 dark:text-red-400 border-red-200 dark:border-red-800';
        default: return 'bg-slate-100 text-slate-700 dark:bg-slate-800 dark:text-slate-300';
    }
}

function getCardBorderClass(status: string) {
    switch (status) {
        case 'Sehat': return 'hover:border-emerald-300/50 dark:hover:border-emerald-700/50';
        case 'Waspada': return 'hover:border-amber-300/50 dark:hover:border-amber-700/50';
        case 'Bahaya': return 'hover:border-red-300/50 dark:hover:border-red-700/50';
        default: return 'hover:border-border';
    }
}

const getIcon = (name: string) => {
    if (name.includes('Savings')) return PiggyBank;
    if (name.includes('Dana Darurat')) return ShieldCheck;
    if (name.includes('Debt')) return Scale;
    return TrendingUp;
}

const getStatusIcon = (status: string) => {
    if (status === 'Sehat') return ShieldCheck;
    if (status === 'Waspada') return AlertTriangle;
    return AlertOctagon;
}
</script>

<template>
  <div class="space-y-8 animate-in fade-in slide-in-from-bottom-4 duration-700">
    <div class="flex items-center justify-between">
      <div>
        <h2 class="text-3xl font-bold tracking-tight text-slate-900 dark:text-white">Kesehatan Keuangan</h2>
        <p class="text-muted-foreground mt-1 text-sm">Analisa kondisi finansial Anda secara objektif.</p>
      </div>
      <button @click="store.fetchFinancialHealth" class="group flex items-center gap-2 px-4 py-2 rounded-full bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-800 shadow-sm hover:shadow-md transition-all text-xs font-medium">
        <RotateCcw class="w-3.5 h-3.5 text-emerald-600 group-hover:-rotate-180 transition-transform duration-500" :class="{'animate-spin': store.isLoading}" />
        <span>Refresh Analisa</span>
      </button>
    </div>

    <div v-if="store.isLoading" class="flex flex-col items-center justify-center py-20 gap-4">
        <div class="relative">
            <div class="absolute inset-0 bg-emerald-500/20 blur-xl rounded-full animate-pulse"></div>
            <HeartPulse class="w-16 h-16 animate-pulse text-emerald-500 relative z-10" />
        </div>
        <p class="text-muted-foreground font-medium animate-pulse text-sm">Sedang memeriksa kesehatan dompetmu...</p>
    </div>

    <div v-else-if="store.data" class="space-y-6">
        <!-- Main Score -->
        <div class="relative overflow-hidden rounded-3xl bg-white dark:bg-slate-950 border border-emerald-100 dark:border-emerald-900/30 shadow-xl shadow-emerald-900/5">
             <!-- Background Decoration -->
            <div class="absolute top-0 right-0 -mt-20 -mr-20 w-96 h-96 bg-gradient-to-br from-emerald-500/10 to-teal-500/5 rounded-full blur-3xl pointer-events-none"></div>
            <div class="absolute bottom-0 left-0 -mb-20 -ml-20 w-80 h-80 bg-gradient-to-tr from-blue-500/10 to-indigo-500/5 rounded-full blur-3xl pointer-events-none"></div>

            <div class="relative z-10 p-6 sm:p-8 flex flex-col md:flex-row items-center justify-between gap-10">
                <div class="flex-1 text-center md:text-left space-y-3">
                    <div class="inline-flex items-center gap-2 px-3 py-1 rounded-full text-[10px] font-bold uppercase tracking-wider bg-white/80 dark:bg-slate-900/80 backdrop-blur-sm border border-slate-200 dark:border-slate-800 shadow-sm">
                        <Scale class="w-3 h-3" />
                        Diagnosa Umum
                    </div>
                     <h3 class="text-2xl sm:text-3xl font-bold text-slate-900 dark:text-white leading-tight">
                         <span v-if="store.data.overall_status === 'Sehat'">Kondisi Keuangan <span class="text-emerald-600">Prima!</span> 🎉</span>
                         <span v-else-if="store.data.overall_status === 'Waspada'">Perlu <span class="text-amber-500">Perhatian</span> ⚠️</span>
                         <span v-else>Kondisi <span class="text-red-500">Kritis</span> 🚨</span>
                     </h3>
                     <p v-if="store.data.overall_status === 'Sehat'" class="text-sm text-slate-600 dark:text-slate-300 max-w-xl leading-relaxed">
                         Luar biasa! Anda telah mengelola keuangan dengan sangat baik. Pertahankan disiplin ini dan fokus pada pertumbuhan aset jangka panjang.
                     </p>
                     <p v-else-if="store.data.overall_status === 'Waspada'" class="text-sm text-slate-600 dark:text-slate-300 max-w-xl leading-relaxed">
                         Anda berada di jalur yang benar, namun ada beberapa indikator yang perlu diperbaiki. Cek detail di bawah untuk kembali ke zona hijau.
                     </p>
                     <p v-else class="text-sm text-slate-600 dark:text-slate-300 max-w-xl leading-relaxed">
                         Saatnya tindakan darurat. Segera evaluasi pengeluaran dan struktur utang Anda untuk menghindari masalah finansial yang lebih serius.
                     </p>
                </div>

                <div class="relative w-72 h-60 flex items-center justify-center shrink-0">
                    <apexchart type="radialBar" height="300" width="100%" :options="chartOptions" :series="series"></apexchart>
                    <div class="absolute bottom-6 text-center pointer-events-none transform translate-y-2">
                        <p class="text-[10px] font-bold text-slate-400 uppercase tracking-[0.2em] mb-1.5">Score</p>
                         <h3 class="text-base font-bold tracking-tight inline-flex items-center gap-2 px-4 py-1.5 rounded-full bg-white/80 dark:bg-slate-900/80 backdrop-blur-md shadow-sm border border-slate-100 dark:border-slate-800" :class="{
                            'text-emerald-600': store.data.overall_status === 'Sehat',
                            'text-amber-600': store.data.overall_status === 'Waspada',
                            'text-red-600': store.data.overall_status === 'Bahaya'
                        }">
                        {{ store.data.overall_status }}
                        </h3>
                    </div>
                </div>
            </div>
        </div>

        <!-- Details Grid -->
        <div class="grid gap-4 md:grid-cols-3">
            <Card v-for="ratio in store.data.ratios" :key="ratio.name" 
                class="overflow-hidden relative group transition-all duration-300 bg-white dark:bg-slate-950 border-slate-200 dark:border-slate-800 hover:shadow-xl hover:-translate-y-1"
                :class="getCardBorderClass(ratio.status)"
            >
                 <!-- Background Icon Faded -->
                 <div class="absolute top-2 right-2 p-4 opacity-[0.03] group-hover:opacity-[0.07] transition-opacity duration-500 scale-150 transform rotate-12 origin-top-right">
                    <component :is="getIcon(ratio.name)" class="w-32 h-32" />
                 </div>
                 
                 <CardHeader class="pb-2">
                     <div class="flex items-center justify-between">
                        <div class="flex items-center gap-3">
                             <div class="p-2 rounded-lg bg-slate-50 dark:bg-slate-900 border border-slate-100 dark:border-slate-800 text-slate-600 dark:text-slate-400 group-hover:text-emerald-600 group-hover:bg-emerald-50 dark:group-hover:bg-emerald-950/30 transition-colors">
                                 <component :is="getIcon(ratio.name)" class="w-4 h-4" />
                             </div>
                             <h4 class="text-sm font-medium text-slate-700 dark:text-slate-200">{{ ratio.name }}</h4>
                        </div>
                        <Badge variant="outline" class="px-2 py-0 border text-[10px]" :class="getStatusColorClass(ratio.status)">
                            <component :is="getStatusIcon(ratio.status)" class="w-3 h-3 mr-1" />
                            {{ ratio.status }}
                        </Badge>
                     </div>
                 </CardHeader>
                 
                 <CardContent class="space-y-4 pt-2">
                     <div>
                         <div class="flex items-baseline gap-1">
                             <span class="text-2xl font-bold tracking-tight text-slate-900 dark:text-white">{{ ratio.formatted_value }}</span>
                         </div>
                         <p class="text-[10px] font-semibold text-muted-foreground mt-1 flex items-center gap-1">
                            TARGET: <span class="bg-slate-100 dark:bg-slate-800 px-1.5 py-0.5 rounded text-slate-700 dark:text-slate-300">{{ ratio.target }}</span>
                         </p>
                     </div>
                     
                     <div class="relative pl-3 py-1">
                        <div class="absolute left-0 top-0 bottom-0 w-1 rounded-full" :class="{
                            'bg-emerald-500': ratio.status === 'Sehat',
                            'bg-amber-400': ratio.status === 'Waspada',
                            'bg-red-500': ratio.status === 'Bahaya'
                        }"></div>
                        <p class="text-xs text-slate-600 dark:text-slate-400 leading-relaxed italic">
                            "{{ ratio.description }}"
                        </p>
                     </div>
                 </CardContent>
            </Card>
        </div>
    </div>
  </div>
</template>
