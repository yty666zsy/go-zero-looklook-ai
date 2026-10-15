<script setup lang="ts">
import { onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { setUnauthorizedHandler } from '@/api/request'
import { useAuthStore } from '@/stores/auth'
import ToastHost from '@/components/ToastHost.vue'

const router = useRouter()
const auth = useAuthStore()

onMounted(() => {
  setUnauthorizedHandler(() => {
    auth.logout()
    router.replace({ name: 'login' })
  })

  if (auth.token && !auth.info) {
    auth.fetchProfile()
  }
})
</script>

<template>
  <router-view v-slot="{ Component, route }">
    <transition name="route-fade" mode="out-in" :duration="{ enter: 220, leave: 140 }">
      <component :is="Component" :key="route.path" />
    </transition>
  </router-view>
  <ToastHost />
</template>
