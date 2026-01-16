<script setup lang="ts">
import { type Debt, useDebtStore } from "@/stores/debt";
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
          // Note: The parent component relies on store data. 
          // Store already refetches data after deletePayment.
          // However, we might need to close/re-open or just rely on reactivity if `debt` prop is reactive from store list.
          // Since `debt` prop is passed from Index.vue which takes it from `filteredItems` (computed from store),
          // it SHOULD update automatically if the parent re-renders or if the object reference is kept.
          // But `filteredItems` creates new array. We might need to ensure the parent updates the selected `debt` prop.
          // Actually, standard pattern is to close details or rely on reactivity. 
          // Let's assume reactivity works if strict ID matching is used, or we just close.
      } catch (e: any) {
          Swal.fire('Gagal', e.message || 'Terjadi kesalahan', 'error');
      }
  }
};

</script>

<template>
  <Dialog :open="open" @update:open="$emit('update:open', $event)">
    <DialogContent class="sm:max-w-[700px]">
      <DialogHeader>
        <DialogTitle class="flex items-center gap-2">
            <span>Detail {{ debt?.type === 'debt' ? 'Utang' : 'Piutang' }}</span>
            <span 
                class="text-xs px-2 py-1 rounded border"
                :class="debt?.is_paid ? 'bg-green-100 text-green-700 border-green-200' : 'bg-yellow-100 text-yellow-700 border-yellow-200'"
            >
                {{ debt?.is_paid ? 'LUNAS' : 'BELUM LUNAS' }}
            </span>
        </DialogTitle>
        <DialogDescription>
            Riwayat pembayaran dan progress pelunasan untuk <strong>{{ debt?.name }}</strong>.
        </DialogDescription>
      </DialogHeader>

      <div v-if="debt" class="space-y-6 pt-4">
          <!-- History Table -->
          <div class="border rounded-md">
              <Table>
                  <TableHeader>
                      <TableRow>
                          <TableHead>Tanggal</TableHead>
                          <TableHead>Dompet</TableHead>
                          <TableHead>Catatan</TableHead>
                          <TableHead class="text-right">Nominal</TableHead>
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
                          <TableCell class="text-sm text-muted-foreground">
                              {{ payment.note || '-' }}
                          </TableCell>
                          <TableCell class="text-right font-medium">
                              {{ formatCurrency(payment.amount) }}
                          </TableCell>
                           <TableCell>
                              <Button 
                                variant="ghost" 
                                size="icon" 
                                class="h-6 w-6 text-red-400 hover:text-red-600 hover:bg-red-50"
                                @click="handleDeletePayment(payment.id)"
                              >
                                  <Trash2 class="h-3 w-3" />
                              </Button>
                          </TableCell>
                      </TableRow>
                  </TableBody>
              </Table>
          </div>
      </div>
    </DialogContent>
  </Dialog>
</template>
