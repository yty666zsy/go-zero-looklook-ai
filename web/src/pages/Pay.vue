<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import NavBar from '@/components/NavBar.vue'
import SmartImage from '@/components/SmartImage.vue'
import { fetchOrderDetail } from '@/api/order'
import { mockPay } from '@/api/payment'
import { useToast } from '@/composables/useToast'
import { date, money, pad } from '@/utils/format'
import type { OrderDetail } from '@/types'

const route = useRoute()
const router = useRouter()
const toast = useToast()

const sn = String(route.params.sn)
const loading = ref(true)
const paying = ref(false)
const done = ref(false)
const order = ref<OrderDetail | null>(null)
const left = ref(0)

let timer: number | undefined

/** 下单后 30 分钟内需完成支付，和后端 DeferCloseHomestayOrder 保持一致 */
const DEADLINE_MINUTES = 30

async function load() {
  loading.value = true
  try {
    order.value = await fetchOrderDetail(sn)
    if (!order.value) {
      toast.error('订单不存在')
      router.replace('/orders')
      return
    }
    if (order.value.tradeState !== 0) {
      done.value = true
      return
    }
    const deadline = order.value.createTime + DEADLINE_MINUTES * 60
    left.value = Math.max(0, deadline - Math.floor(Date.now() / 1000))
    timer = window.setInterval(() => {
      left.value = Math.max(0, left.value - 1)
    }, 1000)
  } catch (e) {
    toast.error((e as Error).message)
  } finally {
    loading.value = false
  }
}

onMounted(load)
onBeforeUnmount(() => timer && window.clearInterval(timer))

const countdown = computed(() => {
  const m = Math.floor(left.value / 60)
  const s = left.value % 60
  return `${pad(m)}:${pad(s)}`
})

async function pay() {
  if (paying.value) return
  paying.value = true
  try {
    await mockPay(sn)
    toast.success('支付成功')
    done.value = true
    timer && window.clearInterval(timer)
    window.setTimeout(() => router.replace(`/orders/${sn}`), 900)
  } catch (e) {
    toast.error((e as Error).message)
  } finally {
    paying.value = false
  }
}
</script>

<template>
  <div class="page-scroll pb-10">
    <NavBar title="收银台" />

    <div class="px-4 pt-4">
      <div v-if="loading" class="space-y-3">
        <div class="skeleton h-28 w-full rounded-card" />
        <div class="skeleton h-40 w-full rounded-card" />
      </div>

      <!-- 支付成功 -->
      <template v-else-if="done">
        <section class="card flex flex-col items-center px-6 py-10 text-center">
          <div class="mb-4 flex h-14 w-14 items-center justify-center rounded-full bg-fill animate-pop">
            <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="#22262B" stroke-width="2.4">
              <path d="M5 13l4 4L19 7" stroke-linecap="round" stroke-linejoin="round" />
            </svg>
          </div>
          <h2 class="text-[17px] font-semibold text-ink">支付成功</h2>
          <p class="mt-1.5 text-[12.5px] text-ink-3">正在跳转到订单详情…</p>
          <div class="mt-6 flex w-full gap-3">
            <button class="btn-line h-10 flex-1 text-[14px]" @click="router.replace('/orders')">查看订单</button>
            <button class="btn-primary h-10 flex-1 text-[14px]" @click="router.replace('/')">回首页</button>
          </div>
        </section>
      </template>

      <template v-else-if="order">
        <!-- 订单摘要 -->
        <section class="card p-4 text-center">
          <p class="text-[12.5px] text-ink-3">待支付金额</p>
          <p class="mt-2 flex items-baseline justify-center gap-1">
            <span class="text-[15px] text-ink">¥</span>
            <span class="text-[34px] font-semibold leading-none tabular-nums text-ink">{{ money(order.orderTotalPrice) }}</span>
          </p>
          <p v-if="left > 0" class="mt-2.5 text-[12px] text-ink-3">
            支付剩余时间 <span class="tabular-nums font-medium text-ink">{{ countdown }}</span>
          </p>
          <p v-else class="mt-2.5 text-[12px] text-ink-3">订单已超时，请重新下单</p>
        </section>

        <!-- 订单信息 -->
        <section class="card mt-3 flex gap-3 p-3">
          <div class="h-[64px] w-[64px] shrink-0 overflow-hidden rounded-xl bg-fill">
            <SmartImage :src="order.cover" :ratio="1" rounded="12px" />
          </div>
          <div class="min-w-0 flex-1">
            <h3 class="ellipsis text-[14px] font-medium text-ink">{{ order.title }}</h3>
            <p class="mt-1 text-[12px] text-ink-3">
              {{ date(order.liveStartDate) }} → {{ date(order.liveEndDate) }} · {{ order.livePeopleNum }} 人
            </p>
            <p class="mt-1 text-[11.5px] text-ink-4">订单号 {{ order.sn }}</p>
          </div>
        </section>

        <!-- 支付方式 -->
        <section class="card mt-3 overflow-hidden">
          <p class="px-4 pb-1 pt-4 text-[13px] font-medium text-ink">选择支付方式</p>
          <div class="flex items-center justify-between px-4 py-3.5">
            <div class="flex items-center gap-3">
              <span class="flex h-8 w-8 items-center justify-center rounded-lg bg-fill">
                <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="#22262B" stroke-width="1.9">
                  <rect x="3" y="5" width="18" height="14" rx="2.5" />
                  <path d="M3 10h18" />
                </svg>
              </span>
              <div>
                <p class="text-[13.5px] text-ink">模拟支付</p>
                <p class="mt-0.5 text-[11px] text-ink-3">开发环境专用 · 跳过真实微信支付</p>
              </div>
            </div>
            <span class="flex h-5 w-5 items-center justify-center rounded-full bg-ink">
              <svg width="11" height="11" viewBox="0 0 24 24" fill="none" stroke="#fff" stroke-width="3.2">
                <path d="M5 13l4 4L19 7" stroke-linecap="round" stroke-linejoin="round" />
              </svg>
            </span>
          </div>
          <div class="mx-4 divider" />
          <div class="flex items-center justify-between px-4 py-3.5 opacity-45">
            <div class="flex items-center gap-3">
              <span class="flex h-8 w-8 items-center justify-center rounded-lg bg-fill">
                <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="#22262B" stroke-width="1.9">
                  <path d="M9 20l1.2-3.2M15 20l-1.2-3.2" stroke-linecap="round" />
                  <circle cx="12" cy="10" r="7" />
                  <path d="M9.2 10h5.6M12 5.5v9" stroke-linecap="round" />
                </svg>
              </span>
              <div>
                <p class="text-[13.5px] text-ink">微信支付</p>
                <p class="mt-0.5 text-[11px] text-ink-3">需在微信内打开并配置商户号</p>
              </div>
            </div>
          </div>
        </section>

        <button class="btn-primary mt-5 h-12 w-full text-[15.5px] disabled:opacity-50" :disabled="paying || left <= 0" @click="pay">
          {{ paying ? '支付中…' : `确认支付 ¥${money(order.orderTotalPrice)}` }}
        </button>
      </template>
    </div>
  </div>
</template>
