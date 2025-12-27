<script setup lang="ts">
import { ref, onMounted, watch } from "vue";
import { useRoute } from "vue-router";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Card, CardContent } from "@/components/ui/card";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";


const route = useRoute();
const activeTab = ref("currency");

const tabs = [
  { id: "profile", label: "Profile" },
  { id: "currency", label: "Currency & Language" },
  { id: "password", label: "Change Password" },
  { id: "whatsapp", label: "WhatsApp Integration" },
];

const formData = ref({
    language: "Indonesia",
    timezone: "Asia/Jakarta (GMT+7)",
    currency: "IDR",
    showDecimal: "Hide"
});

const updateTabFromQuery = () => {
    const tab = route.query.tab as string;
    if (tab && tabs.some(t => t.id === tab)) {
        activeTab.value = tab;
    }
};

onMounted(() => {
    updateTabFromQuery();
});

watch(() => route.query.tab, () => {
    updateTabFromQuery();
});


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
            
            <div v-if="activeTab === 'currency'" class="space-y-6">
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
                    <Button class="w-full">Save</Button>
                </div>
            </div>

            <div v-if="activeTab === 'profile'" class="space-y-6">
                 <div class="space-y-4">
                     <div class="grid w-full items-center gap-1.5">
                        <Label for="name">Full Name</Label>
                        <Input id="name" placeholder="Alimurrofid" />
                     </div>
                      <div class="grid w-full items-center gap-1.5">
                        <Label for="email">Email</Label>
                        <Input id="email" type="email" placeholder="email@example.com" />
                     </div>
                 </div>
                 <Button class="w-full">Save Profile</Button>
            </div>

             <div v-if="activeTab === 'password'" class="space-y-6">
                 <div class="space-y-4">

                      <div class="grid w-full items-center gap-1.5">
                        <Label for="new_pass">New Password</Label>
                        <Input id="new_pass" type="password" />
                     </div>
                      <div class="grid w-full items-center gap-1.5">
                        <Label for="confirm_pass">Confirm Password</Label>
                        <Input id="confirm_pass" type="password" />
                     </div>
                 </div>
                 <Button class="w-full">Update Password</Button>
            </div>

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
