<script setup lang="ts">
import { ref } from "vue";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Loader2, Mail, Lock, Eye, EyeOff } from "lucide-vue-next";

import { useAuthStore } from "@/stores/auth";

const authStore = useAuthStore();
const isLoading = ref(false);
const email = ref("");
const password = ref("");
const showPassword = ref(false);
const errorMessage = ref("");

const handleLogin = async () => {
  isLoading.value = true;
  errorMessage.value = "";
  try {
    await authStore.login({
        email: email.value,
        password: password.value
    });
    // Redirect handled in store or here. Store handles it currently.
  } catch (error: any) {
    if (error.response && error.response.data && error.response.data.error) {
        errorMessage.value = error.response.data.error;
    } else {
        errorMessage.value = "Gagal masuk. Periksa kembali email dan password Anda.";
    }
  } finally {
    isLoading.value = false;
  }
};
</script>

<template>
  <div class="min-h-screen grid lg:grid-cols-2">
    <!-- Left: Branding -->
    <div class="hidden lg:flex flex-col justify-between p-10 bg-gradient-to-br from-emerald-900 to-teal-900 text-white relative overflow-hidden">
        <!-- Decorative Elements -->
        <div class="absolute top-0 left-0 w-full h-full opacity-10 pointer-events-none">
             <svg class="w-full h-full" viewBox="0 0 100 100" preserveAspectRatio="none">
                <path d="M0 100 C 20 0 50 0 100 100 Z" fill="white" />
             </svg>
        </div>

        <div class="relative z-10 flex items-center gap-2">
             <img src="/petualangancuan.svg" alt="Logo" class="w-10 h-10 rounded-full bg-white/10 p-1" />
             <span class="text-xl font-bold tracking-tigher">Petualangan Cuan</span>
        </div>

        <div class="relative z-10 space-y-4 max-w-lg">
            <h1 class="text-5xl font-bold leading-tight">
                Kelola keuangan Anda dengan <span class="text-emerald-400">cerdas</span> dan <span class="text-teal-400">efisien</span>.
            </h1>
            <p class="text-lg text-emerald-100/80">
                Bergabunglah dengan ribuan pengguna lainnya dan mulailah perjalanan finansial Anda menuju kebebasan.
            </p>
        </div>

        <div class="relative z-10 text-sm text-emerald-200/50">
            &copy; 2025 Petualangan Cuan. All rights reserved.
        </div>
    </div>

    <!-- Right: Form -->
    <div class="flex items-center justify-center p-8 bg-background">
        <div class="w-full max-w-md space-y-8">
            <div class="text-center space-y-2">
                <div class="lg:hidden flex justify-center mb-4">
                     <img src="/petualangancuan.svg" alt="Logo" class="w-12 h-12" />
                </div>
                <h2 class="text-3xl font-bold tracking-tight">Selamat Datang Kembali</h2>
                <p class="text-muted-foreground">Masuk ke akun Anda untuk melanjutkan.</p>
            </div>

            <div v-if="errorMessage" class="bg-red-50 text-red-600 p-3 rounded-md text-sm text-center">
                {{ errorMessage }}
            </div>

            <div class="space-y-6">
                <div class="space-y-4">
                     <Button variant="outline" type="button" class="w-full h-12 text-base font-normal flex items-center justify-center gap-3 bg-white hover:bg-gray-50 text-gray-700 border-gray-200 shadow-sm" :disabled="isLoading">
                        <svg class="h-5 w-5" viewBox="0 0 24 24">
                            <path d="M22.56 12.25c0-.78-.07-1.53-.2-2.25H12v4.26h5.92c-.26 1.37-1.04 2.53-2.21 3.31v2.77h3.57c2.08-1.92 3.28-4.74 3.28-8.09z" fill="#4285F4"/>
                            <path d="M12 23c2.97 0 5.46-.98 7.28-2.66l-3.57-2.77c-.98.66-2.23 1.06-3.71 1.06-2.86 0-5.29-1.93-6.16-4.53H2.18v2.84C3.99 20.53 7.7 23 12 23z" fill="#34A853"/>
                            <path d="M5.84 14.09c-.22-.66-.35-1.36-.35-2.09s.13-1.43.35-2.09V7.07H2.18C1.43 8.55 1 10.22 1 12s.43 3.45 1.18 4.93l2.85-2.22.81-.62z" fill="#FBBC05"/>
                            <path d="M12 5.38c1.62 0 3.06.56 4.21 1.64l3.15-3.15C17.45 2.09 14.97 1 12 1 7.7 1 3.99 3.47 2.18 7.07l3.66 2.84c.87-2.6 3.3-4.53 6.16-4.53z" fill="#EA4335"/>
                        </svg>
                        Google
                    </Button>

                    <div class="relative">
                        <div class="absolute inset-0 flex items-center">
                            <span class="w-full border-t"></span>
                        </div>
                        <div class="relative flex justify-center text-xs uppercase">
                            <span class="bg-background px-2 text-muted-foreground lowercase">or login with email</span>
                        </div>
                    </div>
                </div>

                <form @submit.prevent="handleLogin" class="space-y-6">
                    <div class="space-y-4">
                        <div class="space-y-2">
                            <Label for="email">Email</Label>
                            <div class="relative">
                                <Mail class="absolute left-3 top-3 h-4 w-4 text-muted-foreground" />
                                <Input id="email" type="email" placeholder="nama@email.com" class="pl-10 h-10" v-model="email" required />
                            </div>
                        </div>
                        <div class="space-y-2">
                            <Label for="password">Password</Label>
                            <div class="relative">
                                <Lock class="absolute left-3 top-3 h-4 w-4 text-muted-foreground" />
                                <Input id="password" :type="showPassword ? 'text' : 'password'" placeholder="••••••••" class="pl-10 pr-10 h-10" v-model="password" required />
                                <button type="button" @click="showPassword = !showPassword" class="absolute right-3 top-3 text-muted-foreground hover:text-emerald-600 focus:outline-none">
                                    <Eye v-if="!showPassword" class="h-4 w-4" />
                                    <EyeOff v-else class="h-4 w-4" />
                                </button>
                            </div>
                             <div class="flex justify-end pt-1">
                                <a href="#" class="text-xs font-medium text-muted-foreground hover:text-emerald-600">
                                    Forgot password?
                                </a>
                            </div>
                        </div>
                    </div>

                    <Button type="submit" class="w-full bg-emerald-600 hover:bg-emerald-700 font-bold h-10" :disabled="isLoading">
                        <Loader2 v-if="isLoading" class="mr-2 h-4 w-4 animate-spin" />
                        {{ isLoading ? 'Sedang masuk...' : 'Masuk' }}
                    </Button>
                </form>
            </div>

            <div class="text-center text-sm">
                Belum punya akun? 
                <RouterLink to="/register" class="font-medium text-emerald-600 hover:text-emerald-500 hover:underline">
                    Daftar sekarang
                </RouterLink>
            </div>
        </div>
    </div>
  </div>
</template>
