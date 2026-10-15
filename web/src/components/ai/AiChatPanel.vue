<script setup lang="ts">
import { nextTick, ref, watch } from 'vue'
import { useRouter } from 'vue-router'
import { useAiStore } from '@/stores/ai'
import AiBubble from './AiBubble.vue'

const ai = useAiStore()
const router = useRouter()

const draft = ref('')
const scroller = ref<HTMLElement | null>(null)

const QUICK = [
  '上海适合 2 人的民宿，预算 800 以内',
  '取消订单有什么规则？退款多久到账？',
  '帮我看看我的订单',
  '平台支持哪些支付方式？',
]

async function toBottom() {
  await nextTick()
  const el = scroller.value
  if (el) el.scrollTop = el.scrollHeight
}

watch(() => ai.items.length, toBottom)
watch(
  () => ai.items[ai.items.length - 1]?.content,
  () => {
    const el = scroller.value
    // 只有已经贴近底部时才自动跟随，避免用户向上翻阅时被拉回
    if (el && el.scrollHeight - el.scrollTop - el.clientHeight < 120) void toBottom()
  },
)

function send(text?: string) {
  const content = (text ?? draft.value).trim()
  if (!content) return
  draft.value = ''
  void ai.send(content).then(toBottom)
}

function onKeydown(e: KeyboardEvent) {
  if (e.key === 'Enter' && !e.shiftKey && !e.isComposing) {
    e.preventDefault()
    send()
  }
}

function goLogin() {
  ai.close()
  router.push({ name: 'login', query: { redirect: '/' } })
}
</script>

<template>
  <div class="pointer-events-none fixed inset-0 z-50 flex justify-center">
    <div
      class="pointer-events-auto relative flex h-full w-full max-w-[480px] flex-col bg-bg shadow-float"
      style="animation: sheetUp 0.28s cubic-bezier(0.16, 1, 0.3, 1) both"
    >
      <!-- 头部 -->
      <header class="flex items-center gap-2 border-b border-line bg-card px-3.5 py-3">
        <div class="flex h-8 w-8 items-center justify-center rounded-full bg-ink text-[13px] font-medium text-white">
          栖
        </div>
        <div class="min-w-0 flex-1">
          <div class="text-[14px] font-medium leading-tight text-ink">小栖 · 智能助手</div>
          <div class="ellipsis text-[11px] leading-tight text-ink-3">
            <template v-if="ai.health?.mock">本地规则模式 · 知识库 {{ ai.health.knowledgeCount }} 条</template>
            <template v-else-if="ai.health">模型 {{ ai.health.provider }} · 知识库 {{ ai.health.knowledgeCount }} 条</template>
            <template v-else>找房 · 查订单 · 解答政策</template>
          </div>
        </div>

        <button
          class="flex h-8 items-center rounded-btn px-2 text-[12.5px] text-ink-2 transition-colors duration-200 hover:bg-fill"
          @click="ai.historyOpen = true; ai.loadConversations()"
        >
          历史
        </button>
        <button
          class="flex h-8 items-center rounded-btn px-2 text-[12.5px] text-ink-2 transition-colors duration-200 hover:bg-fill"
          @click="ai.reset()"
        >
          新对话
        </button>
        <button
          class="flex h-8 w-8 items-center justify-center rounded-full text-ink-2 transition-colors duration-200 hover:bg-fill"
          aria-label="关闭"
          @click="ai.close()"
        >
          <svg viewBox="0 0 20 20" class="h-4 w-4" fill="none" stroke="currentColor" stroke-width="1.6">
            <path d="M5.5 5.5l9 9M14.5 5.5l-9 9" stroke-linecap="round" />
          </svg>
        </button>
      </header>

      <!-- 消息区 -->
      <div ref="scroller" class="flex-1 space-y-3.5 overflow-y-auto px-3.5 py-4">
        <!-- 未登录 -->
        <div v-if="!ai.loggedIn" class="mx-auto mt-16 max-w-[280px] text-center">
          <div class="mx-auto mb-3 flex h-12 w-12 items-center justify-center rounded-full bg-card text-ink-3">
            <svg viewBox="0 0 24 24" class="h-5 w-5" fill="none" stroke="currentColor" stroke-width="1.5">
              <path d="M12 15v2M12 11a4 4 0 100-8 4 4 0 000 8z" stroke-linecap="round" />
              <rect x="4" y="13" width="16" height="8" rx="2" />
            </svg>
          </div>
          <p class="text-[13px] leading-relaxed text-ink-2">登录后可以让小栖帮你找房、查订单、取消订单</p>
          <button class="btn-primary mt-3 h-9 w-full text-[13.5px]" @click="goLogin">去登录</button>
        </div>

        <!-- 空态 -->
        <template v-else-if="!ai.items.length">
          <div class="pt-6 text-center">
            <div class="mx-auto mb-3 flex h-12 w-12 items-center justify-center rounded-full bg-card text-[15px] font-medium text-ink">
              栖
            </div>
            <p class="text-[13.5px] text-ink-2">你好，我是小栖 👋</p>
            <p class="mt-1 text-[12px] text-ink-3">可以帮你找民宿、查订单、解答预订与退改政策</p>
          </div>
          <div class="mt-5 space-y-2">
            <button
              v-for="q in QUICK"
              :key="q"
              class="w-full rounded-card border border-line bg-card px-3.5 py-2.5 text-left text-[13px] text-ink-2 transition-transform duration-200 active:scale-[0.99]"
              @click="send(q)"
            >
              {{ q }}
            </button>
          </div>
        </template>

        <template v-else>
          <AiBubble v-for="item in ai.items" :key="item.key" :item="item" />
        </template>
      </div>

      <!-- 输入区 -->
      <div class="border-t border-line bg-card px-3.5 py-3" style="padding-bottom: calc(12px + var(--safe-bottom))">
        <div class="flex items-end gap-2">
          <textarea
            v-model="draft"
            rows="1"
            :disabled="!ai.loggedIn"
            placeholder="问点什么…"
            class="max-h-[104px] min-h-[40px] flex-1 resize-none rounded-btn bg-fill px-3 py-2.5 text-[14px] leading-snug text-ink outline-none placeholder:text-ink-4 disabled:opacity-50"
            @keydown="onKeydown"
          />
          <button
            v-if="ai.sending"
            class="btn-ghost h-10 w-10 shrink-0"
            aria-label="停止"
            @click="ai.stop()"
          >
            <svg viewBox="0 0 20 20" class="h-4 w-4" fill="currentColor"><rect x="6" y="6" width="8" height="8" rx="1.5" /></svg>
          </button>
          <button
            v-else
            class="btn flex h-10 w-10 shrink-0 items-center justify-center bg-ink text-white disabled:opacity-40"
            :disabled="!draft.trim() || !ai.loggedIn"
            aria-label="发送"
            @click="send()"
          >
            <svg viewBox="0 0 20 20" class="h-4 w-4" fill="none" stroke="currentColor" stroke-width="1.7">
              <path d="M10 16V4M5.5 8.5L10 4l4.5 4.5" stroke-linecap="round" stroke-linejoin="round" />
            </svg>
          </button>
        </div>
        <p class="mt-2 text-center text-[10.5px] text-ink-4">AI 生成内容仅供参考，房源与订单信息以页面为准</p>
      </div>

      <!-- 历史会话抽屉 -->
      <transition name="ai-fade">
        <div v-if="ai.historyOpen" class="absolute inset-0 z-10 flex flex-col bg-black/25" @click.self="ai.historyOpen = false">
          <div class="mt-auto max-h-[70%] overflow-y-auto rounded-t-card bg-card p-3" style="animation: sheetUp .24s cubic-bezier(.16,1,.3,1) both">
            <div class="mb-2 flex items-center justify-between px-1">
              <span class="text-[13.5px] font-medium text-ink">历史会话</span>
              <button class="text-[12.5px] text-ink-3" @click="ai.historyOpen = false">收起</button>
            </div>
            <div v-if="!ai.conversations.length" class="py-8 text-center text-[12.5px] text-ink-3">
              还没有历史会话
            </div>
            <div
              v-for="c in ai.conversations"
              :key="c.id"
              class="flex items-center gap-2 rounded-btn px-2 py-2.5 transition-colors duration-200 hover:bg-fill-2"
            >
              <button class="min-w-0 flex-1 text-left" @click="ai.pickConversation(c.id)">
                <div class="ellipsis text-[13px] text-ink">{{ c.title || '未命名会话' }}</div>
                <div class="ellipsis text-[11px] text-ink-3">{{ c.msgCount }} 条消息</div>
              </button>
              <button class="shrink-0 rounded-btn px-2 py-1 text-[11.5px] text-ink-3 hover:bg-fill" @click="ai.removeConversation(c.id)">
                删除
              </button>
            </div>
          </div>
        </div>
      </transition>
    </div>
  </div>
</template>

<style scoped>
.ai-fade-enter-active,
.ai-fade-leave-active {
  transition: opacity 0.2s ease;
}
.ai-fade-enter-from,
.ai-fade-leave-to {
  opacity: 0;
}
</style>
