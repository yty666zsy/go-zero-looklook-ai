<script setup lang="ts">
import { computed, onMounted, reactive, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import Modal from '@/components/Modal.vue'
import Pagination from '@/components/Pagination.vue'
import EmptyState from '@/components/EmptyState.vue'
import StateTag from '@/components/StateTag.vue'
import { fetchOrderDetail, fetchOrderStat, fetchOrders, forceCancelOrder } from '@/api/admin'
import type { OrderItem, OrderStat } from '@/types'
import { TRADE_STATES, date, dateTime, money, nights, tradeStateText, tradeStateTone } from '@/utils/format'
import { useToast } from '@/composables/useToast'

const toast = useToast()
const route = useRoute()
const router = useRouter()

const loading = ref(true)
const list = ref<OrderItem[]>([])
const total = ref(0)
const sumPrice = ref(0)
const stat = ref<OrderStat | null>(null)

const query = reactive({
  page: 1,
  pageSize: 10,
  tradeState: -99,
  keyword: '',
})

let timer: number | undefined
watch(
  () => query.keyword,
  () => {
    window.clearTimeout(timer)
    timer = window.setTimeout(() => {
      query.page = 1
      load()
    }, 320)
  },
)

watch(
  () => [query.tradeState, query.page, query.pageSize],
  () => load(),
)

async function load() {
  loading.value = true
  try {
    const [res, s] = await Promise.all([
      fetchOrders({
        page: query.page,
        pageSize: query.pageSize,
        tradeState: query.tradeState,
        keyword: query.keyword || undefined,
      }),
      fetchOrderStat(),
    ])
    list.value = res.list || []
    total.value = res.total
    sumPrice.value = res.sumPrice
    stat.value = s
  } catch (e) {
    toast.error((e as Error).message)
  } finally {
    loading.value = false
  }
}

onMounted(async () => {
  await load()
  const sn = route.query.sn as string | undefined
  if (sn) openDetail({ sn } as OrderItem)
})

function reset() {
  query.keyword = ''
  query.tradeState = -99
  query.page = 1
  load()
}

const hasFilter = computed(() => !!query.keyword || query.tradeState !== -99)

const statChips = computed(() => {
  const s = stat.value
  if (!s) return []
  return [
    { key: -99, label: '全部', value: s.all },
    { key: 0, label: '待支付', value: s.waitPay },
    { key: 1, label: '未使用', value: s.waitUse },
    { key: 2, label: '已使用', value: s.used },
    { key: 3, label: '已退款', value: s.refund },
    { key: 4, label: '已过期', value: s.expired },
    { key: -1, label: '已取消', value: s.cancel },
  ]
})

function pickState(k: number) {
  query.tradeState = k
  query.page = 1
}

// ---------------- 详情 ----------------
const detailOpen = ref(false)
const detailLoading = ref(false)
const detail = ref<OrderItem | null>(null)

async function openDetail(o: OrderItem) {
  detailOpen.value = true
  detailLoading.value = true
  detail.value = null
  try {
    const res = await fetchOrderDetail(o.sn)
    detail.value = res.order
  } catch (e) {
    toast.error((e as Error).message)
    detailOpen.value = false
  } finally {
    detailLoading.value = false
  }
}

function closeDetail() {
  detailOpen.value = false
  if (route.query.sn) router.replace({ name: 'orders' })
}

// ---------------- 强制取消 ----------------
const cancelOpen = ref(false)
const cancelTarget = ref<OrderItem | null>(null)
const cancelReason = ref('')
const cancelSaving = ref(false)
const cancelError = ref('')

function openCancel(o: OrderItem) {
  cancelTarget.value = o
  cancelReason.value = '后台强制取消'
  cancelError.value = ''
  cancelOpen.value = true
}

const cancelable = (s: number) => s === 0 || s === 1

async function doCancel() {
  if (!cancelTarget.value) return
  cancelSaving.value = true
  cancelError.value = ''
  try {
    await forceCancelOrder(cancelTarget.value.sn, cancelReason.value || '后台强制取消')
    toast.success('订单已取消' + (cancelTarget.value.tradeState === 1 ? '，支付流水已标记为退款' : ''))
    cancelOpen.value = false
    detailOpen.value = false
    cancelTarget.value.tradeState = -1
    await load()
  } catch (e) {
    cancelError.value = (e as Error).message
  } finally {
    cancelSaving.value = false
  }
}
</script>

<template>
  <div class="space-y-4">
    <!-- 状态筛选 -->
    <div class="card flex flex-wrap items-center gap-1.5 p-3">
      <button
        v-for="c in statChips"
        :key="c.key"
        class="btn h-8 gap-1.5 rounded-pill px-3.5 text-[12.5px]"
        :class="query.tradeState === c.key ? 'bg-ink text-white' : 'bg-fill text-ink-2 hover:bg-[#e6e9ec]'"
        @click="pickState(c.key)"
      >
        {{ c.label }}
        <span class="num" :class="query.tradeState === c.key ? 'text-white/70' : 'text-ink-4'">{{ c.value }}</span>
      </button>
    </div>

    <!-- 搜索栏 -->
    <div class="card flex flex-wrap items-center gap-3 p-4">
      <div class="relative min-w-[240px] flex-1">
        <svg
          viewBox="0 0 24 24"
          class="pointer-events-none absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-ink-4"
          fill="none"
          stroke="currentColor"
          stroke-width="1.8"
          stroke-linecap="round"
        >
          <circle cx="11" cy="11" r="6.5" />
          <path d="m16 16 4 4" />
        </svg>
        <input v-model="query.keyword" class="input pl-9" placeholder="搜索订单号 / 民宿名称 / 交易码" />
      </div>

      <select v-model.number="query.tradeState" class="select w-[130px]">
        <option v-for="s in TRADE_STATES" :key="s.value" :value="s.value">{{ s.label }}</option>
      </select>

      <button v-if="hasFilter" class="btn btn-ghost btn-md" @click="reset">重置</button>

      <div class="ml-auto flex items-center gap-4 text-[12.5px] text-ink-3">
        <span>筛选结果 <span class="num text-ink-2">{{ total }}</span> 单</span>
        <span>合计 <span class="num font-medium text-ink">{{ money(sumPrice) }}</span></span>
      </div>
    </div>

    <!-- 表格 -->
    <div class="card overflow-hidden">
      <div class="table-wrap">
        <table class="tbl min-w-[1000px]">
          <thead>
            <tr>
              <th class="w-[300px]">订单 / 民宿</th>
              <th class="w-[130px]">用户</th>
              <th class="w-[170px]">入住</th>
              <th class="w-[110px]">金额</th>
              <th class="w-[100px]">状态</th>
              <th class="w-[150px]">下单时间</th>
              <th class="w-[120px] text-right">操作</th>
            </tr>
          </thead>

          <tbody v-if="loading">
            <tr v-for="i in 6" :key="i">
              <td colspan="7"><div class="skeleton h-6 w-full" /></td>
            </tr>
          </tbody>

          <tbody v-else-if="list.length">
            <tr v-for="o in list" :key="o.id">
              <td>
                <div class="flex items-center gap-3">
                  <img :src="o.cover" class="h-11 w-11 shrink-0 rounded-[10px] bg-fill object-cover" loading="lazy" alt="" />
                  <div class="min-w-0">
                    <button
                      class="ellipsis block max-w-[212px] text-left text-[13.5px] font-medium text-ink transition-colors duration-200 hover:text-ink-2"
                      @click="openDetail(o)"
                    >
                      {{ o.title }}
                    </button>
                    <div class="num mt-0.5 text-[11.5px] text-ink-3">{{ o.sn }}</div>
                  </div>
                </div>
              </td>

              <td>
                <div class="num text-[13px] text-ink-2">{{ o.userMobile || '—' }}</div>
                <div class="num text-[11.5px] text-ink-4">uid {{ o.userId }}</div>
              </td>

              <td>
                <div class="num text-[12.5px] text-ink-2">{{ date(o.liveStartDate) }} → {{ date(o.liveEndDate) }}</div>
                <div class="text-[11.5px] text-ink-3">
                  {{ nights(o.liveStartDate, o.liveEndDate) }} 晚 · {{ o.livePeopleNum }} 人
                </div>
              </td>

              <td>
                <div class="num text-[13.5px] font-medium text-ink">{{ money(o.orderTotalPrice) }}</div>
                <div v-if="o.foodPrice > 0" class="num text-[11.5px] text-ink-3">含餐 {{ money(o.foodPrice) }}</div>
              </td>

              <td>
                <StateTag :tone="tradeStateTone(o.tradeState)" :text="tradeStateText(o.tradeState)" />
              </td>

              <td>
                <div class="num text-[12.5px] text-ink-2">{{ dateTime(o.createTime) }}</div>
              </td>

              <td>
                <div class="flex items-center justify-end gap-1.5">
                  <button class="btn btn-line btn-sm" @click="openDetail(o)">详情</button>
                  <button v-if="cancelable(o.tradeState)" class="btn btn-danger btn-sm" @click="openCancel(o)">取消</button>
                </div>
              </td>
            </tr>
          </tbody>
        </table>
      </div>

      <EmptyState v-if="!loading && !list.length" title="没有匹配的订单" desc="换个订单号或切换状态标签看看">
        <button v-if="hasFilter" class="btn btn-ghost btn-sm mt-1" @click="reset">清空筛选</button>
      </EmptyState>

      <Pagination
        v-if="!loading && list.length"
        v-model:page="query.page"
        v-model:page-size="query.pageSize"
        :total="total"
      />
    </div>

    <!-- 订单详情 -->
    <Modal :model-value="detailOpen" title="订单详情" width="580px" @update:model-value="closeDetail">
      <div v-if="detailLoading" class="space-y-3">
        <div class="skeleton h-28 w-full" />
        <div class="skeleton h-6 w-2/3" />
        <div class="skeleton h-6 w-1/2" />
      </div>

      <div v-else-if="detail" class="space-y-4">
        <div class="flex items-center gap-3.5">
          <img :src="detail.cover" class="h-16 w-16 shrink-0 rounded-card bg-fill object-cover" alt="" />
          <div class="min-w-0 flex-1">
            <div class="text-[15px] font-medium text-ink">{{ detail.title }}</div>
            <div class="ellipsis mt-0.5 text-[12.5px] text-ink-3">{{ detail.subTitle }}</div>
          </div>
          <StateTag :tone="tradeStateTone(detail.tradeState)" :text="tradeStateText(detail.tradeState)" />
        </div>

        <div class="grid grid-cols-2 gap-x-4 gap-y-3 rounded-card bg-fill-2 p-4 text-[13px]">
          <div>
            <div class="text-[12px] text-ink-3">订单号</div>
            <div class="num mt-0.5 text-ink">{{ detail.sn }}</div>
          </div>
          <div>
            <div class="text-[12px] text-ink-3">交易码</div>
            <div class="num mt-0.5 text-ink">{{ detail.tradeCode || '—' }}</div>
          </div>
          <div>
            <div class="text-[12px] text-ink-3">下单用户</div>
            <div class="num mt-0.5 text-ink">{{ detail.userMobile || '—' }}</div>
          </div>
          <div>
            <div class="text-[12px] text-ink-3">下单时间</div>
            <div class="num mt-0.5 text-ink">{{ dateTime(detail.createTime) }}</div>
          </div>
          <div>
            <div class="text-[12px] text-ink-3">入住日期</div>
            <div class="num mt-0.5 text-ink">{{ date(detail.liveStartDate) }} → {{ date(detail.liveEndDate) }}</div>
          </div>
          <div>
            <div class="text-[12px] text-ink-3">入住人数</div>
            <div class="mt-0.5 text-ink">
              {{ detail.livePeopleNum }} 人 · {{ nights(detail.liveStartDate, detail.liveEndDate) }} 晚
            </div>
          </div>
        </div>

        <div class="rounded-card border border-line p-4">
          <div class="mb-3 text-[12.5px] font-medium text-ink-2">费用明细</div>
          <div class="space-y-2 text-[13px]">
            <div class="flex justify-between">
              <span class="text-ink-3">房费</span>
              <span class="num text-ink">{{ money(detail.homestayPrice) }}</span>
            </div>
            <div v-if="detail.foodPrice > 0" class="flex justify-between">
              <span class="text-ink-3">餐食</span>
              <span class="num text-ink">{{ money(detail.foodPrice) }}</span>
            </div>
            <div class="divider my-2" />
            <div class="flex items-baseline justify-between">
              <span class="text-ink-2">合计</span>
              <span class="num text-[17px] font-semibold text-ink">{{ money(detail.orderTotalPrice) }}</span>
            </div>
          </div>
        </div>

        <div v-if="detail.remark">
          <div class="mb-1 text-[12px] text-ink-3">备注</div>
          <p class="rounded-btn bg-fill-2 px-3 py-2.5 text-[12.5px] leading-relaxed text-ink-2">{{ detail.remark }}</p>
        </div>
      </div>

      <template #footer>
        <button class="btn btn-ghost btn-md" @click="closeDetail">关闭</button>
        <button
          v-if="detail && cancelable(detail.tradeState)"
          class="btn btn-danger btn-md"
          @click="openCancel(detail)"
        >
          强制取消
        </button>
      </template>
    </Modal>

    <!-- 强制取消确认 -->
    <Modal v-model="cancelOpen" title="强制取消订单" width="430px">
      <div class="space-y-3.5">
        <p class="text-[13px] leading-relaxed text-ink-2">
          即将取消订单 <span class="num font-medium text-ink">{{ cancelTarget?.sn }}</span>
          <template v-if="cancelTarget?.tradeState === 1">
            ，该订单已支付，将同时把支付流水标记为「已退款」（演示环境不会真实退款）。
          </template>
          <template v-else>，该订单尚未支付。</template>
        </p>

        <label class="block">
          <span class="mb-1.5 block text-[12.5px] text-ink-2">取消原因</span>
          <input v-model="cancelReason" class="input" maxlength="40" placeholder="会写入订单备注" />
        </label>

        <p v-if="cancelError" class="rounded-btn bg-[#FBF6F5] px-3 py-2 text-[12.5px] text-st-err">{{ cancelError }}</p>
      </div>

      <template #footer>
        <button class="btn btn-ghost btn-md" @click="cancelOpen = false">再想想</button>
        <button class="btn btn-primary btn-md" :disabled="cancelSaving" @click="doCancel">
          {{ cancelSaving ? '处理中…' : '确认取消' }}
        </button>
      </template>
    </Modal>
  </div>
</template>
