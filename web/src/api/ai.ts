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
}

const P = '/ai/v1'

export const fetchHealth = () => post<AiHealth>(`${P}/health`)

export const listConversations = (limit = 20) => post<{ list: AiConversation[] }>(`${P}/conversation/list`, { limit })

export const listMessages = (conversationId: number, limit = 100) =>
  post<{ list: AiMessage[] }>(`${P}/conversation/messages`, { conversationId, limit })

export const deleteConversation = (conversationId: number) =>
  post<{ ok: boolean }>(`${P}/conversation/delete`, { conversationId })

export const confirmAction = (token: string, confirm: boolean) =>
  post<{ ok: boolean; message: string; cards?: AiCard[] }>(`${P}/action/confirm`, { token, confirm })

/**
 * 流式对话。EventSource 只支持 GET 且不能带自定义头，
 * 所以这里用 fetch + ReadableStream 手工解析 SSE。
 */
export async function chatStream(
  payload: { conversationId: number; content: string },
  onEvent: (e: AiEvent) => void,
  signal?: AbortSignal,
): Promise<void> {
  const token = localStorage.getItem(TOKEN_KEY) || ''
  const res = await fetch(`${P}/chat/stream`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      ...(token ? { Authorization: `Bearer ${token}` } : {}),
    },
    body: JSON.stringify(payload),
    signal,
  })

  if (!res.ok) {
    let msg = 'AI 助手暂时不可用，请稍后再试'
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
          /* 忽略半包/坏包 */
        }
      }
    }
  }
}
