<script setup lang="ts">
const props = withDefaults(
  defineProps<{
    modelValue: number
    min?: number
    max?: number
    step?: number
    suffix?: string
  }>(),
  { min: 1, max: 10, step: 1, suffix: '' },
)

const emit = defineEmits<{ 'update:modelValue': [number] }>()

function change(delta: number) {
  const next = Math.min(props.max, Math.max(props.min, props.modelValue + delta * props.step))
  if (next !== props.modelValue) emit('update:modelValue', next)
}
</script>

<template>
  <div class="flex items-center gap-3">
    <button
      class="flex h-8 w-8 items-center justify-center rounded-full border border-line-2 text-ink-2 transition active:scale-90 disabled:opacity-35"
      :disabled="modelValue <= min"
      aria-label="减少"
      @click="change(-1)"
    >
      <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.2">
        <path d="M5 12h14" stroke-linecap="round" />
      </svg>
    </button>
    <span class="min-w-[52px] text-center text-[14px] tabular-nums text-ink">
      {{ modelValue }}{{ suffix }}
    </span>
    <button
      class="flex h-8 w-8 items-center justify-center rounded-full border border-line-2 text-ink-2 transition active:scale-90 disabled:opacity-35"
      :disabled="modelValue >= max"
      aria-label="增加"
      @click="change(1)"
    >
      <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.2">
        <path d="M12 5v14M5 12h14" stroke-linecap="round" />
      </svg>
    </button>
  </div>
</template>
