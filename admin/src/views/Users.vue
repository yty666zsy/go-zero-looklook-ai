<script setup lang="ts">
import { computed, onMounted, reactive, ref, watch } from 'vue'
import Modal from '@/components/Modal.vue'
import Pagination from '@/components/Pagination.vue'
import EmptyState from '@/components/EmptyState.vue'
import StateTag from '@/components/StateTag.vue'
import { fetchUsers, updateUserStatus } from '@/api/admin'
import type { UserItem } from '@/types'
import { dateTime, fromNow } from '@/utils/format'
import { useToast } from '@/composables/useToast'

const toast = useToast()

const loading = ref(true)
const list = ref<UserItem[]>([])
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
    const res = await fetchUsers(query.page, query.pageSize, query.keyword || undefined)
    list.value = res.list || []
    total.value = res.total
  } catch (e) {
    toast.error((e as Error).message)
  } finally {
    loading.value = false
  }
}

onMounted(load)

// ---------------- 禁用 / 启用 ----------------
const confirmOpen = ref(false)
const target = ref<UserItem | null>(null)
const saving = ref(false)
const errMsg = ref('')

function openConfirm(u: UserItem) {
  target.value = u
  errMsg.value = ''
  confirmOpen.value = true
}

const nextStatus = computed(() => (target.value?.status === 1 ? 0 : 1))

async function doToggle() {
  if (!target.value) return
  saving.value = true
  errMsg.value = ''
  try {
    await updateUserStatus(target.value.id, nextStatus.value)
    target.value.status = nextStatus.value
    toast.success(nextStatus.value === 1 ? '已恢复该用户' : '已禁用该用户，无法再登录 H5 端')
    confirmOpen.value = false
  } catch (e) {
    errMsg.value = (e as Error).message
  } finally {
    saving.value = false
  }
}

const sexText = (s: number) => (s === 1 ? '男' : s === 0 ? '女' : '未知')
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
        <input v-model="query.keyword" class="input pl-9" placeholder="搜索手机号 / 昵称" />
      </div>

      <span class="ml-auto text-[12.5px] text-ink-3">
        共 <span class="num text-ink-2">{{ total }}</span> 位用户
      </span>
    </div>

    <div class="card overflow-hidden">
      <div class="table-wrap">
        <table class="tbl min-w-[900px]">
          <thead>
            <tr>
              <th class="w-[240px]">用户</th>
              <th class="w-[150px]">手机号</th>
              <th class="w-[90px]">性别</th>
              <th class="w-[110px]">下单数</th>
              <th class="w-[110px]">状态</th>
              <th class="w-[180px]">注册时间</th>
              <th class="w-[110px] text-right">操作</th>
            </tr>
          </thead>

          <tbody v-if="loading">
            <tr v-for="i in 5" :key="i">
              <td colspan="7"><div class="skeleton h-6 w-full" /></td>
            </tr>
          </tbody>

          <tbody v-else-if="list.length">
            <tr v-for="u in list" :key="u.id">
              <td>
                <div class="flex items-center gap-3">
                  <img
                    :src="u.avatar"
                    class="h-9 w-9 shrink-0 rounded-full bg-fill object-cover"
                    loading="lazy"
                    alt=""
                  />
                  <div class="min-w-0">
                    <div class="ellipsis max-w-[150px] text-[13.5px] font-medium text-ink">{{ u.nickname || '未命名' }}</div>
                    <div class="num text-[11.5px] text-ink-3">uid {{ u.id }}</div>
                  </div>
                </div>
              </td>

              <td><span class="num text-[13px] text-ink-2">{{ u.mobile }}</span></td>

              <td><span class="text-[13px] text-ink-2">{{ sexText(u.sex) }}</span></td>

              <td>
                <span class="num text-[13.5px] font-medium text-ink">{{ u.orderCount }}</span>
                <span class="ml-1 text-[12px] text-ink-3">笔</span>
              </td>

              <td>
                <StateTag :tone="u.status === 1 ? 'ok' : 'err'" :text="u.status === 1 ? '正常' : '已禁用'" />
              </td>

              <td>
                <div class="num text-[12.5px] text-ink-2">{{ dateTime(u.createTime) }}</div>
                <div class="text-[11.5px] text-ink-4">{{ fromNow(u.createTime) }}</div>
              </td>

              <td>
                <div class="flex justify-end">
                  <button
                    class="btn btn-sm"
                    :class="u.status === 1 ? 'btn-danger' : 'btn-ghost'"
                    @click="openConfirm(u)"
                  >
                    {{ u.status === 1 ? '禁用' : '启用' }}
                  </button>
                </div>
              </td>
            </tr>
          </tbody>
        </table>
      </div>

      <EmptyState v-if="!loading && !list.length" title="没有匹配的用户" desc="换个手机号或昵称试试" />

      <Pagination
        v-if="!loading && list.length"
        v-model:page="query.page"
        v-model:page-size="query.pageSize"
        :total="total"
      />
    </div>

    <Modal v-model="confirmOpen" :title="nextStatus === 1 ? '恢复用户' : '禁用用户'" width="420px">
      <div class="space-y-3.5">
        <div class="flex items-center gap-3 rounded-card bg-fill-2 p-3">
          <img :src="target?.avatar" class="h-10 w-10 rounded-full bg-fill object-cover" alt="" />
          <div class="min-w-0">
            <div class="text-[13.5px] font-medium text-ink">{{ target?.nickname }}</div>
            <div class="num text-[12px] text-ink-3">{{ target?.mobile }}</div>
          </div>
        </div>

        <p class="text-[13px] leading-relaxed text-ink-2">
          <template v-if="nextStatus === 0">
            禁用后该用户将无法登录 H5 端（登录接口会直接拒绝），已有的订单与评价不受影响。
          </template>
          <template v-else>恢复后该用户可以正常登录并下单。</template>
        </p>

        <p v-if="errMsg" class="rounded-btn bg-[#FBF6F5] px-3 py-2 text-[12.5px] text-st-err">{{ errMsg }}</p>
      </div>

      <template #footer>
        <button class="btn btn-ghost btn-md" @click="confirmOpen = false">取消</button>
        <button class="btn btn-primary btn-md" :disabled="saving" @click="doToggle">
          {{ saving ? '处理中…' : nextStatus === 1 ? '确认恢复' : '确认禁用' }}
        </button>
      </template>
    </Modal>
  </div>
</template>
