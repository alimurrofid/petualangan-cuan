<script setup lang="ts">
import { ref, nextTick, onMounted, onUnmounted } from "vue";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Card } from "@/components/ui/card";
import { Send, Bot, User, Sparkles, Image as ImageIcon, Mic, X, Loader2 } from "lucide-vue-next";
import { format } from "date-fns";

interface Message {
  id: number;
  role: "user" | "assistant";
  content: string;
  image?: string;
  audio?: string;
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
const fileInput = ref<HTMLInputElement | null>(null);
const selectedFile = ref<File | null>(null);
const selectedFilePreview = ref<string | null>(null);

// Audio Recording State
const isRecording = ref(false);
const mediaRecorder = ref<MediaRecorder | null>(null);
const audioChunks = ref<Blob[]>([]);
const recordingDuration = ref(0);
const recordingTimer = ref<number | null>(null);
const recordedAudioBlob = ref<Blob | null>(null);
const recordedAudioUrl = ref<string | null>(null);

const scrollToBottom = async () => {
  await nextTick();
  if (chatContainer.value) {
    chatContainer.value.scrollTop = chatContainer.value.scrollHeight;
  }
};

onMounted(scrollToBottom);

const triggerFileInput = () => {
    fileInput.value?.click();
};

const handleFileChange = (event: Event) => {
    const target = event.target as HTMLInputElement;
    if (target.files && target.files[0]) {
        const file = target.files[0];
        selectedFile.value = file;
        
        // Create preview
        const reader = new FileReader();
        reader.onload = (e) => {
            selectedFilePreview.value = e.target?.result as string;
        };
        reader.readAsDataURL(file);
    }
};

const removeSelectedFile = () => {
    selectedFile.value = null;
    selectedFilePreview.value = null;
    if (fileInput.value) fileInput.value.value = "";
};

const removeRecordedAudio = () => {
    if (recordedAudioUrl.value) {
        URL.revokeObjectURL(recordedAudioUrl.value);
    }
    recordedAudioBlob.value = null;
    recordedAudioUrl.value = null;
};

const startRecording = async () => {
    try {
        const stream = await navigator.mediaDevices.getUserMedia({ audio: true });
        mediaRecorder.value = new MediaRecorder(stream);
        audioChunks.value = [];

        mediaRecorder.value.ondataavailable = (event) => {
            audioChunks.value.push(event.data);
        };

        mediaRecorder.value.onstop = () => {
             const audioBlob = new Blob(audioChunks.value, { type: 'audio/webm' });
             recordedAudioBlob.value = audioBlob;
             recordedAudioUrl.value = URL.createObjectURL(audioBlob);
             
             // Stop all tracks
             stream.getTracks().forEach(track => track.stop());
        };

        mediaRecorder.value.start();
        isRecording.value = true;
        removeSelectedFile(); // Clear file if recording starts
        recordingDuration.value = 0;
        recordingTimer.value = window.setInterval(() => {
            recordingDuration.value++;
        }, 1000);

    } catch (err) {
        console.error("Error accessing microphone:", err);
        alert("Gagal mengakses mikrofon. Pastikan izin diberikan.");
    }
};

const stopRecording = () => {
    if (mediaRecorder.value && isRecording.value) {
        mediaRecorder.value.stop();
        isRecording.value = false;
        if (recordingTimer.value) {
            clearInterval(recordingTimer.value);
            recordingTimer.value = null;
        }
    }
};

const formatDuration = (seconds: number) => {
    const mins = Math.floor(seconds / 60);
    const secs = seconds % 60;
    return `${mins}:${secs.toString().padStart(2, '0')}`;
};

const sendMessage = () => {
    if (userInput.value.trim() || selectedFile.value || recordedAudioBlob.value) {
        let fileToSend = selectedFile.value;
        if (recordedAudioBlob.value) {
            fileToSend = new File([recordedAudioBlob.value], "voice_message.webm", { type: 'audio/webm' });
        }
        sendMessageWithFile(fileToSend);
    }
};

const sendMessageWithFile = async (file: File | null) => {
  const text = userInput.value.trim();
  
  // Optimistic UI Update
  const newMessage: Message = {
    id: Date.now(),
    role: "user",
    content: text, // Only use text if present
    image: file && file.type.startsWith('image/') ? selectedFilePreview.value! : undefined,
    audio: file && file.type.startsWith('audio/') && recordedAudioUrl.value ? recordedAudioUrl.value : undefined,
    time: format(new Date(), "HH:mm")
  };
  messages.value.push(newMessage);

  userInput.value = "";
  removeSelectedFile();
  // Don't remove audio immediately if we want to keep it in preview until next record? 
  // No, usually we clear input after send.
  // Note: recordedAudioUrl.value acts as the preview URL for the message too because we just passed it to message.audio
  // However, we should probably NULL it out so the input preview disappears. 
  // But wait, if we null it out, the displayed message using it will break if it uses the same ref (it doesn't, we passed the string value).
  // BUT we should NOT revoke the URL yet if it's used in the message list.
  // Actually, we pass the string value `recordedAudioUrl.value`. If we later `removeRecordedAudio` which calls revoke, it MIGHT look broken if the browser implements it strictly.
  // For safety, we should probably generate a new object URL for the message or just accept it might break on refresh (which is fine).
  // But strictly, clearing the preview involves nulling the ref. 
  
  // To avoid revoking the URL used by the message, we simply won't call removeRecordedAudio() fully, or we just null the ref.
  recordedAudioBlob.value = null;
  recordedAudioUrl.value = null; 
  
  scrollToBottom();

  isTyping.value = true;

  try {
      const formData = new FormData();
      if (text) formData.append("message", text);
      if (file) formData.append("file", file);

      const token = localStorage.getItem("token"); 
      
      const response = await fetch(`${import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080'}/api/chat`, {
          method: "POST",
          headers: {
              "Authorization": `Bearer ${token}`
          },
          body: formData
      });

      if (!response.ok) {
          throw new Error("Failed to send message");
      }

      const data = await response.json();
      
      messages.value.push({
        id: Date.now() + 1,
        role: "assistant",
        content: data.response,
        time: format(new Date(), "HH:mm")
      });

  } catch (error) {
      console.error("Error sending message:", error);
      messages.value.push({
        id: Date.now() + 1,
        role: "assistant",
        content: "Maaf, terjadi kesalahan saat memproses pesanmu. Silakan coba lagi. 😥",
        time: format(new Date(), "HH:mm")
      });
  } finally {
      isTyping.value = false;
      scrollToBottom();
  }
};

onUnmounted(() => {
    if (recordingTimer.value) clearInterval(recordingTimer.value);
});
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
                        <div :class="['px-4 py-2.5 rounded-2xl text-sm shadow-sm leading-relaxed relative group whitespace-pre-wrap', 
                            msg.role === 'user' 
                                ? 'bg-primary text-primary-foreground rounded-tr-sm' 
                                : 'bg-card border border-border rounded-tl-sm']"
                        >
                            <div v-if="msg.image" class="mb-2">
                                <img :src="msg.image" class="rounded-lg max-h-60 w-auto object-contain border border-white/20" alt="Uploaded Image" />
                            </div>
                            <div v-if="msg.audio" class="mb-2">
                                <audio controls :src="msg.audio" class="max-w-[240px] h-10"></audio>
                            </div>
                            <span v-if="msg.content">{{ msg.content }}</span>
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
            
            <!-- Preview Selected File -->
            <div v-if="selectedFilePreview" class="mb-2 relative inline-block">
                <img :src="selectedFilePreview" class="h-20 w-auto rounded-lg border border-border shadow-sm object-cover" />
                <button @click="removeSelectedFile" class="absolute -top-1 -right-1 bg-destructive text-destructive-foreground rounded-full p-0.5 shadow-md">
                    <X class="w-3 h-3" />
                </button>
            </div>

            <!-- Preview Recorded Audio -->
             <div v-if="recordedAudioUrl" class="mb-2 relative inline-flex items-center bg-muted/40 px-3 py-2 rounded-xl border border-border gap-2">
                <audio controls :src="recordedAudioUrl" class="h-8 w-60"></audio>
                <button @click="removeRecordedAudio" class="bg-destructive text-destructive-foreground rounded-full p-1 shadow-sm hover:bg-destructive/90 ml-2" title="Hapus Rekaman">
                    <X class="w-3 h-3" />
                </button>
            </div>

            <div class="flex items-center gap-2 bg-muted/30 p-2 rounded-3xl border border-muted-foreground/10 focus-within:ring-1 focus-within:ring-emerald-500/50 transition-all shadow-sm">
                
                <input 
                    type="file" 
                    ref="fileInput" 
                    class="hidden" 
                    accept="image/*" 
                    @change="handleFileChange"
                    :disabled="isRecording || !!recordedAudioBlob"
                />

                <Input 
                    v-model="userInput" 
                    :placeholder="isRecording ? `Merekam... ${formatDuration(recordingDuration)}` : 'Ketik pesan...'" 
                    class="flex-1 border-none shadow-none focus-visible:ring-0 bg-transparent px-3 h-9 text-base md:text-sm"
                    :disabled="isTyping || isRecording"
                    @keydown.enter.prevent="sendMessage"
                />

                <div class="flex items-center gap-1 pr-1">
                    <Button 
                        v-if="!isRecording"
                        variant="ghost" 
                        size="icon" 
                        class="h-8 w-8 rounded-full text-muted-foreground hover:text-foreground shrink-0 hover:bg-background/80" 
                        title="Kirim Gambar"
                        @click="triggerFileInput"
                    >
                        <ImageIcon class="h-4 w-4" />
                    </Button>

                    <Button 
                        variant="ghost" 
                        size="icon" 
                        :class="[
                            'h-8 w-8 rounded-full shrink-0 transition-all',
                            isRecording ? 'bg-red-100 text-red-600 animate-pulse hover:bg-red-200' : 'text-muted-foreground hover:text-foreground hover:bg-background/80'
                        ]" 
                        title="Rekam Suara"
                        @click="isRecording ? stopRecording() : startRecording()"
                        :disabled="!!recordedAudioBlob"
                    >
                        <Loader2 v-if="isRecording" class="h-4 w-4 animate-spin" />
                        <Mic v-else class="h-4 w-4" />
                    </Button>
                </div>

                <Button 
                    @click="sendMessage"
                    size="icon" 
                    :disabled="(!userInput.trim() && !selectedFile && !recordedAudioBlob) || isTyping || isRecording"
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
