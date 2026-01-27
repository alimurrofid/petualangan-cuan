<script setup lang="ts">
import { computed, ref } from 'vue'
import { Check, ChevronsUpDown } from 'lucide-vue-next'
import { cn } from '@/lib/utils'
import { getEmoji, getIconComponent } from '@/lib/icons'
import { Button } from '@/components/ui/button'
import {
  Command,
  CommandEmpty,
  CommandGroup,
  CommandInput,
  CommandItem,
  CommandList,
} from '@/components/ui/command'
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from '@/components/ui/popover'

interface Option {
  value: string
  label: string
  icon?: string
}

const props = defineProps<{
  options: Option[]
  modelValue: string[]
  placeholder?: string
  countLabel?: string
}>()

const emit = defineEmits<{
  (e: 'update:modelValue', value: string[]): void
}>()

const open = ref(false)

const selectedLabels = computed(() => {
  return props.modelValue
    .map((val) => props.options.find((opt) => opt.value === val)?.label)
    .filter(Boolean) as string[]
})

const displayValue = computed(() => {
  if (selectedLabels.value.length === 0) return props.placeholder || 'Select items...'
  const label = props.countLabel ? ` ${props.countLabel}` : ''
  return `${selectedLabels.value.length}${label} terpilih`
})

const toggleSelection = (value: string) => {
  const newValue = [...props.modelValue]
  if (newValue.includes(value)) {
    newValue.splice(newValue.indexOf(value), 1)
  } else {
    newValue.push(value)
  }
  emit('update:modelValue', newValue)
}
</script>

<template>
  <div class="w-full">
    <Popover v-model:open="open">
        <PopoverTrigger as-child>
        <Button
            variant="outline"
            role="combobox"
            :aria-expanded="open"
            class="w-full justify-between hover:bg-background/50 h-9 rounded-xl text-xs font-semibold px-3 overflow-hidden"
        >
            <span class="truncate block w-full text-left">
                {{ displayValue }}
            </span>
            <ChevronsUpDown class="ml-2 h-4 w-4 shrink-0 opacity-50" />
        </Button>
        </PopoverTrigger>
        <PopoverContent class="w-[--radix-popover-trigger-width] p-0" align="start">
        <Command>
            <CommandInput :placeholder="placeholder || 'Cari...'" />
            <CommandList>
                <CommandEmpty>Tidak ditemukan.</CommandEmpty>
                <CommandGroup>
                <CommandItem
                    v-for="option in options"
                    :key="option.value"
                    :value="option.label"
                    @select="() => toggleSelection(option.value)"
                >
                    <div
                        :class="cn(
                        'mr-2 flex h-4 w-4 items-center justify-center rounded-sm border border-primary',
                        modelValue.includes(option.value)
                            ? 'bg-primary text-primary-foreground'
                            : 'opacity-50 [&_svg]:invisible'
                        )"
                    >
                        <Check class="h-3 w-3" />
                    </div>
                     <span v-if="option.icon && getEmoji(option.icon)" class="mr-2 text-base leading-none">{{ getEmoji(option.icon) }}</span>
                     <component v-else-if="option.icon" :is="getIconComponent(option.icon, 'Circle')" class="mr-2 h-4 w-4" />
                    {{ option.label }}
                </CommandItem>
                </CommandGroup>
            </CommandList>
        </Command>
        </PopoverContent>
    </Popover>
  </div>
</template>
