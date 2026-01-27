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
  [key: string]: any // allow extra properties
}

const props = defineProps<{
  options: Option[]
  modelValue: string | number
  placeholder?: string
  disabled?: boolean
  error?: boolean
}>()

const emit = defineEmits<{
  (e: 'update:modelValue', value: string): void
}>()

const open = ref(false)

const selectedOption = computed(() => {
    return props.options.find((opt) => opt.value === String(props.modelValue))
})

const handleSelect = (value: string) => {
  emit('update:modelValue', value)
  open.value = false
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
          :disabled="disabled"
          :class="cn(
              'w-full justify-between bg-background font-normal', 
              !selectedOption && 'text-muted-foreground',
              error && 'border-red-500 ring-1 ring-red-500'
          )"
        >
          <div class="flex items-center gap-2 truncate">
              <template v-if="selectedOption">
                   <slot name="trigger" :option="selectedOption">
                        <component v-if="selectedOption.icon && getIconComponent(selectedOption.icon)"
                            :is="getIconComponent(selectedOption.icon)" class="h-4 w-4 shrink-0" />
                        <span v-else-if="selectedOption.icon && getEmoji(selectedOption.icon)" class="text-xs shrink-0">{{ getEmoji(selectedOption.icon) }}</span>
                        <span class="truncate">{{ selectedOption.label }}</span>
                   </slot>
              </template>
              <template v-else>
                  {{ placeholder || 'Select item...' }}
              </template>
          </div>
          <ChevronsUpDown class="ml-2 h-4 w-4 shrink-0 opacity-50" />
        </Button>
      </PopoverTrigger>
      <PopoverContent class="w-[--radix-popover-trigger-width] p-0" align="start">
        <Command>
          <CommandInput :placeholder="placeholder ? `Cari ${placeholder.toLowerCase()}...` : 'Cari...'" />
          <CommandList>
            <CommandEmpty>Tidak ditemukan.</CommandEmpty>
            <CommandGroup>
              <CommandItem
                v-for="option in options"
                :key="option.value"
                :value="option.label"
                @select="() => handleSelect(option.value)"
              >
                <Check
                  :class="cn(
                    'mr-2 h-4 w-4',
                    String(modelValue) === option.value ? 'opacity-100' : 'opacity-0'
                  )"
                />
                
                <slot name="option" :option="option">
                     <div class="flex items-center gap-2">
                        <component v-if="option.icon && getIconComponent(option.icon)" :is="getIconComponent(option.icon)" class="h-4 w-4 shrink-0" />
                        <span v-else-if="option.icon && getEmoji(option.icon)" class="text-base leading-none shrink-0">{{ getEmoji(option.icon) }}</span>
                        {{ option.label }}
                     </div>
                </slot>
              </CommandItem>
            </CommandGroup>
          </CommandList>
        </Command>
      </PopoverContent>
    </Popover>
  </div>
</template>
