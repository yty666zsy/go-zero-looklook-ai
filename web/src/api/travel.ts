import { post } from './request'
import type {
  CommentListResp,
  HomeResp,
  Homestay,
  HomestayBussinessListResp,
  HomestayDetailResp,
  SearchReq,
  SearchResp,
} from '@/types'

const PREFIX = '/travel/v1'

/** 首页聚合 */
export function fetchHome(city = '') {
  return post<HomeResp>(`${PREFIX}/homestay/home`, { city })
}

/** 搜索 / 筛选 */
export function searchHomestay(req: SearchReq) {
  return post<SearchResp>(`${PREFIX}/homestay/search`, {
    page: 1,
    pageSize: 10,
    ...req,
  })
}

/** 民宿详情 */
export function fetchHomestayDetail(id: number) {
  return post<HomestayDetailResp>(`${PREFIX}/homestay/homestayDetail`, { id })
}

/** 评价列表 */
export function fetchComments(homestayId: number, page = 1, pageSize = 10) {
  return post<CommentListResp>(`${PREFIX}/homestayComment/commentList`, {
    homestayId,
    page,
    pageSize,
  })
}

/** 优选民宿（活动位） */
export function fetchHomestayList(page = 1, pageSize = 10) {
  return post<{ list: Homestay[] }>(`${PREFIX}/homestay/homestayList`, { page, pageSize })
}

/** 猜你喜欢 */
export function fetchGuessList() {
  return post<{ list: Homestay[] }>(`${PREFIX}/homestay/guessList`, {})
}

/** 店铺列表 */
export function fetchBusinessList(lastId = 0, pageSize = 10) {
  return post<HomestayBussinessListResp>(`${PREFIX}/homestayBussiness/homestayBussinessList`, {
    lastId,
    pageSize,
  })
}
