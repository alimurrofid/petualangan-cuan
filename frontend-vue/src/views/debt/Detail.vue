<script setup lang="ts">
import { type Debt, useDebtStore } from "@/stores/debt";
import { useAuthStore } from "@/stores/auth";
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogDescription } from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { formatCurrency } from "@/lib/utils";
import { format } from "date-fns";
import { id } from "date-fns/locale";
import { Trash2, AlertCircle } from "lucide-vue-next";
import Swal from "sweetalert2";
import { getEmoji, getIconComponent } from "@/lib/icons";

const props = defineProps<{
  open: boolean;
  debt: Debt | null;
}>();

const emit = defineEmits(["update:open"]);

const debtStore = useDebtStore();
const authStore = useAuthStore();

const handleDeletePayment = async (paymentId: number) => {
  const result = await Swal.fire({
      title: 'Hapus Pembayaran?',
      text: "Saldo dompet dan sisa utang akan dikembalikan otomatis.",
      icon: 'warning',
      showCancelButton: true,
      confirmButtonColor: '#ef4444',
      cancelButtonColor: '#64748b',
      confirmButtonText: 'Ya, Hapus!',
      cancelButtonText: 'Batal'
  });

  if (result.isConfirmed) {
      try {
          await debtStore.deletePayment(paymentId);
          Swal.fire('Terhapus!', 'Data pembayaran berhasil dihapus.', 'success');
      } catch (e: any) {
          Swal.fire('Gagal', e.message || 'Terjadi kesalahan', 'error');
      }
  }
};

</script>

<template>
  <Dialog :open="open" @update:open="$emit('update:open', $event)">
    <DialogContent class="sm:max-w-[700px] rounded-3xl bg-card text-foreground">
      <DialogHeader>
        <DialogTitle class="flex items-center gap-2">
            <span>Detail {{ debt?.type === 'debt' ? 'Utang' : 'Piutang' }}</span>
            <span 
                class="text-xs px-2 py-0.5 rounded-md border font-semibold tracking-wide uppercase"
                :class="debt?.is_paid ? 'bg-emerald-50 text-emerald-700 border-emerald-200' : 'bg-yellow-50 text-yellow-700 border-yellow-200'"
            >
                {{ debt?.is_paid ? 'LUNAS' : 'BELUM LUNAS' }}
            </span>
        </DialogTitle>
        <DialogDescription>
            Riwayat pembayaran dan progress pelunasan untuk <strong>{{ debt?.name }}</strong>.
        </DialogDescription>
      </DialogHeader>

      <div v-if="debt" class="space-y-6 pt-4">
          <!-- Desktop Table View -->
          <div class="hidden md:block border rounded-2xl overflow-hidden shadow-sm bg-card">
              <Table>
                  <TableHeader class="bg-muted/30">
                      <TableRow>
                          <TableHead class="text-xs font-semibold uppercase tracking-wider text-muted-foreground py-3">Tanggal</TableHead>
                          <TableHead class="text-xs font-semibold uppercase tracking-wider text-muted-foreground py-3">Dompet</TableHead>
                          <TableHead class="text-xs font-semibold uppercase tracking-wider text-muted-foreground py-3">Catatan</TableHead>
                          <TableHead class="text-xs font-semibold uppercase tracking-wider text-muted-foreground py-3 text-right">Nominal</TableHead>
                          <TableHead class="w-[50px]"></TableHead>
                      </TableRow>
                  </TableHeader>
                  <TableBody>
                      <TableRow v-if="!debt.payments || debt.payments.length === 0">
                          <TableCell colspan="5" class="h-24 text-center">
                              <div class="flex flex-col items-center justify-center gap-2 text-muted-foreground">
                                <AlertCircle class="h-8 w-8 opacity-20" />
                                Belum ada riwayat pembayaran.
                              </div>
                          </TableCell>
                      </TableRow>
                      <TableRow v-for="payment in debt.payments" :key="payment.id">
                          <TableCell class="text-sm">
                              {{ format(new Date(payment.date), "d MMM yyyy HH:mm", { locale: id }) }}
                          </TableCell>
                          <TableCell>
                             <div class="flex items-center gap-2 text-sm">
                                 <component v-if="getIconComponent(payment.wallet.icon)" :is="getIconComponent(payment.wallet.icon)" class="h-3 w-3 text-muted-foreground" />
                                 <span v-else>{{ getEmoji(payment.wallet.icon) || '💼' }}</span>
                                 <span>{{ payment.wallet.name }}</span>
                             </div>
                          </TableCell>
                          <TableCell class="text-sm text-muted-foreground max-w-[150px] truncate">
                              {{ payment.note || '-' }}
                          </TableCell>
                          <TableCell class="text-right font-medium" :class="{ 'privacy-blur': authStore.isPrivacyMode }">
                              {{ formatCurrency(payment.amount) }}
                          </TableCell>
                           <TableCell>
                              <Button 
                                variant="ghost" 
                                size="icon" 
                                class="h-8 w-8 text-red-400 hover:text-red-600 hover:bg-red-50 rounded-lg"
                                @click="handleDeletePayment(payment.id)"
                              >
                                  <Trash2 class="h-3.5 w-3.5" />
                              </Button>
                          </TableCell>
                      </TableRow>
                  </TableBody>
              </Table>
          </div>

          <!-- Mobile Card View -->
          <div class="md:hidden space-y-4">
               <div v-if="!debt.payments || debt.payments.length === 0" class="text-center py-8 text-muted-foreground border-2 border-dashed rounded-xl">
                    <AlertCircle class="h-8 w-8 opacity-20 mx-auto mb-2" />
                    <p class="text-sm">Belum ada riwayat pembayaran.</p>
               </div>
               <div v-for="payment in debt.payments" :key="payment.id" class="bg-card border rounded-xl p-4 shadow-sm space-y-3">
                   <div class="flex justify-between items-start">
                       <div>
                           <p class="text-xs font-bold text-muted-foreground">{{ format(new Date(payment.date), "d MMM yyyy", { locale: id }) }}</p>
                           <p class="text-[10px] text-muted-foreground">{{ format(new Date(payment.date), "HH:mm", { locale: id }) }}</p>
                       </div>
                       <div class="text-right">
                           <p class="font-bold text-base" :class="{ 'privacy-blur': authStore.isPrivacyMode }">{{ formatCurrency(payment.amount) }}</p>
                       </div>
                   </div>
                   
                   <div class="flex items-center gap-2 text-xs bg-muted/50 p-2 rounded-lg">
                        <component v-if="getIconComponent(payment.wallet.icon)" :is="getIconComponent(payment.wallet.icon)" class="h-3.5 w-3.5 text-muted-foreground" />
                        <span v-else>{{ getEmoji(payment.wallet.icon) || '💼' }}</span>
                        <span class="font-medium">{{ payment.wallet.name }}</span>
                   </div>

                   <div class="flex justify-between items-center pt-1 border-t border-border/50">
                       <p class="text-xs text-muted-foreground italic w-full truncate mr-2">{{ payment.note || 'Tanpa catatan' }}</p>
                       <Button 
                            variant="ghost" 
                            size="sm" 
                            class="h-7 px-2 text-red-400 hover:text-red-600 hover:bg-red-50 rounded-md ml-auto"
                            @click="handleDeletePayment(payment.id)"
                        >
                            <Trash2 class="h-3.5 w-3.5 mr-1" /> Hapus
                        </Button>
                   </div>
               </div>
          </div>
      </div>
    </DialogContent>
  </Dialog>
</template>
