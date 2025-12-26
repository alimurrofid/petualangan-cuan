<script setup lang="ts">
import { ref, onMounted } from "vue";
import { Button } from "@/components/ui/button";
import { Moon, Sun } from "lucide-vue-next";

const isDark = ref(false);

const toggleTheme = () => {
  isDark.value = !isDark.value;
  if (isDark.value) {
    document.documentElement.classList.add("dark");
    localStorage.setItem("theme", "dark");
  } else {
    document.documentElement.classList.remove("dark");
    localStorage.setItem("theme", "light");
  }
};

onMounted(() => {
  const savedTheme = localStorage.getItem("theme");
  if (savedTheme === "dark" || (!savedTheme && window.matchMedia("(prefers-color-scheme: dark)").matches)) {
    isDark.value = true;
    document.documentElement.classList.add("dark");
  }
});
</script>

<template>
  <Button variant="ghost" size="icon" @click="toggleTheme" class="rounded-full w-9 h-9">
    <Sun v-if="isDark" class="h-5 w-5 text-yellow-500" />
    <Moon v-else class="h-5 w-5 text-slate-700" />
  </Button>
</template>
