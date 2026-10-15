<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import StatCard from '@/components/StatCard.vue'
import LineChart from '@/components/LineChart.vue'
import BarList from '@/components/BarList.vue'
import { fetchCityStats, fetchOrderStates, fetchOrders, fetchOverview, fetchTrend } from '@/api/admin'
import type { CityStat, OrderItem, Overview, StateStat, TrendPoint } from '@/types'
import { amount, dateTime, money, tradeStateText, tradeStateTone } from '@/utils/format'
import { useToast } from '@/composables/useToast'

const toast = useToast()
const router = useRouter()

const loading = ref(true)
const ov = ref<Overview | null>(null)
const trend = ref<TrendPoint[]>([])
const cities = ref<CityStat[]>([])
const states = ref<StateStat[]>([])
const recent = ref<OrderItem[]>([])
const days = ref(7)

async function loadTrend() {
  const res = await fetchTrend(days.value)
  trend.value = res.list || []
}

async function loadAll() {
  loading.value = true
  try {
    const [o, c, s, r] = await Promise.all([
      fetchOverview(),
      fetchCityStats(),
      fetchOrderStates(),
      fetchOrders({ page: 1, pageSize: 6 }),
    ])
    ov.value = o
    cities.value = c.list || []
    states.value = s.list || []
    recent.value = r.list || []
    await loadTrend()
  } catch (e) {
    toast.error((e as Error).message)
  } finally {
    loading.value = false
  }
}

onMounted(loadAll)

async function switchDays(d: number) {
  days.value = d
  try {
    await loadTrend()
  } catch (e) {
    toast.error((e as Error).message)
  }
}

const trendPoints = computed(() =>
  trend.value.map((p) => ({ label: p.date.slice(5), value: p.orderCount })),
)
const trendGmv = computed(() => trend.value.map((p) => ({ label: p.date.slice(5), value: p.gmv })))

const cityItems = computed(() =>
  cities.value.map((c) => ({ label: c.city, value: c.count, sub: `家 · ${money(c.gmv)}` })),
)

/** 订单状态：拼成一条横向占比条 */
const stateSegments = computed(() => {
  const order = [-1, 0, 1, 2, 3, 4]
  const map = new Map(states.value.map((s) => [s.tradeState, s.count]))
  const total = order.reduce((sum, k) => sum + (map.get(k) || 0), 0) || 1
  const colors: Record<number, string> = {
    0: 'bg-[#D9C49A]',
    1: 'bg-[#8FA48F]',
    2: 'bg-[#6E8570]',
    3: 'bg-[#C4A9A5]',
    4: 'bg-[#C7CCD2]',
    '-1': 'bg-[#A5605A]',
  }
  return order
    .map((k) => ({
      key: k,
      label: tradeStateText(k),
      count: map.get(k) || 0,
      pct: ((map.get(k) || 0) / total) * 100,
      color: colors[k],
    }))
    .filter((s) => s.count > 0)
})

const stateTotal = computed(() => states.value.reduce((s, x) => s + x.count, 0))
</script>

<template>
  <div class="space-y-5">
    <!-- 指标卡 -->
    <div v-if="loading" class="grid grid-cols-2 gap-4 lg:grid-cols-4">
      <div v-for="i in 4" :key="i" class="card h-[104px] p-4">
        <div class="skeleton h-3 w-16" />
        <div class="skeleton mt-3 h-7 w-24" />
      </div>
    </div>

    <div v-else class="stagger grid grid-cols-2 gap-4 lg:grid-cols-4">
      <StatCard
        label="今日订单"
        :value="ov?.todayOrderCount || 0"
        :prev="ov?.yesterdayOrderCount || 0"
        suffix="单"
        hint="较昨日"
      />
      <StatCard
        label="今日成交额"
        :value="ov?.todayGmv || 0"
        :prev="ov?.yesterdayGmv || 0"
        prefix="¥"
        :digits="2"
        hint="较昨日"
      />
      <StatCard label="累计订单" :value="ov?.totalOrderCount || 0" suffix="单" :hint="`其中待支付 ${ov?.waitPayCount || 0} 单`" />
      <StatCard label="累计成交额" :value="ov?.totalGmv || 0" prefix="¥" :digits="2" :hint="`已支付 ${ov?.paidCount || 0} 单`" />
    </div>

    <!-- 趋势 -->
    <div class="card p-5">
      <div class="mb-4 flex flex-wrap items-center justify-between gap-3">
        <div>
          <h2 class="text-[15px] font-semibold tracking-tight text-ink">下单趋势</h2>
          <p class="mt-0.5 text-[12.5px] text-ink-3">实线为订单数，虚线为成交额</p>
        </div>
        <div class="flex rounded-btn bg-fill p-0.5">
          <button
            v-for="d in [7, 14, 30]"
            :key="d"
            class="rounded-[8px] px-3 py-1.5 text-[12.5px] transition-all duration-200"
            :class="days === d ? 'bg-card text-ink shadow-sm font-medium' : 'text-ink-3 hover:text-ink-2'"
            @click="switchDays(d)"
          >
            近 {{ d }} 天
          </button>
        </div>
      </div>

      <LineChart v-if="!loading" :points="trendPoints" :series2="trendGmv" :height="230" unit="单" />
      <div v-else class="skeleton h-[230px] w-full" />
    </div>

    <!-- 城市 / 状态 -->
    <div class="grid gap-5 lg:grid-cols-2">
      <div class="card p-5">
        <div class="mb-4 flex items-baseline justify-between">
          <h2 class="text-[15px] font-semibold tracking-tight text-ink">城市分布</h2>
          <span class="text-[12.5px] text-ink-3">民宿数量 / 成交额</span>
        </div>
        <BarList v-if="!loading" :items="cityItems" />
        <div v-else class="space-y-3">
          <div v-for="i in 5" :key="i" class="skeleton h-6 w-full" />
        </div>
      </div>

      <div class="card flex flex-col p-5">
        <div class="mb-4 flex items-baseline justify-between">
          <h2 class="text-[15px] font-semibold tracking-tight text-ink">订单状态</h2>
          <span class="text-[12.5px] text-ink-3">共 {{ stateTotal }} 单</span>
        </div>

        <div v-if="!loading" class="flex h-2.5 w-full overflow-hidden rounded-pill bg-fill">
          <div
            v-for="s in stateSegments"
            :key="s.key"
            class="h-full transition-[width] duration-500 ease-out"
            :class="s.color"
            :style="{ width: `${s.pct}%` }"
            :title="`${s.label} ${s.count}`"
          />
        </div>

        <ul class="mt-4 grid flex-1 grid-cols-2 gap-x-4 gap-y-3 self-start">
          <li v-for="s in stateSegments" :key="s.key" class="flex items-center gap-2">
            <span class="h-2 w-2 shrink-0 rounded-full" :class="s.color" />
            <span class="text-[12.5px] text-ink-2">{{ s.label }}</span>
            <span class="num ml-auto text-[13px] font-medium text-ink">{{ s.count }}</span>
          </li>
        </ul>

        <div class="mt-4 flex gap-2 border-t border-line pt-4">
          <button class="btn btn-ghost btn-sm flex-1" @click="router.push({ name: 'orders' })">查看全部订单</button>
          <button class="btn btn-line btn-sm flex-1" @click="router.push({ name: 'homestays' })">管理民宿</button>
        </div>
      </div>
    </div>

    <!-- 最新订单 -->
    <div class="card overflow-hidden">
      <div class="flex items-center justify-between border-b border-line px-5 py-4">
        <h2 class="text-[15px] font-semibold tracking-tight text-ink">最新订单</h2>
        <button class="text-[12.5px] text-ink-3 transition-colors duration-200 hover:text-ink" @click="router.push({ name: 'orders' })">
          全部 →
        </button>
      </div>

      <div v-if="!loading && recent.length === 0">
        <p class="px-5 py-10 text-center text-[12.5px] text-ink-3">还没有订单</p>
      </div>

      <ul v-else class="divide-y divide-line">
        <li
          v-for="o in recent"
          :key="o.id"
          class="flex cursor-pointer items-center gap-4 px-5 py-3.5 transition-colors duration-150 hover:bg-fill-2"
          @click="router.push({ name: 'orders', query: { sn: o.sn } })"
        >
          <img :src="o.cover" class="h-10 w-10 shrink-0 rounded-[10px] bg-fill object-cover" loading="lazy" alt="" />
          <div class="min-w-0 flex-1">
            <div class="ellipsis text-[13.5px] font-medium text-ink">{{ o.title }}</div>
            <div class="num mt-0.5 text-[12px] text-ink-3">{{ o.sn }}</div>
          </div>
          <div class="hidden text-right md:block">
            <div class="text-[12.5px] text-ink-2">{{ o.userMobile || '—' }}</div>
            <div class="text-[12px] text-ink-3">{{ dateTime(o.createTime) }}</div>
          </div>
          <div class="num w-20 text-right text-[13.5px] font-medium text-ink">{{ money(o.orderTotalPrice) }}</div>
          <span class="w-16 text-right">
            <span
              class="tag"
              :class="
                tradeStateTone(o.tradeState) === 'ok'
                  ? 'bg-[#EDF2ED] text-st-ok'
                  : tradeStateTone(o.tradeState) === 'warn'
                    ? 'bg-[#F7F1E3] text-st-warn'
                    : tradeStateTone(o.tradeState) === 'err'
                      ? 'bg-[#F7EDEC] text-st-err'
                      : 'bg-fill text-st-off'
              "
            >
              {{ tradeStateText(o.tradeState) }}
            </span>
          </span>
        </li>
      </ul>
    </div>

    <p v-if="!loading" class="pb-2 text-center text-[12px] text-ink-4">
      累计成交额 {{ amount(ov?.totalGmv || 0) }} 元 · 注册用户 {{ ov?.userCount || 0 }} 人 · 民宿
      {{ ov?.homestayCount || 0 }} 家 · 评价 {{ ov?.commentCount || 0 }} 条
    </p>
  </div>
</template>
