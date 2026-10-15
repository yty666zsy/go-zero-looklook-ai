<script setup lang="ts">
import { computed, onMounted, reactive, ref, watch } from 'vue'
import Modal from '@/components/Modal.vue'
import Pagination from '@/components/Pagination.vue'
import EmptyState from '@/components/EmptyState.vue'
import { deleteComment, fetchComments } from '@/api/admin'
import type { CommentItem } from '@/types'
import { dateTime, fromNow } from '@/utils/format'
import { useToast } from '@/composables/useToast'

const toast = useToast()

const loading = ref(true)
const list = ref<CommentItem[]>([])
const total = ref(0)

const query = reactive({ page: 1, pageSize: 10, keyword: '' })

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
  () => [query.page, query.pageSize],
  () => load(),
)

async function load() {
  loading.value = true
  try {
    const res = await fetchComments(query.page, query.pageSize, undefined, query.keyword || undefined)
    list.value = res.list || []
    total.value = res.total
  } catch (e) {
    toast.error((e as Error).message)
  } finally {
    loading.value = false
  }
}

onMounted(load)

const avgScore = computed(() => {
  if (!list.value.length) return 0
  return list.value.reduce((s, c) => s + c.score, 0) / list.value.length
})

const STAR_LABELS = ['卫生', '位置', '服务']

// ---------------- 删除 ----------------
const confirmOpen = ref(false)
const target = ref<CommentItem | null>(null)
const saving = ref(false)
const errMsg = ref('')

function openConfirm(c: CommentItem) {
  target.value = c
  errMsg.value = ''
  confirmOpen.value = true
}

async function doDelete() {
  if (!target.value) return
  saving.value = true
  errMsg.value = ''
  try {
    await deleteComment(target.value.id)
    toast.success('评价已删除，民宿评价数已同步')
    confirmOpen.value = false
    await load()
  } catch (e) {
    errMsg.value = (e as Error).message
  } finally {
    saving.value = false
  }
}
</script>

<template>
  <div class="space-y-4">
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
        <input v-model="query.keyword" class="input pl-9" placeholder="搜索评价内容 / 用户昵称" />
      </div>

      <div class="ml-auto flex items-center gap-4 text-[12.5px] text-ink-3">
        <span>共 <span class="num text-ink-2">{{ total }}</span> 条评价</span>
        <span v-if="list.length">
          本页均分
          <span class="num font-medium text-ink">{{ avgScore.toFixed(1) }}</span>
        </span>
      </div>
    </div>

    <!-- 列表 -->
    <div class="card overflow-hidden">
      <div v-if="loading" class="divide-y divide-line">
        <div v-for="i in 5" :key="i" class="p-4">
          <div class="skeleton h-4 w-24" />
          <div class="skeleton mt-3 h-4 w-full" />
          <div class="skeleton mt-2 h-4 w-2/3" />
        </div>
      </div>

      <ul v-else-if="list.length" class="divide-y divide-line">
        <li v-for="c in list" :key="c.id" class="group flex gap-4 p-4 transition-colors duration-150 hover:bg-fill-2">
          <img :src="c.avatar" class="h-9 w-9 shrink-0 rounded-full bg-fill object-cover" loading="lazy" alt="" />

          <div class="min-w-0 flex-1">
            <div class="flex flex-wrap items-center gap-x-3 gap-y-1">
              <span class="text-[13.5px] font-medium text-ink">{{ c.nickname || '匿名用户' }}</span>
              <span class="num text-[11.5px] text-ink-4">uid {{ c.userId }}</span>
              <span class="text-[12.5px] text-ink-3">
                · <button class="transition-colors duration-200 hover:text-ink">{{ c.homestayTitle || `民宿 #${c.homestayId}` }}</button>
              </span>
              <span class="ml-auto text-[11.5px] text-ink-4" :title="dateTime(c.createTime)">
                {{ fromNow(c.createTime) }}
              </span>
            </div>

            <p class="mt-1.5 text-[13px] leading-relaxed text-ink-2">{{ c.content }}</p>

            <div class="mt-2.5 flex flex-wrap items-center gap-3">
              <div class="flex items-center gap-1">
                <svg viewBox="0 0 24 24" class="h-3.5 w-3.5 text-gold" fill="currentColor">
                  <path d="M12 3.5l2.6 5.3 5.9.9-4.3 4.1 1 5.8-5.2-2.7-5.2 2.7 1-5.8-4.3-4.1 5.9-.9L12 3.5Z" />
                </svg>
                <span class="num text-[12.5px] font-medium text-ink">{{ c.score.toFixed(1) }}</span>
              </div>

              <div class="flex flex-wrap items-center gap-2.5">
                <span v-for="(s, i) in c.star" :key="i" class="text-[11.5px] text-ink-3">
                  {{ STAR_LABELS[i] || '评分' }} <span class="num text-ink-2">{{ s }}</span>
                </span>
              </div>

              <button
                class="ml-auto text-[12px] text-ink-4 opacity-0 transition-all duration-200 hover:text-st-err group-hover:opacity-100"
                @click="openConfirm(c)"
              >
                删除
              </button>
            </div>
          </div>
        </li>
      </ul>

      <EmptyState v-else title="没有匹配的评价" desc="换个关键字试试" />

      <Pagination
        v-if="!loading && list.length"
        v-model:page="query.page"
        v-model:page-size="query.pageSize"
        :total="total"
      />
    </div>

    <Modal v-model="confirmOpen" title="删除评价" width="420px">
      <div class="space-y-3.5">
        <div class="flex items-start gap-3 rounded-card bg-fill-2 p-3">
          <img :src="target?.avatar" class="h-9 w-9 shrink-0 rounded-full bg-fill object-cover" alt="" />
          <div class="min-w-0">
            <div class="text-[13px]">
              <span class="font-medium text-ink">{{ target?.nickname }}</span>
              <span class="text-ink-3"> 在 </span>
              <span class="text-ink">{{ target?.homestayTitle }}</span>
            </div>
            <p class="mt-1 line-clamp-3 text-[12.5px] leading-relaxed text-ink-3">{{ target?.content }}</p>
          </div>
        </div>

        <p class="text-[13px] leading-relaxed text-ink-2">
          删除后会同步扣减该民宿的评价数（软删除，数据仍保留在库中）。此操作不可撤销。
        </p>

        <p v-if="errMsg" class="rounded-btn bg-[#FBF6F5] px-3 py-2 text-[12.5px] text-st-err">{{ errMsg }}</p>
      </div>

      <template #footer>
        <button class="btn btn-ghost btn-md" @click="confirmOpen = false">取消</button>
        <button class="btn btn-primary btn-md" :disabled="saving" @click="doDelete">
          {{ saving ? '删除中…' : '确认删除' }}
        </button>
      </template>
    </Modal>
  </div>
</template>
