<script setup lang="ts">
import { ref, computed, watch } from "vue";
import { 
  format, 
  startOfMonth, 
  endOfMonth, 
  startOfWeek, 
  endOfWeek, 
  eachDayOfInterval, 
  isSameDay, 
  isSameMonth, 
  addMonths, 
  subMonths, 
  isWithinInterval,
  isBefore,
  isAfter
} from "date-fns";
import { id } from "date-fns/locale";
import { ChevronLeft, ChevronRight } from "lucide-vue-next";

const props = defineProps<{
  startDate?: Date;
  endDate?: Date;
}>();

const emit = defineEmits<{
  (e: "update:startDate", value: Date): void;
  (e: "update:endDate", value: Date): void;
  (e: "update:range", value: { start: Date, end: Date }): void;
  (e: "apply"): void;
}>();

const localStartDate = ref<Date | undefined>(props.startDate);
const localEndDate = ref<Date | undefined>(props.endDate);
const currentMonth = ref(props.startDate || new Date());

// Sync props to local if they change externally
watch(() => props.startDate, (val) => { if(val) localStartDate.value = val; currentMonth.value = val || new Date(); });
watch(() => props.endDate, (val) => { if(val) localEndDate.value = val; });

const days = computed(() => {
  const monthStart = startOfMonth(currentMonth.value);
  const monthEnd = endOfMonth(currentMonth.value);
  const startDate = startOfWeek(monthStart);
  const endDate = endOfWeek(monthEnd);
  
  return eachDayOfInterval({ start: startDate, end: endDate });
});

const weekDays = ["Min", "Sen", "Sel", "Rab", "Kam", "Jum", "Sab"];

const isSelected = (date: Date) => {
  if (localStartDate.value && isSameDay(date, localStartDate.value)) return true;
  if (localEndDate.value && isSameDay(date, localEndDate.value)) return true;
  return false;
};

const isInRange = (date: Date) => {
  if (localStartDate.value && localEndDate.value) {
    const start = isBefore(localStartDate.value, localEndDate.value) ? localStartDate.value : localEndDate.value;
    const end = isAfter(localEndDate.value, localStartDate.value) ? localEndDate.value : localStartDate.value;
    return isWithinInterval(date, { start, end });
  }
  return false;
};

const handleDateClick = (date: Date) => {
    // 1. If nothing selected, set start
    if (!localStartDate.value && !localEndDate.value) {
        localStartDate.value = date;
        return;
    }
    
    // 2. If both selected, reset and start new
    if (localStartDate.value && localEndDate.value) {
        localStartDate.value = date;
        localEndDate.value = undefined;
        return;
    }
    
    // 3. If only start selected
    if (localStartDate.value) {
        if (isBefore(date, localStartDate.value)) {
            // If clicked before start, update start
            localStartDate.value = date; 
        } else {
             // Set end
             localEndDate.value = date;
        }
    }
};

const prevMonth = () => {
    currentMonth.value = subMonths(currentMonth.value, 1);
};

const nextMonth = () => {
    currentMonth.value = addMonths(currentMonth.value, 1);
};

const apply = () => {
    if (localStartDate.value && localEndDate.value) {
        // Emit atomic update
        emit("update:range", { start: localStartDate.value, end: localEndDate.value });
        emit("apply");
    }
};

</script>

<template>
  <div class="bg-card border border-border rounded-xl p-4 w-full shadow-sm max-w-[350px]">
    <!-- Header -->
    <div class="flex items-center justify-between mb-4">
        <button @click="prevMonth" class="p-1 hover:bg-muted rounded-md transition-colors">
            <ChevronLeft class="h-5 w-5" />
        </button>
        <span class="font-semibold text-sm">
            {{ format(currentMonth, "MMMM yyyy", { locale: id }) }}
        </span>
        <button @click="nextMonth" class="p-1 hover:bg-muted rounded-md transition-colors">
            <ChevronRight class="h-5 w-5" />
        </button>
    </div>

    <!-- Weekdays -->
    <div class="grid grid-cols-7 mb-2 text-center text-xs text-muted-foreground font-medium">
        <div v-for="day in weekDays" :key="day" class="py-1">{{ day }}</div>
    </div>

    <!-- Calendar Grid -->
    <div class="grid grid-cols-7 gap-y-1 text-sm">
        <div v-for="date in days" :key="date.toString()" class="relative p-[2px]">
            <div 
                :class="[
                    'h-9 w-9 flex items-center justify-center rounded-full cursor-pointer transition-colors relative z-10',
                    isSameMonth(date, currentMonth) ? 'text-foreground' : 'text-muted-foreground opacity-50',
                    isSelected(date) ? 'bg-primary text-primary-foreground font-bold hover:bg-primary/90' : 'hover:bg-muted',
                    isInRange(date) && !isSelected(date) ? 'bg-blue-100 dark:bg-blue-900/30' : ''
                ]"
                @click="handleDateClick(date)"
            >
                {{ format(date, "d") }}
            </div>
             <!-- Background for range connector -->
             <div v-if="isInRange(date)" 
                :class="[
                    'absolute top-[2px] bottom-[2px] h-9 w-full bg-blue-100 dark:bg-blue-900/30 z-0',
                    isSameDay(date, localStartDate!) ? 'rounded-l-full left-1/2 w-1/2' : '',
                    isSameDay(date, localEndDate!) ? 'rounded-r-full right-1/2 w-1/2' : ''
                ]"
            ></div>
        </div>
    </div>

    <!-- Footer actions -->
    <div class="mt-4 pt-4 border-t border-border flex justify-end">
        <button 
            @click="apply"
            :disabled="!localStartDate || !localEndDate"
            class="bg-blue-600 text-white px-4 py-2 rounded-lg text-sm font-semibold hover:bg-blue-700 disabled:opacity-50 disabled:cursor-not-allowed transition-colors w-full"
        >
            Apply
        </button>
    </div>
  </div>
</template>
