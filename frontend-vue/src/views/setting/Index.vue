<script setup lang="ts">
import { ref, onMounted, watch } from "vue";
import { useRoute } from "vue-router";
import { useAuthStore } from "@/stores/auth";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Card, CardContent } from "@/components/ui/card";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Eye, EyeOff, Loader2 } from "lucide-vue-next";
import { useSwal } from "@/composables/useSwal";

const route = useRoute();
const authStore = useAuthStore();
const activeTab = ref("profile");
const isLoading = ref(false);
const swal = useSwal();

const tabs = [
  { id: "profile", label: "Profile" },
  { id: "currency", label: "Currency & Language" },
  { id: "password", label: "Change Password" },
  { id: "whatsapp", label: "WhatsApp Integration" },
];

// Settings Data (Static for now)
const formData = ref({
    language: "Indonesia",
    timezone: "Asia/Jakarta (GMT+7)",
    currency: "IDR",
    showDecimal: "Hide"
});

// Profile Data
const profileForm = ref({
    name: "",
    email: ""
});

// Password Data
const passwordForm = ref({
    new_password: "",
    confirm_password: ""
});

const showPassword = ref({
    new: false,
    confirm: false
});

const updateTabFromQuery = () => {
    const tab = route.query.tab as string;
    if (tab && tabs.some(t => t.id === tab)) {
        activeTab.value = tab;
    }
};

const initProfile = () => {
    if (authStore.user) {
        profileForm.value.name = authStore.user.name;
        profileForm.value.email = authStore.user.email;
    }
};

onMounted(() => {
    updateTabFromQuery();
    initProfile();
});

watch(() => route.query.tab, () => {
    updateTabFromQuery();
});

watch(() => authStore.user, () => {
    initProfile();
}, { deep: true });

// Validation Errors
const errors = ref({
    profile: {
        name: false,
        email: false
    },
    password: {
        new: false,
        confirm: false,
        match: false
    }
});

const handleUpdateProfile = async () => {
    errors.value.profile.name = !profileForm.value.name;
    errors.value.profile.email = !profileForm.value.email;

    if (errors.value.profile.name || errors.value.profile.email) {
        let msg = "Mohon lengkapi data berikut:";
        if (errors.value.profile.name) msg += "<br>- Nama Lengkap";
        if (errors.value.profile.email) msg += "<br>- Email";
        await swal.fire({
            icon: 'error',
            title: 'Validasi Gagal',
            html: msg,
            confirmButtonColor: '#EF4444', 
        });
        return;
    }
    
    isLoading.value = true;
    try {
        await authStore.updateProfile(profileForm.value);
        swal.success("Berhasil Update", "Profil berhasil diperbarui!");
    } catch (error: any) {
        swal.error("Gagal", error.response?.data?.error || "Gagal memperbarui profil");
    } finally {
        isLoading.value = false;
    }
};

const handleUpdatePassword = async () => {
    // Reset errors
    errors.value.password.new = !passwordForm.value.new_password;
    errors.value.password.confirm = !passwordForm.value.confirm_password;
    errors.value.password.match = false;
    
    // Check required fields
    if (errors.value.password.new || errors.value.password.confirm) {
        let msg = "Mohon lengkapi data berikut:";
        if (errors.value.password.new) msg += "<br>- Password Baru";
        if (errors.value.password.confirm) msg += "<br>- Konfirmasi Password";
        
        await swal.fire({
            icon: 'error',
            title: 'Validasi Gagal',
            html: msg,
            confirmButtonColor: '#EF4444', 
        });
        return;
    }

    // Check matching
    if (passwordForm.value.new_password !== passwordForm.value.confirm_password) {
        errors.value.password.match = true;
         await swal.fire({
            icon: 'error',
            title: 'Validasi Gagal',
            text: 'Konfirmasi password tidak cocok dengan password baru',
            confirmButtonColor: '#EF4444', 
        });
        return;
    }
    
    isLoading.value = true;
    try {
        await authStore.changePassword({
            new_password: passwordForm.value.new_password
        });
        swal.success("Berhasil Update", "Password berhasil diperbarui! Silakan login ulang.").then(() => {
            authStore.logout();
        });
    } catch (error: any) {
        swal.error("Gagal", error.response?.data?.error || "Gagal memperbarui password");
    } finally {
        isLoading.value = false;
        passwordForm.value = { new_password: "", confirm_password: "" };
    }
};
</script>

<template>
  <div class="flex-1 space-y-6 pt-2">
    <div class="flex flex-col md:flex-row md:items-center justify-between gap-4 bg-card p-2 rounded-xl border border-border/50 shadow-sm">
        <div class="flex items-center gap-1 overflow-x-auto no-scrollbar">
            <button 
                v-for="tab in tabs" 
                :key="tab.id"
                @click="activeTab = tab.id"
                :class="[
                    'px-4 py-2 rounded-lg text-sm font-medium transition-colors whitespace-nowrap',
                    activeTab === tab.id 
                        ? 'bg-gradient-to-r from-emerald-600 to-teal-500 text-white shadow-sm' 
                        : 'text-muted-foreground hover:bg-muted start-hover'
                ]"
            >
                {{ tab.label }}
            </button>
        </div>
        <Button variant="destructive" size="sm" class="hidden md:flex">Delete Account</Button>
    </div>

    <Card class="border-border/60 shadow-sm overflow-hidden">
        <CardContent class="p-6">
            
            <!-- Currency Tab -->
            <div v-if="activeTab === 'currency'" class="space-y-6">
                <!-- ... Existing Currency Content ... -->
                <div class="space-y-4">
                    <div class="space-y-2">
                        <Label>Language</Label>
                         <Select v-model="formData.language">
                            <SelectTrigger class="w-full">
                                <SelectValue placeholder="Select Language" />
                            </SelectTrigger>
                            <SelectContent>
                                <SelectItem value="English">English</SelectItem>
                                <SelectItem value="Indonesia">Indonesia</SelectItem>
                            </SelectContent>
                        </Select>
                    </div>

                     <div class="space-y-2">
                        <Label>Timezone</Label>
                         <Select v-model="formData.timezone">
                            <SelectTrigger class="w-full">
                                <SelectValue placeholder="Select Timezone" />
                            </SelectTrigger>
                            <SelectContent>
                                <SelectItem value="Asia/Jakarta (GMT+7)">Asia/Jakarta (GMT+7)</SelectItem>
                                <SelectItem value="Asia/Makassar (GMT+8)">Asia/Makassar (GMT+8)</SelectItem>
                                <SelectItem value="Asia/Jayapura (GMT+9)">Asia/Jayapura (GMT+9)</SelectItem>
                            </SelectContent>
                        </Select>
                    </div>

                     <div class="space-y-2">
                        <Label>Currency</Label>
                         <Select v-model="formData.currency">
                            <SelectTrigger class="w-full">
                                <SelectValue placeholder="Select Currency" />
                            </SelectTrigger>
                            <SelectContent>
                                <SelectItem value="IDR">IDR (Rupiah)</SelectItem>
                                <SelectItem value="USD">USD (Dollar)</SelectItem>
                            </SelectContent>
                        </Select>
                    </div>

                     <div class="space-y-2">
                        <Label>Show Decimal</Label>
                         <Select v-model="formData.showDecimal">
                            <SelectTrigger class="w-full">
                                <SelectValue placeholder="Select Option" />
                            </SelectTrigger>
                            <SelectContent>
                                <SelectItem value="Show">Show</SelectItem>
                                <SelectItem value="Hide">Hide</SelectItem>
                            </SelectContent>
                        </Select>
                    </div>
                </div>

                <div class="pt-4">
                    <Button class="w-full bg-gradient-to-r from-emerald-600 to-teal-500 hover:from-emerald-500 hover:to-teal-400 text-white">Save Changes</Button>
                </div>
            </div>

            <!-- Profile Tab -->
            <div v-if="activeTab === 'profile'" class="space-y-6">
                 <div class="space-y-4">
                     <div class="grid w-full items-center gap-1.5">
                        <Label for="name">Full Name</Label>
                        <Input id="name" v-model="profileForm.name" placeholder="Nama Lengkap" :class="errors.profile.name ? 'border-red-500 ring-1 ring-red-500' : ''" />
                        <span v-if="errors.profile.name" class="text-xs text-red-500 font-medium">Nama lengkap wajib diisi</span>
                     </div>
                      <div class="grid w-full items-center gap-1.5">
                        <Label for="email">Email</Label>
                        <Input id="email" type="email" v-model="profileForm.email" placeholder="email@example.com" :class="errors.profile.email ? 'border-red-500 ring-1 ring-red-500' : ''" />
                        <span v-if="errors.profile.email" class="text-xs text-red-500 font-medium">Email wajib diisi</span>
                     </div>
                 </div>
                 <Button class="w-full bg-gradient-to-r from-emerald-600 to-teal-500 text-white hover:from-emerald-500 hover:to-teal-400" @click="handleUpdateProfile" :disabled="isLoading">
                    <Loader2 v-if="isLoading" class="w-4 h-4 mr-2 animate-spin" />
                    Save Profile
                 </Button>
            </div>

            <!-- Password Tab -->
             <div v-if="activeTab === 'password'" class="space-y-6">
                 <div class="space-y-4">
                      <div class="grid w-full items-center gap-1.5">
                        <Label for="new_pass">New Password</Label>
                        <div class="relative">
                            <Input id="new_pass" v-model="passwordForm.new_password" :type="showPassword.new ? 'text' : 'password'" placeholder="Xyz•••••" :class="errors.password.new || errors.password.match ? 'border-red-500 ring-1 ring-red-500' : ''" />
                            <button type="button" @click="showPassword.new = !showPassword.new" class="absolute right-3 top-1/2 -translate-y-1/2 text-muted-foreground hover:text-foreground">
                                <Eye v-if="!showPassword.new" class="w-4 h-4" />
                                <EyeOff v-else class="w-4 h-4" />
                            </button>
                        </div>
                        <span v-if="errors.password.new" class="text-xs text-red-500 font-medium">Password baru wajib diisi</span>
                     </div>
                      <div class="grid w-full items-center gap-1.5">
                        <Label for="confirm_pass">Confirm Password</Label>
                        <div class="relative">
                            <Input id="confirm_pass" v-model="passwordForm.confirm_password" :type="showPassword.confirm ? 'text' : 'password'" placeholder="Xyz•••••" :class="errors.password.confirm || errors.password.match ? 'border-red-500 ring-1 ring-red-500' : ''" />
                             <button type="button" @click="showPassword.confirm = !showPassword.confirm" class="absolute right-3 top-1/2 -translate-y-1/2 text-muted-foreground hover:text-foreground">
                                <Eye v-if="!showPassword.confirm" class="w-4 h-4" />
                                <EyeOff v-else class="w-4 h-4" />
                            </button>
                        </div>
                        <span v-if="errors.password.confirm" class="text-xs text-red-500 font-medium">Konfirmasi password wajib diisi</span>
                        <span v-else-if="errors.password.match" class="text-xs text-red-500 font-medium">Password tidak cocok</span>
                     </div>
                 </div>
                 <Button class="w-full bg-gradient-to-r from-emerald-600 to-teal-500 text-white hover:from-emerald-500 hover:to-teal-400" @click="handleUpdatePassword" :disabled="isLoading">
                    <Loader2 v-if="isLoading" class="w-4 h-4 mr-2 animate-spin" />
                    Update Password
                 </Button>
            </div>

             <!-- WhatsApp Tab -->
             <div v-if="activeTab === 'whatsapp'" class="space-y-6">
                 <div class="p-6 border border-emerald-200 bg-emerald-50 dark:bg-emerald-950/30 rounded-xl text-center space-y-4">
                     <h3 class="text-lg font-bold text-emerald-800 dark:text-emerald-100">WhatsApp Integration</h3>
                     <p class="text-sm text-emerald-600 dark:text-emerald-300">Connect your account to receive daily reports and alerts via WhatsApp.</p>
                     <Button class="bg-emerald-600 hover:bg-emerald-700 text-white">Connect WhatsApp</Button>
                 </div>
            </div>

        </CardContent>
    </Card>

  </div>
</template>
