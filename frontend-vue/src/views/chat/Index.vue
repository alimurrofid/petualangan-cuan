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
  Reply,
  Pencil,
  Copy,
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
  replyToId?: number;
  replyToRole?: string;
  replyToContent?: string;
  isEdited?: boolean;
  isLocal?: boolean;
}

const messages = ref<Message[]>([]);
const isLoadingHistory = ref(false);

const replyingTo = ref<Message | null>(null);
const editingMessageId = ref<number | null>(null);
const editContent = ref("");
const isEditingSubmitting = ref(false);
const selectedMobileActionMessage = ref<Message | null>(null);
const swipingMessageId = ref<number | null>(null);
const swipeTranslateX = ref(0);
let touchStartX = 0;
let touchStartY = 0;
let isSwiping = false;

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
      reply_to_id?: number;
      reply_to_role?: string;
      reply_to_content?: string;
      is_edited?: boolean;
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
        replyToId: m.reply_to_id || undefined,
        replyToRole: m.reply_to_role || undefined,
        replyToContent: m.reply_to_content || undefined,
        isEdited: m.is_edited || false,
      }));
    } else {
      // Pesan sambutan untuk percakapan baru
      messages.value = [
        {
          id: -1,
          role: "assistant",
          content:
            "Halo! Saya Cuan AI, asisten keuangan pribadimu. Kamu bisa kirim teks, foto struk, atau pesan suara! 🤖💰",
          time: format(new Date(), "HH:mm"),
          isLocal: true,
        },
      ];
    }
  } catch {
    messages.value = [
      {
        id: -1,
        role: "assistant",
        content:
          "Halo! Saya Cuan AI, asisten keuangan pribadimu. Kamu bisa kirim teks, foto struk, atau pesan suara! 🤖💰",
        time: format(new Date(), "HH:mm"),
        isLocal: true,
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
        id: -2,
        role: "assistant",
        content:
          "Riwayat percakapan telah dihapus. Halo lagi! Saya siap membantu keuanganmu. 🤖💰",
        time: format(new Date(), "HH:mm"),
        isLocal: true,
      },
    ];
    swal.success('Riwayat percakapan berhasil dihapus');
  } catch {
    swal.error('Gagal', 'Gagal menghapus riwayat percakapan.');
  }
};

const startReply = (msg: Message) => {
  replyingTo.value = msg;
  nextTick(() => {
    const inputEl = document.querySelector('input[placeholder="Ketik pesan..."]') as HTMLInputElement;
    if (inputEl) inputEl.focus();
  });
};

const cancelReply = () => {
  replyingTo.value = null;
};

const startEdit = (msg: Message) => {
  editingMessageId.value = msg.id;
  editContent.value = msg.content;
};

const cancelEdit = () => {
  editingMessageId.value = null;
  editContent.value = "";
};

const scrollToMessage = (msgId?: number) => {
  if (!msgId) return;
  const el = document.getElementById(`msg-${msgId}`);
  if (el) {
    el.scrollIntoView({ behavior: "smooth", block: "center" });
    el.classList.add("ring-2", "ring-emerald-500", "rounded-2xl", "transition-all", "duration-500");
    setTimeout(() => {
      el.classList.remove("ring-2", "ring-emerald-500", "rounded-2xl");
    }, 1500);
  }
};

const copyMessageText = async (msg: Message) => {
  if (!msg.content) return;
  try {
    await navigator.clipboard.writeText(msg.content);
    swal.success("Teks berhasil disalin!");
  } catch {
    swal.error("Gagal", "Tidak dapat menyalin teks.");
  }
};

const openMobileActionMenu = (msg: Message) => {
  selectedMobileActionMessage.value = msg;
};

const closeMobileActionMenu = () => {
  selectedMobileActionMessage.value = null;
};

const isLocalMessage = (id: number) => {
  const msg = messages.value.find((m) => m.id === id);
  return !!msg?.isLocal || id < 0 || id > 100000000000;
};

const confirmDeleteMessage = async (msg: Message) => {
  const hasTx = msg.transactions && msg.transactions.length > 0;

  if (hasTx) {
    const result = await swal.fire({
      title: "Hapus Pesan?",
      text: "Pesan ini memiliki aksi keuangan tersimpan. Ingin membatalkan transaksinya juga?",
      icon: "warning",
      showCancelButton: true,
      showDenyButton: true,
      confirmButtonText: "Hapus & Batalkan Aksi",
      denyButtonText: "Hapus Pesan Saja",
      cancelButtonText: "Batal",
      confirmButtonColor: "#ef4444",
      denyButtonColor: "#64748b",
    });

    if (result.isConfirmed) {
      await deleteMessage(msg.id, true);
    } else if (result.isDenied) {
      await deleteMessage(msg.id, false);
    }
  } else {
    const confirmed = await swal.confirmDelete("pesan ini");
    if (confirmed) {
      await deleteMessage(msg.id, false);
    }
  }
};

const deleteMessage = async (id: number, rollback: boolean) => {
  if (isLocalMessage(id)) {
    messages.value = messages.value.filter((m) => m.id !== id);
    swal.success("Pesan berhasil dihapus");
    return;
  }

  try {
    const token = localStorage.getItem("token");
    const url = `${import.meta.env.VITE_API_BASE_URL}api/ai/chat/messages/${id}${rollback ? "?rollback=true" : ""}`;
    const res = await fetch(url, {
      method: "DELETE",
      headers: { Authorization: `Bearer ${token}` },
    });
    if (!res.ok) {
      const data = await res.json().catch(() => ({}));
      throw new Error(data.error || "Gagal menghapus pesan");
    }

    messages.value = messages.value.filter((m) => m.id !== id);
    swal.success("Pesan berhasil dihapus");
  } catch (err: any) {
    swal.error("Gagal", err.message || "Gagal menghapus pesan");
  }
};

const submitEdit = async (msgId: number) => {
  const newText = editContent.value.trim();
  if (!newText || isEditingSubmitting.value) return;

  isEditingSubmitting.value = true;
  const targetIndex = messages.value.findIndex((m) => m.id === msgId);
  if (targetIndex === -1) {
    cancelEdit();
    isEditingSubmitting.value = false;
    return;
  }

  const targetMsg = messages.value[targetIndex];
  if (!targetMsg) {
    cancelEdit();
    isEditingSubmitting.value = false;
    return;
  }

  // Update pesan user secara lokal
  targetMsg.content = newText;
  targetMsg.isEdited = true;

  // Jika ada pesan asisten setelah pesan ini, cari dan siapkan untuk diganti
  let assistantMsgIndex = -1;
  for (let i = targetIndex + 1; i < messages.value.length; i++) {
    const nextMsg = messages.value[i];
    if (nextMsg && nextMsg.role === "assistant") {
      assistantMsgIndex = i;
      break;
    }
  }

  // Siapkan assistant placeholder baru
  const assistantMsg: Message = reactive({
    id: Date.now(),
    role: "assistant",
    content: "",
    time: format(new Date(), "HH:mm"),
  });

  if (assistantMsgIndex !== -1) {
    messages.value.splice(assistantMsgIndex, 1, assistantMsg);
  } else {
    messages.value.push(assistantMsg);
  }

  cancelEdit();
  isTyping.value = true;
  typingStatus.value = "Memperbarui respons...";

  try {
    const token = localStorage.getItem("token");
    const formData = new FormData();
    formData.append("message", newText);

    await fetchEventSource(`${import.meta.env.VITE_API_BASE_URL}api/ai/chat/messages/${msgId}`, {
      method: "PUT",
      headers: { Authorization: `Bearer ${token}` },
      body: formData,
      async onopen(response) {
        if (!response.ok) throw new Error("Gagal memperbarui pesan");
      },
      onmessage(msg) {
        if (msg.event === "status") {
          typingStatus.value = msg.data;
          isTyping.value = true;
          throttledScrollToBottom();
        } else if (msg.event === "token") {
          try {
            const parsed = JSON.parse(msg.data);
            assistantMsg.content += parsed.content;
            isTyping.value = false;
            throttledScrollToBottom();
          } catch (e) {}
        } else if (msg.event === "done") {
          try {
            const result = JSON.parse(msg.data);
            if (result.assistant_message_id || result.id) {
              assistantMsg.id = result.assistant_message_id || result.id;
              assistantMsg.isLocal = false;
            }
            if (result.transactions?.length) {
              assistantMsg.transactions = result.transactions;
            }
            if (result.reply) {
              assistantMsg.content = result.reply;
            }
            scrollToBottom();
          } catch (e) {}
          isTyping.value = false;
        } else if (msg.event === "error") {
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
        throw err;
      },
    });
  } catch (err: any) {
    isTyping.value = false;
    swal.error("Gagal", err.message || "Gagal memperbarui pesan");
  } finally {
    isEditingSubmitting.value = false;
  }
};

const handleTouchStart = (e: TouchEvent, msg: Message) => {
  const touch = e.touches[0];
  if (!touch || e.touches.length !== 1) return;
  touchStartX = touch.clientX;
  touchStartY = touch.clientY;
  swipingMessageId.value = msg.id;
  swipeTranslateX.value = 0;
  isSwiping = false;
};

const handleTouchMove = (e: TouchEvent, msg: Message) => {
  const touch = e.touches[0];
  if (!touch || swipingMessageId.value !== msg.id || e.touches.length !== 1) return;
  const currentX = touch.clientX;
  const currentY = touch.clientY;
  const diffX = currentX - touchStartX;
  const diffY = currentY - touchStartY;

  // Jika swipe horizontal ke kanan
  if (diffX > 8 && Math.abs(diffX) > Math.abs(diffY)) {
    isSwiping = true;
    swipeTranslateX.value = Math.min(diffX * 0.55, 75);
  }
};

const handleTouchEnd = (msg: Message) => {
  if (swipingMessageId.value === msg.id) {
    if (swipeTranslateX.value >= 40) {
      startReply(msg);
      if (typeof navigator !== "undefined" && navigator.vibrate) {
        navigator.vibrate(25);
      }
    } else if (!isSwiping && swipeTranslateX.value < 8) {
      openMobileActionMenu(msg);
    }
  }
  swipingMessageId.value = null;
  swipeTranslateX.value = 0;
  isSwiping = false;
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

  const userMsg: Message = reactive({
    id: Date.now(),
    role: "user",
    content: text,
    time: format(new Date(), "HH:mm"),
  });
  if (imagePreview.value) {
    userMsg.imageUrl = imagePreview.value;
  }
  if (voiceFile.value && previewVoiceUrl.value) {
    userMsg.audioUrl = previewVoiceUrl.value;
    userMsg.voiceLabel = formatDuration(recordingDuration.value);
  }

  const formData = new FormData();
  if (replyingTo.value) {
    userMsg.replyToId = replyingTo.value.id;
    userMsg.replyToRole = replyingTo.value.role;
    userMsg.replyToContent = replyingTo.value.content || (replyingTo.value.imageUrl ? '📷 Foto' : replyingTo.value.audioUrl ? '🎤 Pesan Suara' : '');
    if (!isLocalMessage(replyingTo.value.id)) {
      formData.append("reply_to_id", replyingTo.value.id.toString());
    }
    replyingTo.value = null;
  }

  messages.value.push(userMsg);

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
        if (msg.event === "user_message") {
          try {
            const data = JSON.parse(msg.data);
            if (data.id || data.user_message_id) {
              userMsg.id = data.user_message_id || data.id;
              userMsg.isLocal = false;
            }
          } catch (e) {}
        } else if (msg.event === "status") {
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

            if (result.user_message_id) {
              userMsg.id = result.user_message_id;
              userMsg.isLocal = false;
            }
            if (result.assistant_message_id || result.id) {
              assistantMsg.id = result.assistant_message_id || result.id;
              assistantMsg.isLocal = false;
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
      isLocal: true,
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
  <div class="flex flex-col h-[calc(100dvh-10.5rem)] md:h-[calc(100dvh-7.5rem)] max-w-5xl mx-auto space-y-3 md:space-y-4">
    <!-- Header -->
    <div class="flex items-center gap-3 sm:gap-4 pb-3 sm:pb-4 border-b border-border">
      <div
        class="overflow-hidden border shadow-sm w-11 h-11 sm:w-14 sm:h-14 shrink-0 rounded-2xl border-emerald-500/20">
        <img src="/cuan-bot.svg" alt="Cuan AI" class="object-cover w-full h-full" />
      </div>
      <div class="flex-1 min-w-0">
        <h2 class="flex items-center gap-2 text-lg sm:text-xl font-bold tracking-tight">
          Cuan AI
          <span
            class="text-[10px] bg-emerald-100/50 dark:bg-emerald-900/30 text-emerald-600 dark:text-emerald-400 px-2 py-0.5 rounded-full border border-emerald-200/50 flex items-center gap-1">
            <Sparkles class="w-3 h-3" /> Beta
          </span>
        </h2>
        <p class="text-xs sm:text-sm text-muted-foreground truncate sm:whitespace-normal">
          Tanya tips keuangan, kirim foto struk, atau rekam suara.
        </p>
      </div>
      <!-- Clear history button -->
      <Button v-if="messages.length > 0 && !isLoadingHistory" variant="ghost" size="icon"
        title="Hapus riwayat percakapan"
        class="transition-colors rounded-full h-8 w-8 sm:h-9 sm:w-9 text-muted-foreground hover:text-destructive hover:bg-destructive/10 shrink-0"
        @click="clearHistory">
        <Trash2 class="w-4 h-4" />
      </Button>
      <!-- Loading history indicator -->
      <div v-if="isLoadingHistory" class="flex items-center gap-2 text-xs text-muted-foreground shrink-0">
        <Loader2 class="w-4 h-4 animate-spin text-emerald-500" />
        <span>Memuat...</span>
      </div>
    </div>

    <Card class="relative flex flex-col flex-1 overflow-hidden shadow-xs bg-muted/20 border-border rounded-2xl md:rounded-3xl">
      <!-- Messages -->
      <div ref="chatContainer" class="flex-1 p-3 sm:p-4 space-y-3 sm:space-y-4 overflow-y-auto custom-scrollbar scroll-smooth">
        <div v-for="msg in messages" :key="msg.id" :id="'msg-' + msg.id"
          class="relative flex w-full group/msg select-text transition-all">
          <!-- Mobile Swipe Indicator Behind Bubble -->
          <div v-if="swipingMessageId === msg.id && swipeTranslateX > 8"
            class="absolute left-2 top-1/2 -translate-y-1/2 flex items-center justify-center w-8 h-8 rounded-full bg-emerald-500 text-white shadow-md pointer-events-none transition-transform"
            :style="{
              transform: `translateY(-50%) scale(${Math.min(swipeTranslateX / 40, 1.15)}) rotate(${Math.min((swipeTranslateX - 8) * 2.5, 90)}deg)`,
              opacity: Math.min(swipeTranslateX / 35, 1)
            }">
            <Reply class="w-4 h-4" />
          </div>

          <div :class="[
            'flex max-w-[90%] sm:max-w-[85%] md:max-w-[70%] min-w-0 gap-2 transition-transform duration-100 ease-out',
            msg.role === 'user'
              ? 'ml-auto flex-row-reverse'
              : 'mr-auto flex-row',
          ]"
          :style="swipingMessageId === msg.id ? {
            transform: `translateX(${swipeTranslateX}px)`,
            transition: isSwiping ? 'none' : 'transform 0.25s cubic-bezier(0.2, 0.8, 0.2, 1)'
          } : {}"
          @touchstart="handleTouchStart($event, msg)"
          @touchmove="handleTouchMove($event, msg)"
          @touchend="handleTouchEnd(msg)">
            <!-- Avatar -->
            <div v-if="msg.role === 'assistant'"
              class="w-8 h-8 sm:w-10 sm:h-10 overflow-hidden border rounded-full shadow-xs shrink-0 border-emerald-200/50">
              <img src="/cuan-bot.svg" alt="Cuan AI" class="object-cover w-full h-full" />
            </div>
            <div v-else
              class="w-8 h-8 sm:w-10 sm:h-10 overflow-hidden border rounded-full shadow-xs shrink-0 border-emerald-500/20">
              <img src="/people.svg" alt="User" class="object-cover w-full h-full" />
            </div>

            <div :class="[
              'flex flex-col relative min-w-0 max-w-full',
              msg.role === 'user' ? 'items-end' : 'items-start',
            ]">
              <!-- Desktop Hover Action Pill -->
              <div v-if="editingMessageId !== msg.id"
                class="hidden md:flex items-center gap-0.5 opacity-0 group-hover/msg:opacity-100 transition-opacity duration-150 absolute top-0 bg-card/95 backdrop-blur-xs border border-border shadow-md rounded-full px-1.5 py-0.5 z-20"
                :class="msg.role === 'user' ? 'right-full mr-2' : 'left-full ml-2'">
                <button @click.stop="startReply(msg)" title="Balas Pesan"
                  class="p-1.5 rounded-full hover:bg-muted text-muted-foreground hover:text-foreground transition-colors">
                  <Reply class="w-3.5 h-3.5" />
                </button>
                <button v-if="msg.role === 'user'" @click.stop="startEdit(msg)" title="Edit Pesan"
                  class="p-1.5 rounded-full hover:bg-muted text-muted-foreground hover:text-foreground transition-colors">
                  <Pencil class="w-3.5 h-3.5" />
                </button>
                <button @click.stop="confirmDeleteMessage(msg)" title="Hapus Pesan"
                  class="p-1.5 rounded-full hover:bg-destructive/10 text-muted-foreground hover:text-destructive transition-colors">
                  <Trash2 class="w-3.5 h-3.5" />
                </button>
              </div>

              <!-- Inline Edit Mode (for user message) -->
              <div v-if="editingMessageId === msg.id" class="w-full min-w-[260px] md:min-w-[340px] space-y-2 mt-1">
                <textarea v-model="editContent" rows="3"
                  class="w-full p-2.5 text-sm rounded-2xl bg-card text-foreground border border-emerald-500 shadow-sm focus:outline-none focus:ring-2 focus:ring-emerald-500/40 custom-scrollbar resize-none"
                  placeholder="Ketik perubahan pesan..."
                  @keydown.enter.exact.prevent="submitEdit(msg.id)"
                  @keydown.esc="cancelEdit"></textarea>
                <div class="flex items-center justify-end gap-2">
                  <Button variant="ghost" size="sm" class="h-7 px-2.5 text-xs rounded-lg" @click="cancelEdit">
                    Batal
                  </Button>
                  <Button size="sm" class="h-7 px-3 text-xs rounded-lg bg-emerald-600 hover:bg-emerald-700 text-white"
                    :disabled="!editContent.trim() || isEditingSubmitting" @click="submitEdit(msg.id)">
                    <Loader2 v-if="isEditingSubmitting" class="w-3 h-3 mr-1 animate-spin" />
                    Simpan & Kirim
                  </Button>
                </div>
              </div>

              <!-- Main Bubble Card (Unified & Responsive) -->
              <div v-else :class="[
                'rounded-2xl text-sm shadow-sm leading-relaxed relative min-w-0 max-w-full overflow-hidden flex flex-col',
                msg.role === 'user'
                  ? 'bg-gradient-to-br from-emerald-500 to-teal-600 text-white rounded-tr-sm'
                  : 'bg-card border border-border rounded-tl-sm text-foreground',
              ]">
                <!-- Quoted message pill inside bubble if this is a reply -->
                <div v-if="msg.replyToContent" @click="scrollToMessage(msg.replyToId)"
                  class="mx-2 mt-2 mb-1 p-2 rounded-xl text-xs cursor-pointer transition-all hover:opacity-90 max-w-full min-w-0 select-none border-l-4"
                  :class="[
                    msg.role === 'user'
                      ? 'bg-black/20 border-white/90 text-white'
                      : 'bg-muted/80 border-emerald-500 text-foreground'
                  ]">
                  <div class="flex items-center gap-1 font-semibold text-[11px] mb-0.5"
                    :class="msg.role === 'user' ? 'text-white/95' : 'text-emerald-600 dark:text-emerald-400'">
                    <Reply class="w-3 h-3 shrink-0" />
                    <span class="truncate">{{ msg.replyToRole === 'user' ? 'Kamu' : 'Cuan AI' }}</span>
                  </div>
                  <p class="text-[11px] opacity-85 line-clamp-2 break-words leading-snug">
                    {{ msg.replyToContent }}
                  </p>
                </div>

                <!-- Image preview in bubble -->
                <div v-if="msg.imageUrl" :class="msg.replyToContent ? 'px-2 pb-1' : 'p-2'">
                  <img :src="getMediaUrl(msg.imageUrl)"
                    class="max-w-[200px] sm:max-w-[240px] rounded-xl shadow-sm border border-border/50 object-cover" alt="Uploaded image" />
                </div>

                <!-- Audio player in bubble -->
                <div v-if="msg.audioUrl" class="p-2 w-full min-w-[220px] sm:min-w-[260px]">
                  <div :class="[
                    'flex items-center gap-2.5 px-3 py-2 rounded-2xl',
                    msg.role === 'user'
                      ? 'bg-white/15 text-white'
                      : 'bg-muted/50 border border-border/60',
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
                    <div class="flex flex-col flex-1 gap-1 min-w-0">
                      <div class="relative h-1 rounded-full cursor-pointer"
                        :class="msg.role === 'user' ? 'bg-white/20' : 'bg-muted'" @click="seekAudio($event, msg.id)">
                        <div class="absolute h-full transition-all rounded-full"
                          :class="msg.role === 'user' ? 'bg-white/70' : 'bg-emerald-500'"
                          :style="{ width: activeAudioId === msg.id ? audioProgress + '%' : '0%' }"></div>
                      </div>
                      <div class="flex justify-between text-[10px]"
                        :class="msg.role === 'user' ? 'text-white/70' : 'text-muted-foreground'">
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
                <span v-else-if="msg.voiceLabel" class="px-3 py-1 text-xs italic opacity-75">
                  {{ msg.voiceLabel }}
                </span>

                <!-- Text content in bubble -->
                <div v-if="msg.content" :class="[
                  'whitespace-pre-wrap break-words leading-relaxed',
                  msg.replyToContent ? 'px-3.5 pt-1 pb-2.5 text-sm' : 'px-4 py-2.5 text-sm',
                ]">
                  {{ msg.content }}
                </div>
              </div>

              <!-- Transaction cards -->
              <div v-if="msg.transactions?.length" class="w-full min-w-0 mt-2 space-y-2">
                <div v-for="tx in msg.transactions" :key="tx.id"
                  class="p-2.5 sm:p-3 text-xs border rounded-2xl shadow-xs transition-all min-w-0 overflow-hidden"
                  :class="[
                    tx.action?.includes('delete')
                      ? 'bg-red-50/80 dark:bg-red-950/30 border-red-200 dark:border-red-900/50'
                      : tx.action?.includes('update')
                        ? 'bg-blue-50/80 dark:bg-blue-950/30 border-blue-200 dark:border-blue-900/50'
                        : 'bg-emerald-50/80 dark:bg-emerald-950/30 border-emerald-200/80 dark:border-emerald-900/50'
                  ]">
                  <!-- Top Row: Description & Amount -->
                  <div class="flex items-start justify-between gap-2 min-w-0">
                    <div class="flex items-center gap-1.5 min-w-0 flex-1">
                      <span class="text-sm shrink-0">
                        {{ tx.action?.includes('delete') ? '🗑️' : tx.action?.includes('update') ? '✏️' : '✅' }}
                      </span>
                      <span class="font-semibold text-xs sm:text-sm truncate min-w-0" :class="{
                        'text-emerald-800 dark:text-emerald-300': !tx.action?.includes('delete') && !tx.action?.includes('update'),
                        'text-blue-800 dark:text-blue-300': tx.action?.includes('update'),
                        'text-red-800 dark:text-red-300 line-through': tx.action?.includes('delete')
                      }" :title="tx.description">
                        {{ tx.description?.replace(/~~/g, '').replace(/\(Dihapus\)/g, '').trim() || '-' }}
                      </span>
                    </div>

                    <span class="font-bold text-xs sm:text-sm shrink-0 text-right whitespace-nowrap" :class="[
                      tx.action?.includes('delete') ? 'line-through opacity-60' : '',
                      (tx.type === 'income' || tx.type === 'receivable') ? 'text-emerald-600 dark:text-emerald-400' : 
                      (tx.type === 'transfer' || tx.type === 'goal' || tx.type === 'wishlist') ? 'text-blue-600 dark:text-blue-400' : 'text-red-600 dark:text-red-400'
                    ]">
                      {{ (tx.type === 'income' || tx.type === 'receivable') ? '+' : (tx.type === 'transfer' || tx.type === 'goal' || tx.type === 'wishlist') ? '' : '-' }}Rp{{ Number(tx.amount || 0).toLocaleString('id-ID') }}
                    </span>
                  </div>

                  <!-- Bottom Meta Row: Wallet, Category, Badge (Wraps gracefully on mobile) -->
                  <div class="flex flex-wrap items-center gap-1.5 sm:gap-2 mt-2 pt-1.5 border-t border-border/40 text-[11px] text-muted-foreground min-w-0">
                    <span v-if="tx.type === 'transfer' && tx.to_wallet_name" class="inline-flex items-center gap-1 min-w-0 truncate max-w-full">
                      <span>🏦</span>
                      <span class="truncate">{{ tx.wallet_name }} ➡️ {{ tx.to_wallet_name }}</span>
                    </span>
                    <span v-else class="inline-flex items-center gap-1 min-w-0 truncate max-w-[130px] sm:max-w-none">
                      <span>🏦</span>
                      <span class="truncate">{{ tx.wallet_name || '-' }}</span>
                    </span>

                    <span class="opacity-40">•</span>

                    <span class="inline-flex items-center gap-1 min-w-0 truncate max-w-[120px] sm:max-w-none">
                      <span>📂</span>
                      <span class="truncate">{{ tx.category_name || '-' }}</span>
                    </span>

                    <span class="ml-auto px-1.5 py-0.5 rounded text-[10px] font-medium shrink-0"
                      :class="[
                        tx.type === 'income' || tx.type === 'receivable'
                          ? 'bg-emerald-100 dark:bg-emerald-900/50 text-emerald-700 dark:text-emerald-300' 
                          : tx.type === 'transfer' || tx.type === 'goal' || tx.type === 'wishlist'
                            ? 'bg-blue-100 dark:bg-blue-900/50 text-blue-700 dark:text-blue-300'
                            : 'bg-red-100 dark:bg-red-900/50 text-red-700 dark:text-red-300'
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

              <!-- Time & edited badge -->
              <div class="flex items-center gap-1 mt-1 text-[10px] text-muted-foreground opacity-70 px-1">
                <span>{{ msg.time }}</span>
                <span v-if="msg.isEdited" class="italic font-medium text-[9px] text-emerald-600 dark:text-emerald-400">
                  • diedit
                </span>
              </div>
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

      <!-- Reply Preview Bar -->
      <div v-if="replyingTo"
        class="flex items-center justify-between gap-3 px-4 py-2 border-t bg-muted/40 border-border text-xs animate-in fade-in duration-150 min-w-0">
        <div class="flex items-center gap-2.5 min-w-0 overflow-hidden border-l-2 border-emerald-500 pl-2.5 flex-1">
          <div class="p-1 rounded-full bg-emerald-100 dark:bg-emerald-900/30 text-emerald-600 dark:text-emerald-400 shrink-0">
            <Reply class="w-3.5 h-3.5" />
          </div>
          <div class="min-w-0 flex-1">
            <div class="font-semibold text-emerald-600 dark:text-emerald-400 truncate text-[11px]">
              Membalas {{ replyingTo.role === 'user' ? 'Kamu' : 'Cuan AI' }}
            </div>
            <p class="text-muted-foreground truncate text-[11px] mt-0.5">
              {{ replyingTo.content || (replyingTo.imageUrl ? '📷 Foto' : replyingTo.audioUrl ? '🎤 Pesan Suara' : '') }}
            </p>
          </div>
        </div>
        <button @click="cancelReply"
          class="p-1 text-muted-foreground hover:text-foreground rounded-full hover:bg-muted transition-colors shrink-0"
          title="Batal membalas">
          <X class="w-4 h-4" />
        </button>
      </div>

      <!-- Input bar -->
      <div class="p-2.5 sm:p-3 border-t md:p-4 bg-card border-border">
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

    <!-- Mobile Action Bottom Sheet -->
    <div v-if="selectedMobileActionMessage" 
      class="fixed inset-0 z-50 flex items-end justify-center bg-black/60 backdrop-blur-xs md:hidden"
      @click.self="closeMobileActionMenu">
      <div class="w-full max-w-md bg-card border-t border-border rounded-t-3xl p-5 space-y-4 animate-in slide-in-from-bottom duration-200">
        <div class="flex items-center justify-between pb-3 border-b border-border">
          <div class="flex items-center gap-2">
            <span class="text-sm font-bold">Menu Pesan</span>
            <span class="text-xs text-muted-foreground">({{ selectedMobileActionMessage.role === 'user' ? 'Kamu' : 'Cuan AI' }})</span>
          </div>
          <button @click="closeMobileActionMenu" class="p-1 text-muted-foreground rounded-full hover:bg-muted">
            <X class="w-4 h-4" />
          </button>
        </div>

        <!-- Quoted message snippet -->
        <div class="p-3 text-xs border rounded-xl bg-muted/30 border-border text-muted-foreground line-clamp-3">
          "{{ selectedMobileActionMessage.content || (selectedMobileActionMessage.imageUrl ? '📷 Foto' : selectedMobileActionMessage.audioUrl ? '🎤 Pesan Suara' : '') }}"
        </div>

        <div class="grid grid-cols-1 gap-1.5 pt-1">
          <button @click="startReply(selectedMobileActionMessage); closeMobileActionMenu()"
            class="flex items-center gap-3 px-4 py-3 text-sm font-medium rounded-2xl hover:bg-muted transition-colors text-foreground">
            <Reply class="w-4 h-4 text-emerald-500" />
            Balas Pesan
          </button>

          <button v-if="selectedMobileActionMessage.role === 'user'" 
            @click="startEdit(selectedMobileActionMessage); closeMobileActionMenu()"
            class="flex items-center gap-3 px-4 py-3 text-sm font-medium rounded-2xl hover:bg-muted transition-colors text-foreground">
            <Pencil class="w-4 h-4 text-blue-500" />
            Edit Pesan
          </button>

          <button @click="copyMessageText(selectedMobileActionMessage); closeMobileActionMenu()"
            class="flex items-center gap-3 px-4 py-3 text-sm font-medium rounded-2xl hover:bg-muted transition-colors text-foreground">
            <Copy class="w-4 h-4 text-muted-foreground" />
            Salin Teks
          </button>

          <button @click="confirmDeleteMessage(selectedMobileActionMessage); closeMobileActionMenu()"
            class="flex items-center gap-3 px-4 py-3 text-sm font-medium rounded-2xl hover:bg-destructive/10 transition-colors text-destructive">
            <Trash2 class="w-4 h-4 text-destructive" />
            Hapus Pesan
          </button>
        </div>
      </div>
    </div>
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
