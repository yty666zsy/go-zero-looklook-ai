/** 与后端 go-zero api 返回结构对齐的类型定义 */

export interface ApiResult<T> {
  code: number
  msg: string
  data: T
}

export interface City {
  code: string
  name: string
  cover: string
  count: number
}

export interface Banner {
  id: number
  cover: string
  title: string
  link: string
}

export interface Homestay {
  id: number
  title: string
  subTitle: string
  banner: string
  cover: string
  images: string[]
  info: string
  peopleNum: number
  homestayBusinessId: number
  userId: number
  rowState: number
  rowType: number
  foodInfo: string
  foodPrice: number
  homestayPrice: number
  marketHomestayPrice: number
  city: string
  address: string
  tags: string[]
  longitude: number
  latitude: number
  score: number
  commentCount: number
}

export interface ScoreStat {
  score: number
  clean: number
  location: number
  service: number
  count: number
}

export interface HomestayBusiness {
  id: number
  title: string
  info: string
  tags: string
  cover: string
  star: number
  isFav: number
  headerImg: string
}

export interface HomeResp {
  cities: City[]
  banners: Banner[]
  hotTags: string[]
  recommend: Homestay[]
  guess: Homestay[]
  recommendCity: string
}

export interface SearchReq {
  keyword?: string
  city?: string
  minPrice?: number
  maxPrice?: number
  peopleNum?: number
  sortBy?: string
  page?: number
  pageSize?: number
}

export interface SearchResp {
  list: Homestay[]
  total: number
}

export interface HomestayDetailResp {
  homestay: Homestay
  images: string[]
  business: HomestayBusiness
  scoreStat: ScoreStat
}

export interface HomestayBusinessListInfo extends HomestayBusiness {
  sellMonth: number
  personConsume: number
}

export interface HomestayBussinessListResp {
  list: HomestayBusinessListInfo[]
}

export interface HomestayComment {
  id: number
  homestayId: number
  content: string
  star: number
  clean: number
  location: number
  service: number
  userId: number
  nickname: string
  avatar: string
  createTime: number
}

export interface CommentListResp {
  list: HomestayComment[]
  total: number
}

/** -1: 已取消 0:待支付 1:未使用 2:已使用 3:已退款 4:已过期 */
export type TradeState = -1 | 0 | 1 | 2 | 3 | 4

export interface OrderListItem {
  id: number
  sn: string
  title: string
  subTitle: string
  homestayId: number
  cover: string
  orderTotalPrice: number
  createTime: number
  tradeState: TradeState
  liveStartDate: number
  liveEndDate: number
  tradeCode: string
}

export interface OrderDetail extends OrderListItem {
  userId: number
  info: string
  foodInfo: string
  foodPrice: number
  homestayPrice: number
  marketHomestayPrice: number
  homestayBusinessId: number
  homestayUserId: number
  tradeCode: string
  foodTotalPrice: number
  homestayTotalPrice: number
  remark: string
  livePeopleNum: number
  needFood: number
  payTime: number
  payType: string
}

export interface UserInfo {
  id: number
  mobile: string
  nickname: string
  sex: number
  avatar: string
  info: string
}

export interface LoginResp {
  accessToken: string
  accessExpire: number
  refreshAfter: number
}
