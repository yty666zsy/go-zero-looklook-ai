<script setup lang="ts">
import { onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { setUnauthorizedHandler } from '@/api/request'
import { useUserStore } from '@/stores/user'
import ToastHost from '@/components/ToastHost.vue'
import AiChatLauncher from '@/components/ai/AiChatLauncher.vue'

const router = useRouter()
const user = useUserStore()

onMounted(() => {
  // 后端返回 401 时统一跳登录
  setUnauthorizedHandler(() => {
    user.logout()
    router.replace({ name: 'login' })
  })

  if (user.token && !user.info) {
    user.fetchInfo()
  }
})
</script>

<template>
  <div class="relative min-h-screen overflow-hidden">
    <router-view v-slot="{ Component, route }">
      <!-- 带 key：各页面根节点都是 div，页面间跳转时能确保走一次完整的 out-in 转场 -->
      <transition name="route-fade" mode="out-in" :duration="{ enter: 240, leave: 160 }">
        <component :is="Component" :key="route.path" />
      </transition>
    </router-view>
    <ToastHost />
    <!-- AI Agent 层：C 端智能助手入口（悬浮球 + 全屏对话面板） -->
    <AiChatLauncher />
  </div>
</template>
