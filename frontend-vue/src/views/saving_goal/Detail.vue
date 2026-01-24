<script setup lang="ts">
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogDescription } from "@/components/ui/dialog";
import { useAuthStore } from "@/stores/auth";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { Button } from "@/components/ui/button";
import { formatCurrency } from "@/lib/utils";
import { format } from "date-fns";
import { id } from "date-fns/locale";
import { AlertCircle, Trash2 } from "lucide-vue-next";
import { getEmoji, getIconComponent } from "@/lib/icons";
import { useSavingGoalStore } from "@/stores/saving_goal";
import Swal from "sweetalert2";

const props = defineProps<{
  open: boolean;
  goal: any | null;
}>();

const emit = defineEmits(["update:open"]);
const store = useSavingGoalStore();
const authStore = useAuthStore();

const handleDeleteContribution = async (contributionId: number) => {
    if (!props.goal) return;
    
    const result = await Swal.fire({
        title: 'Hapus Riwayat?',
        text: "Tabungan akan dikurangi dan saldo dompet akan dikembalikan (jika transaksi dibatalkan).",
        icon: 'warning',
        showCancelButton: true,
        confirmButtonColor: '#ef4444',
        cancelButtonColor: '#64748b',
        confirmButtonText: 'Ya, Hapus!',
        cancelButtonText: 'Batal'
    });

    if (result.isConfirmed) {
        // Optimistic update or wait for re-fetch? Store re-fetches.
        // We need to close or update the local `goal` prop?
        // The parent binds `selectedGoal`. When store updates `goals`, `selectedGoal` might not verify if it's the same refernece.
        // `Index.vue` uses `selectedGoal` ref.
        // If store.fetchGoals() runs, it creates NEW objects. `selectedGoal` will be STALE.
        // We need to emit an event or handle it.
        // Or simpler: close dialog. Use user preference from Debt?
        // Debt re-fetches. Let's try closing dialog or updating local state (harder).
        // Let's just call store and see. Ideally, we emit 'refresh' or close.
        await store.deleteContribution(props.goal.id, contributionId);
        Swal.fire('Terhapus!', 'Data berhasil dihapus.', 'success');
        emit("update:open", false); // Close dialog to avoid stale data
    }
};

</script>

<template>
  <Dialog :open="open" @update:open="$emit('update:open', $event)">
    <DialogContent class="sm:max-w-[700px] rounded-3xl bg-card text-foreground">
      <DialogHeader>
        <DialogTitle class="flex items-center gap-2">
            <span>Detail Target Menabung</span>
            <span 
                class="text-xs px-2 py-0.5 rounded-md border font-semibold tracking-wide uppercase"
                :class="goal?.is_achieved ? 'bg-emerald-50 text-emerald-700 border-emerald-200' : 'bg-blue-50 text-blue-700 border-blue-200'"
            >
                {{ goal?.is_achieved ? 'TERCAPAI' : 'BELUM TERCAPAI' }}
            </span>
        </DialogTitle>
        <DialogDescription>
            Riwayat menabung untuk <strong>{{ goal?.name }}</strong>.
        </DialogDescription>
      </DialogHeader>

      <div v-if="goal" class="space-y-6 pt-4">
          <!-- History Table -->
          <div class="border rounded-2xl overflow-hidden shadow-sm bg-card">
              <Table>
                  <TableHeader class="bg-muted/30">
                      <TableRow>
                          <TableHead class="text-xs font-semibold uppercase tracking-wider text-muted-foreground py-3">Tanggal</TableHead>
                          <TableHead class="text-xs font-semibold uppercase tracking-wider text-muted-foreground py-3">Dompet</TableHead>
                          <TableHead class="text-xs font-semibold uppercase tracking-wider text-muted-foreground py-3">Catatan</TableHead>
                          <TableHead class="text-xs font-semibold uppercase tracking-wider text-muted-foreground py-3 text-right">Nominal</TableHead>
                      </TableRow>
                  </TableHeader>
                  <TableBody>
                      <TableRow v-if="!goal.contributions || goal.contributions.length === 0">
                          <TableCell colspan="4" class="h-24 text-center">
                              <div class="flex flex-col items-center justify-center gap-2 text-muted-foreground">
                                <AlertCircle class="h-8 w-8 opacity-20" />
                                Belum ada riwayat menabung.
                              </div>
                          </TableCell>
                      </TableRow>
                      <TableRow v-for="contribution in goal.contributions" :key="contribution.id">
                          <TableCell class="text-sm">
                              {{ format(new Date(contribution.date), "d MMM yyyy HH:mm", { locale: id }) }}
                          </TableCell>
                          <TableCell>
                             <div class="flex items-center gap-2 text-sm">
                                 <component v-if="contribution.wallet && getIconComponent(contribution.wallet.icon)" :is="getIconComponent(contribution.wallet.icon)" class="h-3 w-3 text-muted-foreground" />
                                 <span v-else-if="contribution.wallet">{{ getEmoji(contribution.wallet.icon) || '💼' }}</span>
                                 <span v-else>💼</span>
                                 <span>{{ contribution.wallet?.name || '-' }}</span>
                             </div>
                          </TableCell>
                          <TableCell class="text-sm text-muted-foreground">
                              {{ contribution.transaction?.description || '-' }}
                          </TableCell>
                          <TableCell class="text-right font-medium" :class="{ 'privacy-blur': authStore.isPrivacyMode }">
                              {{ formatCurrency(contribution.amount) }}
                          </TableCell>
                          <TableCell>
                               <Button 
                                 variant="ghost" 
                                 size="icon" 
                                 class="h-8 w-8 text-red-400 hover:text-red-600 hover:bg-red-50 rounded-lg"
                                 @click="handleDeleteContribution(contribution.id)"
                               >
                                   <Trash2 class="h-3.5 w-3.5" />
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
