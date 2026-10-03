<script setup lang="ts">
import { ref, reactive, nextTick, onMounted, onUnmounted } from "vue";
import { fetchEventSource } from "@microsoft/fetch-event-source";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Card } from "@/components/ui/card";
import {
  Send,
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
import { useSwal } from "@/composables/useSwal";

const swal = useSwal();



interface SavedTransaction {
  id: number;
  action?: string;
  description: string;
  amount: number;
  type: string;
  category_name: string;
  wallet_name: string;
  to_wallet_name?: string;
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

const messages = ref<Message[]>([]);
const isLoadingHistory = ref(false);

const userInput = ref("");
const isTyping = ref(false);
const typingStatus = ref("Sedang berpikir...");
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

let isScrollPending = false;
const scrollToBottom = async () => {
  await nextTick();
  if (chatContainer.value) {
    chatContainer.value.scrollTop = chatContainer.value.scrollHeight;
  }
};

const throttledScrollToBottom = () => {
  if (!isScrollPending && chatContainer.value) {
    const { scrollTop, scrollHeight, clientHeight } = chatContainer.value;
    if (scrollHeight - scrollTop - clientHeight < 150) {
      isScrollPending = true;
      requestAnimationFrame(async () => {
        await scrollToBottom();
        isScrollPending = false;
      });
    }
  }
};

onMounted(async () => {
  await loadHistory();
  scrollToBottom();
});

const loadHistory = async () => {
  isLoadingHistory.value = true;
  try {
    const token = localStorage.getItem("token");
    const res = await fetch(
      import.meta.env.VITE_API_BASE_URL + "api/ai/chat/history?limit=100",
      { headers: { Authorization: `Bearer ${token}` } }
    );
    if (!res.ok) throw new Error("Gagal memuat history");
    const data: Array<{
      id: number;
      role: "user" | "assistant";
      content: string;
      audio_url?: string;
      image_url?: string;
      created_at: string;
      transactions?: SavedTransaction[];
    }> = await res.json();

    if (data && data.length > 0) {
      messages.value = data.map((m) => ({
        id: m.id,
        role: m.role,
        content: m.content,
        time: format(new Date(m.created_at), "HH:mm"),
        audioUrl: m.audio_url || undefined,
        imageUrl: m.image_url || undefined,
        transactions: m.transactions?.length ? m.transactions : undefined,
      }));
    } else {
      // Pesan sambutan untuk percakapan baru
      messages.value = [
        {
          id: 1,
          role: "assistant",
          content:
            "Halo! Saya Cuan AI, asisten keuangan pribadimu. Kamu bisa kirim teks, foto struk, atau pesan suara! 🤖💰",
          time: format(new Date(), "HH:mm"),
        },
      ];
    }
  } catch {
    messages.value = [
      {
        id: 1,
        role: "assistant",
        content:
          "Halo! Saya Cuan AI, asisten keuangan pribadimu. Kamu bisa kirim teks, foto struk, atau pesan suara! 🤖💰",
        time: format(new Date(), "HH:mm"),
      },
    ];
  } finally {
    isLoadingHistory.value = false;
  }
};

const clearHistory = async () => {
  const confirmed = await swal.confirmDelete('riwayat percakapan');
  if (!confirmed) return;

  try {
    const token = localStorage.getItem("token");
    await fetch(import.meta.env.VITE_API_BASE_URL + "api/ai/chat/history", {
      method: "DELETE",
      headers: { Authorization: `Bearer ${token}` },
    });
    messages.value = [
      {
        id: Date.now(),
        role: "assistant",
        content:
          "Riwayat percakapan telah dihapus. Halo lagi! Saya siap membantu keuanganmu. 🤖💰",
        time: format(new Date(), "HH:mm"),
      },
    ];
    swal.success('Riwayat percakapan berhasil dihapus');
  } catch {
    swal.error('Gagal', 'Gagal menghapus riwayat percakapan.');
  }
};

onUnmounted(() => {
  if (previewVoiceUrl.value) {
    URL.revokeObjectURL(previewVoiceUrl.value);
  }
  if (imagePreview.value && imagePreview.value.startsWith("blob:")) {
    URL.revokeObjectURL(imagePreview.value);
  }

  if (activeAudio.value) {
    activeAudio.value.pause();
    activeAudio.value = null;
  }
  if (previewAudio.value) {
    previewAudio.value.pause();
    previewAudio.value = null;
  }
  if (mediaRecorder.value && mediaRecorder.value.state !== "inactive") {
    mediaRecorder.value.stop();
  }
  if (recordingTimer.value) {
    clearInterval(recordingTimer.value);
  }
});

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
  typingStatus.value = "Sedang berpikir...";

  try {
    const token = localStorage.getItem("token");

    const assistantMsg: Message = reactive({
      id: Date.now() + 1,
      role: "assistant",
      content: "",
      time: format(new Date(), "HH:mm"),
    });
    let isMessagePushed = false;

    await fetchEventSource(import.meta.env.VITE_API_BASE_URL + "api/ai/chat/stream", {
      method: "POST",
      headers: { Authorization: `Bearer ${token}` },
      body: formData,
      async onopen(response) {
        if (!response.ok) {
          throw new Error("Gagal memulai streaming");
        }
      },
      onmessage(msg) {
        if (msg.event === "status") {
          typingStatus.value = msg.data;
          isTyping.value = true;
          throttledScrollToBottom();
        } else if (msg.event === "token") {
          try {
            if (!isMessagePushed) {
              messages.value.push(assistantMsg);
              isMessagePushed = true;
              isTyping.value = false;
              scrollToBottom();
            }

            const parsed = JSON.parse(msg.data);
            assistantMsg.content += parsed.content;

            throttledScrollToBottom();
          } catch (e) { }
        } else if (msg.event === "done") {
          try {
            const result = JSON.parse(msg.data);

            if (!isMessagePushed) {
              messages.value.push(assistantMsg);
              isMessagePushed = true;
            }

            if (result.transactions?.length) {
              assistantMsg.transactions = result.transactions;
            }
            if (result.reply) {
              assistantMsg.content = result.reply;
            }

            scrollToBottom();
          } catch (e) { }
          isTyping.value = false;
        } else if (msg.event === "error") {
          if (!isMessagePushed) {
            messages.value.push(assistantMsg);
            isMessagePushed = true;
          }
          try {
            const parsed = JSON.parse(msg.data);
            assistantMsg.content += `\n⚠️ Error: ${parsed.error || parsed || msg.data}`;
          } catch (e) {
            assistantMsg.content += `\n⚠️ Error: ${msg.data}`;
          }
          isTyping.value = false;
        }
      },
      onerror(err) {
        throw err; // Lempar ke catch agar tidak retry berulang kali
      },
    });

  } catch (error: any) {
    isTyping.value = false;
    const errMsg = error.message || "Terjadi kesalahan koneksi.";
    messages.value.push({
      id: Date.now() + 1,
      role: "assistant",
      content: `⚠️ ${errMsg}`,
      time: format(new Date(), "HH:mm"),
    });
  }

  scrollToBottom();
};

const getMediaUrl = (url?: string) => {
  if (!url) return "";
  if (url.startsWith("blob:") || url.startsWith("http") || url.startsWith("data:")) return url;
  return import.meta.env.VITE_API_BASE_URL + url;
};
</script>

<template>
  <div class="flex flex-col h-[calc(100vh-2rem)] md:h-[calc(100vh-3.5rem)] max-w-5xl mx-auto p-4 md:p-6 space-y-4">
    <!-- Header -->
    <div class="flex items-center gap-4 pb-4 border-b border-border">
      <div
        class="overflow-hidden border shadow-lg w-14 h-14 shrink-0 rounded-2xl border-emerald-500/20">
        <img src="/cuan-bot.svg" alt="Cuan AI" class="object-cover w-full h-full" />
      </div>
      <div class="flex-1">
        <h2 class="flex items-center gap-2 text-xl font-bold tracking-tight">
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
      <!-- Clear history button -->
      <Button v-if="messages.length > 0 && !isLoadingHistory" variant="ghost" size="icon"
        title="Hapus riwayat percakapan"
        class="transition-colors rounded-full h-9 w-9 text-muted-foreground hover:text-destructive hover:bg-destructive/10 shrink-0"
        @click="clearHistory">
        <Trash2 class="w-4 h-4" />
      </Button>
      <!-- Loading history indicator -->
      <div v-if="isLoadingHistory" class="flex items-center gap-2 text-xs text-muted-foreground">
        <Loader2 class="w-4 h-4 animate-spin text-emerald-500" />
        <span>Memuat...</span>
      </div>
    </div>

    <Card class="relative flex flex-col flex-1 overflow-hidden shadow-sm bg-muted/30 border-border rounded-3xl">
      <!-- Messages -->
      <div ref="chatContainer" class="flex-1 p-4 space-y-4 overflow-y-auto custom-scrollbar scroll-smooth">
        <div v-for="msg in messages" :key="msg.id" class="flex w-full">
          <div :class="[
            'flex max-w-[80%] md:max-w-[70%] gap-2',
            msg.role === 'user'
              ? 'ml-auto flex-row-reverse'
              : 'mr-auto flex-row',
          ]">
            <!-- Avatar -->
            <div v-if="msg.role === 'assistant'"
              class="w-10 h-10 overflow-hidden border rounded-full shadow-xs shrink-0 border-emerald-200/50">
              <img src="/cuan-bot.svg" alt="Cuan AI" class="object-cover w-full h-full" />
            </div>
            <div v-else
              class="w-10 h-10 overflow-hidden border rounded-full shadow-xs shrink-0 border-emerald-500/20">
              <img src="/people.svg" alt="User" class="object-cover w-full h-full" />
            </div>

            <div :class="[
              'flex flex-col',
              msg.role === 'user' ? 'items-end' : 'items-start',
            ]">
              <!-- Image preview in bubble -->
              <img v-if="msg.imageUrl" :src="getMediaUrl(msg.imageUrl)"
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
                  <button @click="toggleMessageAudio(msg.id, getMediaUrl(msg.audioUrl!))"
                    class="flex items-center justify-center w-8 h-8 transition-colors rounded-full shrink-0"
                    :class="msg.role === 'user'
                      ? 'bg-white/20 hover:bg-white/30 text-white'
                      : 'bg-emerald-100 dark:bg-emerald-900/30 text-emerald-600 dark:text-emerald-400 hover:bg-emerald-200'">
                    <Pause v-if="activeAudioId === msg.id && audioPlaying" class="h-3.5 w-3.5" />
                    <Play v-else class="h-3.5 w-3.5 ml-0.5" />
                  </button>

                  <!-- Progress bar + time -->
                  <div class="flex flex-col flex-1 gap-1">
                    <div class="relative h-1 rounded-full cursor-pointer"
                      :class="msg.role === 'user' ? 'bg-white/20' : 'bg-muted'" @click="seekAudio($event, msg.id)">
                      <div class="absolute h-full transition-all rounded-full"
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
              <span v-else-if="msg.voiceLabel" class="mb-1 text-xs italic text-muted-foreground">
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
              <div v-if="msg.transactions?.length" class="w-full mt-2 space-y-2">
                <div v-for="tx in msg.transactions" :key="tx.id"
                  class="px-3 py-2 text-xs border rounded-xl"
                  :class="[
                    tx.action?.includes('delete')
                      ? 'bg-red-50 dark:bg-red-900/20 border-red-200 dark:border-red-800/50'
                      : tx.action?.includes('update')
                        ? 'bg-blue-50 dark:bg-blue-900/20 border-blue-200 dark:border-blue-800/50'
                        : 'bg-emerald-50 dark:bg-emerald-900/20 border-emerald-200 dark:border-emerald-800/50'
                  ]">
                  <div class="flex items-center justify-between">
                    <span class="font-semibold" :class="{
                      'text-emerald-700 dark:text-emerald-300': !tx.action?.includes('delete') && !tx.action?.includes('update'),
                      'text-blue-700 dark:text-blue-300': tx.action?.includes('update'),
                      'text-red-700 dark:text-red-300 line-through': tx.action?.includes('delete')
                    }">
                      {{ tx.action?.includes('delete') ? '🗑️' : tx.action?.includes('update') ? '✏️' : '✅' }} 
                      {{ tx.description?.replace(/~~/g, '').replace(/\(Dihapus\)/g, '').trim() || '-' }}
                    </span>
                    <span :class="[
                      'font-bold',
                      tx.action?.includes('delete') ? 'line-through opacity-60' : '',
                      (tx.type === 'income' || tx.type === 'receivable') ? 'text-emerald-600' : 
                      (tx.type === 'transfer' || tx.type === 'goal' || tx.type === 'wishlist') ? 'text-blue-500' : 'text-red-500'
                    ]">
                      {{ (tx.type === 'income' || tx.type === 'receivable') ? '+' : (tx.type === 'transfer' || tx.type === 'goal' || tx.type === 'wishlist') ? '' : '-' }}Rp{{ Number(tx.amount || 0).toLocaleString('id-ID') }}
                    </span>
                  </div>
                  <div class="flex items-center gap-2 mt-1 text-muted-foreground">
                    <span v-if="tx.type === 'transfer' && tx.to_wallet_name">
                      🏦 {{ tx.wallet_name }} ➡️ {{ tx.to_wallet_name }}
                    </span>
                    <span v-else>
                      🏦 {{ tx.wallet_name || '-' }}
                    </span>
                    <span>•</span>
                    <span>📂 {{ tx.category_name || '-' }}</span>
                    <span>•</span>
                    <span class="px-1.5 py-0.5 rounded text-[10px] font-medium"
                      :class="[
                        tx.type === 'income' || tx.type === 'receivable'
                          ? 'bg-emerald-100 dark:bg-emerald-900/40 text-emerald-700 dark:text-emerald-300' 
                          : tx.type === 'transfer' || tx.type === 'goal' || tx.type === 'wishlist'
                            ? 'bg-blue-100 dark:bg-blue-900/40 text-blue-700 dark:text-blue-300'
                            : 'bg-red-100 dark:bg-red-900/40 text-red-700 dark:text-red-300'
                      ]">
                      {{ 
                        tx.type === 'income' ? 'Pemasukan' : 
                        tx.type === 'expense' ? 'Pengeluaran' : 
                        tx.type === 'transfer' ? 'Transfer' : 
                        tx.type === 'debt' ? 'Utang' : 
                        tx.type === 'receivable' ? 'Piutang' : 
                        tx.type === 'goal' ? 'Tabungan' : 
                        tx.type === 'wishlist' ? 'Wishlist' : tx.type
                      }}
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
              class="w-10 h-10 overflow-hidden border rounded-full shadow-xs shrink-0 border-emerald-200/50">
              <img src="/cuan-bot.svg" alt="Cuan AI" class="object-cover w-full h-full" />
            </div>
            <div
              class="bg-card border border-border px-4 py-3 rounded-2xl rounded-tl-sm flex items-center gap-1.5 h-10 shadow-sm">
              <Loader2 class="w-4 h-4 text-emerald-500 animate-spin" />
              <span class="text-xs text-muted-foreground">{{ typingStatus }}</span>
            </div>
          </div>
        </div>
      </div>

      <!-- Attachment preview bar -->
      <div v-if="imagePreview && !voiceFile"
        class="flex flex-wrap items-center gap-2 px-4 pt-2 pb-1 border-t bg-card border-border">
        <!-- Image preview -->
        <div class="relative group">
          <img :src="imagePreview" class="object-cover w-16 h-16 border rounded-lg shadow-sm border-border" />
          <button @click="clearImage"
            class="absolute -top-1 -right-1 bg-destructive text-destructive-foreground rounded-full p-0.5 opacity-0 group-hover:opacity-100 transition-opacity">
            <X class="w-3 h-3" />
          </button>
        </div>
      </div>

      <!-- Input bar -->
      <div class="p-3 border-t md:p-4 bg-card border-border">
        <input ref="imageInput" type="file" accept="image/*" class="hidden" @change="onImageSelected" />

        <!-- Voice player bar (replaces input when recording exists) -->
        <div v-if="voiceFile"
          class="flex items-center gap-2 p-2 border shadow-sm bg-muted/30 rounded-3xl border-muted-foreground/10">
          <!-- Play/Pause -->
          <button @click="togglePreviewPlayback"
            class="flex items-center justify-center transition-colors rounded-full h-9 w-9 shrink-0 bg-emerald-100 dark:bg-emerald-900/30 text-emerald-600 dark:text-emerald-400 hover:bg-emerald-200 dark:hover:bg-emerald-900/50">
            <Pause v-if="isPreviewPlaying" class="w-4 h-4" />
            <Play v-else class="h-4 w-4 ml-0.5" />
          </button>

          <!-- Time -->
          <span class="text-xs text-center text-muted-foreground tabular-nums w-9 shrink-0">
            {{ isPreviewPlaying || previewCurrent > 0 ? formatTime(previewCurrent) : '0:00' }}
          </span>

          <!-- Progress bar -->
          <div class="flex-1 relative h-1.5 rounded-full bg-muted cursor-pointer" @click="seekPreviewAudio($event)">
            <div class="absolute h-full transition-all rounded-full bg-emerald-500"
              :style="{ width: previewProgress + '%' }"></div>
            <div class="absolute w-3 h-3 transition-all -translate-y-1/2 rounded-full shadow-sm top-1/2 bg-emerald-500"
              :style="{ left: previewProgress + '%' }" v-show="previewProgress > 0"></div>
          </div>

          <!-- Duration -->
          <span class="text-xs text-center text-muted-foreground tabular-nums w-9 shrink-0">
            {{ previewDurationVal ? formatTime(previewDurationVal) : formatDuration(recordingDuration) }}
          </span>

          <!-- Volume icon -->
          <Volume2 class="w-4 h-4 shrink-0 text-muted-foreground/50" />

          <!-- Divider -->
          <div class="w-px h-6 bg-border"></div>

          <!-- Delete -->
          <button @click="clearVoice"
            class="flex items-center justify-center text-red-500 transition-colors bg-red-100 rounded-full h-9 w-9 shrink-0 dark:bg-red-900/30 hover:bg-red-200 dark:hover:bg-red-900/50">
            <Trash2 class="w-4 h-4" />
          </button>

          <!-- Send -->
          <Button @click="sendMessage" size="icon" :disabled="!canSend()"
            class="text-white transition-all rounded-full shadow-sm h-9 w-9 bg-emerald-600 hover:bg-emerald-700 shrink-0 active:scale-95">
            <Send class="h-4 w-4 ml-0.5" />
          </Button>
        </div>

        <!-- Normal input bar -->
        <div v-else
          class="flex items-center gap-2 p-2 transition-all border shadow-sm bg-muted/30 rounded-3xl border-muted-foreground/10 focus-within:ring-1 focus-within:ring-emerald-500/50">
          <Input v-model="userInput" placeholder="Ketik pesan..."
            class="flex-1 px-3 text-base bg-transparent border-none shadow-none focus-visible:ring-0 h-9 md:text-sm"
            :disabled="isTyping || isRecording" @keydown.enter.prevent="sendMessage" />

          <div class="flex items-center gap-1 pr-1">
            <Button variant="ghost" size="icon"
              class="w-8 h-8 rounded-full text-muted-foreground hover:text-foreground shrink-0 hover:bg-background/80"
              title="Kirim Gambar" :disabled="isTyping || isRecording" @click="triggerImageUpload">
              <ImageIcon class="w-4 h-4" />
            </Button>

            <Button variant="ghost" size="icon" :class="[
              'h-8 w-8 rounded-full shrink-0 transition-all',
              isRecording
                ? 'bg-red-100 dark:bg-red-900/30 text-red-500 hover:bg-red-200 animate-pulse'
                : 'text-muted-foreground hover:text-foreground hover:bg-background/80',
            ]" :title="isRecording ? 'Stop Rekam' : 'Rekam Suara'" :disabled="isTyping" @click="toggleRecording">
              <MicOff v-if="isRecording" class="w-4 h-4" />
              <Mic v-else class="w-4 h-4" />
            </Button>
          </div>

          <Button @click="sendMessage" size="icon" :disabled="!canSend()"
            class="text-white transition-all rounded-full shadow-sm h-9 w-9 bg-emerald-600 hover:bg-emerald-700 shrink-0 active:scale-95">
            <Send class="h-4 w-4 ml-0.5" />
          </Button>
        </div>

        <!-- Recording indicator -->
        <div v-if="isRecording" class="flex items-center justify-center gap-2 mt-2 text-xs text-red-500">
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
