import { defineStore } from 'pinia'
import { ref } from 'vue'
import { TOKEN_KEY } from '@/api/request'
import * as adminApi from '@/api/admin'
import type { AdminInfo } from '@/types'

export const useAuthStore = defineStore('auth', () => {
  const token = ref<string>(localStorage.getItem(TOKEN_KEY) || '')
  const info = ref<AdminInfo | null>(null)
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

  async function fetchProfile() {
    if (!token.value) return null
    loading.value = true
    try {
      const res = await adminApi.fetchProfile()
      info.value = res.admin
      return info.value
    } catch {
      logout()
      return null
    } finally {
      loading.value = false
    }
  }

  async function login(username: string, password: string) {
    const res = await adminApi.login(username, password)
    setToken(res.accessToken)
    info.value = res.admin
  }

  return { token, info, loading, setToken, logout, login, fetchProfile }
})
