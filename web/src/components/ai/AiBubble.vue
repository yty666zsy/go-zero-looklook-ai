<script setup lang="ts">
import { computed, ref } from 'vue'
import type { ChatItem } from '@/stores/ai'
import { useAiStore } from '@/stores/ai'
import AiCards from './AiCards.vue'

const props = defineProps<{ item: ChatItem }>()
const ai = useAiStore()
const showSteps = ref(false)

const TOOL_TEXT: Record<string, string> = {
  search_homestay: '搜索民宿',
  homestay_detail: '查看房源详情',
  my_orders: '查询我的订单',
  order_detail: '查询订单详情',
  cancel_order: '取消订单',
  dashboard_overview: '查询经营概览',
  dashboard_trend: '查询下单趋势',
  dashboard_city: '查询城市分布',
  order_stat: '查询订单统计',
  order_list: '查询订单列表',
  user_list: '查询用户',
  homestay_list: '查询房源',
  force_cancel_order: '强制取消订单',
  update_user_status: '修改用户状态',
  update_homestay_state: '上架/下架房源',
  update_homestay_price: '修改房源价格',
}

function toolName(t?: string) {
  if (!t) return '工具'
  return TOOL_TEXT[t] || t
}

const steps = computed(() => props.item.steps || [])
const cards = computed(() => props.item.cards || [])
const isUser = computed(() => props.item.role === 'user')
const waiting = computed(() => props.item.streaming && !props.item.content && steps.value.length === 0)
</script>

<template>
  <div class="flex" :class="isUser ? 'justify-end' : 'justify-start'">
    <!-- 用户 -->
    <div
      v-if="isUser"
      class="max-w-[80%] rounded-card rounded-br-[6px] bg-ink px-3.5 py-2.5 text-[14px] leading-relaxed text-white"
    >
      {{ item.content }}
    </div>

    <!-- 助手 -->
    <div v-else class="w-full max-w-[92%] space-y-2">
      <!-- 工具调用时间线 -->
      <div v-if="steps.length" class="overflow-hidden rounded-card border border-line bg-fill-2">
        <button
          class="flex w-full items-center gap-2 px-3 py-2 text-left"
          @click="showSteps = !showSteps"
        >
          <span class="flex h-4 w-4 items-center justify-center">
            <span
              v-if="steps.some((s) => s.status === 'running')"
              class="h-3 w-3 animate-spin rounded-full border-[1.5px] border-ink-4 border-t-ink"
            />
            <svg v-else viewBox="0 0 16 16" class="h-3.5 w-3.5 text-ink-2" fill="none" stroke="currentColor" stroke-width="1.6">
              <path d="M3 8.5l3 3 7-7" stroke-linecap="round" stroke-linejoin="round" />
            </svg>
          </span>
          <span class="flex-1 truncate text-[12px] text-ink-2">
            已调用 {{ steps.length }} 个工具 ·
            {{ steps.map((s) => toolName(s.tool)).join('、') }}
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

        <div v-if="showSteps" class="space-y-1.5 border-t border-line px-3 py-2">
          <div v-for="s in steps" :key="s.index" class="flex items-start gap-2">
            <span
              class="mt-[5px] h-1.5 w-1.5 shrink-0 rounded-full"
              :class="{
                'bg-ink-4': s.status === 'running',
                'bg-ink': s.status === 'done',
                'bg-ink-3': s.status === 'pending',
              }"
              :style="s.status === 'fail' ? 'background:#C0563F' : ''"
            />
            <div class="min-w-0 flex-1">
              <div class="text-[12px] text-ink-2">
                {{ toolName(s.tool) }}
                <span v-if="s.costMs" class="text-ink-4">· {{ s.costMs }}ms</span>
              </div>
              <div v-if="s.summary" class="text-[11.5px] text-ink-3">{{ s.summary }}</div>
              <div v-if="s.error" class="text-[11.5px] text-[#C0563F]">{{ s.error }}</div>
            </div>
          </div>
        </div>
      </div>

      <!-- 思考中 -->
      <div v-if="waiting" class="flex items-center gap-2 px-1 py-1">
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
        class="whitespace-pre-wrap rounded-card rounded-bl-[6px] border border-line bg-card px-3.5 py-2.5 text-[14px] leading-relaxed text-ink"
      >
        {{ item.content }}
        <span v-if="item.streaming" class="ml-0.5 inline-block h-3.5 w-[2px] translate-y-[2px] animate-pulse bg-ink" />
      </div>

      <!-- 结构化卡片 -->
      <AiCards v-if="cards.length" :cards="cards" />

      <!-- 待确认写操作 -->
      <div v-if="item.action" class="rounded-card border border-line-2 bg-fill-2 p-3">
        <div class="flex items-start gap-2">
          <svg viewBox="0 0 16 16" class="mt-[1px] h-3.5 w-3.5 shrink-0 text-ink-2" fill="none" stroke="currentColor" stroke-width="1.5">
            <path d="M8 5.5v3.2M8 11h.01" stroke-linecap="round" />
            <circle cx="8" cy="8" r="6.2" />
          </svg>
          <p class="flex-1 text-[12.5px] leading-relaxed text-ink-2">
            需要你确认后才会执行：{{ item.action.confirmText }}
          </p>
        </div>

        <div v-if="item.actionState === 'pending'" class="mt-2.5 flex gap-2">
          <button
            class="btn-primary h-8 flex-1 text-[13px]"
            :disabled="ai.sending"
            @click="ai.confirm(item, true)"
          >
            确认执行
          </button>
          <button class="btn-ghost h-8 flex-1 text-[13px]" @click="ai.confirm(item, false)">取消</button>
        </div>
        <div v-else class="mt-2 text-[12px]" :class="item.actionState === 'done' ? 'text-ink-2' : 'text-ink-3'">
          {{ item.actionState === 'done' ? '已执行' : '已取消' }}<span v-if="item.actionMsg"> · {{ item.actionMsg }}</span>
        </div>
      </div>

      <!-- 错误 -->
      <div v-if="item.error" class="rounded-card border border-line bg-fill-2 px-3 py-2 text-[12.5px] text-[#C0563F]">
        {{ item.error }}
      </div>
    </div>
  </div>
</template>
