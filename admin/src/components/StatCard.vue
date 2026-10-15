<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'

const props = withDefaults(
  defineProps<{
    label: string
    value: number
    /** 环比：昨日值，用于算涨跌 */
    prev?: number
    prefix?: string
    suffix?: string
    /** 保留小数位 */
    digits?: number
    hint?: string
    /** 动画数字 */
    animated?: boolean
  }>(),
  { prefix: '', suffix: '', digits: 0, hint: '', animated: true, prev: undefined },
)

const shown = ref(0)

let token = 0

function run() {
  const mine = ++token
  if (!props.animated) {
    shown.value = props.value
    return
  }
  const target = props.value
  const start = performance.now()
  const dur = 520

  const step = (now: number) => {
    if (mine !== token) return
    const t = Math.min(1, (now - start) / dur)
    const eased = 1 - Math.pow(1 - t, 3)
    shown.value = target * eased
    if (t < 1) requestAnimationFrame(step)
    else shown.value = target
  }
  requestAnimationFrame(step)

  // 兜底：标签页在后台时 rAF 会被节流/暂停，这里保证最终值一定落到位
  window.setTimeout(() => {
    if (mine === token) shown.value = target
  }, dur + 120)
}

onMounted(run)
watch(() => props.value, run)

const display = computed(() => {
  const n = shown.value
  const fixed = n.toFixed(props.digits)
  const [int, dec] = fixed.split('.')
  const sep = int.replace(/\B(?=(\d{3})+(?!\d))/g, ',')
  return dec ? `${sep}.${dec}` : sep
})

const diff = computed(() => {
  if (props.prev === undefined) return null
  const d = props.value - props.prev
  if (props.prev === 0) return d > 0 ? 100 : 0
  return (d / Math.abs(props.prev)) * 100
})
</script>

<template>
  <div class="card group relative overflow-hidden p-4 transition-shadow duration-300 hover:shadow-float">
    <div class="flex items-start justify-between gap-2">
      <span class="text-[12.5px] text-ink-3">{{ label }}</span>
      <span
        v-if="diff !== null"
        class="tag"
        :class="diff > 0 ? 'bg-[#F7F1E3] text-st-warn' : diff < 0 ? 'bg-[#EDF2ED] text-st-ok' : 'bg-fill text-st-off'"
      >
        <svg viewBox="0 0 24 24" class="h-3 w-3" :class="diff < 0 ? 'rotate-180' : ''" fill="currentColor">
          <path d="M12 5l6 8h-4v6h-4v-6H6l6-8Z" />
        </svg>
        {{ diff > 0 ? '+' : '' }}{{ diff.toFixed(1) }}%
      </span>
    </div>

    <div class="mt-2.5 flex items-baseline gap-1">
      <span v-if="prefix" class="text-[15px] font-medium text-ink-2">{{ prefix }}</span>
      <span class="num text-[26px] font-semibold leading-none tracking-tight text-ink">{{ display }}</span>
      <span v-if="suffix" class="text-[13px] text-ink-3">{{ suffix }}</span>
    </div>

    <div v-if="hint" class="mt-2 text-[12px] text-ink-3">{{ hint }}</div>

    <!-- hover 时底部一条细线扫过 -->
    <span
      class="pointer-events-none absolute inset-x-0 bottom-0 h-[2px] origin-left scale-x-0 bg-ink/15 transition-transform duration-300 ease-out group-hover:scale-x-100"
    />
  </div>
</template>
