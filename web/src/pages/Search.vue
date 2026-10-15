<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import NavBar from '@/components/NavBar.vue'
import HomestayCard from '@/components/HomestayCard.vue'
import SkeletonCard from '@/components/SkeletonCard.vue'
import EmptyState from '@/components/EmptyState.vue'
import DateRangeSheet from '@/components/DateRangeSheet.vue'
import FilterSheet from '@/components/FilterSheet.vue'
import BottomSheet from '@/components/BottomSheet.vue'
import { useSearchStore } from '@/stores/search'
import { searchHomestay } from '@/api/travel'
import { useToast } from '@/composables/useToast'
import { dateShort } from '@/utils/format'
import type { Homestay } from '@/types'

const search = useSearchStore()
const toast = useToast()

const keywordInput = ref(search.keyword)
const list = ref<Homestay[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = 10
const loading = ref(false)
const finished = ref(false)

const showDate = ref(false)
const showFilter = ref(false)
const showSort = ref(false)

const SORTS = [
  { key: 'default', label: '综合推荐' },
  { key: 'priceAsc', label: '价格从低到高' },
  { key: 'priceDesc', label: '价格从高到低' },
  { key: 'score', label: '好评优先' },
]

const sortLabel = computed(() => SORTS.find((s) => s.key === search.sortBy)?.label ?? '综合推荐')
const hasFilter = computed(() => search.minPrice > 0 || search.maxPrice > 0 || search.peopleNum !== 2)

async function fetchPage(reset = false) {
  if (loading.value) return
  if (reset) {
    page.value = 1
    finished.value = false
    list.value = []
  }
  if (finished.value) return

  loading.value = true
  try {
    const res = await searchHomestay({
      keyword: search.keyword,
      city: search.city,
      minPrice: search.minPrice,
      maxPrice: search.maxPrice,
      peopleNum: search.peopleNum,
      sortBy: search.sortBy,
      page: page.value,
      pageSize,
    })
    total.value = res.total
    list.value = reset ? res.list : [...list.value, ...res.list]
    if (list.value.length >= res.total || res.list.length < pageSize) finished.value = true
    else page.value += 1
  } catch (e) {
    toast.error((e as Error).message)
    finished.value = true
  } finally {
    loading.value = false
  }
}

onMounted(() => fetchPage(true))

// 筛选条件变化后重新查询
watch(
  () => [search.city, search.minPrice, search.maxPrice, search.peopleNum, search.sortBy],
  () => fetchPage(true),
)

function doSearch() {
  search.keyword = keywordInput.value.trim()
  fetchPage(true)
}

function onRange(start: number, end: number) {
  search.setRange(start, end)
}

function onFilter(p: { minPrice: number; maxPrice: number; peopleNum: number }) {
  search.minPrice = p.minPrice
  search.maxPrice = p.maxPrice
  search.peopleNum = p.peopleNum
}

function pickSort(key: string) {
  search.sortBy = key
  showSort.value = false
}

function onScroll(e: Event) {
  const el = e.target as HTMLElement
  if (el.scrollHeight - el.scrollTop - el.clientHeight < 260) fetchPage()
}
</script>

<template>
  <div class="page-scroll" @scroll.passive="onScroll">
    <NavBar title="找民宿">
      <template #right>
        <span class="pr-1 text-[12px] text-ink-3">{{ search.city }}</span>
      </template>
    </NavBar>

    <div class="sticky top-12 z-30 bg-bg/95 px-4 pb-2.5 pt-2 backdrop-blur-md">
      <div class="flex items-center gap-2 rounded-card bg-card px-3 py-2 shadow-card">
        <svg class="shrink-0 text-ink-3" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <circle cx="11" cy="11" r="6.5" />
          <path d="M15.8 15.8L21 21" stroke-linecap="round" />
        </svg>
        <input
          v-model="keywordInput"
          class="min-w-0 flex-1 bg-transparent text-[14px] text-ink outline-none placeholder:text-ink-3"
          placeholder="搜民宿名 / 地标 / 标签"
          enterkeyhint="search"
          @keyup.enter="doSearch"
        />
        <button v-if="keywordInput" class="text-[12px] text-ink-3" @click="((keywordInput = ''), doSearch())">清除</button>
        <button class="btn-primary h-7 px-3.5 text-[12.5px]" @click="doSearch">搜索</button>
      </div>

      <div class="mt-2.5 flex items-center gap-2 overflow-x-auto pb-0.5">
        <button class="chip shrink-0 whitespace-nowrap" @click="showDate = true">
          {{ dateShort(search.startTs) }} - {{ dateShort(search.endTs) }} · {{ search.nights }}晚
        </button>
        <button class="chip shrink-0 whitespace-nowrap" :class="{ 'chip-active': hasFilter }" @click="showFilter = true">
          筛选{{ hasFilter ? ' ·' : '' }}
        </button>
        <button class="chip shrink-0 whitespace-nowrap" :class="{ 'chip-active': search.sortBy !== 'default' }" @click="showSort = true">
          {{ sortLabel }}
        </button>
      </div>
    </div>

    <section class="px-4 pb-8">
      <p v-if="!loading || list.length" class="mb-2.5 text-[12px] text-ink-3">
        共找到 <span class="font-medium text-ink-2">{{ total }}</span> 间民宿
      </p>

      <SkeletonCard v-if="loading && !list.length" :count="3" />

      <div v-else-if="list.length" class="stagger flex flex-col gap-3">
        <HomestayCard v-for="h in list" :key="h.id" :item="h" />
      </div>

      <EmptyState
        v-else
        text="没有找到符合条件的民宿"
        tip="试试放宽价格或去掉关键词"
        action-text="重置筛选"
        @action="search.reset(), fetchPage(true)"
      />

      <p v-if="list.length && loading" class="py-4 text-center text-[12px] text-ink-3">正在加载…</p>
      <p v-else-if="list.length && finished" class="py-6 text-center text-[11px] text-ink-4">— 已经到底啦 —</p>
    </section>

    <DateRangeSheet v-model="showDate" :start="search.startTs" :end="search.endTs" @change="onRange" />
    <FilterSheet
      v-model="showFilter"
      :min-price="search.minPrice"
      :max-price="search.maxPrice"
      :people-num="search.peopleNum"
      @change="onFilter"
    />

    <BottomSheet v-model="showSort" title="排序方式">
      <ul class="pb-3">
        <li v-for="s in SORTS" :key="s.key">
          <button
            class="flex w-full items-center justify-between px-5 py-3.5 text-left text-[14px] transition active:bg-fill-2"
            @click="pickSort(s.key)"
          >
            <span :class="search.sortBy === s.key ? 'font-medium text-ink' : 'text-ink-2'">{{ s.label }}</span>
            <svg v-if="search.sortBy === s.key" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="#22262B" stroke-width="2.4">
              <path d="M5 13l4 4L19 7" stroke-linecap="round" stroke-linejoin="round" />
            </svg>
          </button>
        </li>
      </ul>
    </BottomSheet>
  </div>
</template>
