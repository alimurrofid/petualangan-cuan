<script setup lang="ts">
import { onMounted } from 'vue';
import { useRouter, useRoute } from 'vue-router';
import { useAuthStore } from '@/stores/auth';

const router = useRouter();
const route = useRoute();
const authStore = useAuthStore();

onMounted(async () => {
    const token = route.query.token as string;
    
    if (token) {
        localStorage.setItem('token', token);
        
        authStore.token = token;
        localStorage.setItem('token', token);
        
        await authStore.fetchUser();
        
        router.push('/dashboard');
    } else {
        router.push('/login?error=Google login failed');
    }
});
</script>

<template>
    <div class="flex items-center justify-center min-h-screen">
        <div class="text-center">
            <h2 class="text-xl font-semibold mb-2">Processing Google Login...</h2>
            <p class="text-muted-foreground">Please wait while we log you in.</p>
        </div>
    </div>
</template>
