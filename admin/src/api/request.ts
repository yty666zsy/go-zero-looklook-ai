import axios, { type AxiosInstance, type AxiosRequestConfig } from 'axios'

/** 后端统一返回 {code,msg,data}，这里把它拆开，业务层只拿 data */
export interface ApiResult<T = unknown> {
  code: number
  msg: string
  data: T
}

export const TOKEN_KEY = 'looklook_admin_token'

const http: AxiosInstance = axios.create({
  timeout: 20000,
  headers: { 'Content-Type': 'application/json' },
})

http.interceptors.request.use((config) => {
  const token = localStorage.getItem(TOKEN_KEY)
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})

/** 401 时通知外层跳登录，同时避免循环依赖 */
let onUnauthorized: (() => void) | null = null
export function setUnauthorizedHandler(fn: () => void) {
  onUnauthorized = fn
}

http.interceptors.response.use(
  (res) => res,
  (error) => {
    const status = error?.response?.status
    const msg = error?.response?.data?.msg || error?.message || '网络异常，请稍后重试'

    if (status === 401) {
      localStorage.removeItem(TOKEN_KEY)
      onUnauthorized?.()
      return Promise.reject(new Error('登录已过期，请重新登录'))
    }

    // 后端错误态是 400 + {code,msg}
    return Promise.reject(new Error(msg))
  },
)

function unwrap<T>(body: unknown): T {
  if (body && typeof body === 'object' && 'code' in (body as Record<string, unknown>)) {
    const r = body as ApiResult<T>
    if (r.code !== 200) {
      throw new Error(r.msg || '请求失败')
    }
    return r.data
  }
  return body as T
}

export async function post<T>(url: string, data: unknown = {}, config?: AxiosRequestConfig): Promise<T> {
  const res = await http.post<ApiResult<T>>(url, data, config)
  return unwrap<T>(res.data)
}

export default http
