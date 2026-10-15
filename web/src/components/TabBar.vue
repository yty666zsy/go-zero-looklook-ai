<script setup lang="ts">
import { useRoute, useRouter } from 'vue-router'
import { computed } from 'vue'

const route = useRoute()
const router = useRouter()

const tabs = [
  { name: 'home', label: '首页', path: '/', icon: 'home' },
  { name: 'orders', label: '订单', path: '/orders', icon: 'order' },
  { name: 'mine', label: '我的', path: '/mine', icon: 'mine' },
] as const

const active = computed(() => route.name as string | undefined)

function go(path: string) {
  if (route.path !== path) router.push(path)
}
</script>

<template>
  <nav
    class="fixed inset-x-0 bottom-0 z-40 mx-auto w-full max-w-[480px] border-t border-line bg-card/92 backdrop-blur-md"
    style="padding-bottom: var(--safe-bottom)"
  >
    <ul class="flex h-14 items-stretch">
      <li v-for="t in tabs" :key="t.name" class="flex-1">
        <button
          class="flex h-full w-full flex-col items-center justify-center gap-1 transition active:scale-95"
          :class="active === t.name ? 'text-ink' : 'text-ink-3'"
          @click="go(t.path)"
        >
          <svg width="21" height="21" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.7">
            <template v-if="t.icon === 'home'">
              <path d="M4 10.5L12 4l8 6.5V19a1 1 0 01-1 1h-4v-6H9v6H5a1 1 0 01-1-1z" stroke-linecap="round" stroke-linejoin="round" />
            </template>
            <template v-else-if="t.icon === 'order'">
              <rect x="4" y="3.5" width="16" height="17" rx="2.5" />
              <path d="M8.5 9h7M8.5 13h7M8.5 17h4" stroke-linecap="round" />
            </template>
            <template v-else>
              <circle cx="12" cy="8.5" r="3.6" />
              <path d="M5 20c1.2-3.5 3.9-5 7-5s5.8 1.5 7 5" stroke-linecap="round" />
            </template>
          </svg>
          <span class="text-[10.5px] leading-none tracking-wide">{{ t.label }}</span>
        </button>
      </li>
    </ul>
  </nav>
</template>
