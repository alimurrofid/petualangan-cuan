import { defineStore } from 'pinia';
import { ref } from 'vue';
import api from '@/lib/api';

export interface Wallet {
  id: number;
  user_id: number;
  name: string;
  type: 'Bank' | 'E-Wallet' | 'Cash';
  balance: number;
  icon?: string;
  created_at?: string;
  updated_at?: string;
}

export interface CreateWalletInput {
  name: string;
  type: string;
  balance: number;
  icon?: string;
}

export interface UpdateWalletInput {
  name: string;
  type: string;
  balance: number;
  icon?: string;
}

export const useWalletStore = defineStore('wallet', () => {
  const wallets = ref<Wallet[]>([]);
  const isLoading = ref(false);
  const error = ref<string | null>(null);

  const fetchWallets = async () => {
    isLoading.value = true;
    error.value = null;
    try {
      const response = await api.get('/api/wallets');
      wallets.value = response.data;
    } catch (err: any) {
      error.value = err.response?.data?.error || 'Failed to fetch wallets';
      console.error(err);
    } finally {
      isLoading.value = false;
    }
  };

  const createWallet = async (input: CreateWalletInput) => {
    isLoading.value = true;
    error.value = null;
    try {
      const response = await api.post('/api/wallets', input);
      wallets.value.push(response.data);
      return response.data;
    } catch (err: any) {
      error.value = err.response?.data?.error || 'Failed to create wallet';
      throw err;
    } finally {
      isLoading.value = false;
    }
  };

  const updateWallet = async (id: number, input: UpdateWalletInput) => {
    isLoading.value = true;
    error.value = null;
    try {
      const response = await api.put(`/api/wallets/${id}`, input);
      const index = wallets.value.findIndex(w => w.id === id);
      if (index !== -1) {
        wallets.value[index] = response.data;
      }
      return response.data;
    } catch (err: any) {
      error.value = err.response?.data?.error || 'Failed to update wallet';
      throw err;
    } finally {
      isLoading.value = false;
    }
  };

  const deleteWallet = async (id: number) => {
    isLoading.value = true;
    error.value = null;
    try {
      await api.delete(`/api/wallets/${id}`);
      wallets.value = wallets.value.filter(w => w.id !== id);
    } catch (err: any) {
      error.value = err.response?.data?.error || 'Failed to delete wallet';
      throw err;
    } finally {
      isLoading.value = false;
    }
  };

  return {
    wallets,
    isLoading,
    error,
    fetchWallets,
    createWallet,
    updateWallet,
    deleteWallet
  };
});
