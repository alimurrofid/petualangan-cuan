<script setup lang="ts">
import { ref, onMounted } from "vue";

const isVisible = ref(false);
const showComponent = ref(true);

const showText1 = ref(false);
const showText2 = ref(false);
const showText3 = ref(false);

const isExiting = ref(false);

onMounted(() => {
  const hasSeenIntro = sessionStorage.getItem("hasSeenIntro");

  if (!hasSeenIntro) {
    isVisible.value = true;
    
    setTimeout(() => {
      showText1.value = true;
    }, 700);

    setTimeout(() => {
      showText2.value = true;
    }, 2200);

    setTimeout(() => {
      showText3.value = true;
    }, 3200);

    setTimeout(() => {
      isExiting.value = true;
      
      setTimeout(() => {
        isVisible.value = false;
        showComponent.value = false;
        sessionStorage.setItem("hasSeenIntro", "true");
      }, 1500);
    }, 4700);

  } else {
    showComponent.value = false;
  }
});
</script>

<template>
  <div
    v-if="showComponent"
    class="fixed inset-0 z-[9999] flex items-center justify-center bg-emerald-950 px-6 sm:px-12 text-orange-50 transition-all duration-1000 ease-in-out"
    style="clip-path: circle(150% at 100% 0);"
    :class="{ 'intro-exit': isExiting }"
  >
    <div class="w-full max-w-4xl flex flex-col space-y-6">
      <h1 
        class="font-serif text-md sm:text-2xl md:text-3xl leading-relaxed text-left transition-all duration-1000"
        :class="showText1 ? 'opacity-100 blur-0 translate-y-0' : 'opacity-0 blur-sm translate-y-4'"
      >
        "Mengelola uang dengan baik tidak ada hubungannya dengan kecerdasan anda ...
      </h1>

      <h1 
        class="font-serif text-md sm:text-2xl md:text-3xl leading-relaxed text-left transition-all duration-1000"
        :class="showText2 ? 'opacity-100 translate-y-0' : 'opacity-0 translate-y-4'"
      >
        dan lebih banyak berhubungan dengan perilaku anda."
      </h1>

      <div 
        class="self-end text-right mt-8 transition-all duration-1000"
        :class="showText3 ? 'opacity-100 translate-y-0' : 'opacity-0 translate-y-4'"
      >
        <p class="font-serif text-sm sm:text-xl font-medium">
          — Morgan Housel
        </p>
        <p class="text-xs sm:text-sm opacity-80 uppercase tracking-widest mt-1">
          Psychology of Money
        </p>
      </div>
    </div>
  </div>
</template>

<style scoped>
.intro-exit {
  clip-path: circle(0% at 100% 0) !important;
}
</style>