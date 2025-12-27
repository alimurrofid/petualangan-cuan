<script setup lang="ts">
import { ref, nextTick, onMounted } from "vue";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Card } from "@/components/ui/card";
import { Send, Bot, User, Sparkles, Image as ImageIcon, Mic } from "lucide-vue-next";
import { format } from "date-fns";

interface Message {
  id: number;
  role: "user" | "assistant";
  content: string;
  time: string;
}

const messages = ref<Message[]>([
  {
    id: 1,
    role: "assistant",
    content: "Halo! Saya asisten keuangan pribadimu. Ada yang bisa saya bantu hari ini? 🤖",
    time: format(new Date(), "HH:mm")
  }
]);

const userInput = ref("");
const isTyping = ref(false);
const chatContainer = ref<HTMLElement | null>(null);

const scrollToBottom = async () => {
  await nextTick();
  if (chatContainer.value) {
    chatContainer.value.scrollTop = chatContainer.value.scrollHeight;
  }
};

onMounted(scrollToBottom);

const sendMessage = async () => {
  const text = userInput.value.trim();
  if (!text) return;

  messages.value.push({
    id: Date.now(),
    role: "user",
    content: text,
    time: format(new Date(), "HH:mm")
  });

  userInput.value = "";
  scrollToBottom();

  isTyping.value = true;
  await new Promise(resolve => setTimeout(resolve, 1500)); 
  
  isTyping.value = false;
  messages.value.push({
    id: Date.now() + 1,
    role: "assistant",
    content: generateMockResponse(text),
    time: format(new Date(), "HH:mm")
  });

  scrollToBottom();
};

const generateMockResponse = (input: string): string => {
  const lower = input.toLowerCase();
  
  if (lower.includes("boros") || lower.includes("pengeluaran")) {
    return "Wah, kalau merasa boros, coba cek lagi kategori pengeluaranmu bulan ini. Mungkin bisa kurangi jajan kopi? ☕😅";
  }
  if (lower.includes("nabung") || lower.includes("investasi")) {
    return "Ide bagus! Menyisihkan 20% pendapatan untuk tabungan adalah awal yang baik.";
  }
  if (lower.includes("halo") || lower.includes("hi")) {
    return "Halo juga! Siap mencatat cuan hari ini? 🚀";
  }

  return "Hmm, menarik. Ceritakan lebih lanjut, atau coba tanya soal tips hemat!";
};
</script>

<template>
  <div class="flex flex-col h-[calc(100vh-2rem)] md:h-[calc(100vh-3.5rem)] max-w-5xl mx-auto p-4 md:p-6 space-y-4">
    
    <div class="flex items-center gap-4 pb-4 border-b border-border">
      <div class="h-12 w-12 rounded-2xl bg-gradient-to-br from-emerald-500 to-teal-600 flex items-center justify-center shadow-lg text-white">
          <Bot class="h-6 w-6" />
      </div>
      <div>
        <h2 class="text-xl font-bold tracking-tight flex items-center gap-2">
            AI Assistant
            <span class="text-[10px] bg-emerald-100/50 dark:bg-emerald-900/30 text-emerald-600 dark:text-emerald-400 px-2 py-0.5 rounded-full border border-emerald-200/50 flex items-center gap-1">
                <Sparkles class="w-3 h-3" /> Beta
            </span>
        </h2>
        <p class="text-sm text-muted-foreground">Tanyakan tips keuangan atau analisis pengeluaranmu.</p>
      </div>
    </div>

    <Card class="flex-1 overflow-hidden flex flex-col bg-muted/30 border-border shadow-sm rounded-3xl relative">
        <div ref="chatContainer" class="flex-1 overflow-y-auto p-4 space-y-4 custom-scrollbar scroll-smooth">
            
            <div v-for="msg in messages" :key="msg.id" class="flex w-full">
                <div :class="[
                  'flex max-w-[80%] md:max-w-[70%] gap-2', 
                  msg.role === 'user' ? 'ml-auto flex-row-reverse' : 'mr-auto flex-row'
                ]">
                    
                    <div v-if="msg.role === 'assistant'" class="h-8 w-8 rounded-full bg-emerald-100 dark:bg-emerald-900/30 flex items-center justify-center shrink-0 border border-emerald-200/50">
                        <Bot class="h-4 w-4 text-emerald-600 dark:text-emerald-400" />
                    </div>
                    <div v-else class="h-8 w-8 rounded-full bg-primary flex items-center justify-center shrink-0 text-primary-foreground shadow-sm">
                        <User class="h-4 w-4" />
                    </div>

                    <div :class="['flex flex-col', msg.role === 'user' ? 'items-end' : 'items-start']">
                        <div :class="['px-4 py-2.5 rounded-2xl text-sm shadow-sm leading-relaxed relative group', 
                            msg.role === 'user' 
                                ? 'bg-primary text-primary-foreground rounded-tr-sm' 
                                : 'bg-card border border-border rounded-tl-sm']"
                        >
                            {{ msg.content }}
                        </div>
                         <span class="text-[10px] text-muted-foreground mt-1 opacity-70 px-1">{{ msg.time }}</span>
                    </div>
                </div>
            </div>

            <div v-if="isTyping" class="flex w-full">
                 <div class="flex max-w-[80%] gap-2 mr-auto flex-row">
                     <div class="h-8 w-8 rounded-full bg-emerald-100 dark:bg-emerald-900/30 flex items-center justify-center shrink-0 border border-emerald-200/50">
                        <Bot class="h-4 w-4 text-emerald-600 dark:text-emerald-400" />
                    </div>
                     <div class="bg-card border border-border px-4 py-3 rounded-2xl rounded-tl-sm flex items-center gap-1 h-10 shadow-sm">
                        <span class="w-1.5 h-1.5 bg-muted-foreground/40 rounded-full animate-bounce"></span>
                        <span class="w-1.5 h-1.5 bg-muted-foreground/40 rounded-full animate-bounce delay-150"></span>
                        <span class="w-1.5 h-1.5 bg-muted-foreground/40 rounded-full animate-bounce delay-300"></span>
                    </div>
                 </div>
            </div>

        </div>

        <div class="p-3 md:p-4 bg-card border-t border-border">
            <div class="flex items-center gap-2 bg-muted/30 p-2 rounded-3xl border border-muted-foreground/10 focus-within:ring-1 focus-within:ring-emerald-500/50 transition-all shadow-sm">
                
                <Input 
                    v-model="userInput" 
                    placeholder="Ketik pesan..." 
                    class="flex-1 border-none shadow-none focus-visible:ring-0 bg-transparent px-3 h-9 text-base md:text-sm"
                    :disabled="isTyping"
                    @keydown.enter.prevent="sendMessage"
                />

                <div class="flex items-center gap-1 pr-1">
                    <Button variant="ghost" size="icon" class="h-8 w-8 rounded-full text-muted-foreground hover:text-foreground shrink-0 hover:bg-background/80" title="Kirim Gambar">
                        <ImageIcon class="h-4 w-4" />
                    </Button>

                    <Button variant="ghost" size="icon" class="h-8 w-8 rounded-full text-muted-foreground hover:text-foreground shrink-0 hover:bg-background/80" title="Rekam Suara">
                        <Mic class="h-4 w-4" />
                    </Button>
                </div>

                <Button 
                    @click="sendMessage"
                    size="icon" 
                    :disabled="!userInput.trim() || isTyping"
                    class="h-9 w-9 rounded-full bg-emerald-600 hover:bg-emerald-700 text-white shadow-sm shrink-0 transition-all active:scale-95"
                >
                    <Send class="h-4 w-4 ml-0.5" />
                </Button>
            </div>
        </div>
    </Card>

  </div>
</template>

<style scoped>
.custom-scrollbar::-webkit-scrollbar {
  width: 4px;
}
.custom-scrollbar::-webkit-scrollbar-track {
  background: transparent;
}
.custom-scrollbar::-webkit-scrollbar-thumb {
  background: hsl(var(--border)); 
  border-radius: 4px;
}
</style>
