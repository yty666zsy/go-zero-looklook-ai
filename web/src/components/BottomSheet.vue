<script setup lang="ts">
import { onBeforeUnmount, watch } from 'vue'

const props = withDefaults(
  defineProps<{
    modelValue: boolean
    title?: string
    /** 是否允许点击遮罩关闭 */
    maskClosable?: boolean
  }>(),
  { title: '', maskClosable: true },
)

const emit = defineEmits<{ 'update:modelValue': [boolean] }>()

function close() {
  emit('update:modelValue', false)
}

function onMask() {
  if (props.maskClosable) close()
}

// 打开时锁滚动
watch(
  () => props.modelValue,
  (v) => {
    document.body.style.overflow = v ? 'hidden' : ''
  },
)

onBeforeUnmount(() => {
  document.body.style.overflow = ''
})
</script>

<template>
  <Teleport to="body">
    <transition name="mask">
      <div
        v-if="modelValue"
        class="fixed inset-0 z-[100] mx-auto flex max-w-[480px] flex-col justify-end bg-black/35"
        @click.self="onMask"
      >
        <div class="sheet max-h-[86vh] overflow-hidden rounded-t-[18px] bg-card">
          <div class="relative flex h-12 items-center justify-center border-b border-line px-4">
            <span class="text-[15px] font-medium text-ink">{{ title }}</span>
            <button
              class="absolute right-3 flex h-8 w-8 items-center justify-center rounded-full text-ink-3 transition active:scale-95"
              aria-label="关闭"
              @click="close"
            >
              <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <path d="M6 6l12 12M18 6L6 18" stroke-linecap="round" />
              </svg>
            </button>
          </div>
          <div class="max-h-[calc(86vh-48px)] overflow-y-auto">
            <slot />
          </div>
        </div>
      </div>
    </transition>
  </Teleport>
</template>

<style scoped>
.mask-enter-active,
.mask-leave-active {
  transition: opacity 0.22s cubic-bezier(0.25, 0.1, 0.25, 1);
}
.mask-enter-from,
.mask-leave-to {
  opacity: 0;
}
.sheet {
  animation: sheetUp 0.3s cubic-bezier(0.16, 1, 0.3, 1) both;
}
@keyframes sheetUp {
  from {
    transform: translateY(100%);
  }
  to {
    transform: translateY(0);
  }
}
</style>
