<script setup lang="ts">
withDefaults(
  defineProps<{
    title?: string
    desc?: string
  }>(),
  { title: '暂无数据', desc: '' },
)
</script>

<template>
  <div class="flex flex-col items-center justify-center gap-3 px-6 py-16 text-center animate-fade-in">
    <div class="flex h-12 w-12 items-center justify-center rounded-full bg-fill">
      <svg viewBox="0 0 24 24" class="h-5 w-5 text-ink-3" fill="none" stroke="currentColor" stroke-width="1.6" stroke-linecap="round" stroke-linejoin="round">
        <path d="M4 7h16M4 12h10M4 17h7" />
      </svg>
    </div>
    <div>
      <div class="text-sm font-medium text-ink-2">{{ title }}</div>
      <div v-if="desc" class="mt-1 text-[12.5px] text-ink-3">{{ desc }}</div>
    </div>
    <slot />
  </div>
</template>
