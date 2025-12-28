<script setup lang="ts">
import { ref } from "vue";
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogFooter, DialogDescription } from "@/components/ui/dialog";
import { Tabs, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Label } from "@/components/ui/label";
import { Input } from "@/components/ui/input";
import { Button } from "@/components/ui/button";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";

const props = defineProps<{
  open: boolean;
}>();

const emit = defineEmits<{
  (e: "update:open", value: boolean): void;
  (e: "save", data: any): void;
}>();

const activeTab = ref("expense");
const date = ref(new Date().toISOString().slice(0, 10));
const amount = ref("");
const selectedWallet = ref("");
const fromWallet = ref("");
const toWallet = ref("");
const feeTransfer = ref("");
const selectedCategory = ref("");
const description = ref("");

const handleSave = () => {
  const transactionData = {
    type: activeTab.value,
    date: date.value,
    amount: amount.value,
    category: selectedCategory.value,
    description: description.value
  };

  if (activeTab.value === 'transfer') {
    Object.assign(transactionData, {
      from_wallet: fromWallet.value,
      to_wallet: toWallet.value,
      fee_transfer: feeTransfer.value
    });
  } else {
    Object.assign(transactionData, {
        wallet: selectedWallet.value
    });
  }
  
  emit("save", transactionData);
  emit("update:open", false);
};
</script>

<template>
  <Dialog :open="open" @update:open="emit('update:open', $event)">
    <DialogContent class="max-w-md bg-card text-foreground">
      <DialogHeader>
        <DialogTitle>Add transaction</DialogTitle>
        <DialogDescription class="hidden">Form for adding manual transaction</DialogDescription>
      </DialogHeader>

      <Tabs v-model="activeTab" class="w-full">
        <TabsList class="grid w-full grid-cols-3 mb-4 h-auto p-1 bg-muted/60 rounded-xl">
          <TabsTrigger value="expense" class="rounded-lg py-2 data-[state=active]:bg-red-500 data-[state=active]:text-white">Expenses</TabsTrigger>
          <TabsTrigger value="income" class="rounded-lg py-2 data-[state=active]:bg-emerald-500 data-[state=active]:text-white">Income</TabsTrigger>
          <TabsTrigger value="transfer" class="rounded-lg py-2 data-[state=active]:bg-blue-500 data-[state=active]:text-white">Transfer</TabsTrigger>
        </TabsList>

        <div class="space-y-4 py-2">
            <div class="space-y-2">
                <Label>Transaction date</Label>
                <div class="relative">
                    <Input type="date" v-model="date" class="block w-full" />
                </div>
            </div>

            <div class="space-y-2">
                <Label>Total transactions</Label>
                <Input type="number" placeholder="Example: 5000" v-model="amount" />
            </div>

            <div v-if="activeTab === 'transfer'" class="grid grid-cols-2 gap-4">
                <div class="space-y-2">
                    <Label>From Wallet</Label>
                    <Select v-model="fromWallet">
                        <SelectTrigger class="w-full">
                            <SelectValue placeholder="From wallet" />
                        </SelectTrigger>
                        <SelectContent>
                            <SelectItem value="cash">💵 Uang Tunai</SelectItem>
                            <SelectItem value="bca">🏦 BCA</SelectItem>
                            <SelectItem value="gopay">📱 GoPay</SelectItem>
                        </SelectContent>
                    </Select>
                </div>
                <div class="space-y-2">
                    <Label>To Wallet</Label>
                    <Select v-model="toWallet">
                        <SelectTrigger class="w-full">
                            <SelectValue placeholder="To wallet" />
                        </SelectTrigger>
                        <SelectContent>
                            <SelectItem value="cash">💵 Uang Tunai</SelectItem>
                            <SelectItem value="bca">🏦 BCA</SelectItem>
                            <SelectItem value="gopay">📱 GoPay</SelectItem>
                        </SelectContent>
                    </Select>
                </div>
            </div>

            <div v-else class="space-y-2">
                <Label>Select Wallet</Label>
                <Select v-model="selectedWallet">
                    <SelectTrigger class="w-full">
                        <SelectValue placeholder="Select wallet" />
                    </SelectTrigger>
                    <SelectContent>
                        <SelectItem value="cash">💵 Uang Tunai</SelectItem>
                        <SelectItem value="bca">🏦 BCA</SelectItem>
                        <SelectItem value="gopay">📱 GoPay</SelectItem>
                    </SelectContent>
                </Select>
            </div>

            <div v-if="activeTab === 'transfer'" class="space-y-2">
                <Label>Fee Transfer</Label>
                <Input type="number" placeholder="Example: 2500" v-model="feeTransfer" />
            </div>

             <div v-if="activeTab !== 'transfer'" class="space-y-2">
                <Label>Select category</Label>
                 <Select v-model="selectedCategory">
                    <SelectTrigger class="w-full">
                        <SelectValue placeholder="Select category" />
                    </SelectTrigger>
                    <SelectContent>
                        <SelectItem value="food">🍔 Makanan</SelectItem>
                        <SelectItem value="transport">🚗 Transport</SelectItem>
                        <SelectItem value="shopping">🛒 Belanja</SelectItem>
                    </SelectContent>
                </Select>
            </div>

            <div class="space-y-2">
                <Label>Description</Label>
                <Input placeholder="Subject / Description" v-model="description" />
            </div>
        </div>
      </Tabs>

      <DialogFooter class="flex gap-2 justify-end mt-4">
        <Button variant="outline" @click="emit('update:open', false)">Cancel</Button>
        <Button @click="handleSave" class="bg-primary text-primary-foreground">Save</Button>
      </DialogFooter>
    </DialogContent>
  </Dialog>
</template>
