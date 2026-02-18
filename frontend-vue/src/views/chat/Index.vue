<script setup lang="ts">
import { ref, nextTick, onMounted } from "vue";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Card } from "@/components/ui/card";
import {
  Send,
  Bot,
  User,
  Sparkles,
  Image as ImageIcon,
  Mic,
  MicOff,
  X,
  Loader2,
  Play,
  Pause,
  Volume2,
  Trash2,
} from "lucide-vue-next";
import { format } from "date-fns";
import api from "@/lib/api";



interface SavedTransaction {
  id: number;
  description: string;
  amount: number;
  type: string;
  category_name: string;
  wallet_name: string;
}

interface Message {
  id: number;
  role: "user" | "assistant";
  content: string;
  time: string;
  imageUrl?: string;
  audioUrl?: string;
  voiceLabel?: string;
  transactions?: SavedTransaction[];
}

const messages = ref<Message[]>([
  {
    id: 1,
    role: "assistant",
    content:
      "Halo! Saya Cuan AI, asisten keuangan pribadimu. Kamu bisa kirim teks, foto struk, atau pesan suara! 🤖💰",
    time: format(new Date(), "HH:mm"),
  },
]);

const userInput = ref("");
const isTyping = ref(false);
const chatContainer = ref<HTMLElement | null>(null);

const imageFile = ref<File | null>(null);
const imagePreview = ref<string | null>(null);
const imageInput = ref<HTMLInputElement | null>(null);

const isRecording = ref(false);
const mediaRecorder = ref<MediaRecorder | null>(null);
const audioChunks = ref<Blob[]>([]);
const voiceFile = ref<Blob | null>(null);
const recordingDuration = ref(0);
const recordingTimer = ref<ReturnType<typeof setInterval> | null>(null);

const activeAudioId = ref<number | null>(null);
const activeAudio = ref<HTMLAudioElement | null>(null);
const audioProgress = ref(0);
const audioCurrent = ref(0);
const audioDurationVal = ref(0);
const audioPlaying = ref(false);

const scrollToBottom = async () => {
  await nextTick();
  if (chatContainer.value) {
    chatContainer.value.scrollTop = chatContainer.value.scrollHeight;
  }
};

onMounted(scrollToBottom);

const triggerImageUpload = () => {
  imageInput.value?.click();
};

const MAX_IMAGE_DIMENSION = 1024;
const IMAGE_QUALITY = 0.7;
const MAX_COMPRESSED_SIZE = 2 * 1024 * 1024; // 2MB after compression

const compressImage = (file: File): Promise<File> => {
  return new Promise((resolve, reject) => {
    const img = new Image();
    const url = URL.createObjectURL(file);

    img.onload = () => {
      URL.revokeObjectURL(url);

      let { width, height } = img;
      if (width > MAX_IMAGE_DIMENSION || height > MAX_IMAGE_DIMENSION) {
        if (width > height) {
          height = Math.round((height * MAX_IMAGE_DIMENSION) / width);
          width = MAX_IMAGE_DIMENSION;
        } else {
          width = Math.round((width * MAX_IMAGE_DIMENSION) / height);
          height = MAX_IMAGE_DIMENSION;
        }
      }

      const canvas = document.createElement("canvas");
      canvas.width = width;
      canvas.height = height;
      const ctx = canvas.getContext("2d")!;
      ctx.drawImage(img, 0, 0, width, height);

      canvas.toBlob(
        (blob) => {
          if (!blob) return reject(new Error("Gagal kompres gambar"));
          const compressed = new File([blob], file.name.replace(/\.\w+$/, ".jpg"), {
            type: "image/jpeg",
          });
          resolve(compressed);
        },
        "image/jpeg",
        IMAGE_QUALITY,
      );
    };

    img.onerror = () => {
      URL.revokeObjectURL(url);
      reject(new Error("Gagal membaca gambar"));
    };

    img.src = url;
  });
};

const onImageSelected = async (event: Event) => {
  const target = event.target as HTMLInputElement;
  const file = target.files?.[0];
  if (!file) return;

  try {
    const compressed = await compressImage(file);
    if (compressed.size > MAX_COMPRESSED_SIZE) {
      alert("Gambar terlalu besar. Mohon gunakan gambar yang lebih kecil.");
      return;
    }
    imageFile.value = compressed;
    const reader = new FileReader();
    reader.onload = (e) => {
      imagePreview.value = e.target?.result as string;
    };
    reader.readAsDataURL(compressed);
  } catch (err) {
    console.error("Image compression failed:", err);
    // Fallback: use original file
    imageFile.value = file;
    const reader = new FileReader();
    reader.onload = (e) => {
      imagePreview.value = e.target?.result as string;
    };
    reader.readAsDataURL(file);
  }
};

const clearImage = () => {
  imageFile.value = null;
  imagePreview.value = null;
  if (imageInput.value) imageInput.value.value = "";
};

const toggleRecording = async () => {
  if (isRecording.value) {
    stopRecording();
  } else {
    await startRecording();
  }
};

const startRecording = async () => {
  try {
    const stream = await navigator.mediaDevices.getUserMedia({ audio: true });
    const recorder = new MediaRecorder(stream);
    audioChunks.value = [];

    recorder.ondataavailable = (e) => {
      if (e.data.size > 0) audioChunks.value.push(e.data);
    };

    recorder.onstop = () => {
      const blob = new Blob(audioChunks.value, { type: "audio/webm" });
      voiceFile.value = blob;
      if (previewVoiceUrl.value) URL.revokeObjectURL(previewVoiceUrl.value);
      previewVoiceUrl.value = URL.createObjectURL(blob);
      stream.getTracks().forEach((track) => track.stop());
    };

    recorder.start();
    mediaRecorder.value = recorder;
    isRecording.value = true;
    recordingDuration.value = 0;
    recordingTimer.value = setInterval(() => {
      recordingDuration.value++;
    }, 1000);
  } catch {
    console.error("Microphone access denied");
  }
};

const stopRecording = () => {
  if (mediaRecorder.value && mediaRecorder.value.state !== "inactive") {
    mediaRecorder.value.stop();
  }
  isRecording.value = false;
  if (recordingTimer.value) {
    clearInterval(recordingTimer.value);
    recordingTimer.value = null;
  }
};

const clearVoice = () => {
  stopPreviewPlayback();
  if (previewVoiceUrl.value) {
    URL.revokeObjectURL(previewVoiceUrl.value);
    previewVoiceUrl.value = null;
  }
  voiceFile.value = null;
  recordingDuration.value = 0;
};

const previewVoiceUrl = ref<string | null>(null);
const isPreviewPlaying = ref(false);
const previewAudio = ref<HTMLAudioElement | null>(null);
const previewProgress = ref(0);
const previewCurrent = ref(0);
const previewDurationVal = ref(0);

const togglePreviewPlayback = () => {
  if (!previewVoiceUrl.value) return;

  if (isPreviewPlaying.value) {
    stopPreviewPlayback();
  } else {
    const audio = new Audio(previewVoiceUrl.value);
    audio.onloadedmetadata = () => {
      previewDurationVal.value = audio.duration;
    };
    audio.ontimeupdate = () => {
      if (audio.duration) {
        previewProgress.value = (audio.currentTime / audio.duration) * 100;
        previewCurrent.value = audio.currentTime;
      }
    };
    audio.onended = () => {
      isPreviewPlaying.value = false;
      previewAudio.value = null;
      previewProgress.value = 0;
      previewCurrent.value = 0;
    };
    audio.play();
    previewAudio.value = audio;
    isPreviewPlaying.value = true;
  }
};

const stopPreviewPlayback = () => {
  if (previewAudio.value) {
    previewAudio.value.pause();
    previewAudio.value = null;
  }
  isPreviewPlaying.value = false;
  previewProgress.value = 0;
  previewCurrent.value = 0;
};

const seekPreviewAudio = (event: MouseEvent) => {
  if (!previewAudio.value) return;
  const target = event.currentTarget as HTMLElement;
  const rect = target.getBoundingClientRect();
  const x = event.clientX - rect.left;
  const pct = x / rect.width;
  previewAudio.value.currentTime = pct * previewAudio.value.duration;
};

const toggleMessageAudio = (msgId: number, url: string) => {
  if (activeAudioId.value === msgId && audioPlaying.value) {
    activeAudio.value?.pause();
    audioPlaying.value = false;
    return;
  }

  if (activeAudio.value) {
    activeAudio.value.pause();
    activeAudio.value = null;
  }

  const audio = new Audio(url);
  activeAudioId.value = msgId;
  activeAudio.value = audio;
  audioProgress.value = 0;
  audioCurrent.value = 0;
  audioDurationVal.value = 0;
  audioPlaying.value = true;

  audio.onloadedmetadata = () => {
    audioDurationVal.value = audio.duration;
  };

  audio.ontimeupdate = () => {
    if (audio.duration) {
      audioProgress.value = (audio.currentTime / audio.duration) * 100;
      audioCurrent.value = audio.currentTime;
    }
  };

  audio.onended = () => {
    audioPlaying.value = false;
    audioProgress.value = 0;
    audioCurrent.value = 0;
    activeAudioId.value = null;
    activeAudio.value = null;
  };

  audio.play();
};

const seekAudio = (event: MouseEvent, msgId: number) => {
  if (activeAudioId.value !== msgId || !activeAudio.value) return;
  const target = event.currentTarget as HTMLElement;
  const rect = target.getBoundingClientRect();
  const x = event.clientX - rect.left;
  const pct = x / rect.width;
  activeAudio.value.currentTime = pct * activeAudio.value.duration;
};

const formatTime = (seconds: number): string => {
  const m = Math.floor(seconds / 60);
  const s = Math.floor(seconds % 60);
  return `${m}:${s.toString().padStart(2, "0")}`;
};

const formatDuration = (seconds: number): string => {
  const m = Math.floor(seconds / 60);
  const s = seconds % 60;
  return `${m}:${s.toString().padStart(2, "0")}`;
};

const canSend = () => {
  return (
    (userInput.value.trim() || imageFile.value || voiceFile.value) && !isTyping.value
  );
};

const sendMessage = async () => {
  if (!canSend()) return;

  const text = userInput.value.trim();

  const userMsg: Message = {
    id: Date.now(),
    role: "user",
    content: text,
    time: format(new Date(), "HH:mm"),
  };
  if (imagePreview.value) {
    userMsg.imageUrl = imagePreview.value;
  }
  if (voiceFile.value && previewVoiceUrl.value) {
    userMsg.audioUrl = previewVoiceUrl.value;
    userMsg.voiceLabel = formatDuration(recordingDuration.value);
  }
  messages.value.push(userMsg);

  const formData = new FormData();
  if (text) formData.append("message", text);
  if (imageFile.value) formData.append("image", imageFile.value);
  if (voiceFile.value) formData.append("voice", voiceFile.value, "voice.webm");

  userInput.value = "";
  clearImage();
  voiceFile.value = null;
  previewVoiceUrl.value = null;
  recordingDuration.value = 0;
  stopPreviewPlayback();
  scrollToBottom();

  isTyping.value = true;
  try {
    const response = await api.post("/api/ai/chat", formData, {
      headers: { "Content-Type": "multipart/form-data" },
      timeout: 180000,
    });

    isTyping.value = false;

    const assistantMsg: Message = {
      id: Date.now() + 1,
      role: "assistant",
      content: response.data.reply,
      time: format(new Date(), "HH:mm"),
    };
    if (response.data.transactions?.length) {
      assistantMsg.transactions = response.data.transactions;
    }
    messages.value.push(assistantMsg);
  } catch (error: any) {
    isTyping.value = false;
    const errMsg =
      error.response?.data?.error || "Terjadi kesalahan. Coba lagi nanti.";
    messages.value.push({
      id: Date.now() + 1,
      role: "assistant",
      content: `⚠️ ${errMsg}`,
      time: format(new Date(), "HH:mm"),
    });
  }

  scrollToBottom();
};
</script>

<template>
  <div class="flex flex-col h-[calc(100vh-2rem)] md:h-[calc(100vh-3.5rem)] max-w-5xl mx-auto p-4 md:p-6 space-y-4">
    <!-- Header -->
    <div class="flex items-center gap-4 pb-4 border-b border-border">
      <div
        class="h-12 w-12 shrink-0 rounded-2xl bg-gradient-to-br from-emerald-500 to-teal-600 flex items-center justify-center shadow-lg text-white">
        <Bot class="h-6 w-6" />
      </div>
      <div>
        <h2 class="text-xl font-bold tracking-tight flex items-center gap-2">
          Cuan AI
          <span
            class="text-[10px] bg-emerald-100/50 dark:bg-emerald-900/30 text-emerald-600 dark:text-emerald-400 px-2 py-0.5 rounded-full border border-emerald-200/50 flex items-center gap-1">
            <Sparkles class="w-3 h-3" /> Beta
          </span>
        </h2>
        <p class="text-sm text-muted-foreground">
          Tanya tips keuangan, kirim foto struk, atau rekam suara.
        </p>
      </div>
    </div>

    <Card class="flex-1 overflow-hidden flex flex-col bg-muted/30 border-border shadow-sm rounded-3xl relative">
      <!-- Messages -->
      <div ref="chatContainer" class="flex-1 overflow-y-auto p-4 space-y-4 custom-scrollbar scroll-smooth">
        <div v-for="msg in messages" :key="msg.id" class="flex w-full">
          <div :class="[
            'flex max-w-[80%] md:max-w-[70%] gap-2',
            msg.role === 'user'
              ? 'ml-auto flex-row-reverse'
              : 'mr-auto flex-row',
          ]">
            <!-- Avatar -->
            <div v-if="msg.role === 'assistant'"
              class="h-8 w-8 rounded-full bg-emerald-100 dark:bg-emerald-900/30 flex items-center justify-center shrink-0 border border-emerald-200/50">
              <Bot class="h-4 w-4 text-emerald-600 dark:text-emerald-400" />
            </div>
            <div v-else
              class="h-8 w-8 rounded-full bg-gradient-to-br from-emerald-500 to-teal-600 flex items-center justify-center shrink-0 text-white shadow-sm">
              <User class="h-4 w-4" />
            </div>

            <div :class="[
              'flex flex-col',
              msg.role === 'user' ? 'items-end' : 'items-start',
            ]">
              <!-- Image preview in bubble -->
              <img v-if="msg.imageUrl" :src="msg.imageUrl"
                class="max-w-[200px] rounded-xl mb-1 shadow-sm border border-border" alt="Uploaded image" />

              <!-- Audio player -->
              <div v-if="msg.audioUrl" class="mb-1 w-full min-w-[240px]">
                <div :class="[
                  'flex items-center gap-2.5 px-3 py-2 rounded-2xl',
                  msg.role === 'user'
                    ? 'bg-emerald-600 text-white'
                    : 'bg-card border border-border',
                ]">
                  <!-- Play/Pause button -->
                  <button @click="toggleMessageAudio(msg.id, msg.audioUrl!)"
                    class="h-8 w-8 rounded-full flex items-center justify-center shrink-0 transition-colors"
                    :class="msg.role === 'user'
                      ? 'bg-white/20 hover:bg-white/30 text-white'
                      : 'bg-emerald-100 dark:bg-emerald-900/30 text-emerald-600 dark:text-emerald-400 hover:bg-emerald-200'">
                    <Pause v-if="activeAudioId === msg.id && audioPlaying" class="h-3.5 w-3.5" />
                    <Play v-else class="h-3.5 w-3.5 ml-0.5" />
                  </button>

                  <!-- Progress bar + time -->
                  <div class="flex-1 flex flex-col gap-1">
                    <div class="relative h-1 rounded-full cursor-pointer"
                      :class="msg.role === 'user' ? 'bg-white/20' : 'bg-muted'" @click="seekAudio($event, msg.id)">
                      <div class="absolute h-full rounded-full transition-all"
                        :class="msg.role === 'user' ? 'bg-white/70' : 'bg-emerald-500'"
                        :style="{ width: activeAudioId === msg.id ? audioProgress + '%' : '0%' }"></div>
                    </div>
                    <div class="flex justify-between text-[10px]"
                      :class="msg.role === 'user' ? 'text-white/60' : 'text-muted-foreground'">
                      <span>{{ activeAudioId === msg.id ? formatTime(audioCurrent) : '0:00' }}</span>
                      <span>{{ activeAudioId === msg.id && audioDurationVal ? formatTime(audioDurationVal) :
                        msg.voiceLabel || '0:00' }}</span>
                    </div>
                  </div>

                  <!-- Volume icon -->
                  <Volume2 class="h-3.5 w-3.5 shrink-0 opacity-50" />
                </div>
              </div>

              <!-- Voice label (fallback if no audioUrl) -->
              <span v-else-if="msg.voiceLabel" class="text-xs text-muted-foreground mb-1 italic">
                {{ msg.voiceLabel }}
              </span>

              <!-- Text bubble -->
              <div v-if="msg.content" :class="[
                'px-4 py-2.5 rounded-2xl text-sm shadow-sm leading-relaxed relative group whitespace-pre-wrap',
                msg.role === 'user'
                  ? 'bg-gradient-to-br from-emerald-500 to-teal-600 text-white rounded-tr-sm'
                  : 'bg-card border border-border rounded-tl-sm',
              ]">
                {{ msg.content }}
              </div>

              <!-- Transaction cards -->
              <div v-if="msg.transactions?.length" class="mt-2 space-y-2 w-full">
                <div v-for="tx in msg.transactions" :key="tx.id"
                  class="bg-emerald-50 dark:bg-emerald-900/20 border border-emerald-200 dark:border-emerald-800/50 rounded-xl px-3 py-2 text-xs">
                  <div class="flex items-center justify-between">
                    <span class="font-semibold text-emerald-700 dark:text-emerald-300">✅ {{ tx.description }}</span>
                    <span :class="[
                      'font-bold',
                      tx.type === 'income' ? 'text-emerald-600' : 'text-red-500'
                    ]">
                      {{ tx.type === 'income' ? '+' : '-' }}Rp{{ Number(tx.amount).toLocaleString('id-ID') }}
                    </span>
                  </div>
                  <div class="flex items-center gap-2 mt-1 text-muted-foreground">
                    <span>🏦 {{ tx.wallet_name }}</span>
                    <span>•</span>
                    <span>📂 {{ tx.category_name }}</span>
                    <span>•</span>
                    <span class="px-1.5 py-0.5 rounded text-[10px] font-medium"
                      :class="tx.type === 'income' ? 'bg-emerald-100 dark:bg-emerald-900/40 text-emerald-700 dark:text-emerald-300' : 'bg-red-100 dark:bg-red-900/40 text-red-700 dark:text-red-300'">
                      {{ tx.type === 'income' ? 'Pemasukan' : 'Pengeluaran' }}
                    </span>
                  </div>
                </div>
              </div>

              <span class="text-[10px] text-muted-foreground mt-1 opacity-70 px-1">
                {{ msg.time }}
              </span>
            </div>
          </div>
        </div>

        <!-- Typing indicator -->
        <div v-if="isTyping" class="flex w-full">
          <div class="flex max-w-[80%] gap-2 mr-auto flex-row">
            <div
              class="h-8 w-8 rounded-full bg-emerald-100 dark:bg-emerald-900/30 flex items-center justify-center shrink-0 border border-emerald-200/50">
              <Bot class="h-4 w-4 text-emerald-600 dark:text-emerald-400" />
            </div>
            <div
              class="bg-card border border-border px-4 py-3 rounded-2xl rounded-tl-sm flex items-center gap-1.5 h-10 shadow-sm">
              <Loader2 class="h-4 w-4 text-emerald-500 animate-spin" />
              <span class="text-xs text-muted-foreground">Sedang berpikir...</span>
            </div>
          </div>
        </div>
      </div>

      <!-- Attachment preview bar -->
      <div v-if="imagePreview && !voiceFile"
        class="px-4 pt-2 pb-1 bg-card border-t border-border flex items-center gap-2 flex-wrap">
        <!-- Image preview -->
        <div class="relative group">
          <img :src="imagePreview" class="h-16 w-16 object-cover rounded-lg border border-border shadow-sm" />
          <button @click="clearImage"
            class="absolute -top-1 -right-1 bg-destructive text-destructive-foreground rounded-full p-0.5 opacity-0 group-hover:opacity-100 transition-opacity">
            <X class="h-3 w-3" />
          </button>
        </div>
      </div>

      <!-- Input bar -->
      <div class="p-3 md:p-4 bg-card border-t border-border">
        <input ref="imageInput" type="file" accept="image/*" class="hidden" @change="onImageSelected" />

        <!-- Voice player bar (replaces input when recording exists) -->
        <div v-if="voiceFile"
          class="flex items-center gap-2 bg-muted/30 p-2 rounded-3xl border border-muted-foreground/10 shadow-sm">
          <!-- Play/Pause -->
          <button @click="togglePreviewPlayback"
            class="h-9 w-9 rounded-full flex items-center justify-center shrink-0 transition-colors bg-emerald-100 dark:bg-emerald-900/30 text-emerald-600 dark:text-emerald-400 hover:bg-emerald-200 dark:hover:bg-emerald-900/50">
            <Pause v-if="isPreviewPlaying" class="h-4 w-4" />
            <Play v-else class="h-4 w-4 ml-0.5" />
          </button>

          <!-- Time -->
          <span class="text-xs text-muted-foreground tabular-nums w-9 text-center shrink-0">
            {{ isPreviewPlaying || previewCurrent > 0 ? formatTime(previewCurrent) : '0:00' }}
          </span>

          <!-- Progress bar -->
          <div class="flex-1 relative h-1.5 rounded-full bg-muted cursor-pointer" @click="seekPreviewAudio($event)">
            <div class="absolute h-full rounded-full bg-emerald-500 transition-all"
              :style="{ width: previewProgress + '%' }"></div>
            <div class="absolute top-1/2 -translate-y-1/2 h-3 w-3 rounded-full bg-emerald-500 shadow-sm transition-all"
              :style="{ left: previewProgress + '%' }" v-show="previewProgress > 0"></div>
          </div>

          <!-- Duration -->
          <span class="text-xs text-muted-foreground tabular-nums w-9 text-center shrink-0">
            {{ previewDurationVal ? formatTime(previewDurationVal) : formatDuration(recordingDuration) }}
          </span>

          <!-- Volume icon -->
          <Volume2 class="h-4 w-4 shrink-0 text-muted-foreground/50" />

          <!-- Divider -->
          <div class="w-px h-6 bg-border"></div>

          <!-- Delete -->
          <button @click="clearVoice"
            class="h-9 w-9 rounded-full flex items-center justify-center shrink-0 transition-colors bg-red-100 dark:bg-red-900/30 text-red-500 hover:bg-red-200 dark:hover:bg-red-900/50">
            <Trash2 class="h-4 w-4" />
          </button>

          <!-- Send -->
          <Button @click="sendMessage" size="icon" :disabled="!canSend()"
            class="h-9 w-9 rounded-full bg-emerald-600 hover:bg-emerald-700 text-white shadow-sm shrink-0 transition-all active:scale-95">
            <Send class="h-4 w-4 ml-0.5" />
          </Button>
        </div>

        <!-- Normal input bar -->
        <div v-else
          class="flex items-center gap-2 bg-muted/30 p-2 rounded-3xl border border-muted-foreground/10 focus-within:ring-1 focus-within:ring-emerald-500/50 transition-all shadow-sm">
          <Input v-model="userInput" placeholder="Ketik pesan..."
            class="flex-1 border-none shadow-none focus-visible:ring-0 bg-transparent px-3 h-9 text-base md:text-sm"
            :disabled="isTyping || isRecording" @keydown.enter.prevent="sendMessage" />

          <div class="flex items-center gap-1 pr-1">
            <Button variant="ghost" size="icon"
              class="h-8 w-8 rounded-full text-muted-foreground hover:text-foreground shrink-0 hover:bg-background/80"
              title="Kirim Gambar" :disabled="isTyping || isRecording" @click="triggerImageUpload">
              <ImageIcon class="h-4 w-4" />
            </Button>

            <Button variant="ghost" size="icon" :class="[
              'h-8 w-8 rounded-full shrink-0 transition-all',
              isRecording
                ? 'bg-red-100 dark:bg-red-900/30 text-red-500 hover:bg-red-200 animate-pulse'
                : 'text-muted-foreground hover:text-foreground hover:bg-background/80',
            ]" :title="isRecording ? 'Stop Rekam' : 'Rekam Suara'" :disabled="isTyping" @click="toggleRecording">
              <MicOff v-if="isRecording" class="h-4 w-4" />
              <Mic v-else class="h-4 w-4" />
            </Button>
          </div>

          <Button @click="sendMessage" size="icon" :disabled="!canSend()"
            class="h-9 w-9 rounded-full bg-emerald-600 hover:bg-emerald-700 text-white shadow-sm shrink-0 transition-all active:scale-95">
            <Send class="h-4 w-4 ml-0.5" />
          </Button>
        </div>

        <!-- Recording indicator -->
        <div v-if="isRecording" class="flex items-center justify-center gap-2 mt-2 text-red-500 text-xs">
          <span class="w-2 h-2 bg-red-500 rounded-full animate-pulse"></span>
          Merekam... {{ formatDuration(recordingDuration) }}
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
