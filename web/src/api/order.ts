import { post } from './request'
import type { OrderDetail, OrderListItem } from '@/types'

const PREFIX = '/order/v1'

export function createHomestayOrder(payload: {
  homestayId: number
  isFood: boolean
  liveStartTime: number
  liveEndTime: number
  livePeopleNum: number
  remark?: string
}) {
  return post<{ orderSn: string }>(`${PREFIX}/homestayOrder/createHomestayOrder`, {
    remark: '',
    ...payload,
  })
}

/** tradeState: -99 全部 / 0 待支付 / 1 有效 */
export function fetchOrderList(lastId = 0, pageSize = 10, tradeState = -99) {
  return post<{ list: OrderListItem[] }>(`${PREFIX}/homestayOrder/userHomestayOrderList`, {
    lastId,
    pageSize,
    tradeState,
  })
}

export function fetchOrderDetail(sn: string) {
  return post<OrderDetail>(`${PREFIX}/homestayOrder/userHomestayOrderDetail`, { sn })
}

export function cancelOrder(sn: string) {
  return post<{ sn: string; tradeState: number }>(`${PREFIX}/homestayOrder/cancelHomestayOrder`, { sn })
}
