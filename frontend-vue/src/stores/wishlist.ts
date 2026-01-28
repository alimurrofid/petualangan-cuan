import { defineStore } from 'pinia';
import { ref } from 'vue';
import api from '@/lib/api';
import { useSwal } from '@/composables/useSwal';

export interface WishlistItem {
    id: number;
    category_id: number;
    category?: {
        id: number;
        name: string;
        type: string;
        icon: string;
        color: string;
    };
    name: string;
    estimated_price: number;
    is_bought: boolean;
    priority: 'low' | 'medium' | 'high';
    created_at: string;
}

export const useWishlistStore = defineStore('wishlist', () => {
    const items = ref<WishlistItem[]>([]);
    const isLoading = ref(false);
    const swal = useSwal();

    const fetchItems = async () => {
        isLoading.value = true;
        try {
            const response = await api.get('/api/wishlist');
            items.value = response.data;
        } catch (error) {
            console.error('Failed to fetch wishlist items:', error);
        } finally {
            isLoading.value = false;
        }
    };

    const createItem = async (data: Partial<WishlistItem>) => {
        try {
            await api.post('/api/wishlist', data);
            await fetchItems();
            swal.toast({ icon: 'success', title: 'Item wishlist berhasil ditambahkan' });
            return true;
        } catch (error) {
            swal.error('Gagal', 'Gagal menambahkan item wishlist');
            return false;
        }
    };

    const updateItem = async (id: number, data: Partial<WishlistItem>) => {
        try {
            await api.put(`/api/wishlist/${id}`, data);
            await fetchItems();
            swal.toast({ icon: 'success', title: 'Item wishlist berhasil diperbarui' });
            return true;
        } catch (error) {
            swal.error('Gagal', 'Gagal memperbarui item wishlist');
            return false;
        }
    };

    const deleteItem = async (id: number) => {
        const confirmed = await swal.confirm('Apakah Anda yakin?', 'Item wishlist akan dihapus permanen');
        if (!confirmed) return false;

        try {
            await api.delete(`/api/wishlist/${id}`);
            await fetchItems();
            swal.toast({ icon: 'success', title: 'Item wishlist berhasil dihapus' });
            return true;
        } catch (error) {
            swal.error('Gagal', 'Gagal menghapus item wishlist');
            return false;
        }
    };

    const markAsBought = async (id: number) => {
        try {
            await api.patch(`/api/wishlist/${id}/bought`, {});
            return true;
        } catch (error) {
            console.error('Failed to mark as bought:', error);
            return false;
        }
    };

    return {
        items,
        isLoading,
        fetchItems,
        createItem,
        updateItem,
        deleteItem,
        markAsBought
    };
});
