import { post } from './request'

const PREFIX = '/payment/v1'

/** 开发环境模拟支付：后端 Mode=dev 时才可用 */
export function mockPay(orderSn: string) {
  return post<{ orderSn: string; tradeState: number }>(`${PREFIX}/thirdPayment/mockPay`, { orderSn })
}
