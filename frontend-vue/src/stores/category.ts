import { defineStore } from 'pinia';
import { ref } from 'vue';
import api from '@/lib/api';

export interface Category {
  id: number;
  user_id: number;
  name: string;
  type: 'income' | 'expense';
  icon?: string;
  budget_limit?: number;
  created_at?: string;
  updated_at?: string;
}

export interface CreateCategoryInput {
  name: string;
  type: string;
  icon?: string;
  budget_limit?: number;
}

export interface UpdateCategoryInput {
  name: string;
  type: string;
  icon?: string;
  budget_limit?: number;
}

export const useCategoryStore = defineStore('category', () => {
  const categories = ref<Category[]>([]);
  const isLoading = ref(false);
  const error = ref<string | null>(null);

  const fetchCategories = async () => {
    isLoading.value = true;
    error.value = null;
    try {
      const response = await api.get('/api/categories');
      categories.value = response.data;
    } catch (err: any) {
      error.value = err.response?.data?.error || 'Failed to fetch categories';
      console.error(err);
    } finally {
      isLoading.value = false;
    }
  };

  const createCategory = async (input: CreateCategoryInput) => {
    isLoading.value = true;
    error.value = null;
    try {
      const response = await api.post('/api/categories', input);
      categories.value.push(response.data);
      return response.data;
    } catch (err: any) {
      error.value = err.response?.data?.error || 'Failed to create category';
      throw err;
    } finally {
      isLoading.value = false;
    }
  };

  const updateCategory = async (id: number, input: UpdateCategoryInput) => {
    isLoading.value = true;
    error.value = null;
    try {
      const response = await api.put(`/api/categories/${id}`, input);
      const index = categories.value.findIndex(c => c.id === id);
      if (index !== -1) {
        categories.value[index] = response.data;
      }
      return response.data;
    } catch (err: any) {
      error.value = err.response?.data?.error || 'Failed to update category';
      throw err;
    } finally {
      isLoading.value = false;
    }
  };

  const deleteCategory = async (id: number) => {
    isLoading.value = true;
    error.value = null;
    try {
      await api.delete(`/api/categories/${id}`);
      categories.value = categories.value.filter(c => c.id !== id);
    } catch (err: any) {
      error.value = err.response?.data?.error || 'Failed to delete category';
      throw err;
    } finally {
      isLoading.value = false;
    }
  };

  return {
    categories,
    isLoading,
    error,
    fetchCategories,
    createCategory,
    updateCategory,
    deleteCategory
  };
});
