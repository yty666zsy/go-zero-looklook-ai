<script setup lang="ts">
import { computed } from 'vue'

const props = defineProps<{
  page: number
  pageSize: number
  total: number
}>()

const emit = defineEmits<{
  (e: 'update:page', v: number): void
  (e: 'update:pageSize', v: number): void
}>()

const pageCount = computed(() => Math.max(1, Math.ceil(props.total / Math.max(1, props.pageSize))))

const from = computed(() => (props.total === 0 ? 0 : (props.page - 1) * props.pageSize + 1))
const to = computed(() => Math.min(props.total, props.page * props.pageSize))

/** 页码窗口：1 … 4 5 6 … 20 */
const pages = computed<(number | '...')[]>(() => {
  const last = pageCount.value
  const cur = props.page
  if (last <= 7) return Array.from({ length: last }, (_, i) => i + 1)
  const out: (number | '...')[] = [1]
  const start = Math.max(2, cur - 1)
  const end = Math.min(last - 1, cur + 1)
  if (start > 2) out.push('...')
  for (let i = start; i <= end; i++) out.push(i)
  if (end < last - 1) out.push('...')
  out.push(last)
  return out
})

function go(p: number) {
  if (p < 1 || p > pageCount.value || p === props.page) return
  emit('update:page', p)
}

function onSize(e: Event) {
  const v = Number((e.target as HTMLSelectElement).value)
  emit('update:pageSize', v)
  emit('update:page', 1)
}
</script>

<template>
  <div class="flex flex-wrap items-center justify-between gap-3 border-t border-line px-4 py-3">
    <div class="text-[12.5px] text-ink-3">
      共 <span class="num text-ink-2">{{ total }}</span> 条 · 当前
      <span class="num text-ink-2">{{ from }}-{{ to }}</span>
    </div>

    <div class="flex items-center gap-3">
      <select class="select h-8 w-[104px] text-[12.5px]" :value="pageSize" @change="onSize">
        <option :value="10">10 条/页</option>
        <option :value="20">20 条/页</option>
        <option :value="50">50 条/页</option>
      </select>

      <div class="flex items-center gap-1">
        <button class="btn btn-line h-8 w-8 p-0" :disabled="page <= 1" @click="go(page - 1)">
          <svg viewBox="0 0 24 24" class="h-3.5 w-3.5" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <path d="M15 5 8 12l7 7" />
          </svg>
        </button>

        <template v-for="(p, i) in pages" :key="`${p}-${i}`">
          <span v-if="p === '...'" class="px-1 text-[12.5px] text-ink-4">…</span>
          <button
            v-else
            class="btn h-8 min-w-8 px-2 text-[12.5px]"
            :class="p === page ? 'bg-ink text-white' : 'border border-line-2 bg-card text-ink-2 hover:bg-fill-2'"
            @click="go(p)"
          >
            {{ p }}
          </button>
        </template>

        <button class="btn btn-line h-8 w-8 p-0" :disabled="page >= pageCount" @click="go(page + 1)">
          <svg viewBox="0 0 24 24" class="h-3.5 w-3.5" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <path d="m9 5 7 7-7 7" />
          </svg>
        </button>
      </div>
    </div>
  </div>
</template>
