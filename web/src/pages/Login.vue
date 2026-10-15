<script setup lang="ts">
import { ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import NavBar from '@/components/NavBar.vue'
import { useUserStore } from '@/stores/user'
import { useToast } from '@/composables/useToast'

const route = useRoute()
const router = useRouter()
const user = useUserStore()
const toast = useToast()

const mode = ref<'login' | 'register'>('login')
const mobile = ref('13800000001')
const password = ref('123456')
const loading = ref(false)

async function submit() {
  if (loading.value) return
  const m = mobile.value.trim()
  if (!/^1\d{10}$/.test(m)) {
    toast.error('请输入正确的手机号')
    return
  }
  if (password.value.length < 6) {
    toast.error('密码至少 6 位')
    return
  }

  loading.value = true
  try {
    if (mode.value === 'login') {
      await user.login(m, password.value)
      toast.success('登录成功')
    } else {
      await user.register(m, password.value)
      toast.success('注册成功')
    }
    const redirect = (route.query.redirect as string) || '/mine'
    router.replace(redirect)
  } catch (e) {
    toast.error((e as Error).message)
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="min-h-screen bg-card">
    <NavBar :title="mode === 'login' ? '登录' : '注册'" />

    <div class="px-6 pt-10">
      <h1 class="text-[24px] font-semibold tracking-tight text-ink">
        {{ mode === 'login' ? '欢迎回来' : '开始你的旅程' }}
      </h1>
      <p class="mt-2 text-[13px] text-ink-3">
        {{ mode === 'login' ? '登录后即可预订与查看订单' : '注册一个账号，开启民宿之旅' }}
      </p>

      <div class="mt-9 space-y-4">
        <label class="block">
          <span class="text-[12.5px] text-ink-3">手机号</span>
          <input
            v-model="mobile"
            type="tel"
            inputmode="numeric"
            maxlength="11"
            class="mt-1.5 w-full rounded-btn bg-fill-2 px-4 py-3.5 text-[15px] text-ink outline-none ring-1 ring-line transition focus:ring-ink-4"
            placeholder="请输入手机号"
          />
        </label>

        <label class="block">
          <span class="text-[12.5px] text-ink-3">密码</span>
          <input
            v-model="password"
            type="password"
            maxlength="20"
            class="mt-1.5 w-full rounded-btn bg-fill-2 px-4 py-3.5 text-[15px] text-ink outline-none ring-1 ring-line transition focus:ring-ink-4"
            placeholder="请输入密码"
            @keyup.enter="submit"
          />
        </label>
      </div>

      <button class="btn-primary mt-7 h-12 w-full text-[15.5px] disabled:opacity-50" :disabled="loading" @click="submit">
        {{ loading ? '请稍候…' : mode === 'login' ? '登录' : '注册' }}
      </button>

      <div class="mt-5 text-center text-[13px] text-ink-3">
        <span>{{ mode === 'login' ? '还没有账号？' : '已有账号？' }}</span>
        <button class="ml-1 font-medium text-ink" @click="mode = mode === 'login' ? 'register' : 'login'">
          {{ mode === 'login' ? '去注册' : '去登录' }}
        </button>
      </div>

      <div class="mt-10 rounded-card bg-fill-2 p-4">
        <p class="text-[12px] leading-relaxed text-ink-3">
          演示账号：<span class="font-medium text-ink-2">13800000001</span> ~ <span class="font-medium text-ink-2">13800000008</span>，
          密码统一为 <span class="font-medium text-ink-2">123456</span>。
        </p>
      </div>
    </div>
  </div>
</template>
