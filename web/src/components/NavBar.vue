<script setup lang="ts">
import { useRouter } from 'vue-router'

withDefaults(
  defineProps<{
    title?: string
    /** 是否显示返回按钮 */
    back?: boolean
    /** 透明浮在内容上（用于详情页大图） */
    floating?: boolean
    /** 是否已经滚过阈值，floating 模式下用来切换实底 */
    solid?: boolean
  }>(),
  { title: '', back: true, floating: false, solid: true },
)

const router = useRouter()

function onBack() {
  if (window.history.length > 1) router.back()
  else router.replace('/')
}
</script>

<template>
  <header
    class="sticky top-0 z-40 h-12 w-full transition-colors duration-300"
    :class="
      floating && !solid
        ? 'bg-transparent'
        : 'bg-card/85 backdrop-blur-md border-b border-line'
    "
  >
    <div class="relative flex h-12 items-center px-3">
      <button
        v-if="back"
        class="flex h-9 w-9 items-center justify-center rounded-full transition active:scale-95"
        :class="floating && !solid ? 'bg-black/25 text-white' : 'text-ink'"
        aria-label="返回"
        @click="onBack"
      >
        <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <path d="M15 5l-7 7 7 7" stroke-linecap="round" stroke-linejoin="round" />
        </svg>
      </button>

      <h1
        class="pointer-events-none absolute inset-x-0 text-center text-[15px] font-medium transition-colors duration-300"
        :class="floating && !solid ? 'text-white/0' : 'text-ink'"
      >
        {{ title }}
      </h1>

      <div class="ml-auto flex items-center gap-2">
        <slot name="right" />
      </div>
    </div>
  </header>
</template>
