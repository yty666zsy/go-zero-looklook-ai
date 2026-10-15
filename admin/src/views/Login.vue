<script setup lang="ts">
import { ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { useToast } from '@/composables/useToast'

const router = useRouter()
const route = useRoute()
const auth = useAuthStore()
const toast = useToast()

const username = ref('admin')
const password = ref('')
const error = ref('')
const submitting = ref(false)

const demos = [
  { label: '超级管理员', username: 'admin', password: 'admin123' },
  { label: '运营', username: 'operator', password: 'operator123' },
]

function fill(u: string, p: string) {
  username.value = u
  password.value = p
  error.value = ''
}

async function submit() {
  error.value = ''
  if (!username.value || !password.value) {
    error.value = '请输入用户名和密码'
    return
  }
  submitting.value = true
  try {
    await auth.login(username.value, password.value)
    toast.success(`欢迎回来，${auth.info?.nickname || username.value}`)
    const redirect = (route.query.redirect as string) || '/'
    router.replace(redirect)
  } catch (e) {
    error.value = (e as Error).message
  } finally {
    submitting.value = false
  }
}
</script>

<template>
  <div class="relative flex min-h-screen items-center justify-center overflow-hidden px-6 py-10">
    <!-- 背景：淡灰网格 + 两块柔光 -->
    <div class="pointer-events-none absolute inset-0 opacity-[0.5]" aria-hidden="true">
      <div
        class="absolute inset-0"
        style="
          background-image: linear-gradient(#e9ebee 1px, transparent 1px),
            linear-gradient(90deg, #e9ebee 1px, transparent 1px);
          background-size: 34px 34px;
          mask-image: radial-gradient(circle at 50% 42%, #000 0%, transparent 72%);
          -webkit-mask-image: radial-gradient(circle at 50% 42%, #000 0%, transparent 72%);
        "
      />
    </div>
    <div class="pointer-events-none absolute -left-24 top-10 h-72 w-72 rounded-full bg-white blur-3xl" aria-hidden="true" />
    <div class="pointer-events-none absolute -right-20 bottom-0 h-80 w-80 rounded-full bg-[#EDEAE3] blur-3xl" aria-hidden="true" />

    <div class="relative w-full max-w-[400px] animate-fade-up">
      <div class="mb-7 text-center">
        <div class="mx-auto flex h-11 w-11 items-center justify-center rounded-[13px] bg-ink text-[17px] font-semibold text-white shadow-float">
          栖
        </div>
        <h1 class="mt-4 text-[22px] font-semibold tracking-tight text-ink">栖居 · 后台管理</h1>
        <p class="mt-1.5 text-[13px] text-ink-3">looklook 民宿预订运营工作台</p>
      </div>

      <form class="card p-6" @submit.prevent="submit">
        <label class="block">
          <span class="mb-1.5 block text-[12.5px] text-ink-2">用户名</span>
          <input
            v-model="username"
            class="input h-10"
            type="text"
            autocomplete="username"
            placeholder="请输入用户名"
          />
        </label>

        <label class="mt-4 block">
          <span class="mb-1.5 block text-[12.5px] text-ink-2">密码</span>
          <input
            v-model="password"
            class="input h-10"
            type="password"
            autocomplete="current-password"
            placeholder="请输入密码"
          />
        </label>

        <transition name="dialog">
          <p v-if="error" class="mt-4 rounded-btn bg-[#FBF6F5] px-3 py-2 text-[12.5px] text-st-err">{{ error }}</p>
        </transition>

        <button type="submit" class="btn btn-primary mt-5 h-10 w-full" :disabled="submitting">
          <span v-if="submitting" class="flex items-center gap-2">
            <span class="h-3 w-3 animate-spin rounded-full border-2 border-white/40 border-t-white" />
            登录中…
          </span>
          <span v-else>登 录</span>
        </button>

        <div class="mt-5 border-t border-line pt-4">
          <div class="mb-2 text-[12px] text-ink-3">演示账号（点击填入）</div>
          <div class="flex gap-2">
            <button
              v-for="d in demos"
              :key="d.username"
              type="button"
              class="btn btn-line btn-sm flex-1"
              @click="fill(d.username, d.password)"
            >
              {{ d.label }}
            </button>
          </div>
        </div>
      </form>

      <p class="mt-5 text-center text-[12px] text-ink-4">后台与 C 端账号体系相互独立</p>
    </div>
  </div>
</template>
