/** 与 app/admin/cmd/api/internal/types 一一对应 */

export interface AdminInfo {
  id: number
  username: string
  nickname: string
  role: string
  lastLoginTime: number
}

export interface LoginResp {
  accessToken: string
  accessExpire: number
  refreshAfter: number
  admin: AdminInfo
}

export interface Overview {
  todayOrderCount: number
  todayGmv: number
  yesterdayOrderCount: number
  yesterdayGmv: number
  totalOrderCount: number
  totalGmv: number
  userCount: number
  homestayCount: number
  businessCount: number
  commentCount: number
  waitPayCount: number
  paidCount: number
  cancelCount: number
}

export interface TrendPoint {
  date: string
  orderCount: number
  gmv: number
}

export interface CityStat {
  city: string
  count: number
  gmv: number
}

export interface StateStat {
  tradeState: number
  count: number
}

export interface HomestayItem {
  id: number
  title: string
  subTitle: string
  cover: string
  city: string
  address: string
  tags: string
  peopleNum: number
  homestayPrice: number
  marketHomestayPrice: number
  foodInfo: string
  foodPrice: number
  rowState: number
  score: number
  commentCount: number
  businessId: number
  businessTitle: string
  createTime: number
}

export interface OrderItem {
  id: number
  sn: string
  userId: number
  userMobile: string
  homestayId: number
  title: string
  subTitle: string
  cover: string
  liveStartDate: number
  liveEndDate: number
  livePeopleNum: number
  needFood: number
  homestayPrice: number
  foodPrice: number
  orderTotalPrice: number
  tradeState: number
  tradeCode: string
  remark: string
  createTime: number
}

export interface UserItem {
  id: number
  mobile: string
  nickname: string
  avatar: string
  sex: number
  info: string
  status: number
  orderCount: number
  createTime: number
}

export interface CommentItem {
  id: number
  homestayId: number
  homestayTitle: string
  userId: number
  nickname: string
  avatar: string
  content: string
  star: number[]
  score: number
  createTime: number
}

export interface Paged<T> {
  list: T[] | null
  page: number
  pageSize: number
  total: number
}

export interface OrderPaged extends Paged<OrderItem> {
  sumPrice: number
}

export interface CityOption {
  city: string
  count: number
}

export interface OrderStat {
  all: number
  waitPay: number
  waitUse: number
  used: number
  cancel: number
  refund: number
  expired: number
}
