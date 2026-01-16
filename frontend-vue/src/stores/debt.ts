import { defineStore } from 'pinia';
import { ref } from 'vue';
import api from '@/lib/api';
import { useWalletStore } from './wallet';
import { useTransactionStore } from './transaction';

export interface DebtPayment {
  id: number;
  debt_id: number;
  transaction_id: number;
  wallet_id: number;
  wallet: {
      id: number;
      name: string;
      icon?: string;
  };
  amount: number;
  date: string;
  note: string;
}

export interface Debt {
  id: number;
  user_id: number;
  name: string;
  amount: number;
  remaining: number;
  type: 'debt' | 'receivable';
  description: string;
  due_date: string;
  is_paid: boolean;
  wallet_id: number;
  wallet: {
    id: number;
    name: string;
    type: string;
    icon?: string;
  };
  payments: DebtPayment[];
  created_at: string;
  updated_at: string;
}

export interface CreateDebtInput {
  wallet_id: number;
  name: string;
  amount: number;
  type: 'debt' | 'receivable';
  description: string;
  due_date?: string | null;
}

export interface PayDebtInput {
  wallet_id: number;
  amount: number;
  note: string;
}


export interface UpdateDebtInput {
  wallet_id: number;
  name: string;
  amount: number;
  description: string;
  due_date: string | null;
}

export const useDebtStore = defineStore('debt', () => {
  const debts = ref<Debt[]>([]);
  const receivables = ref<Debt[]>([]);
  const isLoading = ref(false);
  const error = ref<string | null>(null);

  const fetchDebts = async () => {
    isLoading.value = true;
    error.value = null;
    try {
      // Fetch both types
      const debtsRes = await api.get('/api/debts?type=debt');
      const receivablesRes = await api.get('/api/debts?type=receivable');
      
      debts.value = debtsRes.data.data;
      receivables.value = receivablesRes.data.data;
    } catch (err: any) {
      error.value = err.response?.data?.error || 'Failed to fetch debts';
      console.error(err);
    } finally {
      isLoading.value = false;
    }
  };

  const createDebt = async (input: CreateDebtInput) => {
    isLoading.value = true;
    error.value = null;
    const walletStore = useWalletStore();
    const transactionStore = useTransactionStore();
    try {
      const response = await api.post('/api/debts', input);
      await fetchDebts();
      await walletStore.fetchWallets(); // Update balance
      await transactionStore.fetchTransactions(); // Update history
      return response.data;
    } catch (err: any) {
      error.value = err.response?.data?.error || 'Failed to create debt';
      throw err;
    } finally {
      isLoading.value = false;
    }
  };

  const updateDebt = async (id: number, input: UpdateDebtInput) => {
    isLoading.value = true;
    error.value = null;
    try {
      const response = await api.put(`/api/debts/${id}`, input);
      await fetchDebts();
      return response.data;
    } catch (err: any) {
      error.value = err.response?.data?.error || 'Failed to update debt';
      throw err;
    } finally {
      isLoading.value = false;
    }
  };

  const payDebt = async (id: number, input: PayDebtInput) => {
    isLoading.value = true;
    error.value = null;
    const walletStore = useWalletStore();
    const transactionStore = useTransactionStore();
    try {
      const response = await api.post(`/api/debts/${id}/pay`, input);
      await fetchDebts();
      await walletStore.fetchWallets();
      await transactionStore.fetchTransactions();
      return response.data;
    } catch (err: any) {
      error.value = err.response?.data?.error || 'Failed to pay debt';
      throw err;
    } finally {
      isLoading.value = false;
    }
  };

  const deleteDebt = async (id: number) => {
    isLoading.value = true;
    error.value = null;
    try {
      await api.delete(`/api/debts/${id}`);
      await fetchDebts();
    } catch (err: any) {
      error.value = err.response?.data?.error || 'Failed to delete debt';
      throw err;
    } finally {
      isLoading.value = false;
    }
  };

  const deletePayment = async (id: number) => {
    isLoading.value = true;
    error.value = null;
    try {
      await api.delete(`/api/debts/payments/${id}`);
      await fetchDebts();
      const walletStore = useWalletStore();
      const transactionStore = useTransactionStore();
      await walletStore.fetchWallets();
      await transactionStore.fetchTransactions(); 
    } catch (err: any) {
      error.value = err.response?.data?.error || 'Failed to delete payment';
      throw err;
    } finally {
      isLoading.value = false;
    }
  };

  return {
    debts,
    receivables,
    isLoading,
    error,
    fetchDebts,
    createDebt,
    updateDebt,
    payDebt,
    deleteDebt,
    deletePayment
  };
});
