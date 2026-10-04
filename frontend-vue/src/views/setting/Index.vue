<script setup lang="ts">
import { ref, computed, onMounted, watch } from "vue";
import { useRoute } from "vue-router";
import { useAuthStore } from "@/stores/auth";
import { useSettingsStore, type SupportedLanguage } from "@/stores/settings";
import { useI18n } from "vue-i18n";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Card, CardContent } from "@/components/ui/card";
import { Switch } from "@/components/ui/switch";
import { Eye, EyeOff, Loader2, Smartphone, CheckCircle2, XCircle, LinkIcon, Globe, Coins } from "lucide-vue-next";
import { useSwal } from "@/composables/useSwal";
import { formatCurrency } from "@/lib/utils";

const route = useRoute();
const authStore = useAuthStore();
const settingsStore = useSettingsStore();
const { t } = useI18n();

const activeTab = ref("profile");
const isLoading = ref(false);
const isSavingPreferences = ref(false);
const swal = useSwal();

const tabs = computed(() => [
    { id: "profile", label: t('settings.tabs.profile') },
    { id: "currency", label: t('settings.tabs.format') },
    { id: "password", label: t('settings.tabs.password') },
    { id: "whatsapp", label: t('settings.tabs.whatsapp') },
]);

const formData = ref({
    language: settingsStore.language as SupportedLanguage,
    showDecimal: settingsStore.showDecimal,
});

watch(() => [settingsStore.language, settingsStore.showDecimal], () => {
    formData.value.language = settingsStore.language;
    formData.value.showDecimal = settingsStore.showDecimal;
});

const handleSavePreferences = () => {
    isSavingPreferences.value = true;
    try {
        settingsStore.saveSettings({
            language: formData.value.language,
            showDecimal: formData.value.showDecimal,
        });
        swal.success(t('common.success'), t('settings.savedSuccess'));
    } catch (e: any) {
        swal.error(t('common.failed'), e?.message || "Gagal menyimpan pengaturan");
    } finally {
        setTimeout(() => {
            isSavingPreferences.value = false;
        }, 300);
    }
};

const profileForm = ref({
    name: "",
    email: "",
    payday: 1 as number
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
    if (tab && tabs.value.some((t: { id: string; label: string }) => t.id === tab)) {
        activeTab.value = tab;
    }
};

const initProfile = () => {
    if (authStore.user) {
        profileForm.value.name = authStore.user.name;
        profileForm.value.email = authStore.user.email;
        profileForm.value.payday = authStore.user.payday || 1;
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
        await authStore.updateProfile({
            ...profileForm.value,
            payday: profileForm.value.payday
        });
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
            <Button variant="destructive" size="sm" class="hidden md:flex">{{ t('settings.deleteAccount') }}</Button>
        </div>

        <Card class="border-border/60 shadow-sm overflow-hidden">
            <CardContent class="p-6">

                <!-- Bahasa & Format Tab -->
                <div v-if="activeTab === 'currency'" class="space-y-6">
                    <div class="space-y-4">
                        <!-- Toggle Bahasa -->
                        <div class="flex flex-col sm:flex-row sm:items-center justify-between p-5 rounded-2xl bg-card border border-border/70 gap-4 shadow-xs">
                            <div class="flex items-start gap-3.5">
                                <div class="h-10 w-10 rounded-xl bg-blue-500/10 text-blue-600 dark:text-blue-400 flex items-center justify-center shrink-0 mt-0.5 border border-blue-500/15">
                                    <Globe class="w-5 h-5" />
                                </div>
                                <div>
                                    <h4 class="font-semibold text-base text-foreground">{{ t('settings.languageTitle') }}</h4>
                                    <p class="text-xs text-muted-foreground mt-0.5 leading-relaxed">{{ t('settings.languageDesc') }}</p>
                                </div>
                            </div>
                            
                            <!-- Segmented Pill Toggle with Smooth Sliding Indicator -->
                            <div class="relative inline-flex p-1 rounded-full bg-muted/70 dark:bg-zinc-800/90 border border-border/70 self-start sm:self-center shrink-0 shadow-xs">
                                <!-- Smooth Animated Sliding Pill -->
                                <div 
                                    class="absolute top-1 bottom-1 left-1 w-[calc(50%-4px)] rounded-full bg-background dark:bg-zinc-900 shadow-sm border border-border/70 transition-transform duration-300 ease-[cubic-bezier(0.16,1,0.3,1)] pointer-events-none"
                                    :style="{
                                        transform: formData.language === 'en' ? 'translateX(100%)' : 'translateX(0%)'
                                    }"
                                ></div>

                                <button 
                                    type="button" 
                                    @click="formData.language = 'id'" 
                                    class="relative z-10 flex items-center justify-center gap-2 px-4 py-1.5 rounded-full text-xs transition-colors duration-200 cursor-pointer select-none min-w-[105px]"
                                    :class="formData.language === 'id' 
                                        ? 'text-emerald-600 dark:text-emerald-400 font-bold' 
                                        : 'text-muted-foreground hover:text-foreground font-medium'"
                                >
                                    <svg viewBox="0 0 32 32" class="w-4 h-4 rounded-full overflow-hidden shrink-0 shadow-xs ring-1 ring-black/10 dark:ring-white/15">
                                        <rect width="32" height="16" fill="#E70011"/>
                                        <rect y="16" width="32" height="16" fill="#FFFFFF"/>
                                    </svg>
                                    <span>Indonesia</span>
                                </button>
                                <button 
                                    type="button" 
                                    @click="formData.language = 'en'" 
                                    class="relative z-10 flex items-center justify-center gap-2 px-4 py-1.5 rounded-full text-xs transition-colors duration-200 cursor-pointer select-none min-w-[105px]"
                                    :class="formData.language === 'en' 
                                        ? 'text-emerald-600 dark:text-emerald-400 font-bold' 
                                        : 'text-muted-foreground hover:text-foreground font-medium'"
                                >
                                    <svg viewBox="0 0 32 32" class="w-4 h-4 rounded-full overflow-hidden shrink-0 shadow-xs ring-1 ring-black/10 dark:ring-white/15">
                                        <rect width="32" height="32" fill="#012169"/>
                                        <path d="M0 0 L32 32 M32 0 L0 32" stroke="#FFFFFF" stroke-width="6"/>
                                        <path d="M0 0 L32 32 M32 0 L0 32" stroke="#C8102E" stroke-width="3"/>
                                        <path d="M16 0 V32 M0 16 H32" stroke="#FFFFFF" stroke-width="10"/>
                                        <path d="M16 0 V32 M0 16 H32" stroke="#C8102E" stroke-width="6"/>
                                    </svg>
                                    <span>English</span>
                                </button>
                            </div>
                        </div>

                        <!-- Toggle Desimal -->
                        <div class="flex flex-col sm:flex-row sm:items-center justify-between p-5 rounded-2xl bg-card border border-border/70 gap-4 shadow-xs">
                            <div class="flex items-start gap-3.5">
                                <div class="h-10 w-10 rounded-xl bg-emerald-500/10 text-emerald-600 dark:text-emerald-400 flex items-center justify-center shrink-0 mt-0.5 border border-emerald-500/15">
                                    <Coins class="w-5 h-5" />
                                </div>
                                <div class="space-y-1">
                                    <h4 class="font-semibold text-base text-foreground">{{ t('settings.decimalTitle') }}</h4>
                                    <p class="text-xs text-muted-foreground leading-relaxed">
                                        {{ formData.showDecimal ? t('settings.decimalDescOn') : t('settings.decimalDescOff') }}
                                    </p>
                                </div>
                            </div>
                            
                            <!-- Switch Toggle & Live Preview -->
                            <div class="flex items-center gap-3.5 self-start sm:self-center shrink-0">
                                <div class="flex items-center gap-2 px-3 py-1.5 rounded-lg bg-muted/60 dark:bg-muted/30 border border-border/60 text-xs">
                                    <span class="text-muted-foreground text-[11px] font-medium">Contoh:</span>
                                    <span class="font-bold font-mono text-emerald-600 dark:text-emerald-400">
                                        {{ formatCurrency(1500000, { showDecimal: formData.showDecimal }) }}
                                    </span>
                                </div>
                                <Switch 
                                    v-model="formData.showDecimal" 
                                    id="toggle-decimal"
                                    aria-label="Tampilkan Desimal"
                                />
                            </div>
                        </div>
                    </div>

                    <div class="pt-2">
                        <Button
                            @click="handleSavePreferences"
                            :disabled="isSavingPreferences"
                            class="w-full bg-gradient-to-r from-emerald-600 to-teal-500 hover:from-emerald-500 hover:to-teal-400 text-white shadow-md font-semibold h-11"
                        >
                            <Loader2 v-if="isSavingPreferences" class="w-4 h-4 mr-2 animate-spin" />
                            {{ t('settings.savePreferences') }}
                        </Button>
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

                        <!-- Payday Setting -->
                        <div class="grid w-full items-center gap-1.5">
                            <Label for="payday">Tanggal Gajian</Label>
                            <div class="flex items-center gap-3">
                                <Input id="payday" type="number" min="1" max="31" v-model.number="profileForm.payday"
                                    class="w-24 text-center font-bold text-lg" placeholder="1" />
                                <span class="text-sm text-muted-foreground">setiap bulan</span>
                            </div>
                            <p class="text-xs text-muted-foreground">
                                Atur tanggal gajian Anda (<b>1–31</b>). Semua laporan "Bulanan" dihitung dari tanggal
                                ini.
                                Di bulan yang lebih pendek, sistem otomatis pakai hari terakhir (mis. tgl 31 di Feb →
                                28/29).
                                Contoh: gajian tgl 25 → siklus <b>25 Feb – 24 Mar</b>.
                            </p>
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
