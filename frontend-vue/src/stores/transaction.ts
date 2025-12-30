import { defineStore } from 'pinia';
import { ref } from 'vue';
import api from '@/lib/api';
import { useWalletStore } from './wallet';

export interface Transaction {
  id: number;
  user_id: number;
  wallet_id: number;
  category_id: number;
  amount: number;
  type: 'income' | 'expense' | 'transfer_in' | 'transfer_out';
  description: string;
  date: string;
  wallet: {
    name: string;
    icon: string;
    type: string;
  };
  category: {
    name: string;
    icon: string;
    type: string;
  };
}

export interface CreateTransactionInput {
  wallet_id: number;
  category_id: number;
  amount: number;
  type: string;
  description: string;
  date: string;
}

export interface TransferInput {
  from_wallet_id: number;
  to_wallet_id: number;
  amount: number;
  description: string;
  date: string;
}

export interface TransactionSummary {
  date: string;
  income: number;
  expense: number;
}

export interface CategoryBreakdown {
  category_name: string;
  category_icon: string;
  type: string;
  total_amount: number;
  budget_limit: number;
  is_over_budget: boolean;
}

export interface TransactionFilterParams {
  page?: number;
  limit?: number;
  start_date?: string;
  end_date?: string;
  wallet_id?: number | string;
  category_id?: number | string;
  search?: string;
  type?: string;
}

export interface PaginationMeta {
  total: number;
  page: number;
  limit: number;
}

export const useTransactionStore = defineStore('transaction', () => {
  const transactions = ref<Transaction[]>([]);
  const paginationMeta = ref<PaginationMeta>({ total: 0, page: 1, limit: 10 });
  const reportData = ref<CategoryBreakdown[]>([]);
  const isLoading = ref(false);
  const error = ref<string | null>(null);

  // Filter State
  const filters = ref<TransactionFilterParams>({
    page: 1,
    limit: 10,
    start_date: '',
    end_date: '',
    wallet_id: 'all',
    category_id: 'all',
    search: '',
    type: ''
  });

  const setFilters = (newFilters: Partial<TransactionFilterParams>) => {
    filters.value = { ...filters.value, ...newFilters };
  };

  const fetchTransactions = async (params: TransactionFilterParams = {}) => {
    isLoading.value = true;
    error.value = null;
    
    // Merge provided params with current store filters
    const finalParams = { ...filters.value, ...params };

    try {
      const queryParams = new URLSearchParams();
      if (finalParams.page) queryParams.append('page', finalParams.page.toString());
      if (finalParams.limit) queryParams.append('limit', finalParams.limit.toString());
      if (finalParams.start_date) queryParams.append('start_date', finalParams.start_date);
      if (finalParams.end_date) queryParams.append('end_date', finalParams.end_date);
      if (finalParams.wallet_id && finalParams.wallet_id !== 'all') queryParams.append('wallet_id', finalParams.wallet_id.toString());
      if (finalParams.category_id && finalParams.category_id !== 'all') queryParams.append('category_id', finalParams.category_id.toString());
      if (finalParams.search) queryParams.append('search', finalParams.search);
      if (finalParams.type) queryParams.append('type', finalParams.type);

      const response = await api.get(`/api/transactions?${queryParams.toString()}`);
      if (response.data.status === 'success') {
          transactions.value = response.data.data;
          if (response.data.meta) {
              paginationMeta.value = response.data.meta;
          }
      }
    } catch (err: any) {
      error.value = err.response?.data?.message || 'Failed to fetch transactions';
      console.error(err);
    } finally {
      isLoading.value = false;
    }
  };


  const fetchReport = async (startDate: string, endDate: string, walletId?: number, type?: string) => {
    isLoading.value = true;
    error.value = null;
    try {
        let url = `/api/transactions/report?start_date=${startDate}&end_date=${endDate}`;
        if (walletId) url += `&wallet_id=${walletId}`;
        if (type) url += `&type=${type}`;

        const response = await api.get(url);
        if (response.data.status === 'success') {
          reportData.value = response.data.data;
          return reportData.value;
        }
        return null;
    } catch (err: any) {
      error.value = err.response?.data?.message || 'Failed to fetch report';
      console.error(err);
      throw err;
    } finally {
      isLoading.value = false;
    }
  };

  const createTransaction = async (input: CreateTransactionInput) => {
    isLoading.value = true;
    error.value = null;
    const walletStore = useWalletStore();
    try {
      const response = await api.post('/api/transactions', input);
      // Refresh transactions and wallets (balance changed)
      await refreshData();
      await walletStore.fetchWallets();
      return response.data;
    } catch (err: any) {
      error.value = err.response?.data?.error || 'Failed to create transaction';
      throw err;
    } finally {
      isLoading.value = false;
    }
  };

  const transfer = async (input: TransferInput) => {
    isLoading.value = true;
    error.value = null;
    const walletStore = useWalletStore();
    try {
      const response = await api.post('/api/transactions/transfer', input);
      await refreshData();
      await walletStore.fetchWallets();
      return response.data;
    } catch (err: any) {
      error.value = err.response?.data?.error || 'Failed to transfer';
      throw err;
    } finally {
      isLoading.value = false;
    }
  };

  const deleteTransaction = async (id: number) => {
    isLoading.value = true;
    error.value = null;
    const walletStore = useWalletStore();
    try {
      await api.delete(`/api/transactions/${id}`);
      // Refresh transactions and wallets (balance reverted)
      await refreshData();
      await walletStore.fetchWallets();
    } catch (err: any) {
      error.value = err.response?.data?.error || 'Failed to delete transaction';
      throw err;
    } finally {
      isLoading.value = false;
    }
  };

  const calendarData = ref<TransactionSummary[]>([]);

  const fetchCalendarData = async (startDate?: string, endDate?: string, walletId?: number | string, categoryId?: number | string, search?: string) => {
    isLoading.value = true;
    try {
      // Use provided params or store filters
      const start = startDate || filters.value.start_date;
      const end = endDate || filters.value.end_date;
      const wId = walletId || filters.value.wallet_id;
      const cId = categoryId || filters.value.category_id;
      const sTerm = search !== undefined ? search : filters.value.search;

      if (!start || !end) return [];

      let url = `/api/transactions/calendar?start_date=${start}&end_date=${end}`;
      if (wId && wId !== 'all') url += `&wallet_id=${wId}`;
      if (cId && cId !== 'all') url += `&category_id=${cId}`;
      if (sTerm) url += `&search=${encodeURIComponent(sTerm)}`;

      const response = await api.get(url);
      if (response.data.status === 'success' && Array.isArray(response.data.data)) {
          calendarData.value = response.data.data as TransactionSummary[];
          return calendarData.value;
      }
      calendarData.value = [];
      return [];
    } catch (err: any) {
      console.error(err);
      calendarData.value = [];
      return [];
    } finally {
      isLoading.value = false;
    }
  };

  const refreshData = async () => {
      // Refresh with current filters
      await Promise.all([
          fetchTransactions(),
          fetchCalendarData()
      ]);
  };

  return {
    transactions,
    calendarData,
    paginationMeta,
    filters,
    isLoading,
    error,
    setFilters,
    fetchTransactions,
    createTransaction,
    deleteTransaction,
    transfer,
    fetchCalendarData,
    fetchReport,
    reportData,
    refreshData
  };
});
