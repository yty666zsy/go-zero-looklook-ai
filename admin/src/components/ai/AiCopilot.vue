<script setup lang="ts">
import { computed } from 'vue'
import { useAiStore } from '@/stores/ai'
import AiCopilotCore from './AiCopilotCore.vue'

const ai = useAiStore()

const modeText = computed(() => {
  if (!ai.health) return '连接中…'
  return ai.health.mock ? '本地规则模式' : `模型 ${ai.health.model || ai.health.provider}`
})
</script>

<template>
  <div class="fixed inset-0 z-50 flex justify-end">
    <!-- 遮罩 -->
    <div class="absolute inset-0 bg-black/20 animate-fade-in" @click="ai.close()" />

    <!-- 抽屉 -->
    <aside
      class="relative flex h-full w-full max-w-[480px] flex-col border-l border-line bg-bg shadow-float"
      style="animation: drawerIn 0.28s var(--ease-out) both"
    >
      <header class="flex shrink-0 items-center gap-2.5 border-b border-line bg-card px-4 py-3">
        <span class="flex h-8 w-8 items-center justify-center rounded-[9px] bg-ink text-[12px] font-semibold text-white">AI</span>
        <div class="min-w-0 flex-1 leading-tight">
          <div class="text-[14px] font-semibold text-ink">运营 Copilot</div>
          <div class="ellipsis text-[11px] text-ink-3">
            {{ modeText }} · 知识库 {{ ai.health?.knowledgeCount ?? 0 }} 条
          </div>
        </div>

        <button class="btn btn-line btn-sm" @click="ai.historyOpen = !ai.historyOpen; ai.loadConversations()">历史</button>
        <button class="btn btn-line btn-sm" @click="ai.reset()">新对话</button>
        <button class="btn btn-ghost btn-sm !px-2" title="关闭" @click="ai.close()">
          <svg viewBox="0 0 20 20" class="h-4 w-4" fill="none" stroke="currentColor" stroke-width="1.7">
            <path d="M5.5 5.5l9 9M14.5 5.5l-9 9" stroke-linecap="round" />
          </svg>
        </button>
      </header>

      <div class="relative flex min-h-0 flex-1 flex-col">
        <AiCopilotCore variant="drawer" />

        <!-- 历史会话 -->
        <transition name="overlay">
          <div v-if="ai.historyOpen" class="absolute inset-0 z-10 flex flex-col bg-black/20" @click.self="ai.historyOpen = false">
            <div class="mt-auto max-h-[72%] overflow-y-auto rounded-t-card bg-card p-3 animate-fade-up">
              <div class="mb-2 flex items-center justify-between px-1">
                <span class="text-[13.5px] font-medium text-ink">历史会话</span>
                <button class="text-[12.5px] text-ink-3 hover:text-ink" @click="ai.historyOpen = false">收起</button>
              </div>
              <div v-if="!ai.conversations.length" class="py-8 text-center text-[12.5px] text-ink-3">还没有历史会话</div>
              <div
                v-for="c in ai.conversations"
                :key="c.id"
                class="group flex items-center gap-2 rounded-btn px-2 py-2.5 transition-colors duration-150 hover:bg-fill-2"
              >
                <button class="min-w-0 flex-1 text-left" @click="ai.pickConversation(c.id)">
                  <div class="ellipsis text-[13px] text-ink">{{ c.title || '未命名会话' }}</div>
                  <div class="ellipsis text-[11.5px] text-ink-3">{{ c.msgCount }} 条消息 · {{ c.lastMsg }}</div>
                </button>
                <button
                  class="shrink-0 rounded-btn px-2 py-1 text-[11.5px] text-ink-3 opacity-0 transition-opacity duration-150 hover:bg-fill hover:text-st-err group-hover:opacity-100"
                  @click="ai.removeConversation(c.id)"
                >
                  删除
                </button>
              </div>
            </div>
          </div>
        </transition>
      </div>
    </aside>
  </div>
</template>

<style scoped>
@keyframes drawerIn {
  from {
    transform: translateX(24px);
    opacity: 0;
  }
  to {
    transform: translateX(0);
    opacity: 1;
  }
}
</style>
