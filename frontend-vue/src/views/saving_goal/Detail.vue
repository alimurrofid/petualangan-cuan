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
        await store.deleteContribution(props.goal.id, contributionId);
        Swal.fire('Terhapus!', 'Data berhasil dihapus.', 'success');
        emit("update:open", false);
    }
};

</script>

<template>
  <Dialog :open="open" @update:open="$emit('update:open', $event)">
    <DialogContent class="sm:max-w-[800px] rounded-3xl bg-card text-foreground">
      <DialogHeader class="pr-6">
        <DialogTitle class="flex flex-wrap items-center gap-2 leading-normal">
            <span>Detail Target Menabung</span>
            <span 
                class="text-xs px-2 py-0.5 rounded-md border font-semibold tracking-wide uppercase whitespace-nowrap"
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
                      <TableRow v-if="!goal.contributions || goal.contributions.length === 0">
                          <TableCell colspan="5" class="h-24 text-center">
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
                             <div class="flex items-center gap-2 text-sm max-w-[200px]">
                                 <component v-if="contribution.wallet && getIconComponent(contribution.wallet.icon)" :is="getIconComponent(contribution.wallet.icon)" class="h-3 w-3 text-muted-foreground shrink-0" />
                                 <span v-else-if="contribution.wallet">{{ getEmoji(contribution.wallet.icon) || '💼' }}</span>
                                 <span v-else>💼</span>
                                 <span class="truncate" :title="contribution.wallet?.name">{{ contribution.wallet?.name || '-' }}</span>
                             </div>
                          </TableCell>
                          <TableCell class="text-sm text-muted-foreground max-w-[150px] truncate">
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

          <!-- Mobile Card View -->
          <div class="md:hidden space-y-4">
               <div v-if="!goal.contributions || goal.contributions.length === 0" class="text-center py-8 text-muted-foreground border-2 border-dashed rounded-xl">
                    <AlertCircle class="h-8 w-8 opacity-20 mx-auto mb-2" />
                    <p class="text-sm">Belum ada riwayat menabung.</p>
               </div>
               <div v-for="contribution in goal.contributions" :key="contribution.id" class="bg-card border rounded-xl p-4 shadow-sm space-y-3">
                   <div class="flex justify-between items-start">
                       <div>
                           <p class="text-xs font-bold text-muted-foreground">{{ format(new Date(contribution.date), "d MMM yyyy", { locale: id }) }}</p>
                           <p class="text-[10px] text-muted-foreground">{{ format(new Date(contribution.date), "HH:mm", { locale: id }) }}</p>
                       </div>
                       <div class="text-right">
                           <p class="font-bold text-base text-emerald-600" :class="{ 'privacy-blur': authStore.isPrivacyMode }">{{ formatCurrency(contribution.amount) }}</p>
                       </div>
                   </div>
                   
                   <div class="flex items-center gap-2 text-xs bg-muted/50 p-2 rounded-lg">
                        <component v-if="contribution.wallet && getIconComponent(contribution.wallet.icon)" :is="getIconComponent(contribution.wallet.icon)" class="h-3.5 w-3.5 text-muted-foreground" />
                        <span v-else-if="contribution.wallet">{{ getEmoji(contribution.wallet.icon) || '💼' }}</span>
                        <span v-else>💼</span>
                        <span class="font-medium">{{ contribution.wallet?.name || '-' }}</span>
                   </div>

                   <div class="flex justify-between items-center pt-1 border-t border-border/50">
                       <p class="text-xs text-muted-foreground italic w-full truncate mr-2">{{ contribution.transaction?.description || 'Tanpa catatan' }}</p>
                       <Button 
                            variant="ghost" 
                            size="sm" 
                            class="h-7 px-2 text-red-400 hover:text-red-600 hover:bg-red-50 rounded-md ml-auto"
                            @click="handleDeleteContribution(contribution.id)"
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
