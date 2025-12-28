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

export const useTransactionStore = defineStore('transaction', () => {
  const transactions = ref<Transaction[]>([]);
  const reportData = ref<CategoryBreakdown[]>([]);
  const isLoading = ref(false);
  const error = ref<string | null>(null);

  const fetchTransactions = async () => {
    isLoading.value = true;
    error.value = null;
    try {
      const response = await api.get('/api/transactions');
      if (response.data.status === 'success') {
          transactions.value = response.data.data;
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
      await fetchTransactions();
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
      await fetchTransactions();
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
      await fetchTransactions();
      await walletStore.fetchWallets();
    } catch (err: any) {
      error.value = err.response?.data?.error || 'Failed to delete transaction';
      throw err;
    } finally {
      isLoading.value = false;
    }
  };

  const fetchCalendarData = async (startDate: string, endDate: string) => {
    isLoading.value = true;
    try {
      const response = await api.get(`/api/transactions/calendar?start_date=${startDate}&end_date=${endDate}`);
      if (response.data.status === 'success') {
          return response.data.data as TransactionSummary[];
      }
      return [];
    } catch (err: any) {
      console.error(err);
      return [];
    } finally {
      isLoading.value = false;
    }
  };

  return {
    transactions,
    isLoading,
    error,
    fetchTransactions,
    createTransaction,
    deleteTransaction,
    transfer,
    fetchCalendarData,
    fetchReport,
    reportData
  };
});
