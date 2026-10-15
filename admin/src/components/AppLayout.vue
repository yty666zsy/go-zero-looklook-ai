<script setup lang="ts">
import { computed, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { useAiStore } from '@/stores/ai'
import { useToast } from '@/composables/useToast'
import ChangePasswordDialog from '@/components/ChangePasswordDialog.vue'
import AiCopilot from '@/components/ai/AiCopilot.vue'

const route = useRoute()
const router = useRouter()
const auth = useAuthStore()
const ai = useAiStore()
const toast = useToast()

interface NavItem {
  name: string
  path: string
  label: string
  icon: string
}

const nav: NavItem[] = [
  {
    name: 'dashboard',
    path: '/',
    label: '数据看板',
    icon: 'M3 13h4v8H3v-8Zm7-9h4v17h-4V4Zm7 5h4v12h-4V9Z',
  },
  {
    name: 'homestays',
    path: '/homestays',
    label: '民宿管理',
    icon: 'M4 10.5 12 4l8 6.5V20a1 1 0 0 1-1 1h-4v-6H9v6H5a1 1 0 0 1-1-1v-9.5Z',
  },
  {
    name: 'orders',
    path: '/orders',
    label: '订单管理',
    icon: 'M6 3h9l4 4v14H6V3Zm8 1.5V8h3.5L14 4.5ZM9 12h7M9 16h5',
  },
  {
    name: 'users',
    path: '/users',
    label: '用户管理',
    icon: 'M12 12a4 4 0 1 0 0-8 4 4 0 0 0 0 8Zm8 8c0-3.3-3.6-6-8-6s-8 2.7-8 6',
  },
  {
    name: 'comments',
    path: '/comments',
    label: '评价管理',
    icon: 'M5 4h14a1 1 0 0 1 1 1v10a1 1 0 0 1-1 1H9l-4 4V5a1 1 0 0 1 1-1Z',
  },
  {
    name: 'agent',
    path: '/agent',
    label: 'AI 助手',
    icon: 'M12 3.5l1.7 4.1 4.1 1.7-4.1 1.7L12 15l-1.7-4-4.1-1.7 4.1-1.7L12 3.5ZM18 14.5l.8 1.9 1.9.8-1.9.8-.8 1.9-.8-1.9-1.9-.8 1.9-.8.8-1.9Z',
  },
  {
    name: 'knowledge',
    path: '/knowledge',
    label: '知识库',
    icon: 'M5 4.5h9.5L19 9v10.5H5V4.5Zm8 1.5V9h3.5L13 6ZM8.5 12h7M8.5 15.5h4.5',
  },
  {
    name: 'monitor',
    path: '/monitor',
    label: 'AI 值守',
    icon: 'M12 3a6 6 0 0 0-6 6c0 3.2-1 4.6-1.7 5.4-.4.5-.1 1.3.6 1.3h14.2c.7 0 1-.8.6-1.3C19 13.6 18 12.2 18 9a6 6 0 0 0-6-6Zm-2.4 15a2.5 2.5 0 0 0 4.8 0',
  },
]

const title = computed(() => (route.meta.title as string) || '')
const subtitle = computed(() => (route.meta.subtitle as string) || '')

const pwdOpen = ref(false)
const menuOpen = ref(false)

function isActive(item: NavItem) {
  return route.name === item.name
}

function onLogout() {
  auth.logout()
  toast.info('已退出登录')
  router.replace({ name: 'login' })
}

const roleText: Record<string, string> = {
  super: '超级管理员',
  operator: '运营',
}
</script>

<template>
  <div class="min-h-screen">
    <!-- 侧边栏 -->
    <aside
      class="fixed inset-y-0 left-0 z-30 flex w-[var(--sidebar-w)] flex-col border-r border-line bg-card"
    >
      <div class="flex h-[var(--topbar-h)] items-center gap-2.5 border-b border-line px-5">
        <div class="flex h-7 w-7 items-center justify-center rounded-[9px] bg-ink text-[13px] font-semibold text-white">
          栖
        </div>
        <div class="leading-tight">
          <div class="text-[15px] font-semibold tracking-tight text-ink">栖居后台</div>
          <div class="text-[11px] text-ink-3">looklook admin</div>
        </div>
      </div>

      <nav class="flex-1 space-y-1 overflow-y-auto p-3">
        <router-link
          v-for="item in nav"
          :key="item.name"
          :to="item.path"
          class="group relative flex h-10 items-center gap-3 rounded-btn px-3 text-sm transition-colors duration-200"
          :class="isActive(item) ? 'bg-fill text-ink font-medium' : 'text-ink-2 hover:bg-fill-2 hover:text-ink'"
        >
          <span
            class="absolute left-0 top-1/2 h-5 w-[3px] -translate-y-1/2 rounded-r-pill bg-ink transition-all duration-200"
            :class="isActive(item) ? 'opacity-100 scale-y-100' : 'opacity-0 scale-y-50'"
          />
          <svg viewBox="0 0 24 24" class="h-[18px] w-[18px] shrink-0" fill="none" stroke="currentColor" stroke-width="1.7" stroke-linecap="round" stroke-linejoin="round">
            <path :d="item.icon" />
          </svg>
          <span>{{ item.label }}</span>
        </router-link>
      </nav>

      <div class="border-t border-line p-3">
        <div class="flex items-center gap-2.5 rounded-btn px-2 py-2">
          <div class="flex h-8 w-8 items-center justify-center rounded-full bg-fill text-[13px] font-medium text-ink-2">
            {{ (auth.info?.nickname || 'A').slice(0, 1) }}
          </div>
          <div class="min-w-0 flex-1 leading-tight">
            <div class="ellipsis text-[13px] font-medium text-ink">{{ auth.info?.nickname || '未登录' }}</div>
            <div class="ellipsis text-[11px] text-ink-3">
              {{ roleText[auth.info?.role || ''] || auth.info?.username || '-' }}
            </div>
          </div>
          <button
            class="relative flex h-7 w-7 items-center justify-center rounded-btn text-ink-3 transition-colors duration-200 hover:bg-fill hover:text-ink"
            title="账号"
            @click="menuOpen = !menuOpen"
          >
            <svg viewBox="0 0 24 24" class="h-4 w-4" fill="currentColor">
              <circle cx="5" cy="12" r="1.6" />
              <circle cx="12" cy="12" r="1.6" />
              <circle cx="19" cy="12" r="1.6" />
            </svg>
          </button>

          <transition name="dialog">
            <div
              v-if="menuOpen"
              class="absolute bottom-16 right-3 w-36 overflow-hidden rounded-btn border border-line bg-card py-1 shadow-float"
              @mouseleave="menuOpen = false"
            >
              <button
                class="flex w-full items-center px-3 py-2 text-left text-[13px] text-ink-2 transition-colors duration-150 hover:bg-fill"
                @click="pwdOpen = true; menuOpen = false"
              >
                修改密码
              </button>
              <button
                class="flex w-full items-center px-3 py-2 text-left text-[13px] text-st-err transition-colors duration-150 hover:bg-fill"
                @click="onLogout"
              >
                退出登录
              </button>
            </div>
          </transition>
        </div>
      </div>
    </aside>

    <!-- 主区域 -->
    <div class="ml-[var(--sidebar-w)]">
      <header
        class="sticky top-0 z-20 flex h-[var(--topbar-h)] items-center justify-between border-b border-line bg-card/85 px-6 backdrop-blur"
      >
        <div class="flex items-baseline gap-3">
          <h1 class="text-[17px] font-semibold tracking-tight text-ink">{{ title }}</h1>
          <span v-if="subtitle" class="text-[13px] text-ink-3">{{ subtitle }}</span>
        </div>
        <div class="flex items-center gap-3">
          <button
            v-if="route.name !== 'agent'"
            class="btn btn-line btn-sm"
            title="在任何页面唤起 AI 助手"
            @click="ai.toggle()"
          >
            <svg viewBox="0 0 24 24" class="h-3.5 w-3.5" fill="none" stroke="currentColor" stroke-width="1.7" stroke-linecap="round" stroke-linejoin="round">
              <path d="M12 3.5l1.7 4.1 4.1 1.7-4.1 1.7L12 15l-1.7-4-4.1-1.7 4.1-1.7L12 3.5ZM18 14.5l.8 1.9 1.9.8-1.9.8-.8 1.9-.8-1.9-1.9-.8 1.9-.8.8-1.9Z" />
            </svg>
            AI 助手
          </button>
          <a
            class="btn btn-line btn-sm"
            href="http://127.0.0.1:5173"
            target="_blank"
            rel="noreferrer"
          >
            打开 H5 端
          </a>
          <button class="btn btn-sm bg-ink text-white hover:bg-[#31363d]" @click="pwdOpen = true">修改密码</button>
        </div>
      </header>

      <main class="p-6">
        <router-view v-slot="{ Component }">
          <transition name="route-fade" mode="out-in" :duration="{ enter: 200, leave: 120 }">
            <component :is="Component" />
          </transition>
        </router-view>
      </main>
    </div>

    <ChangePasswordDialog v-model:open="pwdOpen" />

    <!-- 全局运营 Copilot 抽屉（自身带入场动画，这里不再包 transition，避免离场残留） -->
    <AiCopilot v-if="ai.open" />
  </div>
</template>
