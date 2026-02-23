<script setup lang="ts">
import { ref, onMounted, computed, watch } from "vue";
import heic2any from "heic2any";
import { format, parseISO } from "date-fns";
import { useWalletStore } from "@/stores/wallet";
import { useCategoryStore } from "@/stores/category";
import { useTransactionStore } from "@/stores/transaction";
import { useWishlistStore } from "@/stores/wishlist";
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogFooter, DialogDescription } from "@/components/ui/dialog";
import { Tabs, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Label } from "@/components/ui/label";
import { Input } from "@/components/ui/input";
import { Button } from "@/components/ui/button";
import SearchableSelect from "@/components/ui/searchable-select/SearchableSelect.vue";
import { getEmoji, getIconComponent } from "@/lib/icons";
import { Calendar } from "lucide-vue-next";
import { parseCurrencyInput, formatCurrencyInput, formatCurrencyLive, formatDate } from "@/lib/utils";
import { useSwal } from "@/composables/useSwal";

const props = defineProps<{
    open: boolean;
    transactionToEdit?: any | null;
    initialData?: {
        amount?: number;
        category_id?: number;
        description?: string;
    } | null;
    wishlistItemId?: number | null;
    savingGoalTarget?: any | null;
}>();

const emit = defineEmits<{
    (e: "update:open", value: boolean): void;
    (e: "save", data: any): void;
}>();

const walletStore = useWalletStore();
const categoryStore = useCategoryStore();
const transactionStore = useTransactionStore();
const wishlistStore = useWishlistStore();
const swal = useSwal();
const baseUrl = import.meta.env.VITE_API_BASE_URL;

const isSubmitting = ref(false);

const activeTab = ref<"expense" | "income" | "transfer" | "saving">("expense");
const date = ref(format(new Date(), "yyyy-MM-dd"));
const amount = ref("");
const amountDisplay = ref("");
const selectedWallet = ref("");
const toWallet = ref("");
const transferFee = ref("");
const transferFeeDisplay = ref("");
const selectedCategory = ref("");
const description = ref("");
const file = ref<File | null>(null);
const existingAttachment = ref("");

const isProcessingFile = ref(false);

const handleFileChange = async (event: Event) => {
    const target = event.target as HTMLInputElement;
    if (target.files && target.files.length > 0) {
        const selectedFile = target.files[0];
        if (!selectedFile) return;

        if (selectedFile.size > 5 * 1024 * 1024) {
            await swal.error("Gagal", "Ukuran file maksimal 5MB");
            target.value = "";
            file.value = null;
            return;
        }

        const fileType = selectedFile.type.toLowerCase();
        const fileName = selectedFile.name.toLowerCase();

        if (fileType === "image/heic" || fileName.endsWith(".heic")) {
            isProcessingFile.value = true;
            try {
                const convertedBlob = await heic2any({
                    blob: selectedFile,
                    toType: "image/jpeg",
                    quality: 0.8
                });

                const finalBlob = Array.isArray(convertedBlob) ? convertedBlob[0] : convertedBlob;

                if (!finalBlob) {
                    throw new Error("Conversion resulted in empty blob");
                }

                const newName = fileName.replace(/\.heic$/, ".jpg");
                file.value = new File([finalBlob], newName, { type: "image/jpeg" });
            } catch (e) {
                console.error("HEIC Conversion failed", e);
                await swal.error("Gagal", "Gagal memproses file HEIC. Silakan gunakan JPG/PNG.");
                target.value = "";
                file.value = null;
            } finally {
                isProcessingFile.value = false;
            }
        } else {
            file.value = selectedFile;
        }
    }
};

onMounted(() => {
    walletStore.fetchWallets();
    categoryStore.fetchCategories();
});
watch(() => props.transactionToEdit, (newVal) => {
    if (newVal) {
        activeTab.value = newVal.type === 'transfer_in' || newVal.type === 'transfer_out' ? 'transfer' : newVal.type;
        if (newVal.type === 'transfer_out' || newVal.type === 'transfer_in') {
            activeTab.value = 'transfer';
            if (newVal.related_transaction_id) {
                transactionStore.fetchTransaction(newVal.related_transaction_id).then(related => {
                    if (related) {
                        if (newVal.type === 'transfer_out') {
                            toWallet.value = String(related.wallet_id);
                        } else {
                            toWallet.value = String(newVal.wallet_id);
                            selectedWallet.value = String(related.wallet_id);
                        }
                    }
                });
            }
        }

        if (newVal.type === 'transfer_out') {
            selectedWallet.value = String(newVal.wallet_id);
        } else if (newVal.type === 'transfer_in') {
            if (activeTab.value === 'transfer') {
                toWallet.value = String(newVal.wallet_id);
            } else {
                selectedWallet.value = String(newVal.wallet_id);
            }
        }

        date.value = format(parseISO(newVal.date), 'yyyy-MM-dd');
        amount.value = newVal.amount.toString();
        amountDisplay.value = formatCurrencyInput(newVal.amount);
        selectedWallet.value = String(newVal.wallet_id);
        selectedCategory.value = String(newVal.category_id);
        description.value = newVal.description;
        existingAttachment.value = newVal.attachment || "";
        file.value = null;
    } else {
        // Reset defaults
        activeTab.value = "expense";
        date.value = format(new Date(), "yyyy-MM-dd");
        amount.value = "";
        amountDisplay.value = "";
        transferFee.value = "";
        transferFeeDisplay.value = "";
        selectedWallet.value = "";
        toWallet.value = "";
        selectedCategory.value = "";
        selectedCategory.value = "";
        description.value = "";
        file.value = null;
        existingAttachment.value = "";
    }
});

watch(() => props.initialData, (newVal) => {
    if (newVal && !props.transactionToEdit) {
        amount.value = newVal.amount?.toString() || "";
        amountDisplay.value = newVal.amount ? formatCurrencyInput(newVal.amount) : "";
        selectedCategory.value = newVal.category_id?.toString() || "";
        description.value = newVal.description || "";
        activeTab.value = "expense";
    }

    if (props.wishlistItemId) {
        activeTab.value = "expense";
    }
}, { immediate: true });

watch(() => props.savingGoalTarget, (newVal) => {
    if (isSubmitting.value) {
        return;
    }

    if (newVal) {
        activeTab.value = "saving";
        description.value = `Alokasi ke ${newVal.name}`;
        if (newVal.category_id) {
            selectedCategory.value = String(newVal.category_id);
        }
    } else {
        if (activeTab.value === 'saving') {
            activeTab.value = "expense";
            description.value = "";
            selectedCategory.value = "";
            amount.value = "";
            amountDisplay.value = "";
        }
    }
}, { immediate: true });

const filteredCategories = computed(() => {
    if (activeTab.value === 'saving') {
        return categoryStore.categories.filter(c => c.type === 'expense');
    }
    return categoryStore.categories.filter(c => c.type === activeTab.value);
});

const walletOptions = computed(() => walletStore.wallets.map(w => ({
    value: String(w.id),
    label: w.name,
    icon: w.icon,
    balance: w.balance,
    available_balance: w.available_balance
})));

const categoryOptions = computed(() => filteredCategories.value.map(c => ({
    value: String(c.id),
    label: c.name,
    icon: c.icon
})));

watch(() => props.open, (newVal) => {
    if (!newVal) {
        resetForm();
    }
});

const errors = ref({
    date: false,
    amount: false,
    wallet: false,
    toWallet: false,
    category: false,
});

const handleSave = async () => {
    isSubmitting.value = true;

    Object.keys(errors.value).forEach(k => (errors.value as any)[k] = false);

    errors.value.date = !date.value;
    errors.value.amount = !amount.value;

    if (activeTab.value === 'transfer') {
        errors.value.wallet = !selectedWallet.value;
        errors.value.toWallet = !toWallet.value;
    } else {
        errors.value.wallet = !selectedWallet.value;
        errors.value.category = !selectedCategory.value;
    }

    const hasError = Object.values(errors.value).some(v => v);
    if (hasError) {
        let msg = "Mohon lengkapi data berikut:";
        if (errors.value.date) msg += "<br>- Tanggal";
        if (errors.value.amount) msg += "<br>- Nominal";
        if (errors.value.wallet) msg += "<br>- " + (activeTab.value === 'transfer' ? 'Dompet Asal' : 'Dompet');
        if (errors.value.toWallet) msg += "<br>- Dompet Tujuan";
        if (errors.value.category) msg += "<br>- Kategori";

        await swal.fire({
            icon: 'error',
            title: 'Validasi Gagal',
            html: msg,
            confirmButtonColor: '#EF4444',
        });
        setTimeout(() => { isSubmitting.value = false; }, 300);
        return;
    }

    if (activeTab.value === 'transfer') {
        if (selectedWallet.value === toWallet.value) {
            await swal.error("Gagal", "Dompet asal dan tujuan tidak boleh sama");
            setTimeout(() => { isSubmitting.value = false; }, 300);
            return;
        }

        const wallet = walletStore.wallets.find(w => String(w.id) === selectedWallet.value);
        if (wallet) {
            const totalRequired = Number(amount.value) + Number(transferFee.value || 0);
            if (wallet.balance < totalRequired) {
                await swal.error("Gagal", `Saldo tidak cukup untuk Transfer + Biaya (Total: ${new Intl.NumberFormat("id-ID", { style: "currency", currency: "IDR" }).format(totalRequired)})`);
                setTimeout(() => { isSubmitting.value = false; }, 300);
                return;
            }
        }
    }

    try {
        const now = new Date();
        const [year = now.getFullYear(), month = now.getMonth() + 1, day = now.getDate()] = date.value.split('-').map(Number);
        const finalDate = new Date(year, month - 1, day, now.getHours(), now.getMinutes(), now.getSeconds());

        if (activeTab.value === 'saving' && props.savingGoalTarget) {
            const { useSavingGoalStore } = await import("@/stores/saving_goal");
            const savingStore = useSavingGoalStore();

            await savingStore.addContribution(props.savingGoalTarget.id, {
                wallet_id: Number(selectedWallet.value),
                amount: Number(amount.value),
                date: finalDate.toISOString(),
                description: description.value
            });
            swal.success('Berhasil menabung!');
            emit("update:open", false);
            emit("save", {});

            return;
        }

        if (props.transactionToEdit) {
            let finalType: string = activeTab.value;
            let finalWalletId = Number(selectedWallet.value);

            if (activeTab.value === 'transfer') {
                if (props.transactionToEdit.type === 'transfer_in') {
                    finalType = 'transfer_in';
                    finalWalletId = Number(toWallet.value);
                } else if (props.transactionToEdit.type === 'transfer_out') {
                    finalType = 'transfer_out';
                    finalWalletId = Number(selectedWallet.value);
                } else {
                    finalType = 'transfer_out';
                }
            }

            const payload = new FormData();
            payload.append('wallet_id', String(finalWalletId));
            payload.append('category_id', String(selectedCategory.value));
            payload.append('amount', String(amount.value));
            payload.append('type', finalType);
            payload.append('description', description.value || "");
            payload.append('date', format(finalDate, "yyyy-MM-dd'T'HH:mm:ssXXX"));

            if (file.value) {
                payload.append('attachment', file.value);
            }

            await transactionStore.updateTransaction(props.transactionToEdit.id, payload);
            swal.success('Transaksi berhasil diperbarui');
        } else if (activeTab.value === 'transfer') {
            await transactionStore.transfer({
                date: format(finalDate, "yyyy-MM-dd'T'HH:mm:ssXXX"),
                amount: Number(amount.value),
                transfer_fee: Number(transferFee.value || 0),
                from_wallet_id: Number(selectedWallet.value),
                to_wallet_id: Number(toWallet.value),
                description: description.value || "Transfer Antar Dompet"
            });
            swal.success("Berhasil", "Transfer berhasil dilakukan");
        } else {
            const payload = new FormData();
            payload.append('type', activeTab.value);
            payload.append('date', format(finalDate, "yyyy-MM-dd'T'HH:mm:ssXXX"));
            payload.append('amount', String(amount.value));
            payload.append('category_id', String(selectedCategory.value));
            payload.append('wallet_id', String(selectedWallet.value));
            payload.append('description', description.value || "");

            if (file.value) {
                payload.append('attachment', file.value);
            }

            await transactionStore.createTransaction(payload);
            swal.success('Transaksi berhasil disimpan');

            if (props.wishlistItemId) {
                await wishlistStore.markAsBought(props.wishlistItemId);
            }
        }

        emit("update:open", false);
        emit("save", {});
    } catch (error: any) {
        console.error("Error in handleSave:", error);

        const errMsg = error.response?.data?.error || "";
        if (errMsg.toLowerCase().includes('insufficient')) {
            return;
        }

        swal.error("Gagal", props.transactionToEdit ? "Gagal memperbarui transaksi" : "Gagal melakukan transaksi");
    } finally {
        isSubmitting.value = false;
    }
};

const resetForm = () => {
    amount.value = "";
    amountDisplay.value = "";
    transferFee.value = "";
    transferFeeDisplay.value = "";
    description.value = "";
    file.value = null;
    existingAttachment.value = "";

    if (!props.savingGoalTarget && !props.transactionToEdit) {
        activeTab.value = "expense";
    }

    if (!props.savingGoalTarget) {
        selectedCategory.value = "";
    }
};

watch(amountDisplay, (val) => {
    const formatted = formatCurrencyLive(val);
    if (formatted !== val) {
        amountDisplay.value = formatted;
        return;
    }

    const num = parseCurrencyInput(val);
    amount.value = num.toString();
});
watch(amount, (val) => {
    const num = Number(val);
    const currentParsed = parseCurrencyInput(amountDisplay.value);

    if (Math.abs(currentParsed - num) > 0.001) {
        amountDisplay.value = val ? formatCurrencyInput(val) : "";
    }
});

watch(transferFeeDisplay, (val) => {
    const formatted = formatCurrencyLive(val);
    if (formatted !== val) {
        transferFeeDisplay.value = formatted;
        return;
    }
    const num = parseCurrencyInput(val);
    transferFee.value = num.toString();
});

watch(transferFee, (val) => {
    const num = Number(val);
    const currentParsed = parseCurrencyInput(transferFeeDisplay.value);
    if (Math.abs(currentParsed - num) > 0.001) {
        transferFeeDisplay.value = val ? formatCurrencyInput(val) : "";
    }
});

const onAmountBlur = () => {
    const num = parseCurrencyInput(amountDisplay.value);
    if (num) amountDisplay.value = formatCurrencyInput(num);
};

const onFeeBlur = () => {
    const num = parseCurrencyInput(transferFeeDisplay.value);
    if (num) transferFeeDisplay.value = formatCurrencyInput(num);
};
</script>

<template>
    <Dialog :open="props.open" @update:open="emit('update:open', $event)">
        <DialogContent class="max-w-md bg-card text-foreground" @interact-outside="swal.handleSwalInteractOutside">
            <DialogHeader>
                <DialogTitle>{{ transactionToEdit ? 'Edit Transaksi' : 'Tambah Transaksi' }}</DialogTitle>
                <DialogDescription>Catat pengeluaran atau pemasukan baru.</DialogDescription>
            </DialogHeader>

            <div v-if="transactionToEdit && (transactionToEdit.type === 'transfer_in' || transactionToEdit.type === 'transfer_out')"
                class="bg-blue-50 dark:bg-blue-950/30 p-3 rounded-lg text-xs text-blue-700 dark:text-blue-300 mb-4 border border-blue-200 dark:border-blue-800">
                <strong>Info:</strong> Transaksi ini terhubung dengan transaksi transfer lainnya. Perubahan nominal,
                tanggal, atau deskripsi akan otomatis diterapkan pada pasangannya.
            </div>

            <Tabs v-model="activeTab" class="w-full">
                <TabsList class="grid w-full grid-cols-3 mb-4 h-auto p-1 bg-muted/60 rounded-xl"
                    v-if="activeTab !== 'saving' && !props.wishlistItemId">
                    <TabsTrigger value="expense"
                        class="rounded-lg py-2 data-[state=active]:bg-red-500 data-[state=active]:text-white dark:data-[state=active]:bg-red-600 dark:text-muted-foreground dark:data-[state=active]:text-white transition-all">
                        Pengeluaran</TabsTrigger>
                    <TabsTrigger value="income"
                        class="rounded-lg py-2 data-[state=active]:bg-emerald-500 data-[state=active]:text-white dark:data-[state=active]:bg-emerald-600 dark:text-muted-foreground dark:data-[state=active]:text-white transition-all">
                        Pemasukan</TabsTrigger>
                    <TabsTrigger value="transfer"
                        class="rounded-lg py-2 data-[state=active]:bg-blue-500 data-[state=active]:text-white dark:data-[state=active]:bg-blue-600 dark:text-muted-foreground dark:data-[state=active]:text-white transition-all">
                        Transfer
                    </TabsTrigger>
                </TabsList>

                <div v-if="activeTab === 'saving'"
                    class="bg-emerald-50 dark:bg-emerald-950/30 text-emerald-700 dark:text-emerald-300 p-3 rounded-lg mb-4 text-center font-medium border border-emerald-200 dark:border-emerald-800">
                    Menabung untuk: {{ props.savingGoalTarget?.name }}
                </div>

                <div v-if="props.wishlistItemId"
                    class="bg-red-50 dark:bg-red-950/30 text-red-700 dark:text-red-300 p-3 rounded-lg mb-4 text-center font-medium border border-red-200 dark:border-red-800">
                    Pembelian Wishlist
                </div>

                <div class="space-y-4 py-2">
                    <div class="space-y-2">
                        <Label>Tanggal</Label>
                        <div class="relative">
                            <input type="date" v-model="date"
                                class="peer absolute inset-0 w-full h-full opacity-0 cursor-pointer z-10"
                                :disabled="isSubmitting" @click="($event.target as HTMLInputElement).showPicker()" />
                            <Input type="text" readonly tabindex="-1"
                                :value="date ? formatDate(date, 'dd/MM/yyyy') : ''" placeholder="dd/mm/yyyy"
                                :class="['block w-full bg-background cursor-pointer peer-focus-visible:ring-2 peer-focus-visible:ring-ring peer-focus-visible:ring-offset-2', errors.date ? 'border-red-500 ring-1 ring-red-500' : '']"
                                :disabled="isSubmitting" />
                            <div
                                class="absolute inset-y-0 right-0 flex items-center pr-3 pointer-events-none text-muted-foreground z-20">
                                <Calendar class="w-4 h-4" />
                            </div>
                        </div>
                        <span v-if="errors.date" class="text-xs text-red-500 font-medium">Tanggal wajib diisi</span>
                    </div>

                    <div class="space-y-2">
                        <Label>Nominal</Label>
                        <Input type="text" inputmode="decimal" placeholder="0" v-model="amountDisplay"
                            @blur="onAmountBlur"
                            :class="['bg-background', errors.amount ? 'border-red-500 ring-1 ring-red-500' : '']"
                            :disabled="isSubmitting" />
                        <span v-if="errors.amount" class="text-xs text-red-500 font-medium">Nominal wajib diisi</span>
                    </div>

                    <div class="space-y-2">
                        <Label>{{ activeTab === 'transfer' ? 'Dari Dompet' : 'Dompet' }}</Label>
                        <SearchableSelect v-model="selectedWallet" :options="walletOptions" :disabled="isSubmitting"
                            :error="errors.wallet" placeholder="Pilih Dompet">
                            <template #option="{ option }">
                                <div class="flex items-center gap-2 justify-between w-full">
                                    <div class="flex items-center gap-2">
                                        <component v-if="getIconComponent(option.icon)"
                                            :is="getIconComponent(option.icon)" class="h-4 w-4 shrink-0" />
                                        <span v-else class="text-xs shrink-0">{{ getEmoji(option.icon) || '💼' }}</span>
                                        <span>{{ option.label }}</span>
                                    </div>
                                    <div v-if="option.available_balance !== undefined && option.available_balance !== option.balance"
                                        class="text-xs text-muted-foreground whitespace-nowrap">
                                        (Tersedia: {{ new Intl.NumberFormat("id-ID", {
                                            style: "currency", currency:
                                                "IDR", maximumFractionDigits: 0
                                        }).format(option.available_balance) }})
                                    </div>
                                </div>
                            </template>
                        </SearchableSelect>
                        <span v-if="errors.wallet" class="text-xs text-red-500 font-medium">{{ activeTab === 'transfer'
                            ? 'Dompet asal' : 'Dompet' }} wajib dipilih</span>
                    </div>

                    <div v-if="activeTab === 'transfer'" class="space-y-2">
                        <Label>Ke Dompet</Label>
                        <SearchableSelect v-model="toWallet" :options="walletOptions" :disabled="isSubmitting"
                            :error="errors.toWallet" placeholder="Pilih Dompet Tujuan">
                            <template #option="{ option }">
                                <div class="flex items-center gap-2 justify-between w-full">
                                    <div class="flex items-center gap-2">
                                        <component v-if="getIconComponent(option.icon)"
                                            :is="getIconComponent(option.icon)" class="h-4 w-4 shrink-0" />
                                        <span v-else class="text-xs shrink-0">{{ getEmoji(option.icon) || '💼' }}</span>
                                        <span>{{ option.label }}</span>
                                    </div>
                                    <div v-if="option.available_balance !== undefined && option.available_balance !== option.balance"
                                        class="text-xs text-muted-foreground whitespace-nowrap">
                                        (Tersedia: {{ new Intl.NumberFormat("id-ID", {
                                            style: "currency", currency:
                                                "IDR", maximumFractionDigits: 0
                                        }).format(option.available_balance) }})
                                    </div>
                                </div>
                            </template>
                        </SearchableSelect>
                        <span v-if="errors.toWallet" class="text-xs text-red-500 font-medium">Dompet tujuan wajib
                            dipilih</span>
                    </div>

                    <div v-if="activeTab === 'transfer'" class="space-y-2">
                        <Label>Biaya Admin (Opsional)</Label>
                        <Input type="text" inputmode="decimal" placeholder="0" v-model="transferFeeDisplay"
                            @blur="onFeeBlur" class="bg-background" :disabled="isSubmitting" />
                        <span class="text-[10px] text-muted-foreground">Biaya ini akan ditarik dari dompet asal
                            (Pengeluaran baru).</span>
                    </div>

                    <div v-if="activeTab !== 'transfer'" class="space-y-2">
                        <Label>Kategori</Label>
                        <SearchableSelect v-model="selectedCategory" :options="categoryOptions"
                            :disabled="isSubmitting || activeTab === 'saving'" :error="errors.category"
                            placeholder="Pilih Kategori">
                            <template #option="{ option }">
                                <div class="flex items-center gap-2">
                                    <component v-if="getIconComponent(option.icon)" :is="getIconComponent(option.icon)"
                                        class="h-4 w-4 shrink-0" />
                                    <span v-else>{{ getEmoji(option.icon) || option.icon }}</span>
                                    <span>{{ option.label }}</span>
                                </div>
                            </template>
                        </SearchableSelect>
                        <span v-if="errors.category" class="text-xs text-red-500 font-medium">Kategori wajib
                            dipilih</span>
                    </div>
                    <div class="space-y-2">
                        <Label>Deskripsi (Opsional)</Label>
                        <Input placeholder="Misal: Makan siang, Gaji bulanan" v-model="description"
                            class="bg-background" :disabled="isSubmitting || activeTab === 'saving'" />
                    </div>

                    <div v-if="activeTab !== 'transfer'" class="space-y-2">
                        <Label>Lampiran (Foto/Gambar)</Label>
                        <div v-if="existingAttachment && !file" class="mb-2 relative w-fit">
                            <img :src="`${baseUrl.replace(/\/$/, '')}/${existingAttachment.replace(/^\//, '')}`"
                                alt="Lampiran" class="h-20 w-auto rounded-md border border-border" />
                            <Button type="button" variant="destructive" size="icon"
                                class="h-5 w-5 absolute -top-2 -right-2 rounded-full" @click="existingAttachment = ''">
                                <span class="text-xs">x</span>
                            </Button>
                        </div>
                        <Input type="file" accept="image/png, image/jpeg, image/jpg, .heic" @change="handleFileChange"
                            class="bg-background" :disabled="isSubmitting || isProcessingFile" />
                        <p v-if="isProcessingFile" class="text-xs text-blue-500 font-medium animate-pulse mt-1">Sedang
                            memproses gambar...</p>
                        <p class="text-[10px] text-muted-foreground">Maksimal 5MB. Format: JPG, PNG, HEIC.</p>
                    </div>
                </div>
            </Tabs>

            <DialogFooter class="flex gap-2 justify-end mt-4">
                <Button variant="outline" @click="emit('update:open', false)"
                    :disabled="isSubmitting || isProcessingFile">Batal</Button>
                <Button @click="handleSave"
                    class="bg-gradient-to-r from-emerald-600 to-teal-500 text-white hover:from-emerald-500 hover:to-teal-400"
                    :disabled="isSubmitting || isProcessingFile" :loading="isSubmitting || isProcessingFile">
                    {{ isProcessingFile ? 'Memproses...' : (transactionToEdit ? 'Simpan Perubahan' : 'Simpan') }}
                </Button>
            </DialogFooter>
        </DialogContent>
    </Dialog>
</template>
