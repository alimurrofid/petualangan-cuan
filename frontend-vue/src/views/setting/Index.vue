<script setup lang="ts">
import { ref, computed, onMounted, watch } from "vue";
import { useRoute } from "vue-router";
import { useAuthStore } from "@/stores/auth";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Card, CardContent } from "@/components/ui/card";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Eye, EyeOff, Loader2, Smartphone, CheckCircle2, XCircle, LinkIcon } from "lucide-vue-next";
import { useSwal } from "@/composables/useSwal";

const route = useRoute();
const authStore = useAuthStore();
const activeTab = ref("profile");
const isLoading = ref(false);
const swal = useSwal();

const tabs = [
    { id: "profile", label: "Profil" },
    { id: "currency", label: "Bahasa & Mata Uang" },
    { id: "password", label: "Ganti Kata Sandi" },
    { id: "whatsapp", label: "Integrasi WhatsApp" },
];

const formData = ref({
    language: "Indonesia",
    timezone: "Asia/Jakarta (GMT+7)",
    currency: "IDR",
    showDecimal: "Hide"
});

const profileForm = ref({
    name: "",
    email: ""
});

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
        // Sync phone ke whatsappForm
        whatsappForm.value.phone = authStore.user.phone || "";
    }
};

// ─── WhatsApp Integration ─────────────────────────────────────────────────────
const whatsappForm = ref({ phone: "" });
const whatsappError = ref("");
const isWhatsappLoading = ref(false);

const isPhoneConnected = computed(() => !!authStore.user?.phone);

const formatPhoneDisplay = (phone: string) => {
    // "628xxx" → "+62 8xx-xxxx-xxxx"
    if (!phone) return "";
    const num = phone.startsWith("62") ? phone.slice(2) : phone;
    return `+62 ${num.slice(0, 3)}-${num.slice(3, 7)}-${num.slice(7)}`;
};

const handleConnectWhatsApp = async () => {
    whatsappError.value = "";
    const raw = whatsappForm.value.phone.replace(/\D/g, "");
    const phone = raw.startsWith("0") ? "62" + raw.slice(1) : raw.startsWith("62") ? raw : "62" + raw;

    if (phone.length < 10 || phone.length > 15) {
        whatsappError.value = "Nomor HP tidak valid. Contoh: 08123456789 atau 628123456789";
        return;
    }

    isWhatsappLoading.value = true;
    try {
        await authStore.updateProfile({
            name: authStore.user?.name,
            email: authStore.user?.email,
            phone
        });
        whatsappForm.value.phone = phone;
        swal.success("Berhasil!", `WhatsApp <b>${formatPhoneDisplay(phone)}</b> berhasil dihubungkan.`);
    } catch (error: any) {
        swal.error("Gagal", error.response?.data?.error || "Gagal menghubungkan WhatsApp");
    } finally {
        isWhatsappLoading.value = false;
    }
};

const handleDisconnectWhatsApp = async () => {
    const result = await swal.fire({
        icon: "warning",
        title: "Putuskan WhatsApp?",
        text: "Bot AI tidak akan bisa membalas pesan WhatsApp Anda setelah ini.",
        showCancelButton: true,
        confirmButtonText: "Ya, Putuskan",
        cancelButtonText: "Batal",
        confirmButtonColor: "#EF4444",
    });
    if (!result.isConfirmed) return;

    isWhatsappLoading.value = true;
    try {
        await authStore.updateProfile({
            name: authStore.user?.name,
            email: authStore.user?.email,
            phone: ""
        });
        whatsappForm.value.phone = "";
        swal.success("Berhasil", "WhatsApp berhasil diputus dari akun Anda.");
    } catch (error: any) {
        swal.error("Gagal", error.response?.data?.error || "Gagal memutus WhatsApp");
    } finally {
        isWhatsappLoading.value = false;
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
    errors.value.password.new = !passwordForm.value.new_password;
    errors.value.password.confirm = !passwordForm.value.confirm_password;
    errors.value.password.match = false;

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
        <div
            class="flex flex-col md:flex-row md:items-center justify-between gap-4 bg-card p-2 rounded-xl border border-border/50 shadow-sm">
            <div class="flex items-center gap-1 overflow-x-auto no-scrollbar">
                <button v-for="tab in tabs" :key="tab.id" @click="activeTab = tab.id" :class="[
                    'px-4 py-2 rounded-lg text-sm font-medium transition-colors whitespace-nowrap',
                    activeTab === tab.id
                        ? 'bg-gradient-to-r from-emerald-600 to-teal-500 text-white shadow-sm'
                        : 'text-muted-foreground hover:bg-muted start-hover'
                ]">
                    {{ tab.label }}
                </button>
            </div>
            <Button variant="destructive" size="sm" class="hidden md:flex">Hapus Akun</Button>
        </div>

        <Card class="border-border/60 shadow-sm overflow-hidden">
            <CardContent class="p-6">

                <!-- Currency Tab -->
                <div v-if="activeTab === 'currency'" class="space-y-6">
                    <!-- ... Existing Currency Content ... -->
                    <div class="space-y-4">
                        <div class="space-y-2">
                            <Label>Bahasa</Label>
                            <Select v-model="formData.language">
                                <SelectTrigger class="w-full">
                                    <SelectValue placeholder="Pilih Bahasa" />
                                </SelectTrigger>
                                <SelectContent>
                                    <SelectItem value="English">English</SelectItem>
                                    <SelectItem value="Indonesia">Indonesia</SelectItem>
                                </SelectContent>
                            </Select>
                        </div>

                        <div class="space-y-2">
                            <Label>Zona Waktu</Label>
                            <Select v-model="formData.timezone">
                                <SelectTrigger class="w-full">
                                    <SelectValue placeholder="Pilih Zona Waktu" />
                                </SelectTrigger>
                                <SelectContent>
                                    <SelectItem value="Asia/Jakarta (GMT+7)">Asia/Jakarta (GMT+7)</SelectItem>
                                    <SelectItem value="Asia/Makassar (GMT+8)">Asia/Makassar (GMT+8)</SelectItem>
                                    <SelectItem value="Asia/Jayapura (GMT+9)">Asia/Jayapura (GMT+9)</SelectItem>
                                </SelectContent>
                            </Select>
                        </div>

                        <div class="space-y-2">
                            <Label>Mata Uang</Label>
                            <Select v-model="formData.currency">
                                <SelectTrigger class="w-full">
                                    <SelectValue placeholder="Pilih Mata Uang" />
                                </SelectTrigger>
                                <SelectContent>
                                    <SelectItem value="IDR">IDR (Rupiah)</SelectItem>
                                    <SelectItem value="USD">USD (Dollar)</SelectItem>
                                </SelectContent>
                            </Select>
                        </div>

                        <div class="space-y-2">
                            <Label>Tampilkan Desimal</Label>
                            <Select v-model="formData.showDecimal">
                                <SelectTrigger class="w-full">
                                    <SelectValue placeholder="Pilih Opsi" />
                                </SelectTrigger>
                                <SelectContent>
                                    <SelectItem value="Show">Tampilkan</SelectItem>
                                    <SelectItem value="Hide">Sembunyikan</SelectItem>
                                </SelectContent>
                            </Select>
                        </div>
                    </div>

                    <div class="pt-4">
                        <Button
                            class="w-full bg-gradient-to-r from-emerald-600 to-teal-500 hover:from-emerald-500 hover:to-teal-400 text-white">Simpan
                            Perubahan</Button>
                    </div>
                </div>

                <!-- Profile Tab -->
                <div v-if="activeTab === 'profile'" class="space-y-6">
                    <div class="space-y-4">
                        <div class="grid w-full items-center gap-1.5">
                            <Label for="name">Nama Lengkap</Label>
                            <Input id="name" v-model="profileForm.name" placeholder="Nama Lengkap"
                                :class="errors.profile.name ? 'border-red-500 ring-1 ring-red-500' : ''" />
                            <span v-if="errors.profile.name" class="text-xs text-red-500 font-medium">Nama lengkap wajib
                                diisi</span>
                        </div>
                        <div class="grid w-full items-center gap-1.5">
                            <Label for="email">Email</Label>
                            <Input id="email" type="email" v-model="profileForm.email" placeholder="email@example.com"
                                :class="errors.profile.email ? 'border-red-500 ring-1 ring-red-500' : ''" />
                            <span v-if="errors.profile.email" class="text-xs text-red-500 font-medium">Email wajib
                                diisi</span>
                        </div>
                    </div>
                    <Button
                        class="w-full bg-gradient-to-r from-emerald-600 to-teal-500 text-white hover:from-emerald-500 hover:to-teal-400"
                        @click="handleUpdateProfile" :disabled="isLoading">
                        <Loader2 v-if="isLoading" class="w-4 h-4 mr-2 animate-spin" />
                        Simpan Profil
                    </Button>
                </div>

                <!-- Password Tab -->
                <div v-if="activeTab === 'password'" class="space-y-6">
                    <div class="space-y-4">
                        <div class="grid w-full items-center gap-1.5">
                            <Label for="new_pass">Kata Sandi Baru</Label>
                            <div class="relative">
                                <Input id="new_pass" v-model="passwordForm.new_password"
                                    :type="showPassword.new ? 'text' : 'password'" placeholder="Xyz•••••"
                                    :class="errors.password.new || errors.password.match ? 'border-red-500 ring-1 ring-red-500' : ''" />
                                <button type="button" @click="showPassword.new = !showPassword.new"
                                    class="absolute right-3 top-1/2 -translate-y-1/2 text-muted-foreground hover:text-foreground">
                                    <Eye v-if="!showPassword.new" class="w-4 h-4" />
                                    <EyeOff v-else class="w-4 h-4" />
                                </button>
                            </div>
                            <span v-if="errors.password.new" class="text-xs text-red-500 font-medium">Password baru
                                wajib diisi</span>
                        </div>
                        <div class="grid w-full items-center gap-1.5">
                            <Label for="confirm_pass">Konfirmasi Kata Sandi</Label>
                            <div class="relative">
                                <Input id="confirm_pass" v-model="passwordForm.confirm_password"
                                    :type="showPassword.confirm ? 'text' : 'password'" placeholder="Xyz•••••"
                                    :class="errors.password.confirm || errors.password.match ? 'border-red-500 ring-1 ring-red-500' : ''" />
                                <button type="button" @click="showPassword.confirm = !showPassword.confirm"
                                    class="absolute right-3 top-1/2 -translate-y-1/2 text-muted-foreground hover:text-foreground">
                                    <Eye v-if="!showPassword.confirm" class="w-4 h-4" />
                                    <EyeOff v-else class="w-4 h-4" />
                                </button>
                            </div>
                            <span v-if="errors.password.confirm" class="text-xs text-red-500 font-medium">Konfirmasi
                                password wajib diisi</span>
                            <span v-else-if="errors.password.match" class="text-xs text-red-500 font-medium">Password
                                tidak cocok</span>
                        </div>
                    </div>
                    <Button
                        class="w-full bg-gradient-to-r from-emerald-600 to-teal-500 text-white hover:from-emerald-500 hover:to-teal-400"
                        @click="handleUpdatePassword" :disabled="isLoading">
                        <Loader2 v-if="isLoading" class="w-4 h-4 mr-2 animate-spin" />
                        Perbarui Kata Sandi
                    </Button>
                </div>

                <!-- WhatsApp Tab -->
                <div v-if="activeTab === 'whatsapp'" class="space-y-6">

                    <!-- Status Card -->
                    <div :class="[
                        'flex items-center gap-4 p-4 rounded-xl border',
                        isPhoneConnected
                            ? 'bg-emerald-50 border-emerald-200 dark:bg-emerald-950/30 dark:border-emerald-800'
                            : 'bg-muted/40 border-border'
                    ]">
                        <div
                            :class="['p-3 rounded-full', isPhoneConnected ? 'bg-emerald-100 dark:bg-emerald-900' : 'bg-muted']">
                            <CheckCircle2 v-if="isPhoneConnected" class="w-6 h-6 text-emerald-600" />
                            <XCircle v-else class="w-6 h-6 text-muted-foreground" />
                        </div>
                        <div class="flex-1 min-w-0">
                            <p class="font-semibold text-sm">
                                {{ isPhoneConnected ? 'WhatsApp Terhubung' : 'WhatsApp Belum Terhubung' }}
                            </p>
                            <p class="text-xs text-muted-foreground mt-0.5">
                                {{ isPhoneConnected
                                    ? formatPhoneDisplay(authStore.user?.phone)
                                    : 'Hubungkan nomor WhatsApp Anda agar bot AI bisa membalas pesan.'
                                }}
                            </p>
                        </div>
                        <Button v-if="isPhoneConnected" variant="outline" size="sm"
                            class="text-red-500 border-red-200 hover:bg-red-50 hover:text-red-600 shrink-0"
                            @click="handleDisconnectWhatsApp" :disabled="isWhatsappLoading">
                            <Loader2 v-if="isWhatsappLoading" class="w-3 h-3 mr-1 animate-spin" />
                            Putuskan
                        </Button>
                    </div>

                    <!-- Form input nomor -->
                    <div class="space-y-4">
                        <div class="space-y-1.5">
                            <Label for="wa-phone">Nomor WhatsApp</Label>
                            <div class="relative">
                                <Smartphone
                                    class="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-muted-foreground" />
                                <Input id="wa-phone" v-model="whatsappForm.phone"
                                    placeholder="08123456789 atau 628123456789" class="pl-10"
                                    :class="whatsappError ? 'border-red-500 ring-1 ring-red-500' : ''"
                                    @keyup.enter="handleConnectWhatsApp" />
                            </div>
                            <span v-if="whatsappError" class="text-xs text-red-500 font-medium">{{ whatsappError
                            }}</span>
                            <p class="text-xs text-muted-foreground">
                                Masukkan nomor HP yang terdaftar di WhatsApp, tanpa tanda baca.<br>
                                Contoh: <code class="bg-muted px-1 rounded">081234567890</code> atau <code
                                    class="bg-muted px-1 rounded">6281234567890</code>
                            </p>
                        </div>

                        <Button
                            class="w-full bg-gradient-to-r from-emerald-600 to-teal-500 hover:from-emerald-500 hover:to-teal-400 text-white"
                            @click="handleConnectWhatsApp" :disabled="isWhatsappLoading">
                            <Loader2 v-if="isWhatsappLoading" class="w-4 h-4 mr-2 animate-spin" />
                            <LinkIcon v-else class="w-4 h-4 mr-2" />
                            {{ isPhoneConnected ? 'Perbarui Nomor WhatsApp' : 'Hubungkan WhatsApp' }}
                        </Button>
                    </div>

                    <!-- Info box -->
                    <div
                        class="p-4 rounded-xl bg-blue-50 border border-blue-100 dark:bg-blue-950/20 dark:border-blue-900 text-sm text-blue-700 dark:text-blue-300 space-y-1">
                        <p class="font-semibold">💡 Cara kerja integrasi WhatsApp:</p>
                        <ol class="list-decimal list-inside space-y-1 text-xs">
                            <li>Hubungkan nomor WA Anda di halaman ini</li>
                            <li>Kirim pesan ke nomor WA bot (tanyakan ke admin)</li>
                            <li>Bot AI akan membalas & mencatat transaksi otomatis</li>
                            <li>Anda bisa kirim pesan suara atau foto struk</li>
                        </ol>
                    </div>

                </div>

            </CardContent>
        </Card>

    </div>
</template>
