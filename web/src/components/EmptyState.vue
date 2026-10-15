<script setup lang="ts">
withDefaults(
  defineProps<{
    text?: string
    tip?: string
    actionText?: string
  }>(),
  { text: '这里还什么都没有', tip: '', actionText: '' },
)

defineEmits<{ action: [] }>()
</script>

<template>
  <div class="flex flex-col items-center justify-center px-8 py-16 text-center">
    <div class="mb-4 flex h-16 w-16 items-center justify-center rounded-full bg-fill">
      <svg width="26" height="26" viewBox="0 0 24 24" fill="none" stroke="#9AA0A8" stroke-width="1.5">
        <circle cx="11" cy="11" r="6.5" />
        <path d="M15.8 15.8L21 21" stroke-linecap="round" />
      </svg>
    </div>
    <p class="text-[14px] text-ink-2">{{ text }}</p>
    <p v-if="tip" class="mt-1 text-[12.5px] text-ink-3">{{ tip }}</p>
    <button v-if="actionText" class="btn-ghost mt-5 px-5 py-2 text-[13px]" @click="$emit('action')">
      {{ actionText }}
    </button>
  </div>
</template>
