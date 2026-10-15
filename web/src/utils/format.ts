import type { TradeState } from '@/types'

/** 价格显示：整数不带小数，小数保留两位 */
export function money(v: number | undefined | null): string {
  if (v === undefined || v === null || Number.isNaN(v)) return '0'
  return Number.isInteger(v) ? String(v) : v.toFixed(2)
}

/** 秒级时间戳 -> 2026-10-14 */
export function date(ts: number): string {
  if (!ts) return ''
  const d = new Date(ts * 1000)
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())}`
}

/** 秒级时间戳 -> 10-14 */
export function dateShort(ts: number): string {
  if (!ts) return ''
  const d = new Date(ts * 1000)
  return `${pad(d.getMonth() + 1)}-${pad(d.getDate())}`
}

/** 秒级时间戳 -> 2026-10-14 18:30 */
export function dateTime(ts: number): string {
  if (!ts) return ''
  const d = new Date(ts * 1000)
  return `${date(ts)} ${pad(d.getHours())}:${pad(d.getMinutes())}`
}

/** 入住晚数 */
export function nights(startTs: number, endTs: number): number {
  if (!startTs || !endTs) return 0
  return Math.max(0, Math.round((endTs - startTs) / 86400))
}

/** 相对时间：3天前 / 刚刚 */
export function fromNow(ts: number): string {
  if (!ts) return ''
  const diff = Date.now() / 1000 - ts
  if (diff < 60) return '刚刚'
  if (diff < 3600) return `${Math.floor(diff / 60)}分钟前`
  if (diff < 86400) return `${Math.floor(diff / 3600)}小时前`
  if (diff < 86400 * 30) return `${Math.floor(diff / 86400)}天前`
  return date(ts)
}

export const TRADE_STATE_TEXT: Record<number, string> = {
  [-1]: '已取消',
  [0]: '待支付',
  [1]: '待入住',
  [2]: '已入住',
  [3]: '已退款',
  [4]: '已过期',
}

export function tradeStateText(state: TradeState | number): string {
  return TRADE_STATE_TEXT[state] ?? '未知'
}

export function pad(n: number): string {
  return n < 10 ? `0${n}` : String(n)
}

/** 取当天 0 点的时间戳（秒），用于日期选择 */
export function todayStart(): number {
  const d = new Date()
  d.setHours(0, 0, 0, 0)
  return Math.floor(d.getTime() / 1000)
}

/** 后端是按房间/人次计价的整数天数，这里统一按天 */
export function daysBetweenTs(startTs: number, endTs: number): number {
  return Math.max(1, Math.round((endTs - startTs) / 86400))
}
