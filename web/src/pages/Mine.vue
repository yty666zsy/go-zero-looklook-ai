<script setup lang="ts">
import { computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import TabBar from '@/components/TabBar.vue'
import SmartImage from '@/components/SmartImage.vue'
import BottomSheet from '@/components/BottomSheet.vue'
import { ref } from 'vue'
import { useUserStore } from '@/stores/user'
import { useSearchStore } from '@/stores/search'
import { useToast } from '@/composables/useToast'

const router = useRouter()
const user = useUserStore()
const search = useSearchStore()
const toast = useToast()

const showAbout = ref(false)

const nickname = computed(() => user.info?.nickname || '未登录')

onMounted(() => {
  if (user.token && !user.info) user.fetchInfo()
})

const shortcuts = [
  { label: '待付款', state: 0, icon: 'wallet' },
  { label: '待入住', state: 1, icon: 'key' },
  { label: '已入住', state: 2, icon: 'bed' },
]

function goOrders() {
  if (!user.token) {
    router.push({ name: 'login', query: { redirect: '/orders' } })
    return
  }
  router.push('/orders')
}

function logout() {
  user.logout()
  toast.success('已退出登录')
}
</script>

<template>
  <div class="page-scroll pb-[68px]">
    <header class="bg-card px-4 pb-5 pt-8">
      <div class="flex items-center gap-3.5">
        <div class="h-14 w-14 shrink-0 overflow-hidden rounded-full bg-fill ring-1 ring-line">
          <SmartImage v-if="user.info?.avatar" :src="user.info.avatar" :ratio="1" rounded="999px" />
          <div v-else class="flex h-full w-full items-center justify-center text-[18px] text-ink-3">栖</div>
        </div>
        <div class="min-w-0 flex-1">
          <template v-if="user.token">
            <h1 class="ellipsis text-[17px] font-semibold text-ink">{{ nickname }}</h1>
            <p class="mt-1 text-[12.5px] text-ink-3">{{ user.info?.mobile || '' }}</p>
          </template>
          <template v-else>
            <h1 class="text-[17px] font-semibold text-ink">登录 / 注册</h1>
            <p class="mt-1 text-[12.5px] text-ink-3">登录后可管理订单与收藏</p>
          </template>
        </div>
        <button
          v-if="!user.token"
          class="btn-primary h-9 shrink-0 px-5 text-[13px]"
          @click="router.push({ name: 'login', query: { redirect: '/mine' } })"
        >
          去登录
        </button>
      </div>
    </header>

    <!-- 订单快捷入口 -->
    <section class="mx-4 -mt-2 card px-2 py-4">
      <button class="flex w-full items-center justify-between px-2 pb-3.5" @click="goOrders">
        <span class="text-[14.5px] font-medium text-ink">我的订单</span>
        <span class="text-[12px] text-ink-3">全部订单 ›</span>
      </button>
      <div class="flex">
        <button
          v-for="s in shortcuts"
          :key="s.label"
          class="flex flex-1 flex-col items-center gap-2 transition active:scale-95"
          @click="goOrders"
        >
          <span class="flex h-9 w-9 items-center justify-center rounded-xl bg-fill">
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="#5A6068" stroke-width="1.8">
              <template v-if="s.icon === 'wallet'">
                <rect x="3" y="6" width="18" height="13" rx="2.5" />
                <path d="M3 10h18" />
              </template>
              <template v-else-if="s.icon === 'key'">
                <circle cx="8" cy="12" r="3.4" />
                <path d="M11.4 12H21M18 12v3M15.4 12v2.4" stroke-linecap="round" />
              </template>
              <template v-else>
                <path d="M3 18v-5a2 2 0 012-2h14a2 2 0 012 2v5" stroke-linecap="round" />
                <path d="M3 18h18M6 11V7a2 2 0 012-2h8a2 2 0 012 2v4" stroke-linecap="round" />
              </template>
            </svg>
          </span>
          <span class="text-[12px] text-ink-2">{{ s.label }}</span>
        </button>
      </div>
    </section>

    <!-- 常用 -->
    <section class="mx-4 mt-3 card overflow-hidden">
      <button class="flex w-full items-center justify-between px-4 py-3.5 transition active:bg-fill-2" @click="goOrders">
        <span class="text-[14px] text-ink">行程管理</span>
        <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="#9AA0A8" stroke-width="2"><path d="M9 5l7 7-7 7" stroke-linecap="round" /></svg>
      </button>
      <div class="mx-4 divider" />
      <button
        class="flex w-full items-center justify-between px-4 py-3.5 transition active:bg-fill-2"
        @click="(search.reset(), toast.success('搜索条件已重置'))"
      >
        <span class="text-[14px] text-ink">重置搜索条件</span>
        <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="#9AA0A8" stroke-width="2"><path d="M9 5l7 7-7 7" stroke-linecap="round" /></svg>
      </button>
      <div class="mx-4 divider" />
      <button class="flex w-full items-center justify-between px-4 py-3.5 transition active:bg-fill-2" @click="showAbout = true">
        <span class="text-[14px] text-ink">关于栖居</span>
        <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="#9AA0A8" stroke-width="2"><path d="M9 5l7 7-7 7" stroke-linecap="round" /></svg>
      </button>
    </section>

    <section v-if="user.token" class="mx-4 mt-3 card overflow-hidden">
      <button class="w-full px-4 py-3.5 text-center text-[14px] text-ink-2 transition active:bg-fill-2" @click="logout">
        退出登录
      </button>
    </section>

    <p class="py-8 text-center text-[11px] text-ink-4">栖居 · go-zero + Vue3 全栈示例</p>

    <TabBar />

    <BottomSheet v-model="showAbout" title="关于栖居">
      <div class="px-5 py-5 text-[13px] leading-relaxed text-ink-2">
        <p>栖居是一个民宿预订的示例项目：后端基于 go-zero 微服务（travel / order / payment / usercenter），前端为 Vue 3 + Vite + TailwindCSS 的移动端 H5。</p>
        <p class="mt-3">支付环节在开发环境使用内置的模拟支付，方便本地联调。</p>
      </div>
    </BottomSheet>
  </div>
</template>
