import { defineStore } from 'pinia';
import { ref } from 'vue';
import api from '@/lib/api';
import type { Wallet } from './wallet';
import type { Transaction, CategoryBreakdown } from './transaction';

export interface MonthlyTrend {
  date: string;
  income: number;
  expense: number;
}

export interface PaceMetric {
  total_budget: number;
  remaining_budget: number;
  days_remaining: number;
  daily_safe_to_spend: number;
  today_expense: number;
  today_remaining_safe: number;
  status: 'SAFE' | 'WARNING' | 'OVERPACE';
}

export interface DashboardData {
  total_balance: number;
  total_available_balance: number;
  total_income_month: number;
  total_expense_month: number;
  wallets: Wallet[];
  recent_transactions: Transaction[];
  monthly_trend: MonthlyTrend[];
  expense_breakdown: CategoryBreakdown[];
  pace_metric?: PaceMetric | null;
}

export const useDashboardStore = defineStore('dashboard', () => {
    const data = ref<DashboardData | null>(null);
    const isLoading = ref(false);
    const error = ref<string | null>(null);

    const fetchDashboard = async () => {
        isLoading.value = true;
        error.value = null;
        try {
            const response = await api.get('/api/dashboard');
            if (response.data.status === 'success') {
                data.value = response.data.data;
            }
        } catch (err: any) {
             error.value = err.response?.data?.message || 'Failed to fetch dashboard data';
             console.error(err);
        } finally {
            isLoading.value = false;
        }
    };

    return {
        data,
        isLoading,
        error,
        fetchDashboard
    };
});
