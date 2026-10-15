import { defineStore } from 'pinia'
import { computed, ref } from 'vue'
import * as aiApi from '@/api/ai'
import type { AiAction, AiCard, AiConversation, AiEvent, AiHealth, AiStep } from '@/api/ai'

const CONV_KEY = 'looklook_admin_ai_conv'

export interface ChatItem {
  key: string
  role: 'user' | 'assistant'
  content: string
  steps: AiStep[]
  cards: AiCard[]
  action?: AiAction
  actionState?: 'pending' | 'done' | 'rejected'
  actionMsg?: string
  streaming?: boolean
  error?: string
  createTime?: number
}

let seq = 0
const nextKey = () => `m${Date.now()}_${seq++}`

/** 运营 Copilot：对话状态 + SSE 事件归并（抽屉与整页共用同一份状态） */
export const useAiStore = defineStore('ai', () => {
  const open = ref(false)
  const historyOpen = ref(false)
  const conversationId = ref<number>(Number(localStorage.getItem(CONV_KEY) || 0))
  const items = ref<ChatItem[]>([])
  const sending = ref(false)
  const conversations = ref<AiConversation[]>([])
  const health = ref<AiHealth | null>(null)

  let controller: AbortController | null = null

  const toolCount = computed(() => health.value?.toolsAdmin.length ?? 0)

  function setConversation(id: number) {
    conversationId.value = id
    if (id > 0) localStorage.setItem(CONV_KEY, String(id))
    else localStorage.removeItem(CONV_KEY)
  }

  function toggle() {
    open.value = !open.value
    if (open.value) void warmUp()
  }

  function close() {
    open.value = false
    historyOpen.value = false
  }

  /** 打开时补一次元信息 / 恢复上次会话 */
  async function warmUp() {
    if (!health.value) void loadHealth()
    if (!items.value.length && conversationId.value > 0) await restore()
    if (!items.value.length) void loadConversations()
  }

  async function loadHealth() {
    try {
      health.value = await aiApi.fetchHealth()
    } catch {
      health.value = null
    }
  }

  async function restore() {
    if (!conversationId.value) return
    try {
      const res = await aiApi.listMessages(conversationId.value)
      items.value = res.list.map((m) => ({
        key: `h${m.id}`,
        role: m.role === 'user' ? 'user' : 'assistant',
        content: m.content,
        steps: m.steps || [],
        cards: m.cards || [],
        createTime: m.createTime,
      }))
    } catch {
      setConversation(0)
      items.value = []
    }
  }

  function reset() {
    controller?.abort()
    controller = null
    sending.value = false
    setConversation(0)
    items.value = []
    historyOpen.value = false
  }

  function stop() {
    controller?.abort()
    controller = null
    sending.value = false
    const last = items.value[items.value.length - 1]
    if (last?.streaming) last.streaming = false
    if (last && !last.content) last.content = '（已停止）'
  }

  function handleEvent(item: ChatItem, e: AiEvent) {
    switch (e.type) {
      case 'meta':
        if (e.conversationId) setConversation(e.conversationId)
        break
      case 'step': {
        if (!e.step) break
        const i = item.steps.findIndex((s) => s.index === e.step!.index)
        if (i >= 0) item.steps[i] = { ...item.steps[i], ...e.step }
        else item.steps.push(e.step)
        break
      }
      case 'delta':
        item.content += e.text || ''
        break
      case 'retract': {
        // 模型调工具前的过渡语，按长度撤回已流式展示的部分
        const n = (e.text || '').length
        if (n > 0) item.content = item.content.slice(0, Math.max(0, item.content.length - n))
        break
      }
      case 'card':
        if (e.card) item.cards.push(e.card)
        break
      case 'action':
        if (e.action) {
          item.action = e.action
          item.actionState = 'pending'
        }
        break
      case 'error':
        item.error = e.message || 'AI 服务出错了'
        break
      default:
        break
    }
  }

  async function send(text: string) {
    const content = text.trim()
    if (!content || sending.value) return

    items.value.push({
      key: nextKey(),
      role: 'user',
      content,
      steps: [],
      cards: [],
      createTime: Math.floor(Date.now() / 1000),
    })
    items.value.push({
      key: nextKey(),
      role: 'assistant',
      content: '',
      steps: [],
      cards: [],
      streaming: true,
    })
    // 关键：push 进去的是原始对象，必须从响应式数组里取回代理再改，
    // 否则后续的 content/steps 变更不会触发视图更新。
    const assistant = items.value[items.value.length - 1]
    sending.value = true

    controller = new AbortController()
    try {
      await aiApi.chatStream(
        { conversationId: conversationId.value, content },
        (e) => handleEvent(assistant, e),
        controller.signal,
      )
    } catch (err) {
      const msg = err instanceof Error ? err.message : '网络异常'
      if (msg.includes('abort')) {
        if (!assistant.content) assistant.content = '（已停止）'
      } else {
        assistant.error = msg
      }
    } finally {
      assistant.streaming = false
      sending.value = false
      controller = null
      void loadConversations()
    }
  }

  /** 写操作二次确认：确认后才真正落到业务系统 */
  async function confirm(item: ChatItem, ok: boolean) {
    if (!item.action || item.actionState !== 'pending') return
    const act = item.action
    item.actionState = ok ? 'done' : 'rejected'
    try {
      const res = await aiApi.confirmAction(act.token, ok)
      item.actionMsg = res.message
      if (res.cards?.length) {
        items.value.push({
          key: nextKey(),
          role: 'assistant',
          content: '',
          steps: [],
          cards: res.cards,
        })
      }
    } catch (err) {
      item.actionState = 'pending'
      item.actionMsg = err instanceof Error ? err.message : '操作失败'
    }
  }

  async function loadConversations() {
    try {
      const res = await aiApi.listConversations(30)
      conversations.value = res.list
    } catch {
      conversations.value = []
    }
  }

  async function pickConversation(id: number) {
    if (sending.value) stop()
    setConversation(id)
    items.value = []
    historyOpen.value = false
    await restore()
  }

  async function removeConversation(id: number) {
    try {
      await aiApi.deleteConversation(id)
    } catch {
      /* ignore */
    }
    conversations.value = conversations.value.filter((c) => c.id !== id)
    if (conversationId.value === id) reset()
  }

  return {
    open,
    historyOpen,
    conversationId,
    items,
    sending,
    conversations,
    health,
    toolCount,
    toggle,
    close,
    warmUp,
    loadHealth,
    restore,
    reset,
    stop,
    send,
    confirm,
    loadConversations,
    pickConversation,
    removeConversation,
  }
})
