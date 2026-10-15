<script setup lang="ts">
import { watch } from 'vue'

const props = withDefaults(
  defineProps<{
    modelValue: boolean
    title?: string
    desc?: string
    width?: string
    closeOnOverlay?: boolean
  }>(),
  {
    title: '',
    desc: '',
    width: '440px',
    closeOnOverlay: true,
  },
)

const emit = defineEmits<{
  (e: 'update:modelValue', v: boolean): void
}>()

function close() {
  emit('update:modelValue', false)
}

function onOverlay() {
  if (props.closeOnOverlay) close()
}

function onKey(e: KeyboardEvent) {
  if (e.key === 'Escape') close()
}

watch(
  () => props.modelValue,
  (open) => {
    if (open) window.addEventListener('keydown', onKey)
    else window.removeEventListener('keydown', onKey)
  },
)
</script>

<template>
  <teleport to="body">
    <transition name="overlay">
      <div v-if="modelValue" class="fixed inset-0 z-50 bg-ink/25 backdrop-blur-[2px]" @click="onOverlay" />
    </transition>
    <transition name="dialog">
      <div
        v-if="modelValue"
        class="fixed left-1/2 top-1/2 z-50 max-h-[86vh] -translate-x-1/2 -translate-y-1/2 overflow-hidden rounded-card border border-line bg-card shadow-float"
        :style="{ width }"
        role="dialog"
        aria-modal="true"
      >
        <div class="flex items-start justify-between gap-4 border-b border-line px-5 py-4">
          <div class="min-w-0">
            <h3 class="text-[15px] font-semibold tracking-tight text-ink">{{ title }}</h3>
            <p v-if="desc" class="mt-0.5 text-[12.5px] leading-relaxed text-ink-3">{{ desc }}</p>
          </div>
          <button
            class="flex h-7 w-7 shrink-0 items-center justify-center rounded-btn text-ink-3 transition-colors duration-200 hover:bg-fill hover:text-ink"
            @click="close"
          >
            <svg viewBox="0 0 24 24" class="h-4 w-4" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round">
              <path d="M6 6l12 12M18 6 6 18" />
            </svg>
          </button>
        </div>

        <div class="max-h-[62vh] overflow-y-auto px-5 py-4">
          <slot />
        </div>

        <div v-if="$slots.footer" class="flex items-center justify-end gap-2 border-t border-line bg-fill-2 px-5 py-3.5">
          <slot name="footer" />
        </div>
      </div>
    </transition>
  </teleport>
</template>
