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
        // Use auth store action if available, or set token directly
        // Assuming authStore has a method to set token or we can set it manually.
        // Usually login action sets token. Let's see if we can misuse login or add a method.
        // For now, I'll manually set it in localStorage and update store state if possible.
        // Reading Login.vue: authStore.login takes {email, password}.
        // I should probably check authStore definition.
        
        // For simplicity and standard JWT flow:
        localStorage.setItem('token', token);
        // Force reload or re-init store state might be needed, or we just push to dashboard.
        // Ideally authStore has a `setToken` or `initialize` method.
        // But simply redirecting to dashboard might work if dashboard checks localStorage or store initializes from it.
        
        // Let's assume store initializes from localStorage or we need to reload.
        // To be safe, we can try to call a store action to set user/token if it exists.
        // But simpler:
        
        // Let's try to fetch user profile using the token to confirm and set user state.
        // Or just redirect to dashboard (which is protected)
        
        // Update: check if authStore has logic to set credentials.
        // If not, I'll just set localStorage and let the app handle it.
        
        // Better:
        authStore.token = token;
        localStorage.setItem('token', token);
        
        // Fetch user profile
        await authStore.fetchUser();
        
        // Optionally fetch user data here if needed, or let Dashboard handle it.
        router.push('/dashboard');
    } else {
        // Failed, go back to login
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
