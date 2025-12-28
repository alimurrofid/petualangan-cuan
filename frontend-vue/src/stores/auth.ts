import { defineStore } from 'pinia';
import { ref } from 'vue';
import api from '@/lib/api';
import router from '@/router';

export const useAuthStore = defineStore('auth', () => {
    const token = ref(localStorage.getItem('token') || '');
    const user = ref(JSON.parse(localStorage.getItem('user') || 'null'));

    const login = async (credentials: any) => {
        try {
            const response = await api.post('/auth/login', credentials);
            token.value = response.data.token;
            user.value = response.data.user;
            localStorage.setItem('token', token.value);
            localStorage.setItem('user', JSON.stringify(user.value));
            
            router.push('/dashboard');
        } catch (error) {
            throw error;
        }
    };

    const register = async (credentials: any) => {
        try {
            const response = await api.post('/auth/register', credentials);
            token.value = response.data.token;
            user.value = response.data.user;
            localStorage.setItem('token', token.value);
            localStorage.setItem('user', JSON.stringify(user.value));
            router.push('/dashboard');
        } catch (error) {
           throw error;
        }
    };

    const logout = () => {
        token.value = '';
        user.value = null;
        localStorage.removeItem('token');
        localStorage.removeItem('user');
        router.push('/login');
    };

    return { token, user, login, register, logout };
});
