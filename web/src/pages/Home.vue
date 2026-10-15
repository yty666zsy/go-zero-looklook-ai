<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import TabBar from '@/components/TabBar.vue'
import SmartImage from '@/components/SmartImage.vue'
import HomestayCard from '@/components/HomestayCard.vue'
import SkeletonCard from '@/components/SkeletonCard.vue'
import EmptyState from '@/components/EmptyState.vue'
import DateRangeSheet from '@/components/DateRangeSheet.vue'
import { useSearchStore } from '@/stores/search'
import { fetchHome } from '@/api/travel'
import { useToast } from '@/composables/useToast'
import { dateShort } from '@/utils/format'
import type { HomeResp } from '@/types'

const router = useRouter()
const search = useSearchStore()
const toast = useToast()

const loading = ref(true)
const data = ref<HomeResp | null>(null)
const showDate = ref(false)
const bannerIdx = ref(0)

const cityOptions = ['上海', '杭州', '成都', '厦门', '丽江', '三亚', '大理', '北京']

async function load() {
  loading.value = true
  try {
    data.value = await fetchHome(search.city)
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

function goSearch() {
  router.push('/search')
}

function pickCity(city: string) {
  search.city = city
  router.push('/search')
}

function pickTag(tag: string) {
  search.keyword = tag
  router.push('/search')
}

const banners = computed(() => data.value?.banners ?? [])
const cities = computed(() => (data.value?.cities ?? []).filter((c) => cityOptions.includes(c.name)))
</script>

<template>
  <div class="page-scroll pb-[68px]">
    <!-- 顶部搜索区 -->
    <header class="sticky top-0 z-30 bg-bg/95 px-4 pb-3 pt-4 backdrop-blur-md">
      <div class="mb-3 flex items-center justify-between">
        <div class="flex items-baseline gap-2">
          <h1 class="text-[19px] font-semibold tracking-tight text-ink">栖居</h1>
          <span class="text-[11.5px] text-ink-3">住进风景里</span>
        </div>
        <button
          class="flex items-center gap-1 rounded-pill bg-card px-3 py-1.5 text-[12.5px] text-ink shadow-card transition active:scale-95"
          @click="goSearch"
        >
          <svg width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.2">
            <path d="M12 21s7-5.5 7-11a7 7 0 10-14 0c0 5.5 7 11 7 11z" stroke-linejoin="round" />
            <circle cx="12" cy="10" r="2.6" />
          </svg>
          {{ search.city }}
        </button>
      </div>

      <button
        class="flex w-full items-center gap-3 rounded-card bg-card p-3 text-left shadow-card transition active:scale-[0.99]"
        @click="goSearch"
      >
        <svg class="shrink-0 text-ink-3" width="17" height="17" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <circle cx="11" cy="11" r="6.5" />
          <path d="M15.8 15.8L21 21" stroke-linecap="round" />
        </svg>
        <div class="flex min-w-0 flex-1 flex-col gap-1">
          <span class="text-[14px] text-ink-3">想住哪儿？搜民宿名 / 地标 / 标签</span>
          <div class="flex items-center gap-2 text-[12px] text-ink-2">
            <span class="rounded-pill bg-fill px-2 py-0.5">{{ dateShort(search.startTs) }} 入住</span>
            <span class="rounded-pill bg-fill px-2 py-0.5">{{ search.nights }} 晚</span>
            <span class="rounded-pill bg-fill px-2 py-0.5">{{ search.peopleNum }} 人</span>
          </div>
        </div>
        <span class="btn-primary h-8 shrink-0 px-4 text-[13px]">搜索</span>
      </button>
    </header>

    <main class="px-4">
      <!-- 轮播 -->
      <section v-if="banners.length" class="mt-1">
        <div class="relative overflow-hidden rounded-card shadow-card">
          <div
            class="flex transition-transform duration-500 ease-out"
            :style="{ transform: `translateX(-${bannerIdx * 100}%)` }"
          >
            <div v-for="b in banners" :key="b.id" class="w-full shrink-0" @click="router.push(`/detail/${b.id}`)">
              <SmartImage :src="b.cover" :ratio="0.42" rounded="0" />
            </div>
          </div>
          <div class="absolute bottom-2.5 left-1/2 flex -translate-x-1/2 gap-1.5">
            <button
              v-for="(b, i) in banners"
              :key="b.id"
              class="h-1.5 rounded-pill transition-all duration-300"
              :class="i === bannerIdx ? 'w-4 bg-white' : 'w-1.5 bg-white/55'"
              :aria-label="`第 ${i + 1} 张`"
              @click="bannerIdx = i"
            />
          </div>
          <div v-if="banners.length > 1" class="absolute inset-y-0 left-0 flex w-1/2" @click="bannerIdx = Math.max(0, bannerIdx - 1)" />
          <div
            v-if="banners.length > 1"
            class="absolute inset-y-0 right-0 flex w-1/2 justify-end"
            @click="bannerIdx = Math.min(banners.length - 1, bannerIdx + 1)"
          />
        </div>
      </section>

      <!-- 日期快捷入口 -->
      <section class="mt-4">
        <div class="card flex items-center divide-x divide-line px-1 py-3">
          <button class="flex flex-1 flex-col items-center gap-1 transition active:scale-95" @click="showDate = true">
            <span class="text-[11px] text-ink-3">入住</span>
            <span class="text-[14px] font-medium text-ink">{{ dateShort(search.startTs) }}</span>
          </button>
          <button class="flex flex-1 flex-col items-center gap-1 transition active:scale-95" @click="showDate = true">
            <span class="text-[11px] text-ink-3">离店</span>
            <span class="text-[14px] font-medium text-ink">{{ dateShort(search.endTs) }}</span>
          </button>
          <button class="flex flex-1 flex-col items-center gap-1 transition active:scale-95" @click="goSearch">
            <span class="text-[11px] text-ink-3">共</span>
            <span class="text-[14px] font-medium text-ink">{{ search.nights }} 晚</span>
          </button>
        </div>
      </section>

      <!-- 热门城市 -->
      <section class="mt-6">
        <div class="mb-3 flex items-baseline justify-between">
          <h2 class="text-[16px] font-semibold text-ink">热门城市</h2>
          <span class="text-[12px] text-ink-3">按当地房源数排序</span>
        </div>
        <div class="grid grid-cols-4 gap-2.5">
          <button
            v-for="c in cities.slice(0, 8)"
            :key="c.code"
            class="group flex flex-col items-center gap-2 transition active:scale-95"
            @click="pickCity(c.name)"
          >
            <div class="h-14 w-14 overflow-hidden rounded-full bg-fill ring-1 ring-line transition group-active:ring-line-2">
              <SmartImage :src="c.cover" :ratio="1" rounded="999px" />
            </div>
            <span class="text-[12.5px] text-ink">{{ c.name }}</span>
          </button>
        </div>
      </section>

      <!-- 热门标签 -->
      <section v-if="data?.hotTags?.length" class="mt-6">
        <h2 class="mb-3 text-[16px] font-semibold text-ink">大家都在找</h2>
        <div class="flex flex-wrap gap-2">
          <button v-for="t in data.hotTags" :key="t" class="chip" @click="pickTag(t)">{{ t }}</button>
        </div>
      </section>

      <!-- 优选民宿 -->
      <section class="mt-6">
        <div class="mb-3 flex items-baseline justify-between">
          <h2 class="text-[16px] font-semibold text-ink">优选民宿</h2>
          <button class="text-[12px] text-ink-3 transition active:opacity-60" @click="goSearch">全部 ›</button>
        </div>
        <SkeletonCard v-if="loading" :count="2" />
        <div v-else-if="data?.recommend?.length" class="stagger flex flex-col gap-3">
          <HomestayCard v-for="h in data.recommend" :key="h.id" :item="h" />
        </div>
        <EmptyState v-else text="暂无优选房源" />
      </section>

      <!-- 猜你喜欢 -->
      <section class="mt-6">
        <div class="mb-3 flex items-baseline justify-between">
          <h2 class="text-[16px] font-semibold text-ink">猜你喜欢</h2>
          <span class="text-[12px] text-ink-3">根据你的浏览推荐</span>
        </div>
        <SkeletonCard v-if="loading" :count="2" />
        <div v-else-if="data?.guess?.length" class="stagger grid grid-cols-2 gap-3">
          <HomestayCard v-for="h in data.guess" :key="h.id" :item="h" layout="grid" />
        </div>
        <EmptyState v-else text="暂无推荐" />
      </section>

      <p class="py-8 text-center text-[11px] text-ink-4">— 已经到底啦 —</p>
    </main>

    <TabBar />
    <DateRangeSheet v-model="showDate" :start="search.startTs" :end="search.endTs" @change="onRange" />
  </div>
</template>
