<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useAiStore } from '@/stores/ai'
import AiCopilotCore from '@/components/ai/AiCopilotCore.vue'

const ai = useAiStore()

interface ToolMeta {
  name: string
  label: string
  desc: string
  write: boolean
}

/** 与 app/aiagent/internal/tools 一一对应：write=true 的工具必须经过二次确认 */
const TOOLS: ToolMeta[] = [
  { name: 'dashboard_overview', label: '经营概览', desc: '今日成交额、订单量、用户数', write: false },
  { name: 'dashboard_trend', label: '下单趋势', desc: '近 N 天下单量与 GMV', write: false },
  { name: 'dashboard_city', label: '城市分布', desc: '各城市订单与成交额', write: false },
  { name: 'order_stat', label: '订单状态统计', desc: '各状态订单数量与金额', write: false },
  { name: 'order_list', label: '订单列表', desc: '按状态 / 关键词分页查询', write: false },
  { name: 'order_detail', label: '订单详情', desc: '按单号查订单详情', write: false },
  { name: 'user_list', label: '用户列表', desc: '按手机号 / 昵称查询', write: false },
  { name: 'homestay_list', label: '房源列表', desc: '按城市 / 状态查询房源', write: false },
  { name: 'force_cancel_order', label: '强制取消订单', desc: '运营侧取消用户订单', write: true },
  { name: 'update_user_status', label: '修改用户状态', desc: '禁用 / 启用用户', write: true },
  { name: 'update_homestay_state', label: '上下架房源', desc: '修改房源在售状态', write: true },
  { name: 'update_homestay_price', label: '修改房源价格', desc: '修改房源售价', write: true },
]

const readTools = computed(() => TOOLS.filter((t) => !t.write))
const writeTools = computed(() => TOOLS.filter((t) => t.write))

/** 服务端真实注册的工具名（用于校验前端清单没有过期） */
const registered = computed(() => new Set(ai.health?.toolsAdmin || []))
const missing = computed(() => (ai.health ? TOOLS.filter((t) => !registered.value.has(t.name)).map((t) => t.name) : []))

const tab = ref<'tools' | 'history'>('tools')

onMounted(() => {
  void ai.warmUp()
})
</script>

<template>
  <div class="grid grid-cols-1 gap-4 xl:grid-cols-[300px_minmax(0,1fr)]">
    <!-- 左：能力与历史 -->
    <div class="space-y-4">
      <div class="card p-4">
        <div class="flex items-center gap-2.5">
          <span class="flex h-9 w-9 items-center justify-center rounded-[10px] bg-ink text-[12px] font-semibold text-white">AI</span>
          <div class="leading-tight">
            <div class="text-[14px] font-semibold text-ink">运营 Copilot</div>
            <div class="text-[11.5px] text-ink-3">
              {{ ai.health?.mock ? '本地规则模式（未配置大模型）' : ai.health ? `模型 ${ai.health.model || ai.health.provider}` : '连接中…' }}
            </div>
          </div>
        </div>

        <dl class="mt-3.5 grid grid-cols-3 gap-2 border-t border-line pt-3.5">
          <div>
            <dt class="text-[11px] text-ink-3">只读工具</dt>
            <dd class="num text-[16px] font-semibold text-ink">{{ readTools.length }}</dd>
          </div>
          <div>
            <dt class="text-[11px] text-ink-3">写操作</dt>
            <dd class="num text-[16px] font-semibold text-ink">{{ writeTools.length }}</dd>
          </div>
          <div>
            <dt class="text-[11px] text-ink-3">知识库</dt>
            <dd class="num text-[16px] font-semibold text-ink">{{ ai.health?.knowledgeCount ?? 0 }}</dd>
          </div>
        </dl>

        <div class="mt-3.5 flex gap-2">
          <button class="btn btn-line btn-sm flex-1" @click="ai.reset()">新对话</button>
          <router-link class="btn btn-line btn-sm flex-1" :to="{ name: 'knowledge' }">知识库</router-link>
          <button class="btn btn-line btn-sm flex-1" @click="tab = 'history'; ai.loadConversations()">历史</button>
        </div>

        <p v-if="missing.length" class="mt-3 rounded-btn bg-[#FBF6F5] px-2.5 py-2 text-[11.5px] leading-relaxed text-st-err">
          服务端未注册：{{ missing.join('、') }}
        </p>
      </div>

      <div class="card overflow-hidden">
        <div class="flex border-b border-line">
          <button
            class="flex-1 px-3 py-2.5 text-[13px] transition-colors duration-150"
            :class="tab === 'tools' ? 'bg-fill-2 font-medium text-ink' : 'text-ink-3 hover:text-ink'"
            @click="tab = 'tools'"
          >
            工具清单
          </button>
          <button
            class="flex-1 px-3 py-2.5 text-[13px] transition-colors duration-150"
            :class="tab === 'history' ? 'bg-fill-2 font-medium text-ink' : 'text-ink-3 hover:text-ink'"
            @click="tab = 'history'; ai.loadConversations()"
          >
            历史会话
          </button>
        </div>

        <!-- 工具清单 -->
        <div v-if="tab === 'tools'" class="max-h-[440px] overflow-y-auto p-3">
          <div class="mb-1.5 px-1 text-[11px] font-medium uppercase tracking-wide text-ink-3">只读 · 自动执行</div>
          <div class="space-y-1">
            <div
              v-for="t in readTools"
              :key="t.name"
              class="rounded-btn px-2 py-1.5 transition-colors duration-150 hover:bg-fill-2"
            >
              <div class="flex items-center gap-2">
                <span class="h-1.5 w-1.5 rounded-full bg-st-ok" />
                <span class="text-[12.5px] text-ink">{{ t.label }}</span>
                <span class="num ml-auto text-[10.5px] text-ink-4">{{ t.name }}</span>
              </div>
              <div class="ml-3.5 text-[11px] text-ink-3">{{ t.desc }}</div>
            </div>
          </div>

          <div class="mb-1.5 mt-3 px-1 text-[11px] font-medium uppercase tracking-wide text-ink-3">写操作 · 需二次确认</div>
          <div class="space-y-1">
            <div
              v-for="t in writeTools"
              :key="t.name"
              class="rounded-btn px-2 py-1.5 transition-colors duration-150 hover:bg-fill-2"
            >
              <div class="flex items-center gap-2">
                <span class="h-1.5 w-1.5 rounded-full bg-st-warn" />
                <span class="text-[12.5px] text-ink">{{ t.label }}</span>
                <span class="num ml-auto text-[10.5px] text-ink-4">{{ t.name }}</span>
              </div>
              <div class="ml-3.5 text-[11px] text-ink-3">{{ t.desc }}</div>
            </div>
          </div>
        </div>

        <!-- 历史会话 -->
        <div v-else class="max-h-[440px] overflow-y-auto p-2">
          <div v-if="!ai.conversations.length" class="py-10 text-center text-[12.5px] text-ink-3">还没有历史会话</div>
          <div
            v-for="c in ai.conversations"
            :key="c.id"
            class="group flex items-center gap-2 rounded-btn px-2 py-2.5 transition-colors duration-150 hover:bg-fill-2"
          >
            <button class="min-w-0 flex-1 text-left" @click="ai.pickConversation(c.id)">
              <div class="ellipsis text-[13px]" :class="c.id === ai.conversationId ? 'font-medium text-ink' : 'text-ink-2'">
                {{ c.title || '未命名会话' }}
              </div>
              <div class="ellipsis text-[11px] text-ink-3">{{ c.msgCount }} 条消息 · {{ c.lastMsg }}</div>
            </button>
            <button
              class="shrink-0 rounded-btn px-2 py-1 text-[11.5px] text-ink-3 opacity-0 transition-opacity duration-150 hover:bg-fill hover:text-st-err group-hover:opacity-100"
              @click="ai.removeConversation(c.id)"
            >
              删除
            </button>
          </div>
        </div>
      </div>
    </div>

    <!-- 右：对话 -->
    <div class="card flex h-[calc(100vh-var(--topbar-h)-48px)] min-h-[560px] flex-col overflow-hidden">
      <header class="flex shrink-0 items-center gap-2 border-b border-line px-4 py-3">
        <span class="text-[14px] font-semibold text-ink">对话</span>
        <span v-if="ai.conversationId" class="tag bg-fill text-ink-3">会话 #{{ ai.conversationId }}</span>
        <span class="ml-auto text-[11.5px] text-ink-3">
          {{ ai.health?.mock ? '本地规则模式' : ai.health ? ai.health.model || ai.health.provider : '' }}
        </span>
      </header>

      <AiCopilotCore variant="page" />
    </div>
  </div>
</template>
