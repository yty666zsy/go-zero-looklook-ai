<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import BottomSheet from './BottomSheet.vue'
import { todayStart } from '@/utils/format'

const props = defineProps<{
  modelValue: boolean
  start: number
  end: number
}>()

const emit = defineEmits<{
  'update:modelValue': [boolean]
  change: [start: number, end: number]
}>()

const DAY = 86400

/** 选中的起点，用来判断第一次点击是「选开始」还是「选结束」 */
const picking = ref<'start' | 'end'>('start')
const tempStart = ref(props.start)
const tempEnd = ref(props.end)

watch(
  () => props.modelValue,
  (v) => {
    if (v) {
      tempStart.value = props.start
      tempEnd.value = props.end
      picking.value = 'start'
      // 默认滚到已选月份
      monthOffset.value = monthsBetween(new Date(), new Date(props.start * 1000))
    }
  },
)

const monthOffset = ref(0)

function monthsBetween(a: Date, b: Date) {
  return (b.getFullYear() - a.getFullYear()) * 12 + (b.getMonth() - a.getMonth())
}

interface Cell {
  ts: number
  day: number
  disabled: boolean
  isStart: boolean
  isEnd: boolean
  inRange: boolean
  isToday: boolean
}

const today = todayStart()

/** 展示 3 个月 */
const months = computed(() => {
  const list: { label: string; cells: (Cell | null)[] }[] = []
  const base = new Date()
  base.setDate(1)
  base.setHours(0, 0, 0, 0)
  base.setMonth(base.getMonth() + monthOffset.value)

  for (let m = 0; m < 3; m++) {
    const d = new Date(base)
    d.setMonth(d.getMonth() + m)
    const year = d.getFullYear()
    const month = d.getMonth()
    const firstWeekday = new Date(year, month, 1).getDay()
    const daysInMonth = new Date(year, month + 1, 0).getDate()

    const cells: (Cell | null)[] = []
    for (let i = 0; i < firstWeekday; i++) cells.push(null)

    for (let day = 1; day <= daysInMonth; day++) {
      const ts = Math.floor(new Date(year, month, day).getTime() / 1000)
      const s = Math.min(tempStart.value, tempEnd.value)
      const e = Math.max(tempStart.value, tempEnd.value)
      cells.push({
        ts,
        day,
        disabled: ts < today,
        isStart: ts === s,
        isEnd: ts === e,
        inRange: ts > s && ts < e,
        isToday: ts === today,
      })
    }
    list.push({ label: `${year}年${month + 1}月`, cells })
  }
  return list
})

function pick(cell: Cell) {
  if (cell.disabled) return

  if (picking.value === 'start') {
    tempStart.value = cell.ts
    tempEnd.value = 0
    picking.value = 'end'
    return
  }

  if (cell.ts <= tempStart.value) {
    // 点到开始之前，直接重开一轮
    tempStart.value = cell.ts
    tempEnd.value = 0
    picking.value = 'end'
    return
  }

  tempEnd.value = cell.ts
  picking.value = 'start'
}

const nights = computed(() => {
  if (!tempStart.value || !tempEnd.value) return 0
  return Math.round((tempEnd.value - tempStart.value) / DAY)
})

function confirm() {
  if (!tempStart.value || !tempEnd.value) return
  emit('change', tempStart.value, tempEnd.value)
  emit('update:modelValue', false)
}

const weekNames = ['日', '一', '二', '三', '四', '五', '六']

function label(ts: number) {
  if (!ts) return '请选择'
  const d = new Date(ts * 1000)
  return `${d.getMonth() + 1}月${d.getDate()}日`
}
</script>

<template>
  <BottomSheet :model-value="modelValue" title="选择入住日期" @update:model-value="$emit('update:modelValue', $event)">
    <div class="flex items-center justify-between px-5 py-3.5 text-[13px]">
      <div class="flex flex-col">
        <span class="text-[11px] text-ink-3">入住</span>
        <span class="mt-0.5 font-medium" :class="tempStart ? 'text-ink' : 'text-ink-3'">{{ label(tempStart) }}</span>
      </div>
      <div class="mx-4 flex flex-1 items-center gap-2">
        <span class="h-px flex-1 bg-line" />
        <span class="rounded-pill bg-fill px-2.5 py-1 text-[11px] text-ink-2">{{ nights }} 晚</span>
        <span class="h-px flex-1 bg-line" />
      </div>
      <div class="flex flex-col text-right">
        <span class="text-[11px] text-ink-3">离店</span>
        <span class="mt-0.5 font-medium" :class="tempEnd ? 'text-ink' : 'text-ink-3'">{{ label(tempEnd) }}</span>
      </div>
    </div>

    <div class="border-t border-line px-4 pb-4">
      <div v-for="m in months" :key="m.label" class="pt-4">
        <p class="mb-2 text-center text-[13px] font-medium text-ink">{{ m.label }}</p>
        <div class="mb-1 grid grid-cols-7">
          <span v-for="w in weekNames" :key="w" class="py-1 text-center text-[11px] text-ink-3">{{ w }}</span>
        </div>
        <div class="grid grid-cols-7 gap-y-1">
          <template v-for="(c, i) in m.cells" :key="i">
            <span v-if="!c" />
            <button
              v-else
              class="relative mx-auto flex h-9 w-9 items-center justify-center rounded-full text-[13px] transition duration-150"
              :class="[
                c.disabled ? 'text-ink-4' : 'text-ink active:scale-90',
                c.inRange ? 'bg-fill' : '',
                c.isStart || c.isEnd ? 'bg-ink text-white' : '',
              ]"
              :disabled="c.disabled"
              @click="pick(c)"
            >
              {{ c.day }}
              <span v-if="c.isToday && !c.isStart && !c.isEnd" class="absolute bottom-0.5 h-1 w-1 rounded-full bg-ink-3" />
            </button>
          </template>
        </div>
      </div>
    </div>

    <div class="sticky bottom-0 border-t border-line bg-card px-4 py-3" style="padding-bottom: calc(12px + var(--safe-bottom))">
      <button
        class="btn-primary h-11 w-full text-[15px] disabled:opacity-40"
        :disabled="!tempStart || !tempEnd"
        @click="confirm"
      >
        确定
      </button>
    </div>
  </BottomSheet>
</template>
