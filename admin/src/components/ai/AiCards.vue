<script setup lang="ts">
import { useRouter } from 'vue-router'
import type { AiCard } from '@/api/ai'
import { date } from '@/utils/format'

interface MetricItem {
  label: string
  value: string
  hint?: string
}

interface TrendItem {
  date: string
  orderCount: number
  gmv: number
}

interface AdminOrderItem {
  id: number
  sn: string
  userId: number
  userMobile: string
  homestayId: number
  title: string
  liveStartDate: number
  liveEndDate: number
  livePeopleNum: number
  orderTotalPrice: number
  tradeState: number
  remark: string
  createTime: number
}

interface AdminUserItem {
  id: number
  mobile: string
  nickname: string
  status: number
  orderCount: number
  createTime: number
}

interface AdminHomestayItem {
  id: number
  title: string
  city: string
  address: string
  tags: string
  peopleNum: number
  homestayPrice: number
  marketHomestayPrice: number
  rowState: number
  score: number
  commentCount: number
  createTime: number
}

const props = defineProps<{ cards: AiCard[] }>()
const router = useRouter()

function itemsOf<T>(card: AiCard): T[] {
  return Array.isArray(card.items) ? (card.items as T[]) : []
}

const ORDER_STATE: Record<number, { text: string; cls: string }> = {
  [-1]: { text: '已取消', cls: 'bg-fill text-st-off' },
  0: { text: '待支付', cls: 'bg-[#F7F1E3] text-st-warn' },
  1: { text: '未使用', cls: 'bg-[#EDF2ED] text-st-ok' },
  2: { text: '已使用', cls: 'bg-fill text-ink-2' },
  3: { text: '已退款', cls: 'bg-fill text-st-off' },
  4: { text: '已过期', cls: 'bg-fill text-st-off' },
}

function stateOf(s: number) {
  return ORDER_STATE[s] || { text: `状态 ${s}`, cls: 'bg-fill text-ink-2' }
}

function yuan(v: number) {
  return `¥${Number(v || 0).toLocaleString('zh-CN', { maximumFractionDigits: 2 })}`
}

function openOrder(sn: string) {
  router.push({ name: 'orders', query: { sn } })
}

function openHomestay() {
  router.push({ name: 'homestays' })
}

function openUser() {
  router.push({ name: 'users' })
}

function bars(card: AiCard): TrendItem[] {
  return itemsOf<TrendItem>(card)
}

function barMax(card: AiCard) {
  const list = bars(card)
  return Math.max(1, ...list.map((x) => x.orderCount))
}
</script>

<template>
  <div class="space-y-2.5">
    <!-- 指标组 -->
    <template v-for="(card, ci) in props.cards" :key="ci">
      <div v-if="card.kind === 'metrics'" class="grid grid-cols-2 gap-2.5 lg:grid-cols-4">
        <div
          v-for="(m, mi) in itemsOf<MetricItem>(card)"
          :key="mi"
          class="rounded-card border border-line bg-card px-3 py-2.5"
        >
          <div class="text-[12px] text-ink-3">{{ m.label }}</div>
          <div class="num mt-1 text-[17px] font-semibold leading-none tracking-tight text-ink">{{ m.value }}</div>
          <div v-if="m.hint" class="mt-1 text-[11.5px] text-ink-3">{{ m.hint }}</div>
        </div>
      </div>

      <!-- 趋势 -->
      <div v-else-if="card.kind === 'trend'" class="rounded-card border border-line bg-card p-3.5">
        <div class="mb-3 flex items-center justify-between">
          <span class="text-[13px] font-medium text-ink">{{ card.title }}</span>
          <div class="flex items-center gap-3 text-[11.5px] text-ink-3">
            <span class="flex items-center gap-1.5">
              <i class="inline-block h-2 w-2 rounded-[2px] bg-ink" />下单量
            </span>
            <span>合计 {{ bars(card).reduce((a, b) => a + b.orderCount, 0) }} 单 ·
              {{ yuan(bars(card).reduce((a, b) => a + b.gmv, 0)) }}</span>
          </div>
        </div>
        <div class="flex h-[92px] items-end gap-2">
          <div v-for="(p, pi) in bars(card)" :key="pi" class="group flex flex-1 flex-col items-center gap-1.5">
            <span class="text-[10.5px] text-ink-4 opacity-0 transition-opacity duration-200 group-hover:opacity-100">
              {{ p.orderCount }}
            </span>
            <div
              class="w-full rounded-t-[4px] bg-ink/75 transition-all duration-300 ease-out group-hover:bg-ink"
              :style="{ height: `${Math.max(3, (p.orderCount / barMax(card)) * 60)}px` }"
              :title="`${p.date} · ${p.orderCount} 单 · ${yuan(p.gmv)}`"
            />
            <span class="text-[10px] text-ink-4">{{ p.date.slice(5) }}</span>
          </div>
        </div>
      </div>

      <!-- 订单 -->
      <div v-else-if="card.kind === 'admin_orders'" class="overflow-hidden rounded-card border border-line bg-card">
        <div class="border-b border-line px-3.5 py-2.5 text-[13px] font-medium text-ink">{{ card.title }}</div>
        <button
          v-for="o in itemsOf<AdminOrderItem>(card)"
          :key="o.id"
          class="flex w-full items-center gap-3 border-b border-line px-3.5 py-2.5 text-left transition-colors duration-150 last:border-0 hover:bg-fill-2"
          @click="openOrder(o.sn)"
        >
          <div class="min-w-0 flex-1">
            <div class="flex items-center gap-2">
              <span class="ellipsis text-[13px] font-medium text-ink">{{ o.title }}</span>
              <span class="tag shrink-0" :class="stateOf(o.tradeState).cls">{{ stateOf(o.tradeState).text }}</span>
            </div>
            <div class="mt-1 flex items-center gap-3 text-[11.5px] text-ink-3">
              <span class="num">{{ o.sn }}</span>
              <span>{{ o.userMobile }}</span>
              <span>{{ date(o.liveStartDate) }} ~ {{ date(o.liveEndDate) }}</span>
            </div>
          </div>
          <span class="num shrink-0 text-[13.5px] font-semibold text-ink">{{ yuan(o.orderTotalPrice) }}</span>
        </button>
      </div>

      <!-- 用户 -->
      <div v-else-if="card.kind === 'admin_users'" class="overflow-hidden rounded-card border border-line bg-card">
        <div class="border-b border-line px-3.5 py-2.5 text-[13px] font-medium text-ink">{{ card.title }}</div>
        <button
          v-for="u in itemsOf<AdminUserItem>(card)"
          :key="u.id"
          class="flex w-full items-center gap-3 border-b border-line px-3.5 py-2.5 text-left transition-colors duration-150 last:border-0 hover:bg-fill-2"
          @click="openUser()"
        >
          <span class="flex h-7 w-7 shrink-0 items-center justify-center rounded-full bg-fill text-[12px] font-medium text-ink-2">
            {{ (u.nickname || 'U').slice(0, 1) }}
          </span>
          <div class="min-w-0 flex-1">
            <div class="ellipsis text-[13px] font-medium text-ink">
              {{ u.nickname }} <span class="num text-[11.5px] font-normal text-ink-3">#{{ u.id }}</span>
            </div>
            <div class="num mt-0.5 text-[11.5px] text-ink-3">{{ u.mobile }} · {{ u.orderCount }} 单 · 注册 {{ date(u.createTime) }}</div>
          </div>
          <span class="tag shrink-0" :class="u.status === 1 ? 'bg-[#EDF2ED] text-st-ok' : 'bg-fill text-st-off'">
            {{ u.status === 1 ? '正常' : '已禁用' }}
          </span>
        </button>
      </div>

      <!-- 房源 -->
      <div v-else-if="card.kind === 'admin_homestays'" class="overflow-hidden rounded-card border border-line bg-card">
        <div class="border-b border-line px-3.5 py-2.5 text-[13px] font-medium text-ink">{{ card.title }}</div>
        <button
          v-for="h in itemsOf<AdminHomestayItem>(card)"
          :key="h.id"
          class="flex w-full items-center gap-3 border-b border-line px-3.5 py-2.5 text-left transition-colors duration-150 last:border-0 hover:bg-fill-2"
          @click="openHomestay()"
        >
          <div class="min-w-0 flex-1">
            <div class="flex items-center gap-2">
              <span class="ellipsis text-[13px] font-medium text-ink">{{ h.title }}</span>
              <span class="tag shrink-0" :class="h.rowState === 1 ? 'bg-[#EDF2ED] text-st-ok' : 'bg-fill text-st-off'">
                {{ h.rowState === 1 ? '在售' : '已下架' }}
              </span>
            </div>
            <div class="mt-1 flex items-center gap-3 text-[11.5px] text-ink-3">
              <span>#{{ h.id }} · {{ h.city }} {{ h.address }}</span>
              <span>可住 {{ h.peopleNum }} 人</span>
              <span>评分 {{ h.score.toFixed(1) }}</span>
            </div>
          </div>
          <span class="num shrink-0 text-[13.5px] font-semibold text-ink">{{ yuan(h.homestayPrice) }}</span>
        </button>
      </div>

      <!-- 兜底 -->
      <div v-else class="rounded-card border border-line bg-card px-3.5 py-2.5 text-[12.5px] text-ink-3">
        {{ card.title || card.kind }}
      </div>
    </template>
  </div>
</template>
