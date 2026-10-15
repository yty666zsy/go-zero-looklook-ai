import { post } from './request'
import type { LoginResp, UserInfo } from '@/types'

const PREFIX = '/usercenter/v1'

export function login(mobile: string, password: string) {
  return post<LoginResp>(`${PREFIX}/user/login`, { mobile, password })
}

export function register(mobile: string, password: string) {
  return post<LoginResp>(`${PREFIX}/user/register`, { mobile, password })
}

export function fetchUserInfo() {
  return post<{ userInfo: UserInfo }>(`${PREFIX}/user/detail`, {})
}
