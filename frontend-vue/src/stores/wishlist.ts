import { defineStore } from 'pinia';
import { ref } from 'vue';
import axios from 'axios';
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
    const baseUrl = import.meta.env.VITE_API_BASE_URL;
    const swal = useSwal();

    const fetchItems = async () => {
        try {
            const token = localStorage.getItem('token');
            const response = await axios.get(`${baseUrl}/api/wishlist`, {
                headers: { Authorization: `Bearer ${token}` }
            });
            items.value = response.data;
        } catch (error) {
            console.error('Failed to fetch wishlist items:', error);
        }
    };

    const createItem = async (data: Partial<WishlistItem>) => {
        try {
            const token = localStorage.getItem('token');
            await axios.post(`${baseUrl}/api/wishlist`, data, {
                headers: { Authorization: `Bearer ${token}` }
            });
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
            const token = localStorage.getItem('token');
            await axios.put(`${baseUrl}/api/wishlist/${id}`, data, {
                headers: { Authorization: `Bearer ${token}` }
            });
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
            const token = localStorage.getItem('token');
            await axios.delete(`${baseUrl}/api/wishlist/${id}`, {
                headers: { Authorization: `Bearer ${token}` }
            });
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
            const token = localStorage.getItem('token');
            await axios.patch(`${baseUrl}/api/wishlist/${id}/bought`, {}, {
                headers: { Authorization: `Bearer ${token}` }
            });
            await fetchItems();
            // Toast is handled by the calling component usually, but we can add one here if needed
            return true;
        } catch (error) {
            console.error('Failed to mark as bought:', error);
            return false;
        }
    };

    return {
        items,
        fetchItems,
        createItem,
        updateItem,
        deleteItem,
        markAsBought
    };
});
