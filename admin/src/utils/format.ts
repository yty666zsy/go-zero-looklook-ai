/** 通用格式化：金额 / 时间 / 状态文案 */

export function fen2yuan(v: number | undefined | null): string {
  if (v === undefined || v === null || Number.isNaN(v)) return '0'
  const n = Math.round(Number(v) * 100) / 100
  return n.toFixed(2).replace(/\.00$/, '')
}

/** 千分位金额，带 ¥ 前缀 */
export function money(v: number | undefined | null): string {
  return `¥${fen2yuan(v)}`
}

/** 千分位（不带货币符号），用于 GMV 大数 */
export function amount(v: number | undefined | null): string {
  const n = Number(v || 0)
  const s = (Math.round(n * 100) / 100).toFixed(2).replace(/\.00$/, '')
  const [int, dec] = s.split('.')
  const withSep = int.replace(/\B(?=(\d{3})+(?!\d))/g, ',')
  return dec ? `${withSep}.${dec}` : withSep
}

function pad(n: number) {
  return n < 10 ? `0${n}` : `${n}`
}

/** unix 秒 -> YYYY-MM-DD */
export function date(unix?: number | null): string {
  if (!unix) return '-'
  const d = new Date(unix * 1000)
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())}`
}

/** unix 秒 -> YYYY-MM-DD HH:mm */
export function dateTime(unix?: number | null): string {
  if (!unix) return '-'
  const d = new Date(unix * 1000)
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}`
}

/** unix 秒 -> MM-DD */
export function md(unix?: number | null): string {
  if (!unix) return '-'
  const d = new Date(unix * 1000)
  return `${pad(d.getMonth() + 1)}-${pad(d.getDate())}`
}

/** 相对时间：刚刚 / 5 分钟前 / 3 天前 */
export function fromNow(unix?: number | null): string {
  if (!unix) return '-'
  const diff = Math.floor(Date.now() / 1000) - unix
  if (diff < 60) return '刚刚'
  if (diff < 3600) return `${Math.floor(diff / 60)} 分钟前`
  if (diff < 86400) return `${Math.floor(diff / 3600)} 小时前`
  if (diff < 86400 * 30) return `${Math.floor(diff / 86400)} 天前`
  return date(unix)
}

/** 入住天数 */
export function nights(start?: number | null, end?: number | null): number {
  if (!start || !end || end <= start) return 1
  return Math.max(1, Math.round((end - start) / 86400))
}

export const TRADE_STATES: { value: number; label: string }[] = [
  { value: -99, label: '全部状态' },
  { value: 0, label: '待支付' },
  { value: 1, label: '未使用' },
  { value: 2, label: '已使用' },
  { value: 3, label: '已退款' },
  { value: 4, label: '已过期' },
  { value: -1, label: '已取消' },
]

export function tradeStateText(state: number): string {
  switch (state) {
    case -1:
      return '已取消'
    case 0:
      return '待支付'
    case 1:
      return '未使用'
    case 2:
      return '已使用'
    case 3:
      return '已退款'
    case 4:
      return '已过期'
    default:
      return '未知'
  }
}

/** 状态 -> 低饱和灰调色板，避免出现蓝色 */
export function tradeStateTone(state: number): 'ok' | 'warn' | 'err' | 'off' {
  switch (state) {
    case 1:
    case 2:
      return 'ok'
    case 0:
      return 'warn'
    case -1:
    case 3:
      return 'err'
    default:
      return 'off'
  }
}

/** 把 "成都,春熙路,市中心" 之类的 tags 拆成数组 */
export function splitTags(tags?: string | null): string[] {
  if (!tags) return []
  return tags
    .split(/[,，|]/)
    .map((t) => t.trim())
    .filter(Boolean)
}
