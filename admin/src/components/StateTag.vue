<script setup lang="ts">
import { computed } from 'vue'

const props = withDefaults(
  defineProps<{
    tone?: 'ok' | 'warn' | 'err' | 'off'
    text: string
    dot?: boolean
  }>(),
  { tone: 'off', dot: true },
)

const cls = computed(() => {
  switch (props.tone) {
    case 'ok':
      return 'bg-[#EDF2ED] text-st-ok'
    case 'warn':
      return 'bg-[#F7F1E3] text-st-warn'
    case 'err':
      return 'bg-[#F7EDEC] text-st-err'
    default:
      return 'bg-fill text-st-off'
  }
})

const dotCls = computed(() => {
  switch (props.tone) {
    case 'ok':
      return 'bg-st-ok'
    case 'warn':
      return 'bg-st-warn'
    case 'err':
      return 'bg-st-err'
    default:
      return 'bg-st-off'
  }
})
</script>

<template>
  <span class="tag" :class="cls">
    <span v-if="dot" :class="['h-1.5 w-1.5 rounded-full', dotCls]" />
    {{ text }}
  </span>
</template>
