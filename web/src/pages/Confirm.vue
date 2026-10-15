<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import NavBar from '@/components/NavBar.vue'
import SmartImage from '@/components/SmartImage.vue'
import Stepper from '@/components/Stepper.vue'
import DateRangeSheet from '@/components/DateRangeSheet.vue'
import { fetchHomestayDetail } from '@/api/travel'
import { createHomestayOrder } from '@/api/order'
import { useSearchStore } from '@/stores/search'
import { useToast } from '@/composables/useToast'
import { date, money } from '@/utils/format'
import type { HomestayDetailResp } from '@/types'

const route = useRoute()
const router = useRouter()
const search = useSearchStore()
const toast = useToast()

const id = Number(route.params.id)
const loading = ref(true)
const submitting = ref(false)
const detail = ref<HomestayDetailResp | null>(null)
const showDate = ref(false)
const isFood = ref(false)
const peopleNum = ref(search.peopleNum)
const remark = ref('')

const homestay = computed(() => detail.value?.homestay)

const nights = computed(() => search.nights)
const roomTotal = computed(() => (homestay.value ? homestay.value.homestayPrice * nights.value : 0))
const foodTotal = computed(() =>
  isFood.value && homestay.value ? homestay.value.foodPrice * peopleNum.value * nights.value : 0,
)
const total = computed(() => roomTotal.value + foodTotal.value)

async function load() {
  loading.value = true
  try {
    detail.value = await fetchHomestayDetail(id)
    peopleNum.value = Math.min(search.peopleNum, detail.value.homestay.peopleNum || search.peopleNum)
    isFood.value = detail.value.homestay.foodPrice > 0
  } catch (e) {
    toast.error((e as Error).message)
  } finally {
    loading.value = false
  }
}

onMounted(load)

function onRange(start: number, end: number) {
  search.setRange(start, end)
}

async function submit() {
  if (!homestay.value || submitting.value) return
  if (search.endTs <= search.startTs) {
    toast.error('请选择至少 1 晚的入住日期')
    return
  }

  submitting.value = true
  try {
    const res = await createHomestayOrder({
      homestayId: id,
      isFood: isFood.value,
      liveStartTime: search.startTs,
      liveEndTime: search.endTs,
      livePeopleNum: peopleNum.value,
      remark: remark.value,
    })
    search.peopleNum = peopleNum.value
    toast.success('下单成功，请尽快支付')
    router.replace(`/pay/${res.orderSn}`)
  } catch (e) {
    toast.error((e as Error).message)
  } finally {
    submitting.value = false
  }
}
</script>

<template>
  <div class="page-scroll pb-[84px]">
    <NavBar title="确认订单" />

    <div v-if="loading" class="space-y-3 px-4 pt-3">
      <div class="skeleton h-24 w-full rounded-card" />
      <div class="skeleton h-32 w-full rounded-card" />
      <div class="skeleton h-28 w-full rounded-card" />
    </div>

    <main v-else-if="homestay" class="space-y-3 px-4 pt-3">
      <!-- 民宿信息 -->
      <section class="card flex gap-3 p-3">
        <div class="h-[76px] w-[76px] shrink-0 overflow-hidden rounded-xl bg-fill">
          <SmartImage :src="homestay.cover" :ratio="1" rounded="12px" />
        </div>
        <div class="min-w-0 flex-1">
          <h2 class="ellipsis text-[14.5px] font-medium text-ink">{{ homestay.title }}</h2>
          <p class="mt-1 ellipsis text-[12px] text-ink-3">{{ homestay.city }} · {{ homestay.address }}</p>
          <p class="mt-1.5 text-[12px] text-ink-2">可住 {{ homestay.peopleNum }} 人 · {{ homestay.foodInfo || '不含餐食' }}</p>
        </div>
      </section>

      <!-- 入住信息 -->
      <section class="card overflow-hidden">
        <button class="flex w-full items-center justify-between px-4 py-3.5 text-left transition active:bg-fill-2" @click="showDate = true">
          <span class="text-[13.5px] text-ink-2">入住 / 离店</span>
          <span class="flex items-center gap-1.5 text-[13.5px] text-ink">
            {{ date(search.startTs) }} → {{ date(search.endTs) }}
            <span class="text-ink-3">共{{ nights }}晚</span>
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="#9AA0A8" stroke-width="2"><path d="M9 5l7 7-7 7" stroke-linecap="round" /></svg>
          </span>
        </button>
        <div class="mx-4 divider" />
        <div class="flex items-center justify-between px-4 py-3.5">
          <span class="text-[13.5px] text-ink-2">入住人数</span>
          <Stepper v-model="peopleNum" :min="1" :max="homestay.peopleNum || 10" suffix=" 人" />
        </div>
        <template v-if="homestay.foodPrice > 0">
          <div class="mx-4 divider" />
          <div class="flex items-center justify-between px-4 py-3.5">
            <div class="min-w-0 pr-3">
              <p class="text-[13.5px] text-ink">需要餐食</p>
              <p class="mt-0.5 text-[11.5px] text-ink-3">{{ homestay.foodInfo }} · ¥{{ money(homestay.foodPrice) }}/人/晚</p>
            </div>
            <button
              class="relative h-7 w-12 shrink-0 rounded-pill transition-colors duration-200"
              :class="isFood ? 'bg-ink' : 'bg-line-2'"
              role="switch"
              :aria-checked="isFood"
              @click="isFood = !isFood"
            >
              <span
                class="absolute top-0.5 h-6 w-6 rounded-full bg-white shadow transition-all duration-200 ease-out"
                :class="isFood ? 'left-[22px]' : 'left-0.5'"
              />
            </button>
          </div>
        </template>
        <div class="mx-4 divider" />
        <div class="flex items-center px-4 py-3.5">
          <span class="shrink-0 text-[13.5px] text-ink-2">备注</span>
          <input
            v-model="remark"
            class="ml-3 min-w-0 flex-1 bg-transparent text-right text-[13.5px] text-ink outline-none placeholder:text-ink-3"
            maxlength="40"
            placeholder="到店时间等，可留空"
          />
        </div>
      </section>

      <!-- 费用明细 -->
      <section class="card px-4 py-4">
        <h3 class="mb-3 text-[14px] font-medium text-ink">费用明细</h3>
        <div class="space-y-2.5 text-[13px]">
          <div class="flex justify-between">
            <span class="text-ink-2">房费 ¥{{ money(homestay.homestayPrice) }} × {{ nights }} 晚</span>
            <span class="tabular-nums text-ink">¥{{ money(roomTotal) }}</span>
          </div>
          <div v-if="isFood" class="flex justify-between">
            <span class="text-ink-2">餐费 ¥{{ money(homestay.foodPrice) }} × {{ peopleNum }} 人 × {{ nights }} 晚</span>
            <span class="tabular-nums text-ink">¥{{ money(foodTotal) }}</span>
          </div>
          <div class="divider my-1" />
          <div class="flex items-baseline justify-between">
            <span class="text-[13px] text-ink-2">合计</span>
            <span class="flex items-baseline gap-0.5">
              <span class="text-[12px] text-ink">¥</span>
              <span class="text-[20px] font-semibold tabular-nums leading-none text-ink">{{ money(total) }}</span>
            </span>
          </div>
        </div>
      </section>

      <p class="px-1 pb-2 text-[11px] leading-relaxed text-ink-3">
        提交订单后请在 30 分钟内完成支付，超时订单将自动关闭。取消政策以房东设定为准。
      </p>
    </main>

    <div
      v-if="homestay"
      class="fixed inset-x-0 bottom-0 z-40 mx-auto w-full max-w-[480px] border-t border-line bg-card/96 backdrop-blur-md"
      style="padding-bottom: var(--safe-bottom)"
    >
      <div class="flex items-center gap-3 px-4 py-2.5">
        <div class="min-w-0 flex-1">
          <div class="flex items-baseline gap-1">
            <span class="text-[11px] text-ink-3">合计</span>
            <span class="text-[11px] text-ink">¥</span>
            <span class="text-[20px] font-semibold leading-none text-ink">{{ money(total) }}</span>
          </div>
          <p class="mt-0.5 text-[11px] text-ink-3">{{ nights }} 晚 · {{ peopleNum }} 人{{ isFood ? ' · 含餐食' : '' }}</p>
        </div>
        <button class="btn-primary h-11 px-8 text-[15px] disabled:opacity-50" :disabled="submitting" @click="submit">
          {{ submitting ? '提交中…' : '提交订单' }}
        </button>
      </div>
    </div>

    <DateRangeSheet v-model="showDate" :start="search.startTs" :end="search.endTs" @change="onRange" />
  </div>
</template>
