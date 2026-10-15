import { defineStore } from 'pinia'
import { ref } from 'vue'
import { TOKEN_KEY } from '@/api/request'
import * as userApi from '@/api/user'
import type { UserInfo } from '@/types'

export const useUserStore = defineStore('user', () => {
  const token = ref<string>(localStorage.getItem(TOKEN_KEY) || '')
  const info = ref<UserInfo | null>(null)
  const loading = ref(false)

  function setToken(t: string) {
    token.value = t
    if (t) localStorage.setItem(TOKEN_KEY, t)
    else localStorage.removeItem(TOKEN_KEY)
  }

  function logout() {
    setToken('')
    info.value = null
  }

  async function fetchInfo() {
    if (!token.value) return null
    loading.value = true
    try {
      const res = await userApi.fetchUserInfo()
      info.value = res.userInfo
      return info.value
    } catch {
      // token 失效等场景直接登出，交给路由守卫处理
      logout()
      return null
    } finally {
      loading.value = false
    }
  }

  async function login(mobile: string, password: string) {
    const res = await userApi.login(mobile, password)
    setToken(res.accessToken)
    await fetchInfo()
  }

  async function register(mobile: string, password: string) {
    const res = await userApi.register(mobile, password)
    setToken(res.accessToken)
    await fetchInfo()
  }

  return { token, info, loading, setToken, logout, login, register, fetchInfo }
})
