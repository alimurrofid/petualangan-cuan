<script setup lang="ts">
import { computed } from "vue"
import type { HTMLAttributes } from "vue"
import { SwitchRoot, SwitchThumb } from "reka-ui"
import { cn } from "@/lib/utils"

const props = withDefaults(
  defineProps<{
    modelValue?: boolean
    checked?: boolean
    disabled?: boolean
    id?: string
    name?: string
    class?: HTMLAttributes["class"]
  }>(),
  {
    modelValue: undefined,
    checked: undefined,
    disabled: false,
  }
)

const emit = defineEmits<{
  (e: "update:modelValue", value: boolean): void
  (e: "update:checked", value: boolean): void
  (e: "change", value: boolean): void
}>()

const isChecked = computed({
  get() {
    if (props.checked !== undefined) return Boolean(props.checked)
    if (props.modelValue !== undefined) return Boolean(props.modelValue)
    return false
  },
  set(val: boolean) {
    emit("update:modelValue", val)
    emit("update:checked", val)
    emit("change", val)
  },
})
</script>

<template>
  <SwitchRoot
    :model-value="isChecked"
    @update:model-value="(val: boolean) => isChecked = val"
    :disabled="disabled"
    :id="id"
    :name="name"
    :class="
      cn(
        'peer inline-flex h-7 w-12 shrink-0 cursor-pointer items-center rounded-full p-0.5 transition-colors duration-200 ease-in-out focus-visible:outline-hidden focus-visible:ring-2 focus-visible:ring-emerald-500/50 focus-visible:ring-offset-2 focus-visible:ring-offset-background disabled:cursor-not-allowed disabled:opacity-50 select-none',
        'data-[state=checked]:bg-emerald-600 dark:data-[state=checked]:bg-emerald-500 shadow-xs',
        'data-[state=unchecked]:bg-zinc-300 dark:data-[state=unchecked]:bg-zinc-700 hover:data-[state=unchecked]:bg-zinc-400/80 dark:hover:data-[state=unchecked]:bg-zinc-600',
        props.class,
      )
    "
  >
    <SwitchThumb
      :class="
        cn(
          'pointer-events-none block size-6 rounded-full bg-white shadow-md shadow-black/20 ring-1 ring-black/5 transition-transform duration-200 ease-in-out',
          'data-[state=checked]:translate-x-5 data-[state=unchecked]:translate-x-0',
        )
      "
    />
  </SwitchRoot>
</template>
