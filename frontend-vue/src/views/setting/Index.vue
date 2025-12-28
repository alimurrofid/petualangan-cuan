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

const route = useRoute();
const authStore = useAuthStore();
const activeTab = ref("profile");
const isLoading = ref(false);

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
    old_password: "",
    new_password: "",
    confirm_password: ""
});

const showPassword = ref({
    old: false,
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

const handleUpdateProfile = async () => {
    if (!profileForm.value.name || !profileForm.value.email) return alert("Nama dan Email wajib diisi");
    
    isLoading.value = true;
    try {
        await authStore.updateProfile(profileForm.value);
        alert("Profil berhasil diperbarui!");
    } catch (error: any) {
        alert(error.response?.data?.error || "Gagal memperbarui profil");
    } finally {
        isLoading.value = false;
    }
};

const handleUpdatePassword = async () => {
    if (!passwordForm.value.old_password || !passwordForm.value.new_password) return alert("Mohon lengkapi data");
    if (passwordForm.value.new_password !== passwordForm.value.confirm_password) return alert("Konfirmasi password tidak cocok");
    
    isLoading.value = true;
    try {
        await authStore.changePassword({
            old_password: passwordForm.value.old_password,
            new_password: passwordForm.value.new_password
        });
        alert("Password berhasil diperbarui! Silakan login ulang.");
        authStore.logout();
    } catch (error: any) {
        alert(error.response?.data?.error || "Gagal memperbarui password");
    } finally {
        isLoading.value = false;
        passwordForm.value = { old_password: "", new_password: "", confirm_password: "" };
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
                        ? 'bg-foreground text-background shadow-sm' 
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
                    <Button class="w-full">Save Changes</Button>
                </div>
            </div>

            <!-- Profile Tab -->
            <div v-if="activeTab === 'profile'" class="space-y-6">
                 <div class="space-y-4">
                     <div class="grid w-full items-center gap-1.5">
                        <Label for="name">Full Name</Label>
                        <Input id="name" v-model="profileForm.name" placeholder="Nama Lengkap" />
                     </div>
                      <div class="grid w-full items-center gap-1.5">
                        <Label for="email">Email</Label>
                        <Input id="email" type="email" v-model="profileForm.email" placeholder="email@example.com" />
                     </div>
                 </div>
                 <Button class="w-full" @click="handleUpdateProfile" :disabled="isLoading">
                    <Loader2 v-if="isLoading" class="w-4 h-4 mr-2 animate-spin" />
                    Save Profile
                 </Button>
            </div>

            <!-- Password Tab -->
             <div v-if="activeTab === 'password'" class="space-y-6">
                 <div class="space-y-4">
                      
                      <div class="grid w-full items-center gap-1.5">
                        <Label for="old_pass">Old Password</Label>
                        <div class="relative">
                            <Input id="old_pass" v-model="passwordForm.old_password" :type="showPassword.old ? 'text' : 'password'" placeholder="••••••••" />
                            <button type="button" @click="showPassword.old = !showPassword.old" class="absolute right-3 top-1/2 -translate-y-1/2 text-muted-foreground hover:text-foreground">
                                <Eye v-if="!showPassword.old" class="w-4 h-4" />
                                <EyeOff v-else class="w-4 h-4" />
                            </button>
                        </div>
                     </div>

                      <div class="grid w-full items-center gap-1.5 grayscale opacity-50"><div class="h-px bg-border my-2"></div></div>

                      <div class="grid w-full items-center gap-1.5">
                        <Label for="new_pass">New Password</Label>
                        <div class="relative">
                            <Input id="new_pass" v-model="passwordForm.new_password" :type="showPassword.new ? 'text' : 'password'" placeholder="••••••••" />
                            <button type="button" @click="showPassword.new = !showPassword.new" class="absolute right-3 top-1/2 -translate-y-1/2 text-muted-foreground hover:text-foreground">
                                <Eye v-if="!showPassword.new" class="w-4 h-4" />
                                <EyeOff v-else class="w-4 h-4" />
                            </button>
                        </div>
                     </div>
                      <div class="grid w-full items-center gap-1.5">
                        <Label for="confirm_pass">Confirm Password</Label>
                        <div class="relative">
                            <Input id="confirm_pass" v-model="passwordForm.confirm_password" :type="showPassword.confirm ? 'text' : 'password'" placeholder="••••••••" />
                             <button type="button" @click="showPassword.confirm = !showPassword.confirm" class="absolute right-3 top-1/2 -translate-y-1/2 text-muted-foreground hover:text-foreground">
                                <Eye v-if="!showPassword.confirm" class="w-4 h-4" />
                                <EyeOff v-else class="w-4 h-4" />
                            </button>
                        </div>
                     </div>
                 </div>
                 <Button class="w-full" @click="handleUpdatePassword" :disabled="isLoading">
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
