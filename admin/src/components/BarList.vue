<script setup lang="ts">
import { computed } from 'vue'

interface Item {
  label: string
  value: number
  sub?: string
}

const props = withDefaults(
  defineProps<{
    items: Item[]
    /** 数值展示前缀，如 ¥ */
    prefix?: string
    /** 条的颜色，默认墨色 */
    tone?: 'ink' | 'gold' | 'soft'
  }>(),
  { prefix: '', tone: 'ink' },
)

const max = computed(() => Math.max(1, ...props.items.map((i) => i.value)))

const barCls = computed(() => {
  if (props.tone === 'gold') return 'bg-[#D9C49A]'
  if (props.tone === 'soft') return 'bg-ink-4'
  return 'bg-ink/70'
})

function fmt(v: number) {
  const s = (Math.round(v * 100) / 100).toString()
  const [int, dec] = s.split('.')
  return int.replace(/\B(?=(\d{3})+(?!\d))/g, ',') + (dec ? `.${dec}` : '')
}
</script>

<template>
  <ul class="space-y-3">
    <li v-for="(it, i) in items" :key="it.label" class="group">
      <div class="mb-1.5 flex items-baseline justify-between gap-3 text-[12.5px]">
        <span class="text-ink-2">{{ it.label }}</span>
        <span class="num text-ink-3">
          {{ prefix }}{{ fmt(it.value) }}<span v-if="it.sub" class="ml-1 text-ink-4">{{ it.sub }}</span>
        </span>
      </div>
      <div class="h-1.5 w-full overflow-hidden rounded-pill bg-fill">
        <div
          class="h-full rounded-pill transition-[width] duration-500 ease-out"
          :class="barCls"
          :style="{ width: `${(it.value / max) * 100}%`, transitionDelay: `${i * 40}ms` }"
        />
      </div>
    </li>
    <li v-if="items.length === 0" class="py-6 text-center text-[12.5px] text-ink-3">暂无数据</li>
  </ul>
</template>
