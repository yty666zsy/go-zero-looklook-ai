<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import NavBar from '@/components/NavBar.vue'
import SmartImage from '@/components/SmartImage.vue'
import BottomSheet from '@/components/BottomSheet.vue'
import { cancelOrder, fetchOrderDetail } from '@/api/order'
import { useToast } from '@/composables/useToast'
import { date, dateTime, money, tradeStateText } from '@/utils/format'
import type { OrderDetail } from '@/types'

const route = useRoute()
const router = useRouter()
const toast = useToast()

const sn = String(route.params.sn)
const loading = ref(true)
const order = ref<OrderDetail | null>(null)
const showCancel = ref(false)
const acting = ref(false)

async function load() {
  loading.value = true
  try {
    order.value = await fetchOrderDetail(sn)
    if (!order.value) {
      toast.error('订单不存在')
      router.replace('/orders')
    }
  } catch (e) {
    toast.error((e as Error).message)
  } finally {
    loading.value = false
  }
}

onMounted(load)

async function doCancel() {
  if (acting.value) return
  acting.value = true
  try {
    await cancelOrder(sn)
    toast.success('订单已取消')
    showCancel.value = false
    await load()
  } catch (e) {
    toast.error((e as Error).message)
  } finally {
    acting.value = false
  }
}

const STATE_TIP: Record<number, { title: string; desc: string }> = {
  [-1]: { title: '订单已取消', desc: '如有疑问可联系房东或客服' },
  [0]: { title: '等待支付', desc: '请在 30 分钟内完成支付，超时自动关闭' },
  [1]: { title: '预订成功', desc: '请按约定时间入住，祝您住得愉快' },
  [2]: { title: '已入住', desc: '感谢入住，欢迎留下你的评价' },
  [3]: { title: '已退款', desc: '退款将原路返回' },
  [4]: { title: '订单已过期', desc: '超时未支付，订单自动关闭' },
}
</script>

<template>
  <div class="page-scroll pb-24">
    <NavBar title="订单详情" />

    <div v-if="loading" class="space-y-3 px-4 pt-3">
      <div class="skeleton h-20 w-full rounded-card" />
      <div class="skeleton h-28 w-full rounded-card" />
      <div class="skeleton h-36 w-full rounded-card" />
    </div>

    <main v-else-if="order" class="space-y-3 px-4 pt-3">
      <!-- 状态 -->
      <section class="card px-4 py-4">
        <h2 class="text-[17px] font-semibold text-ink">{{ STATE_TIP[order.tradeState]?.title ?? tradeStateText(order.tradeState) }}</h2>
        <p class="mt-1.5 text-[12.5px] text-ink-3">{{ STATE_TIP[order.tradeState]?.desc }}</p>
      </section>

      <!-- 民宿 -->
      <section class="card p-4">
        <div class="flex gap-3" @click="router.push(`/detail/${order.homestayId}`)">
          <div class="h-[76px] w-[76px] shrink-0 overflow-hidden rounded-xl bg-fill">
            <SmartImage :src="order.cover" :ratio="1" rounded="12px" />
          </div>
          <div class="min-w-0 flex-1">
            <h3 class="ellipsis text-[14.5px] font-medium text-ink">{{ order.title }}</h3>
            <p class="mt-1 ellipsis text-[12px] text-ink-3">{{ order.subTitle }}</p>
            <p class="mt-1.5 text-[12px] text-ink-2">{{ date(order.liveStartDate) }} → {{ date(order.liveEndDate) }}</p>
          </div>
        </div>
      </section>

      <!-- 入住信息 -->
      <section class="card px-4 py-4 text-[13px]">
        <h3 class="mb-3 text-[14px] font-medium text-ink">入住信息</h3>
        <div class="space-y-2.5">
          <div class="flex justify-between gap-4">
            <span class="shrink-0 text-ink-3">入住 / 离店</span>
            <span class="text-right text-ink">{{ date(order.liveStartDate) }} → {{ date(order.liveEndDate) }}</span>
          </div>
          <div class="flex justify-between gap-4">
            <span class="text-ink-3">入住人数</span>
            <span class="text-ink">{{ order.livePeopleNum }} 人</span>
          </div>
          <div class="flex justify-between gap-4">
            <span class="text-ink-3">餐食</span>
            <span class="text-ink">{{ order.foodTotalPrice > 0 ? order.foodInfo || '含餐食' : '不含餐食' }}</span>
          </div>
          <div v-if="order.remark" class="flex justify-between gap-4">
            <span class="shrink-0 text-ink-3">备注</span>
            <span class="text-right text-ink">{{ order.remark }}</span>
          </div>
        </div>
      </section>

      <!-- 费用 -->
      <section class="card px-4 py-4 text-[13px]">
        <h3 class="mb-3 text-[14px] font-medium text-ink">费用明细</h3>
        <div class="space-y-2.5">
          <div class="flex justify-between">
            <span class="text-ink-3">房费</span>
            <span class="tabular-nums text-ink">¥{{ money(order.homestayTotalPrice) }}</span>
          </div>
          <div v-if="order.foodTotalPrice > 0" class="flex justify-between">
            <span class="text-ink-3">餐费</span>
            <span class="tabular-nums text-ink">¥{{ money(order.foodTotalPrice) }}</span>
          </div>
          <div class="divider my-1" />
          <div class="flex items-baseline justify-between">
            <span class="text-ink-2">实付金额</span>
            <span class="flex items-baseline gap-0.5">
              <span class="text-[12px] text-ink">¥</span>
              <span class="text-[19px] font-semibold leading-none tabular-nums text-ink">{{ money(order.orderTotalPrice) }}</span>
            </span>
          </div>
        </div>
      </section>

      <!-- 订单信息 -->
      <section class="card px-4 py-4 text-[12.5px]">
        <h3 class="mb-3 text-[14px] font-medium text-ink">订单信息</h3>
        <div class="space-y-2.5">
          <div class="flex justify-between gap-4">
            <span class="shrink-0 text-ink-3">订单号</span>
            <span class="select-all text-right text-ink-2">{{ order.sn }}</span>
          </div>
          <div class="flex justify-between gap-4">
            <span class="text-ink-3">下单时间</span>
            <span class="text-ink-2">{{ dateTime(order.createTime) }}</span>
          </div>
          <div v-if="order.payTime" class="flex justify-between gap-4">
            <span class="text-ink-3">支付时间</span>
            <span class="text-ink-2">{{ dateTime(order.payTime) }}</span>
          </div>
          <div v-if="order.payType" class="flex justify-between gap-4">
            <span class="text-ink-3">支付方式</span>
            <span class="text-ink-2">{{ order.payType }}</span>
          </div>
          <div v-if="order.tradeState === 0" class="flex justify-between gap-4">
            <span class="text-ink-3">入住凭证</span>
            <span class="text-ink-2">支付后生成</span>
          </div>
          <div v-else-if="order.tradeCode" class="flex justify-between gap-4">
            <span class="text-ink-3">入住凭证</span>
            <span class="select-all font-medium tracking-[2px] text-ink">{{ order.tradeCode }}</span>
          </div>
        </div>
      </section>
    </main>

    <!-- 底部操作 -->
    <div
      v-if="order"
      class="fixed inset-x-0 bottom-0 z-40 mx-auto w-full max-w-[480px] border-t border-line bg-card/96 px-4 py-2.5 backdrop-blur-md"
      style="padding-bottom: calc(10px + var(--safe-bottom))"
    >
      <div class="flex justify-end gap-2.5">
        <template v-if="order.tradeState === 0">
          <button class="btn-line h-10 px-5 text-[13.5px]" @click="showCancel = true">取消订单</button>
          <button class="btn-primary h-10 px-6 text-[13.5px]" @click="router.push(`/pay/${order.sn}`)">去支付</button>
        </template>
        <button v-else class="btn-ghost h-10 px-6 text-[13.5px]" @click="router.push('/orders')">返回订单列表</button>
      </div>
    </div>

    <BottomSheet v-model="showCancel" title="取消订单">
      <div class="px-5 py-4">
        <p class="text-[13.5px] leading-relaxed text-ink-2">
          确定要取消这笔订单吗？取消后不可恢复，需要重新下单。
        </p>
        <div class="mt-5 flex gap-3">
          <button class="btn-line h-11 flex-1 text-[14.5px]" @click="showCancel = false">再想想</button>
          <button class="btn-primary h-11 flex-1 text-[14.5px] disabled:opacity-50" :disabled="acting" @click="doCancel">
            {{ acting ? '处理中…' : '确认取消' }}
          </button>
        </div>
      </div>
    </BottomSheet>
  </div>
</template>
