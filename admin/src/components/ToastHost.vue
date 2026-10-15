<script setup lang="ts">
import { useToast } from '@/composables/useToast'

const { toasts, dismiss } = useToast()

const dot: Record<string, string> = {
  success: 'bg-st-ok',
  error: 'bg-st-err',
  info: 'bg-ink-2',
}
</script>

<template>
  <div class="pointer-events-none fixed inset-x-0 top-4 z-[100] flex flex-col items-center gap-2">
    <transition-group name="toast">
      <div
        v-for="t in toasts"
        :key="t.id"
        class="pointer-events-auto flex items-center gap-2.5 rounded-btn border border-line bg-card/95 px-4 py-2.5 text-sm text-ink shadow-float backdrop-blur"
        @click="dismiss(t.id)"
      >
        <span :class="['h-1.5 w-1.5 rounded-full', dot[t.type]]" />
        <span>{{ t.text }}</span>
      </div>
    </transition-group>
  </div>
</template>

<style scoped>
.toast-enter-active,
.toast-leave-active {
  transition: opacity 0.22s cubic-bezier(0.25, 0.1, 0.25, 1), transform 0.24s cubic-bezier(0.16, 1, 0.3, 1);
}
.toast-enter-from,
.toast-leave-to {
  opacity: 0;
  transform: translateY(-8px);
}
.toast-leave-active {
  position: absolute;
}
</style>
