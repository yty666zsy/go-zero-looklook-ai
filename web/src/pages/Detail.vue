<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import NavBar from '@/components/NavBar.vue'
import SmartImage from '@/components/SmartImage.vue'
import StarRating from '@/components/StarRating.vue'
import BottomSheet from '@/components/BottomSheet.vue'
import EmptyState from '@/components/EmptyState.vue'
import { fetchComments, fetchHomestayDetail } from '@/api/travel'
import { useSearchStore } from '@/stores/search'
import { useToast } from '@/composables/useToast'
import { fromNow, money } from '@/utils/format'
import type { HomestayComment, HomestayDetailResp } from '@/types'

const route = useRoute()
const router = useRouter()
const search = useSearchStore()
const toast = useToast()

const id = Number(route.params.id)
const loading = ref(true)
const detail = ref<HomestayDetailResp | null>(null)
const comments = ref<HomestayComment[]>([])
const commentTotal = ref(0)
const imageIdx = ref(0)
const scrolled = ref(false)
const showComments = ref(false)
const showIntro = ref(false)

const homestay = computed(() => detail.value?.homestay)
const images = computed(() => detail.value?.images ?? [])
const business = computed(() => detail.value?.business)
const stat = computed(() => detail.value?.scoreStat)

async function load() {
  loading.value = true
  try {
    const [d, c] = await Promise.all([fetchHomestayDetail(id), fetchComments(id, 1, 3)])
    detail.value = d
    comments.value = c.list ?? []
    commentTotal.value = c.total ?? 0
    search.keyword = ''
  } catch (e) {
    toast.error((e as Error).message)
  } finally {
    loading.value = false
  }
}

async function loadAllComments() {
  try {
    const c = await fetchComments(id, 1, 50)
    comments.value = c.list ?? []
    commentTotal.value = c.total ?? 0
    showComments.value = true
  } catch (e) {
    toast.error((e as Error).message)
  }
}

onMounted(load)

function onScroll(e: Event) {
  scrolled.value = (e.target as HTMLElement).scrollTop > 120
}

function book() {
  router.push(`/confirm/${id}`)
}
</script>

<template>
  <div class="page-scroll pb-[76px]" @scroll.passive="onScroll">
    <NavBar title="民宿详情" :floating="true" :solid="scrolled" />

    <div v-if="loading" class="px-4 pt-3">
      <div class="skeleton h-52 w-full rounded-card" />
      <div class="mt-4 space-y-2.5">
        <div class="skeleton h-5 w-2/3" />
        <div class="skeleton h-3.5 w-1/2" />
        <div class="skeleton h-20 w-full" />
      </div>
    </div>

    <EmptyState v-else-if="!homestay" text="民宿不存在或已下架" action-text="返回首页" @action="router.replace('/')" />

    <template v-else>
      <!-- 图片轮播 -->
      <section class="relative">
        <div
          class="flex transition-transform duration-500 ease-out"
          :style="{ transform: `translateX(-${imageIdx * 100}%)` }"
        >
          <div v-for="(img, i) in images" :key="i" class="w-full shrink-0">
            <SmartImage :src="img" :ratio="0.72" rounded="0" />
          </div>
        </div>
        <div class="absolute bottom-3 left-1/2 flex -translate-x-1/2 gap-1.5">
          <button
            v-for="(_, i) in images"
            :key="i"
            class="h-1.5 rounded-pill transition-all duration-300"
            :class="i === imageIdx ? 'w-4 bg-white' : 'w-1.5 bg-white/55'"
            :aria-label="`第 ${i + 1} 张`"
            @click="imageIdx = i"
          />
        </div>
        <span class="absolute right-3 top-3 rounded-pill bg-black/35 px-2.5 py-1 text-[11px] text-white backdrop-blur">
          {{ imageIdx + 1 }} / {{ images.length || 1 }}
        </span>
      </section>

      <!-- 标题 + 评分 -->
      <section class="bg-card px-4 pb-4 pt-4">
        <h1 class="text-[19px] font-semibold leading-snug text-ink">{{ homestay.title }}</h1>
        <p class="mt-1.5 text-[13px] leading-relaxed text-ink-2">{{ homestay.subTitle }}</p>

        <div class="mt-3 flex items-center gap-2.5">
          <div class="flex items-center gap-1.5 rounded-pill bg-fill px-2.5 py-1">
            <span class="text-[13px] font-semibold text-ink">{{ (stat?.score ?? homestay.score).toFixed(1) }}</span>
            <StarRating :score="stat?.score ?? homestay.score" :size="11" />
          </div>
          <button class="text-[12.5px] text-ink-2 underline-offset-2 active:underline" @click="loadAllComments">
            {{ commentTotal || homestay.commentCount }} 条真实评价
          </button>
        </div>

        <div v-if="homestay.tags?.length" class="mt-3 flex flex-wrap gap-1.5">
          <span v-for="t in homestay.tags" :key="t" class="rounded-pill bg-fill px-2.5 py-1 text-[11.5px] leading-none text-ink-2">
            {{ t }}
          </span>
        </div>
      </section>

      <div class="mx-4 my-2.5 divider" />

      <!-- 位置 -->
      <section class="bg-card px-4 py-4">
        <div class="flex items-start gap-3">
          <svg class="mt-0.5 shrink-0 text-ink-3" width="17" height="17" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8">
            <path d="M12 21s7-5.5 7-11a7 7 0 10-14 0c0 5.5 7 11 7 11z" stroke-linejoin="round" />
            <circle cx="12" cy="10" r="2.6" />
          </svg>
          <div class="min-w-0 flex-1">
            <p class="text-[14px] text-ink">{{ homestay.city }} · {{ homestay.address }}</p>
            <p class="mt-1 text-[12px] text-ink-3">
              可住 {{ homestay.peopleNum }} 人 · {{ homestay.foodInfo || '不含餐食' }}
            </p>
          </div>
        </div>
      </section>

      <div class="mx-4 my-2.5 divider" />

      <!-- 评分维度 -->
      <section v-if="stat && stat.count > 0" class="bg-card px-4 py-4">
        <h2 class="mb-3 text-[15px] font-semibold text-ink">住客评分</h2>
        <div class="flex items-center gap-2.5">
          <div class="text-center">
            <p class="text-[26px] font-semibold leading-none text-ink">{{ stat.score.toFixed(1) }}</p>
            <StarRating class="mt-1.5" :score="stat.score" :size="11" />
          </div>
          <div class="ml-3 flex-1 space-y-1.5">
            <div v-for="d in [
              { label: '卫生', value: stat.clean },
              { label: '位置', value: stat.location },
              { label: '服务', value: stat.service },
            ]" :key="d.label" class="flex items-center gap-2">
              <span class="w-7 shrink-0 text-[11.5px] text-ink-3">{{ d.label }}</span>
              <span class="h-1 flex-1 overflow-hidden rounded-pill bg-fill">
                <span class="block h-full rounded-pill bg-ink-2 transition-all duration-700 ease-out" :style="{ width: `${(d.value / 5) * 100}%` }" />
              </span>
              <span class="w-6 text-right text-[11.5px] tabular-nums text-ink-2">{{ d.value.toFixed(1) }}</span>
            </div>
          </div>
        </div>
      </section>

      <div class="mx-4 my-2.5 divider" />

      <!-- 简介 -->
      <section class="bg-card px-4 py-4">
        <h2 class="mb-2.5 text-[15px] font-semibold text-ink">房源介绍</h2>
        <p class="ellipsis-2 text-[13px] leading-relaxed text-ink-2" :class="{ '!line-clamp-none': showIntro }" :style="showIntro ? 'display:block' : ''">
          {{ homestay.info }}
        </p>
        <button v-if="!showIntro" class="mt-2 text-[12.5px] text-ink-2" @click="showIntro = true">展开全部 ⌄</button>
      </section>

      <div class="mx-4 my-2.5 divider" />

      <!-- 评价 -->
      <section class="bg-card px-4 py-4">
        <div class="mb-3 flex items-baseline justify-between">
          <h2 class="text-[15px] font-semibold text-ink">住客评价</h2>
          <button class="text-[12px] text-ink-3" @click="loadAllComments">全部 {{ commentTotal }} 条 ›</button>
        </div>

        <div v-if="comments.length" class="space-y-4">
          <article v-for="c in comments.slice(0, 3)" :key="c.id" class="flex gap-3">
            <div class="h-8 w-8 shrink-0 overflow-hidden rounded-full bg-fill">
              <SmartImage :src="c.avatar" :ratio="1" rounded="999px" />
            </div>
            <div class="min-w-0 flex-1">
              <div class="flex items-center gap-2">
                <span class="text-[13px] text-ink">{{ c.nickname || '匿名住客' }}</span>
                <span class="text-[11px] text-ink-4">{{ fromNow(c.createTime) }}</span>
              </div>
              <p class="mt-1.5 text-[13px] leading-relaxed text-ink-2">{{ c.content }}</p>
            </div>
          </article>
        </div>
        <EmptyState v-else text="还没有评价" tip="成为第一个入住的客人吧" />
      </section>

      <div class="mx-4 my-2.5 divider" />

      <!-- 房东 / 店铺 -->
      <section v-if="business?.id" class="bg-card px-4 py-4">
        <h2 class="mb-3 text-[15px] font-semibold text-ink">房源由房东提供</h2>
        <div class="flex items-center gap-3">
          <div class="h-11 w-11 shrink-0 overflow-hidden rounded-full bg-fill">
            <SmartImage :src="business.cover" :ratio="1" rounded="999px" />
          </div>
          <div class="min-w-0 flex-1">
            <p class="ellipsis text-[14px] font-medium text-ink">{{ business.title }}</p>
            <p class="mt-0.5 ellipsis text-[12px] text-ink-3">{{ business.info }}</p>
          </div>
          <div class="shrink-0 text-right">
            <p class="text-[15px] font-semibold text-ink">{{ (business.star ?? 0).toFixed(1) }}</p>
            <p class="text-[11px] text-ink-3">房东评分</p>
          </div>
        </div>
      </section>

      <p class="py-8 text-center text-[11px] text-ink-4">— 栖居 · 住进风景里 —</p>
    </template>

    <!-- 底部预订栏 -->
    <div
      v-if="homestay"
      class="fixed inset-x-0 bottom-0 z-40 mx-auto w-full max-w-[480px] border-t border-line bg-card/96 backdrop-blur-md"
      style="padding-bottom: var(--safe-bottom)"
    >
      <div class="flex items-center gap-3 px-4 py-2.5">
        <div class="min-w-0 flex-1">
          <div class="flex items-baseline gap-1">
            <span class="text-[11px] text-ink-3">¥</span>
            <span class="text-[20px] font-semibold leading-none text-ink">{{ money(homestay.homestayPrice) }}</span>
            <span class="text-[11.5px] text-ink-3">/ 晚起</span>
          </div>
          <p v-if="homestay.marketHomestayPrice > homestay.homestayPrice" class="mt-0.5 text-[11px] text-ink-3 line-through">
            门市价 ¥{{ money(homestay.marketHomestayPrice) }}
          </p>
        </div>
        <button class="btn-primary h-11 px-8 text-[15px]" @click="book">立即预订</button>
      </div>
    </div>

    <!-- 全部评价 -->
    <BottomSheet v-model="showComments" :title="`全部评价（${commentTotal}）`">
      <div class="space-y-4 px-5 py-4">
        <article v-for="c in comments" :key="c.id" class="flex gap-3">
          <div class="h-8 w-8 shrink-0 overflow-hidden rounded-full bg-fill">
            <SmartImage :src="c.avatar" :ratio="1" rounded="999px" />
          </div>
          <div class="min-w-0 flex-1 border-b border-line pb-4 last:border-0">
            <div class="flex items-center justify-between">
              <span class="text-[13px] text-ink">{{ c.nickname || '匿名住客' }}</span>
              <StarRating :score="c.star" :size="11" />
            </div>
            <p class="mt-1.5 text-[13px] leading-relaxed text-ink-2">{{ c.content }}</p>
            <p class="mt-1 text-[11px] text-ink-4">{{ fromNow(c.createTime) }}</p>
          </div>
        </article>
      </div>
    </BottomSheet>
  </div>
</template>
