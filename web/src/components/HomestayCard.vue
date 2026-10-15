<script setup lang="ts">
import { useRouter } from 'vue-router'
import SmartImage from './SmartImage.vue'
import { money } from '@/utils/format'
import type { Homestay } from '@/types'

const props = defineProps<{ item: Homestay; layout?: 'row' | 'grid' }>()
const router = useRouter()

function open() {
  router.push(`/detail/${props.item.id}`)
}
</script>

<template>
  <article
    class="card overflow-hidden transition-transform duration-200 ease-out active:scale-[0.985]"
    :class="layout === 'grid' ? '' : ''"
    @click="open"
  >
    <SmartImage :src="item.cover || item.images?.[0]" :ratio="0.62" rounded="0" />

    <div class="p-3">
      <h3 class="ellipsis text-[15px] font-medium leading-snug text-ink">{{ item.title }}</h3>
      <p class="mt-1 ellipsis text-[12.5px] text-ink-2">{{ item.subTitle }}</p>

      <div v-if="item.tags?.length" class="mt-2 flex flex-wrap gap-1.5">
        <span
          v-for="tag in item.tags.slice(0, 3)"
          :key="tag"
          class="rounded-pill bg-fill px-2 py-[3px] text-[11px] leading-none text-ink-2"
        >
          {{ tag }}
        </span>
      </div>

      <div class="mt-2.5 flex items-end justify-between">
        <div class="flex items-center gap-1.5">
          <span class="text-[13px] font-semibold text-ink">{{ (item.score ?? 0).toFixed(1) }}</span>
          <span class="text-[11.5px] text-ink-3">分 · {{ item.commentCount }} 条评价</span>
        </div>
        <div class="flex items-baseline gap-1">
          <span class="text-[11px] text-ink-3">¥</span>
          <span class="text-[17px] font-semibold leading-none text-ink">{{ money(item.homestayPrice) }}</span>
          <span class="text-[11px] text-ink-3">起</span>
        </div>
      </div>
    </div>
  </article>
</template>
