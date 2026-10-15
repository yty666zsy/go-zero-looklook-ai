<script setup lang="ts">
import { useToast } from '@/composables/useToast'

const { items } = useToast()
</script>

<template>
  <Teleport to="body">
    <div class="pointer-events-none fixed inset-x-0 top-6 z-[120] flex flex-col items-center gap-2">
      <transition-group name="toast">
        <div
          v-for="item in items"
          :key="item.id"
          class="max-w-[80vw] rounded-pill px-4 py-2.5 text-[13px] leading-none shadow-float
                 backdrop-blur bg-ink/90 text-white"
          :class="{
            'bg-ink/90': item.type === 'info',
            'bg-[#3F4B3F]/92': item.type === 'success',
            'bg-[#5A3A3A]/92': item.type === 'error',
          }"
        >
          {{ item.text }}
        </div>
      </transition-group>
    </div>
  </Teleport>
</template>

<style scoped>
.toast-enter-active,
.toast-leave-active {
  transition: opacity 0.22s cubic-bezier(0.16, 1, 0.3, 1), transform 0.22s cubic-bezier(0.16, 1, 0.3, 1);
}
.toast-enter-from,
.toast-leave-to {
  opacity: 0;
  transform: translateY(-8px) scale(0.96);
}
</style>
