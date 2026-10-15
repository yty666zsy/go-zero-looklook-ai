import { post, TOKEN_KEY } from './request'

/** 与 app/aiagent 的 SSE 事件一一对应 */
export interface AiStep {
  index: number
  type: string
  tool?: string
  args?: Record<string, unknown>
  status: 'running' | 'done' | 'fail' | 'pending' | string
  summary?: string
  costMs?: number
  error?: string
}

export interface AiCard {
  kind: string
  title?: string
  items?: unknown
}

export interface AiAction {
  token: string
  tool: string
  args?: Record<string, unknown>
  confirmText: string
  expireIn: number
}

export interface AiEvent {
  type: 'ping' | 'meta' | 'step' | 'delta' | 'retract' | 'card' | 'action' | 'done' | 'error' | string
  conversationId?: number
  messageId?: number
  step?: AiStep
  text?: string
  card?: AiCard
  action?: AiAction
  model?: string
  message?: string
}

export interface AiConversation {
  id: number
  title: string
  msgCount: number
  lastMsg: string
  createTime: number
  updateTime: number
}

export interface AiMessage {
  id: number
  role: string
  content: string
  steps?: AiStep[]
  cards?: AiCard[]
  createTime: number
}

export interface AiHealth {
  provider: string
  model: string
  mock: boolean
  knowledgeCount: number
  toolsCustomer: string[]
  toolsAdmin: string[]
}

export interface KnowledgeItem {
  id: number
  scene: string
  category: string
  title: string
  content: string
  keywords: string
  enabled: number
  updateTime: number
}

export interface KnowledgeSavePayload {
  id?: number
  scene: string
  category?: string
  title: string
  content: string
  keywords?: string
  enabled?: number
}

/** 管理端场景前缀：权限走 AdminJwtAuth，与 C 端完全隔离 */
const B = '/ai/v1'
const A = '/ai/v1/admin'

export const fetchHealth = () => post<AiHealth>(`${B}/health`)

export const listConversations = (limit = 30) => post<{ list: AiConversation[] }>(`${A}/conversation/list`, { limit })

export const listMessages = (conversationId: number, limit = 100) =>
  post<{ list: AiMessage[] }>(`${A}/conversation/messages`, { conversationId, limit })

export const deleteConversation = (conversationId: number) =>
  post<{ ok: boolean }>(`${A}/conversation/delete`, { conversationId })

export const confirmAction = (token: string, confirm: boolean) =>
  post<{ ok: boolean; message: string; cards?: AiCard[] }>(`${A}/action/confirm`, { token, confirm })

export const listKnowledge = (keyword = '', scene = '') =>
  post<{ list: KnowledgeItem[] }>(`${A}/knowledge/list`, { keyword, scene })

export const saveKnowledge = (payload: KnowledgeSavePayload) =>
  post<{ id: number }>(`${A}/knowledge/save`, payload)

export const deleteKnowledge = (id: number) => post<{ id: number }>(`${A}/knowledge/delete`, { id })

/* ---------------- 主动型 Agent：经营日报 + 异常巡检 ---------------- */

export interface AiReport {
  id: number
  type: string
  date: string
  title: string
  summary: string
  content?: string
  metrics?: AiCard[]
  cards?: AiCard[]
  steps?: AiStep[]
  highlights?: string[]
  model?: string
  generatedBy: string
  latencyMs: number
  createTime: number
}

export interface AiAlert {
  id: number
  rule: string
  ruleTitle: string
  level: 'info' | 'warn' | 'error' | string
  category: string
  title: string
  detail: string
  suggestion: string
  actionTool?: string
  actionArgs?: Record<string, unknown>
  actionText?: string
  relatedType?: string
  relatedId?: string
  status: number
  hitCount: number
  lastHitTime: number
  createTime: number
}

export interface AiPatrolRule {
  rule: string
  title: string
  count: number
}

export interface AiPatrolResult {
  batch: string
  startedAt: number
  costMs: number
  scanned: { orders: number; homestays: number; users: number }
  findings: number
  created: number
  updated: number
  cleared: number
  pending: number
  rules: AiPatrolRule[]
  alerts: AiAlert[]
}

export const listReports = (limit = 30) => post<{ list: AiReport[] }>(`${A}/report/list`, { limit })

export const reportDetail = (id: number) => post<{ report: AiReport }>(`${A}/report/detail`, { id })

export const generateReport = (payload: { type?: string; date?: string; force?: boolean } = {}) =>
  post<{ report: AiReport; created: boolean }>(`${A}/report/generate`, payload)

export const runPatrol = () => post<AiPatrolResult>(`${A}/patrol/run`, {})

/** status: -1 全部 / 0 待处理 / 1 已处置 / 2 已忽略 / 3 已消除 */
export const listAlerts = (status = -1, level = '', limit = 50) =>
  post<{ list: AiAlert[]; pendingCount: number; totalCount: number }>(`${A}/alert/list`, { status, level, limit })

/** 一键处置：只生成待确认动作，真正执行要走 confirmAction */
export const alertAction = (id: number) =>
  post<{ alertId: number; token: string; confirmText: string; expireIn: number }>(`${A}/alert/action`, { id })

export const setAlertStatus = (id: number, status: number) =>
  post<{ ok: boolean; status: number }>(`${A}/alert/status`, { id, status })

/**
 * 运营 Copilot 流式对话。
 * EventSource 不支持 POST/自定义头，所以用 fetch + ReadableStream 手工解析 SSE。
 */
export async function chatStream(
  payload: { conversationId: number; content: string },
  onEvent: (e: AiEvent) => void,
  signal?: AbortSignal,
): Promise<void> {
  const token = localStorage.getItem(TOKEN_KEY) || ''
  const res = await fetch(`${A}/chat/stream`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      ...(token ? { Authorization: `Bearer ${token}` } : {}),
    },
    body: JSON.stringify(payload),
    signal,
  })

  if (!res.ok) {
    let msg = 'AI 服务暂时不可用，请稍后再试'
    try {
      const j = (await res.json()) as { msg?: string }
      if (j?.msg) msg = j.msg
    } catch {
      /* ignore */
    }
    throw new Error(msg)
  }
  if (!res.body) throw new Error('当前浏览器不支持流式响应')

  const reader = res.body.getReader()
  const decoder = new TextDecoder('utf-8')
  let buf = ''

  for (;;) {
    const { value, done } = await reader.read()
    if (done) break
    buf += decoder.decode(value, { stream: true })

    const blocks = buf.split('\n\n')
    buf = blocks.pop() ?? ''
    for (const block of blocks) {
      for (const line of block.split('\n')) {
        if (!line.startsWith('data:')) continue
        const raw = line.slice(5).trim()
        if (!raw) continue
        try {
          onEvent(JSON.parse(raw) as AiEvent)
        } catch {
          /* 忽略半包 */
        }
      }
    }
  }
}
