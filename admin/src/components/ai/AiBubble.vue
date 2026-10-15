<script setup lang="ts">
import { computed, ref } from 'vue'
import type { ChatItem } from '@/stores/ai'
import { useAiStore } from '@/stores/ai'
import AiCards from './AiCards.vue'

const props = defineProps<{ item: ChatItem }>()
const ai = useAiStore()
const showSteps = ref(true)

const TOOL_TEXT: Record<string, string> = {
  dashboard_overview: '经营概览',
  dashboard_trend: '下单趋势',
  dashboard_city: '城市分布',
  order_stat: '订单状态统计',
  order_list: '订单列表',
  order_detail: '订单详情',
  user_list: '用户列表',
  homestay_list: '房源列表',
  force_cancel_order: '强制取消订单',
  update_user_status: '修改用户状态',
  update_homestay_state: '上下架房源',
  update_homestay_price: '修改房源价格',
}

function toolName(t?: string) {
  if (!t) return '工具'
  return TOOL_TEXT[t] || t
}

function argsText(args?: Record<string, unknown>) {
  if (!args) return ''
  const parts = Object.entries(args)
    .filter(([, v]) => v !== '' && v !== 0 && v !== null && v !== undefined)
    .map(([k, v]) => `${k}=${String(v)}`)
  return parts.join('  ')
}

const isUser = computed(() => props.item.role === 'user')
const steps = computed(() => props.item.steps || [])
const cards = computed(() => props.item.cards || [])
const running = computed(() => steps.value.some((s) => s.status === 'running'))
</script>

<template>
  <div class="flex" :class="isUser ? 'justify-end' : 'justify-start'">
    <div v-if="isUser" class="max-w-[76%] rounded-card rounded-br-[5px] bg-ink px-3.5 py-2.5 text-[13.5px] leading-relaxed text-white">
      {{ item.content }}
    </div>

    <div v-else class="w-full space-y-2.5">
      <!-- 执行过程时间线 -->
      <div v-if="steps.length" class="overflow-hidden rounded-card border border-line bg-fill-2">
        <button class="flex w-full items-center gap-2 px-3 py-2 text-left" @click="showSteps = !showSteps">
          <span class="flex h-4 w-4 items-center justify-center">
            <span v-if="running" class="h-3 w-3 animate-spin rounded-full border-[1.5px] border-ink-4 border-t-ink" />
            <svg v-else viewBox="0 0 16 16" class="h-3.5 w-3.5 text-st-ok" fill="none" stroke="currentColor" stroke-width="1.7">
              <path d="M3 8.5l3 3 7-7" stroke-linecap="round" stroke-linejoin="round" />
            </svg>
          </span>
          <span class="flex-1 text-[12px] text-ink-2">
            {{ running ? '正在调用工具…' : `已调用 ${steps.length} 个工具` }}
            <span class="text-ink-3">· {{ steps.map((s) => toolName(s.tool)).join('、') }}</span>
          </span>
          <svg
            viewBox="0 0 16 16"
            class="h-3.5 w-3.5 text-ink-3 transition-transform duration-200"
            :class="showSteps ? 'rotate-180' : ''"
            fill="none"
            stroke="currentColor"
            stroke-width="1.6"
          >
            <path d="M4 6.5l4 4 4-4" stroke-linecap="round" stroke-linejoin="round" />
          </svg>
        </button>

        <div v-if="showSteps" class="space-y-0 border-t border-line px-3 py-2.5">
          <div v-for="(s, si) in steps" :key="s.index" class="flex gap-2.5">
            <!-- 竖向连接线 -->
            <div class="flex flex-col items-center">
              <span
                class="mt-[5px] h-2 w-2 shrink-0 rounded-full"
                :class="{
                  'bg-ink-4 animate-pulse': s.status === 'running',
                  'bg-st-ok': s.status === 'done',
                  'bg-ink-3': s.status === 'pending',
                  'bg-st-err': s.status === 'fail',
                }"
              />
              <span v-if="si < steps.length - 1" class="mt-1 w-px flex-1 bg-line" />
            </div>
            <div class="min-w-0 flex-1 pb-2">
              <div class="flex items-center gap-2 text-[12px] text-ink">
                <span>{{ toolName(s.tool) }}</span>
                <span v-if="s.costMs" class="num text-[11px] text-ink-4">{{ s.costMs }}ms</span>
                <span v-if="s.status === 'fail'" class="tag bg-[#FBF6F5] text-st-err">失败</span>
              </div>
              <div v-if="argsText(s.args)" class="num mt-0.5 text-[11px] text-ink-4">{{ argsText(s.args) }}</div>
              <div v-if="s.summary" class="mt-0.5 text-[11.5px] text-ink-3">{{ s.summary }}</div>
              <div v-if="s.error" class="mt-0.5 text-[11.5px] text-st-err">{{ s.error }}</div>
            </div>
          </div>
        </div>
      </div>

      <!-- 等待中 -->
      <div v-if="item.streaming && !item.content && !steps.length" class="flex items-center gap-2 px-0.5 py-1">
        <span class="flex gap-1">
          <i class="h-1.5 w-1.5 animate-bounce rounded-full bg-ink-4" style="animation-delay: 0ms" />
          <i class="h-1.5 w-1.5 animate-bounce rounded-full bg-ink-4" style="animation-delay: 120ms" />
          <i class="h-1.5 w-1.5 animate-bounce rounded-full bg-ink-4" style="animation-delay: 240ms" />
        </span>
        <span class="text-[12px] text-ink-3">正在思考…</span>
      </div>

      <!-- 正文 -->
      <div
        v-if="item.content"
        class="whitespace-pre-wrap rounded-card rounded-bl-[5px] border border-line bg-card px-3.5 py-2.5 text-[13.5px] leading-relaxed text-ink"
      >
        {{ item.content }}
        <span v-if="item.streaming" class="ml-0.5 inline-block h-3.5 w-[2px] translate-y-[2px] animate-pulse bg-ink" />
      </div>

      <AiCards v-if="cards.length" :cards="cards" />

      <!-- 写操作二次确认 -->
      <div v-if="item.action" class="rounded-card border border-line-2 bg-fill-2 p-3">
        <div class="flex items-start gap-2.5">
          <span class="mt-[1px] flex h-4 w-4 shrink-0 items-center justify-center rounded-full bg-[#F7F1E3] text-st-warn">
            <svg viewBox="0 0 16 16" class="h-3 w-3" fill="none" stroke="currentColor" stroke-width="1.8">
              <path d="M8 5v3.4M8 11h.01" stroke-linecap="round" />
            </svg>
          </span>
          <div class="flex-1">
            <div class="text-[12.5px] font-medium text-ink">待确认的写操作</div>
            <p class="mt-0.5 text-[12.5px] leading-relaxed text-ink-2">{{ item.action.confirmText }}</p>
            <p class="mt-1 text-[11px] text-ink-4">
              确认前不会改动任何业务数据；令牌 {{ item.action.expireIn }} 秒内有效、且只能执行一次，操作会写入审计日志。
            </p>
          </div>
        </div>

        <div v-if="item.actionState === 'pending'" class="mt-2.5 flex gap-2 pl-6.5">
          <button class="btn btn-primary btn-sm" :disabled="ai.sending" @click="ai.confirm(item, true)">确认执行</button>
          <button class="btn btn-line btn-sm" @click="ai.confirm(item, false)">取消</button>
        </div>
        <div
          v-else
          class="mt-2 pl-6.5 text-[12px]"
          :class="item.actionState === 'done' ? 'text-st-ok' : 'text-ink-3'"
        >
          {{ item.actionState === 'done' ? '已执行' : '已取消' }}
          <span v-if="item.actionMsg" class="text-ink-3">· {{ item.actionMsg }}</span>
        </div>
      </div>

      <div v-if="item.error" class="rounded-card border border-[#E7D3D1] bg-[#FBF6F5] px-3 py-2 text-[12.5px] text-st-err">
        {{ item.error }}
      </div>
    </div>
  </div>
</template>
