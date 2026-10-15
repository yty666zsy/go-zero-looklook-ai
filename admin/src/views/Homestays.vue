<script setup lang="ts">
import { computed, onMounted, reactive, ref, watch } from 'vue'
import Modal from '@/components/Modal.vue'
import Pagination from '@/components/Pagination.vue'
import EmptyState from '@/components/EmptyState.vue'
import StateTag from '@/components/StateTag.vue'
import { fetchCityOptions, fetchHomestays, updateHomestayPrice, updateHomestayState } from '@/api/admin'
import type { CityOption, HomestayItem } from '@/types'
import { money, splitTags } from '@/utils/format'
import { useToast } from '@/composables/useToast'

const toast = useToast()

const loading = ref(true)
const list = ref<HomestayItem[]>([])
const total = ref(0)
const cities = ref<CityOption[]>([])

const query = reactive({
  page: 1,
  pageSize: 10,
  keyword: '',
  city: '',
  rowState: -1,
})

/** 输入关键字做防抖 */
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
  () => [query.city, query.rowState, query.page, query.pageSize],
  () => load(),
)

async function load() {
  loading.value = true
  try {
    const res = await fetchHomestays({
      page: query.page,
      pageSize: query.pageSize,
      keyword: query.keyword || undefined,
      city: query.city || undefined,
      rowState: query.rowState,
    })
    list.value = res.list || []
    total.value = res.total
  } catch (e) {
    toast.error((e as Error).message)
  } finally {
    loading.value = false
  }
}

async function loadCities() {
  try {
    const res = await fetchCityOptions()
    cities.value = res.list || []
  } catch {
    /* 城市下拉失败不影响主流程 */
  }
}

onMounted(() => {
  loadCities()
  load()
})

function reset() {
  query.keyword = ''
  query.city = ''
  query.rowState = -1
  query.page = 1
  load()
}

const hasFilter = computed(() => !!query.keyword || !!query.city || query.rowState !== -1)

// ---------------- 上下架 ----------------
const toggling = ref<number | null>(null)

async function toggleState(item: HomestayItem) {
  const next = item.rowState === 1 ? 0 : 1
  toggling.value = item.id
  try {
    await updateHomestayState(item.id, next)
    item.rowState = next
    toast.success(next === 1 ? `《${item.title}》已上架` : `《${item.title}》已下架`)
  } catch (e) {
    toast.error((e as Error).message)
  } finally {
    toggling.value = null
  }
}

// ---------------- 改价 ----------------
const priceOpen = ref(false)
const priceTarget = ref<HomestayItem | null>(null)
const priceForm = reactive({ homestayPrice: 0, marketHomestayPrice: 0 })
const priceSaving = ref(false)
const priceError = ref('')

function openPrice(item: HomestayItem) {
  priceTarget.value = item
  priceForm.homestayPrice = item.homestayPrice
  priceForm.marketHomestayPrice = item.marketHomestayPrice
  priceError.value = ''
  priceOpen.value = true
}

const discount = computed(() => {
  if (!priceForm.marketHomestayPrice) return null
  const d = (priceForm.homestayPrice / priceForm.marketHomestayPrice) * 10
  return Number.isFinite(d) && d > 0 ? d.toFixed(1) : null
})

async function savePrice() {
  if (!priceTarget.value) return
  priceError.value = ''
  if (!(priceForm.homestayPrice > 0)) {
    priceError.value = '售价必须大于 0'
    return
  }
  if (priceForm.marketHomestayPrice && priceForm.marketHomestayPrice < priceForm.homestayPrice) {
    priceError.value = '市场价不能低于售价'
    return
  }

  priceSaving.value = true
  try {
    await updateHomestayPrice(priceTarget.value.id, priceForm.homestayPrice, priceForm.marketHomestayPrice)
    priceTarget.value.homestayPrice = priceForm.homestayPrice
    priceTarget.value.marketHomestayPrice = priceForm.marketHomestayPrice
    toast.success('价格已更新，H5 端实时生效')
    priceOpen.value = false
  } catch (e) {
    priceError.value = (e as Error).message
  } finally {
    priceSaving.value = false
  }
}

// ---------------- 详情抽屉（复用 Modal） ----------------
const detailOpen = ref(false)
const detail = ref<HomestayItem | null>(null)

function openDetail(item: HomestayItem) {
  detail.value = item
  detailOpen.value = true
}

const detailTags = computed(() => splitTags(detail.value?.tags))
</script>

<template>
  <div class="space-y-4">
    <!-- 筛选条 -->
    <div class="card flex flex-wrap items-center gap-3 p-4">
      <div class="relative min-w-[220px] flex-1">
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
        <input v-model="query.keyword" class="input pl-9" placeholder="搜索民宿名称 / 城市 / 地址" />
      </div>

      <select v-model="query.city" class="select w-[150px]">
        <option value="">全部城市</option>
        <option v-for="c in cities" :key="c.city" :value="c.city">{{ c.city }}（{{ c.count }}）</option>
      </select>

      <select v-model.number="query.rowState" class="select w-[120px]">
        <option :value="-1">全部状态</option>
        <option :value="1">在售</option>
        <option :value="0">已下架</option>
      </select>

      <button v-if="hasFilter" class="btn btn-ghost btn-md" @click="reset">重置</button>

      <span class="ml-auto text-[12.5px] text-ink-3">
        共 <span class="num text-ink-2">{{ total }}</span> 家民宿
      </span>
    </div>

    <!-- 表格 -->
    <div class="card overflow-hidden">
      <div class="table-wrap">
        <table class="tbl min-w-[940px]">
          <thead>
            <tr>
              <th class="w-[320px]">民宿</th>
              <th class="w-[110px]">城市</th>
              <th class="w-[150px]">价格</th>
              <th class="w-[110px]">评分 / 评价</th>
              <th class="w-[110px]">状态</th>
              <th class="w-[120px] text-right">操作</th>
            </tr>
          </thead>

          <tbody v-if="loading">
            <tr v-for="i in 5" :key="i">
              <td colspan="6">
                <div class="skeleton h-6 w-full" />
              </td>
            </tr>
          </tbody>

          <tbody v-else-if="list.length">
            <tr v-for="item in list" :key="item.id">
              <td>
                <div class="flex items-center gap-3">
                  <img
                    :src="item.cover"
                    class="h-11 w-11 shrink-0 rounded-[10px] bg-fill object-cover"
                    loading="lazy"
                    alt=""
                  />
                  <div class="min-w-0">
                    <button
                      class="ellipsis block max-w-[228px] text-left text-[13.5px] font-medium text-ink transition-colors duration-200 hover:text-ink-2"
                      @click="openDetail(item)"
                    >
                      {{ item.title }}
                    </button>
                    <div class="ellipsis mt-0.5 max-w-[228px] text-[12px] text-ink-3">
                      {{ item.businessTitle || '—' }}
                    </div>
                  </div>
                </div>
              </td>

              <td>
                <span class="text-[13px] text-ink-2">{{ item.city || '—' }}</span>
              </td>

              <td>
                <div class="num text-[13.5px] font-medium text-ink">{{ money(item.homestayPrice) }}</div>
                <div class="num text-[12px] text-ink-4 line-through">{{ money(item.marketHomestayPrice) }}</div>
              </td>

              <td>
                <div class="flex items-center gap-1.5">
                  <svg viewBox="0 0 24 24" class="h-3.5 w-3.5 text-gold" fill="currentColor">
                    <path d="M12 3.5l2.6 5.3 5.9.9-4.3 4.1 1 5.8-5.2-2.7-5.2 2.7 1-5.8-4.3-4.1 5.9-.9L12 3.5Z" />
                  </svg>
                  <span class="num text-[13px] text-ink">{{ item.score || '—' }}</span>
                </div>
                <div class="num text-[12px] text-ink-3">{{ item.commentCount }} 条评价</div>
              </td>

              <td>
                <StateTag
                  :tone="item.rowState === 1 ? 'ok' : 'off'"
                  :text="item.rowState === 1 ? '在售' : '已下架'"
                />
              </td>

              <td>
                <div class="flex items-center justify-end gap-1.5">
                  <button class="btn btn-line btn-sm" @click="openPrice(item)">改价</button>
                  <button
                    class="btn btn-sm"
                    :class="item.rowState === 1 ? 'btn-danger' : 'btn-ghost'"
                    :disabled="toggling === item.id"
                    @click="toggleState(item)"
                  >
                    {{ item.rowState === 1 ? '下架' : '上架' }}
                  </button>
                </div>
              </td>
            </tr>
          </tbody>
        </table>
      </div>

      <EmptyState v-if="!loading && !list.length" title="没有匹配的民宿" desc="换个关键字或清空筛选条件试试">
        <button v-if="hasFilter" class="btn btn-ghost btn-sm mt-1" @click="reset">清空筛选</button>
      </EmptyState>

      <Pagination
        v-if="!loading && list.length"
        v-model:page="query.page"
        v-model:page-size="query.pageSize"
        :total="total"
      />
    </div>

    <!-- 改价弹窗 -->
    <Modal
      v-model="priceOpen"
      :title="`调整价格 · ${priceTarget?.title || ''}`"
      desc="单位：元 / 晚。保存后会立即清掉 C 端缓存，H5 端马上生效。"
      width="420px"
    >
      <div class="space-y-4">
        <label class="block">
          <span class="mb-1.5 block text-[12.5px] text-ink-2">售价（元 / 晚）</span>
          <input v-model.number="priceForm.homestayPrice" class="input" type="number" min="1" step="1" />
        </label>
        <label class="block">
          <span class="mb-1.5 block text-[12.5px] text-ink-2">市场价（元 / 晚）</span>
          <input v-model.number="priceForm.marketHomestayPrice" class="input" type="number" min="1" step="1" />
        </label>

        <div class="flex items-center justify-between rounded-btn bg-fill-2 px-3 py-2.5 text-[12.5px]">
          <span class="text-ink-3">折后</span>
          <span class="text-ink-2">
            <span v-if="discount" class="num font-medium text-ink">{{ discount }} 折</span>
            <span v-else class="text-ink-4">—</span>
            <span class="mx-2 text-ink-4">·</span>
            当前售价 <span class="num text-ink">{{ money(priceForm.homestayPrice) }}</span>
          </span>
        </div>

        <p v-if="priceError" class="rounded-btn bg-[#FBF6F5] px-3 py-2 text-[12.5px] text-st-err">{{ priceError }}</p>
      </div>

      <template #footer>
        <button class="btn btn-ghost btn-md" @click="priceOpen = false">取消</button>
        <button class="btn btn-primary btn-md" :disabled="priceSaving" @click="savePrice">
          {{ priceSaving ? '保存中…' : '保存' }}
        </button>
      </template>
    </Modal>

    <!-- 详情弹窗 -->
    <Modal v-model="detailOpen" :title="detail?.title || '民宿详情'" width="560px">
      <div v-if="detail" class="space-y-4">
        <img :src="detail.cover" class="h-44 w-full rounded-card bg-fill object-cover" alt="" />

        <div class="grid grid-cols-2 gap-x-4 gap-y-3 text-[13px]">
          <div>
            <div class="text-[12px] text-ink-3">城市</div>
            <div class="mt-0.5 text-ink">{{ detail.city || '—' }}</div>
          </div>
          <div>
            <div class="text-[12px] text-ink-3">地址</div>
            <div class="mt-0.5 text-ink">{{ detail.address || '—' }}</div>
          </div>
          <div>
            <div class="text-[12px] text-ink-3">售价 / 市场价</div>
            <div class="num mt-0.5 text-ink">{{ money(detail.homestayPrice) }} / {{ money(detail.marketHomestayPrice) }}</div>
          </div>
          <div>
            <div class="text-[12px] text-ink-3">可住人数</div>
            <div class="mt-0.5 text-ink">{{ detail.peopleNum || '—' }} 人</div>
          </div>
          <div>
            <div class="text-[12px] text-ink-3">所属店铺</div>
            <div class="mt-0.5 text-ink">{{ detail.businessTitle || '—' }}</div>
          </div>
          <div>
            <div class="text-[12px] text-ink-3">评分 / 评价</div>
            <div class="mt-0.5 text-ink">
              <span class="num">{{ detail.score || '—' }}</span> · {{ detail.commentCount }} 条
            </div>
          </div>
        </div>

        <div v-if="detailTags.length" class="flex flex-wrap gap-1.5">
          <span v-for="t in detailTags" :key="t" class="tag bg-fill text-ink-2">{{ t }}</span>
        </div>

        <div v-if="detail.foodInfo">
          <div class="mb-1 text-[12px] text-ink-3">餐食说明</div>
          <p class="rounded-btn bg-fill-2 px-3 py-2.5 text-[12.5px] leading-relaxed text-ink-2">{{ detail.foodInfo }}</p>
        </div>

        <div v-if="detail.subTitle" class="text-[12.5px] leading-relaxed text-ink-3">{{ detail.subTitle }}</div>
      </div>
    </Modal>
  </div>
</template>
