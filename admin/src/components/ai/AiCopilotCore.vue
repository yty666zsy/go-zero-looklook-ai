<script setup lang="ts">
import { nextTick, ref, watch } from 'vue'
import { useAiStore } from '@/stores/ai'
import AiBubble from './AiBubble.vue'

const props = withDefaults(defineProps<{ variant?: 'drawer' | 'page' }>(), { variant: 'drawer' })

const ai = useAiStore()

const draft = ref('')
const scroller = ref<HTMLElement | null>(null)

const QUICK = [
  '今天成交额和订单量怎么样？',
  '看一下最近 7 天的下单趋势',
  '订单状态占比是什么情况',
  '上海有哪些房源',
  '查一下最近的订单',
]

async function toBottom(force = false) {
  await nextTick()
  const el = scroller.value
  if (!el) return
  if (force || el.scrollHeight - el.scrollTop - el.clientHeight < 160) {
    el.scrollTop = el.scrollHeight
  }
}

watch(() => ai.items.length, () => void toBottom(true))
watch(
  () => ai.items[ai.items.length - 1]?.content,
  () => void toBottom(),
)

function send(text?: string) {
  const content = (text ?? draft.value).trim()
  if (!content) return
  draft.value = ''
  void ai.send(content).then(() => toBottom(true))
}

function onKeydown(e: KeyboardEvent) {
  if (e.key === 'Enter' && !e.shiftKey && !e.isComposing) {
    e.preventDefault()
    send()
  }
}

defineExpose({ focusInput: () => toBottom(true) })
</script>

<template>
  <div class="flex min-h-0 flex-1 flex-col">
    <!-- 消息区 -->
    <div
      ref="scroller"
      class="min-h-0 flex-1 space-y-4 overflow-y-auto"
      :class="props.variant === 'page' ? 'px-6 py-5' : 'px-4 py-4'"
    >
      <!-- 空态 -->
      <template v-if="!ai.items.length">
        <div class="pt-4">
          <div class="flex items-center gap-2.5">
            <span class="flex h-9 w-9 items-center justify-center rounded-[10px] bg-ink text-[13px] font-semibold text-white">
              AI
            </span>
            <div class="leading-tight">
              <div class="text-[14px] font-semibold text-ink">运营 Copilot</div>
              <div class="text-[11.5px] text-ink-3">
                {{ ai.health?.mock ? '本地规则模式' : ai.health ? `模型 ${ai.health.model || ai.health.provider}` : '加载中…' }}
                · 知识库 {{ ai.health?.knowledgeCount ?? 0 }} 条 · 可调用 {{ ai.toolCount }} 个业务接口
              </div>
            </div>
          </div>

          <p class="mt-4 text-[12.5px] leading-relaxed text-ink-2">
            我会直接读取后台的真实数据来回答，涉及「改数据」的操作（改价、上下架、禁用用户、强制取消订单）会先给你一个确认卡片，确认后才会执行，并写入审计日志。
          </p>

          <div class="mt-4 space-y-2">
            <button
              v-for="q in QUICK"
              :key="q"
              class="flex w-full items-center gap-2 rounded-btn border border-line bg-card px-3 py-2.5 text-left text-[13px] text-ink-2 transition-all duration-200 hover:border-ink-4 hover:bg-fill-2"
              @click="send(q)"
            >
              <svg viewBox="0 0 16 16" class="h-3.5 w-3.5 shrink-0 text-ink-4" fill="none" stroke="currentColor" stroke-width="1.5">
                <path d="M8 2.5l1.1 2.9 2.9 1.1-2.9 1.1L8 10.5 6.9 7.6 4 6.5l2.9-1.1L8 2.5Z" stroke-linejoin="round" />
              </svg>
              <span class="flex-1">{{ q }}</span>
              <svg viewBox="0 0 16 16" class="h-3.5 w-3.5 shrink-0 text-ink-4" fill="none" stroke="currentColor" stroke-width="1.7">
                <path d="M6.5 4l4 4-4 4" stroke-linecap="round" stroke-linejoin="round" />
              </svg>
            </button>
          </div>
        </div>
      </template>

      <template v-else>
        <AiBubble v-for="item in ai.items" :key="item.key" :item="item" />
      </template>
    </div>

    <!-- 输入区 -->
    <div class="shrink-0 border-t border-line bg-card px-4 py-3">
      <div class="flex items-end gap-2">
        <textarea
          v-model="draft"
          rows="1"
          placeholder="问点什么…（Enter 发送，Shift+Enter 换行）"
          class="input max-h-[120px] min-h-[38px] flex-1 resize-none py-2 leading-snug"
          @keydown="onKeydown"
        />
        <button v-if="ai.sending" class="btn btn-line btn-md shrink-0" title="停止" @click="ai.stop()">
          <svg viewBox="0 0 20 20" class="h-3.5 w-3.5" fill="currentColor"><rect x="6" y="6" width="8" height="8" rx="1.5" /></svg>
          停止
        </button>
        <button v-else class="btn btn-primary btn-md shrink-0" :disabled="!draft.trim()" @click="send()">
          <svg viewBox="0 0 20 20" class="h-3.5 w-3.5" fill="none" stroke="currentColor" stroke-width="1.8">
            <path d="M10 16V4M5.5 8.5L10 4l4.5 4.5" stroke-linecap="round" stroke-linejoin="round" />
          </svg>
          发送
        </button>
      </div>
      <p class="mt-2 text-[11px] text-ink-4">
        AI 生成内容仅供参考，涉及数据的判断请以业务页面为准；写操作需要你在卡片上二次确认。
      </p>
    </div>
  </div>
</template>
