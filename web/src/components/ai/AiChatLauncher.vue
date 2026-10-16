<script setup lang="ts">
import { computed } from 'vue'
import { useRoute } from 'vue-router'
import { useAiStore } from '@/stores/ai'
import AiChatPanel from './AiChatPanel.vue'

const ai = useAiStore()
const route = useRoute()

// 登录页不展示助手入口
const hidden = computed(() => route.name === 'login')
// 有底部 TabBar / 固定操作栏的页面把按钮抬高一点，避免遮挡
const bottom = computed(() => (route.meta.tab ? 84 : route.meta.bottomBar ? 88 : 26))
</script>

<template>
  <template v-if="!hidden">
    <transition name="launcher">
      <div v-show="!ai.open" class="pointer-events-none fixed inset-x-0 bottom-0 z-40 flex justify-center">
        <div class="relative w-full max-w-[480px]">
          <button
            class="pointer-events-auto absolute right-4 flex h-12 items-center gap-2 rounded-pill bg-ink pl-3.5 pr-4 text-white shadow-float transition-transform duration-200 active:scale-[0.96]"
            :style="{ bottom: `${bottom}px` }"
            aria-label="打开智能助手"
            @click="ai.toggle()"
          >
            <svg viewBox="0 0 20 20" class="h-4 w-4" fill="none" stroke="currentColor" stroke-width="1.5">
              <path d="M10 3.5l1.3 3.2 3.2 1.3-3.2 1.3L10 12.5 8.7 9.3 5.5 8l3.2-1.3L10 3.5z" stroke-linejoin="round" />
              <circle cx="15.2" cy="14.6" r="1.6" />
            </svg>
            <span class="text-[13px] font-medium">小栖</span>
          </button>
        </div>
      </div>
    </transition>

    <!-- 面板自带入场动画，这里不再包 transition，避免离场残留 -->
    <AiChatPanel v-if="ai.open" />
  </template>
</template>

<style scoped>
.launcher-enter-active,
.launcher-leave-active {
  transition: opacity 0.2s ease, transform 0.24s cubic-bezier(0.16, 1, 0.3, 1);
}
.launcher-enter-from,
.launcher-leave-to {
  opacity: 0;
  transform: translateY(12px) scale(0.96);
}
</style>
