import { post } from './request'
import type {
  AdminInfo,
  CityOption,
  CityStat,
  CommentItem,
  HomestayItem,
  LoginResp,
  OrderItem,
  OrderPaged,
  OrderStat,
  Overview,
  Paged,
  StateStat,
  TrendPoint,
  UserItem,
} from '@/types'

const P = '/admin/v1'

// ---------------- 管理员 ----------------
export const login = (username: string, password: string) =>
  post<LoginResp>(`${P}/admin/login`, { username, password })

export const fetchProfile = () => post<{ admin: AdminInfo }>(`${P}/admin/profile`)

export const changePassword = (oldPassword: string, newPassword: string) =>
  post<{ ok: boolean }>(`${P}/admin/changePassword`, { oldPassword, newPassword })

// ---------------- 看板 ----------------
export const fetchOverview = () => post<Overview>(`${P}/dashboard/overview`)

export const fetchTrend = (days = 7) => post<{ list: TrendPoint[] }>(`${P}/dashboard/trend`, { days })

export const fetchCityStats = () => post<{ list: CityStat[] }>(`${P}/dashboard/city`)

export const fetchOrderStates = () => post<{ list: StateStat[] }>(`${P}/dashboard/orderState`)

// ---------------- 民宿 ----------------
export interface HomestayQuery {
  page: number
  pageSize: number
  keyword?: string
  city?: string
  rowState?: number
}

export const fetchHomestays = (q: HomestayQuery) => post<Paged<HomestayItem>>(`${P}/homestay/list`, q)

export const fetchHomestayDetail = (id: number) =>
  post<{ homestay: HomestayItem }>(`${P}/homestay/detail`, { id })

export const updateHomestayState = (id: number, rowState: number) =>
  post<{ id: number; rowState: number }>(`${P}/homestay/updateState`, { id, rowState })

export const updateHomestayPrice = (id: number, homestayPrice: number, marketHomestayPrice?: number) =>
  post<{ id: number }>(`${P}/homestay/updatePrice`, { id, homestayPrice, marketHomestayPrice })

export const fetchCityOptions = () => post<{ list: CityOption[] }>(`${P}/homestay/cityOptions`)

// ---------------- 订单 ----------------
export interface OrderQuery {
  page: number
  pageSize: number
  tradeState?: number
  keyword?: string
}

export const fetchOrders = (q: OrderQuery) => post<OrderPaged>(`${P}/order/list`, q)

export const fetchOrderDetail = (sn: string) => post<{ order: OrderItem }>(`${P}/order/detail`, { sn })

export const forceCancelOrder = (sn: string, reason?: string) =>
  post<{ sn: string; tradeState: number }>(`${P}/order/forceCancel`, { sn, reason })

export const fetchOrderStat = () => post<OrderStat>(`${P}/order/stat`)

// ---------------- 用户 ----------------
export const fetchUsers = (page: number, pageSize: number, keyword?: string) =>
  post<Paged<UserItem>>(`${P}/user/list`, { page, pageSize, keyword })

export const updateUserStatus = (id: number, status: number) =>
  post<{ id: number; status: number }>(`${P}/user/updateStatus`, { id, status })

// ---------------- 评价 ----------------
export const fetchComments = (page: number, pageSize: number, homestayId?: number, keyword?: string) =>
  post<Paged<CommentItem>>(`${P}/comment/list`, { page, pageSize, homestayId, keyword })

export const deleteComment = (id: number) => post<{ id: number }>(`${P}/comment/delete`, { id })
