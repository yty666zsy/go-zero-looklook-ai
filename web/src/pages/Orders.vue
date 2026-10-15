<script setup lang="ts">
import { onMounted, ref, watch } from 'vue'
import { useRouter } from 'vue-router'
import NavBar from '@/components/NavBar.vue'
import TabBar from '@/components/TabBar.vue'
import SmartImage from '@/components/SmartImage.vue'
import SkeletonCard from '@/components/SkeletonCard.vue'
import EmptyState from '@/components/EmptyState.vue'
import BottomSheet from '@/components/BottomSheet.vue'
import { cancelOrder, fetchOrderList } from '@/api/order'
import { useToast } from '@/composables/useToast'
import { date, dateTime, money, tradeStateText } from '@/utils/format'
import type { OrderListItem, TradeState } from '@/types'

const router = useRouter()
const toast = useToast()

const TABS = [
  { label: '全部', state: -99 },
  { label: '待付款', state: 0 },
  { label: '待入住', state: 1 },
  { label: '已入住', state: 2 },
  { label: '已取消', state: -1 },
]

const tabIdx = ref(0)
const loading = ref(true)
const list = ref<OrderListItem[]>([])
const confirming = ref<OrderListItem | null>(null)
const acting = ref(false)

async function load() {
  loading.value = true
  list.value = []
  try {
    const res = await fetchOrderList(0, 50, TABS[tabIdx.value].state)
    list.value = res.list ?? []
  } catch (e) {
    toast.error((e as Error).message)
  } finally {
    loading.value = false
  }
}

onMounted(load)
watch(tabIdx, load)

function stateClass(state: TradeState | number) {
  if (state === 0) return 'text-ink'
  if (state === 1) return 'text-ink'
  return 'text-ink-3'
}

async function doCancel() {
  if (!confirming.value || acting.value) return
  acting.value = true
  try {
    await cancelOrder(confirming.value.sn)
    toast.success('订单已取消')
    confirming.value = null
    await load()
  } catch (e) {
    toast.error((e as Error).message)
  } finally {
    acting.value = false
  }
}
</script>

<template>
  <div class="page-scroll pb-[68px]">
    <NavBar title="我的订单" :back="false" />

    <div class="sticky top-12 z-30 flex gap-1 overflow-x-auto border-b border-line bg-bg/95 px-4 py-2 backdrop-blur-md">
      <button
        v-for="(t, i) in TABS"
        :key="t.label"
        class="relative shrink-0 whitespace-nowrap px-3 py-1.5 text-[13.5px] transition-colors duration-200"
        :class="tabIdx === i ? 'font-medium text-ink' : 'text-ink-3'"
        @click="tabIdx = i"
      >
        {{ t.label }}
        <span
          class="absolute inset-x-2.5 -bottom-2 h-[2px] rounded-pill bg-ink transition-all duration-250 ease-out"
          :class="tabIdx === i ? 'opacity-100' : 'opacity-0'"
        />
      </button>
    </div>

    <main class="px-4 pt-3">
      <SkeletonCard v-if="loading" :count="3" />

      <div v-else-if="list.length" class="stagger flex flex-col gap-3">
        <article v-for="o in list" :key="o.sn" class="card overflow-hidden" @click="router.push(`/orders/${o.sn}`)">
          <div class="flex items-center justify-between px-4 py-2.5">
            <span class="text-[11.5px] text-ink-3">{{ dateTime(o.createTime) }}</span>
            <span class="text-[12.5px] font-medium" :class="stateClass(o.tradeState)">
              {{ tradeStateText(o.tradeState) }}
            </span>
          </div>
          <div class="mx-4 divider" />

          <div class="flex gap-3 p-4">
            <div class="h-[76px] w-[76px] shrink-0 overflow-hidden rounded-xl bg-fill">
              <SmartImage :src="o.cover" :ratio="1" rounded="12px" />
            </div>
            <div class="min-w-0 flex-1">
              <h3 class="ellipsis text-[14.5px] font-medium text-ink">{{ o.title }}</h3>
              <p class="mt-1 ellipsis text-[12px] text-ink-3">{{ o.subTitle }}</p>
              <p class="mt-1.5 text-[12px] text-ink-2">{{ date(o.liveStartDate) }} → {{ date(o.liveEndDate) }}</p>
            </div>
          </div>

          <div class="mx-4 divider" />
          <div class="flex items-center justify-between px-4 py-3">
            <div class="flex items-baseline gap-1">
              <span class="text-[11.5px] text-ink-3">合计</span>
              <span class="text-[12px] text-ink">¥</span>
              <span class="text-[16px] font-semibold leading-none tabular-nums text-ink">{{ money(o.orderTotalPrice) }}</span>
            </div>
            <div class="flex gap-2">
              <button
                v-if="o.tradeState === 0"
                class="btn-line h-8 px-4 text-[12.5px]"
                @click.stop="confirming = o"
              >
                取消订单
              </button>
              <button
                v-if="o.tradeState === 0"
                class="btn-primary h-8 px-4 text-[12.5px]"
                @click.stop="router.push(`/pay/${o.sn}`)"
              >
                去支付
              </button>
              <button v-else class="btn-ghost h-8 px-4 text-[12.5px]" @click.stop="router.push(`/orders/${o.sn}`)">
                查看详情
              </button>
            </div>
          </div>
        </article>
      </div>

      <EmptyState
        v-else
        text="还没有订单"
        tip="挑一间喜欢的民宿，住进风景里"
        action-text="去找民宿"
        @action="router.push('/')"
      />
    </main>

    <TabBar />

    <BottomSheet :model-value="!!confirming" title="取消订单" @update:model-value="confirming = null">
      <div class="px-5 py-4">
        <p class="text-[13.5px] leading-relaxed text-ink-2">
          确定要取消「{{ confirming?.title }}」这笔订单吗？取消后不可恢复，需要重新下单。
        </p>
        <div class="mt-5 flex gap-3">
          <button class="btn-line h-11 flex-1 text-[14.5px]" @click="confirming = null">再想想</button>
          <button class="btn-primary h-11 flex-1 text-[14.5px] disabled:opacity-50" :disabled="acting" @click="doCancel">
            {{ acting ? '处理中…' : '确认取消' }}
          </button>
        </div>
      </div>
    </BottomSheet>
  </div>
</template>
