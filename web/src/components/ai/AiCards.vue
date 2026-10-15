<script setup lang="ts">
import { useRouter } from 'vue-router'
import type { AiCard } from '@/api/ai'
import { date, money } from '@/utils/format'

interface HomestayItem {
  id: number
  title: string
  cover: string
  city: string
  address: string
  tags?: string[]
  peopleNum: number
  price: number
  score: number
  commentCount: number
}

interface OrderItem {
  sn: string
  title: string
  cover: string
  tradeState: number
  tradeStateText: string
  total: number
  liveStartDate: number
  liveEndDate: number
}

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

const props = defineProps<{ cards: AiCard[] }>()
const router = useRouter()

function itemsOf<T>(card: AiCard): T[] {
  return Array.isArray(card.items) ? (card.items as T[]) : []
}

function openHomestay(id: number) {
  router.push({ name: 'detail', params: { id } })
}

function openOrder(sn: string) {
  router.push({ name: 'orderDetail', params: { sn } })
}

function stateTone(state: number) {
  if (state === 0) return 'bg-ink text-white'
  if (state < 0 || state === 3 || state === 4) return 'bg-fill text-ink-3'
  return 'bg-fill text-ink'
}
</script>

<template>
  <div class="space-y-2">
    <div v-for="(card, ci) in props.cards" :key="ci">
      <!-- 民宿：横向卡片 -->
      <template v-if="card.kind === 'homestay_list'">
        <div class="-mx-1 flex gap-2 overflow-x-auto px-1 pb-1">
          <button
            v-for="h in itemsOf<HomestayItem>(card)"
            :key="h.id"
            class="w-[208px] shrink-0 overflow-hidden rounded-card border border-line bg-card text-left transition-transform duration-200 active:scale-[0.985]"
            @click="openHomestay(h.id)"
          >
            <img :src="h.cover" class="h-[104px] w-full object-cover" loading="lazy" alt="" />
            <div class="space-y-1 p-2.5">
              <div class="ellipsis text-[13px] font-medium text-ink">{{ h.title }}</div>
              <div class="ellipsis text-[11.5px] text-ink-3">{{ h.city }} · {{ h.address }}</div>
              <div class="flex items-center justify-between pt-0.5">
                <span class="text-[14px] font-semibold text-ink">¥{{ money(h.price) }}</span>
                <span class="text-[11px] text-ink-3">可住{{ h.peopleNum }}人 · {{ h.score.toFixed(1) }}分</span>
              </div>
            </div>
          </button>
        </div>
      </template>

      <!-- 订单 -->
      <template v-else-if="card.kind === 'order_list'">
        <button
          v-for="o in itemsOf<OrderItem>(card)"
          :key="o.sn"
          class="mb-2 flex w-full gap-3 rounded-card border border-line bg-card p-2.5 text-left transition-transform duration-200 active:scale-[0.99]"
          @click="openOrder(o.sn)"
        >
          <img :src="o.cover" class="h-[60px] w-[60px] shrink-0 rounded-lg object-cover" loading="lazy" alt="" />
          <div class="min-w-0 flex-1">
            <div class="flex items-start gap-2">
              <span class="ellipsis-2 flex-1 text-[13px] font-medium text-ink">{{ o.title }}</span>
              <span class="shrink-0 rounded-pill px-2 py-0.5 text-[11px]" :class="stateTone(o.tradeState)">
                {{ o.tradeStateText }}
              </span>
            </div>
            <div class="mt-1 text-[11.5px] text-ink-3">
              {{ date(o.liveStartDate) }} ~ {{ date(o.liveEndDate) }}
            </div>
            <div class="mt-0.5 flex items-center justify-between">
              <span class="ellipsis text-[11px] text-ink-4">单号 {{ o.sn }}</span>
              <span class="text-[13px] font-semibold text-ink">¥{{ money(o.total) }}</span>
            </div>
          </div>
        </button>
      </template>

      <!-- 指标 -->
      <template v-else-if="card.kind === 'metrics'">
        <div class="grid grid-cols-2 gap-2">
          <div
            v-for="(m, mi) in itemsOf<MetricItem>(card)"
            :key="mi"
            class="rounded-card border border-line bg-card px-3 py-2.5"
          >
            <div class="text-[11.5px] text-ink-3">{{ m.label }}</div>
            <div class="mt-0.5 text-[16px] font-semibold tracking-tight text-ink">{{ m.value }}</div>
            <div v-if="m.hint" class="mt-0.5 text-[11px] text-ink-3">{{ m.hint }}</div>
          </div>
        </div>
      </template>

      <!-- 趋势（手绘 SVG，无第三方图表库） -->
      <template v-else-if="card.kind === 'trend'">
        <div class="rounded-card border border-line bg-card p-3">
          <div class="mb-2 text-[12px] text-ink-3">{{ card.title }}</div>
          <div class="flex h-[76px] items-end gap-1.5">
            <div
              v-for="(p, pi) in itemsOf<TrendItem>(card)"
              :key="pi"
              class="group flex flex-1 flex-col items-center gap-1"
            >
              <div
                class="w-full rounded-t-[4px] bg-ink/80 transition-all duration-300"
                :style="{ height: `${Math.max(3, (p.orderCount / Math.max(1, Math.max(...itemsOf<TrendItem>(card).map((x) => x.orderCount)))) * 52)}px` }"
              />
              <span class="text-[9px] text-ink-4">{{ p.date.slice(5) }}</span>
            </div>
          </div>
        </div>
      </template>

      <!-- 兜底：未知卡片只显示标题 -->
      <div v-else class="rounded-card border border-line bg-card px-3 py-2 text-[12px] text-ink-3">
        {{ card.title || card.kind }}
      </div>
    </div>
  </div>
</template>
