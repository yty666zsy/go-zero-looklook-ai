<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import * as aiApi from '@/api/ai'
import type { AiAlert, AiCard, AiReport, AiPatrolResult, AiStep } from '@/api/ai'
import { useToast } from '@/composables/useToast'
import AiCards from '@/components/ai/AiCards.vue'
import Modal from '@/components/Modal.vue'
import EmptyState from '@/components/EmptyState.vue'
import { dateTime } from '@/utils/format'

const toast = useToast()

/* ================= Tab ================= */
const tab = ref<'alert' | 'report'>('alert')

/* ================= 巡检与告警 ================= */
const alerts = ref<AiAlert[]>([])
const alertLoading = ref(false)
const pendingCount = ref(0)
const totalCount = ref(0)
const patrolStatus = ref(-1)
const patrolLevel = ref('')
const patroling = ref(false)
const lastPatrol = ref<AiPatrolResult | null>(null)

const STATUS_OPTIONS = [
  { value: -1, label: '全部' },
  { value: 0, label: '待处理' },
  { value: 1, label: '已处置' },
  { value: 2, label: '已忽略' },
  { value: 3, label: '已消除' },
]

const LEVEL_META: Record<string, { text: string; cls: string; bar: string }> = {
  error: { text: '严重', cls: 'bg-[#FBF6F5] text-st-err', bar: 'border-l-[#D8B4AE]' },
  warn: { text: '警告', cls: 'bg-[#F7F1E3] text-st-warn', bar: 'border-l-[#E2CBA3]' },
  info: { text: '提示', cls: 'bg-fill text-ink-2', bar: 'border-l-[#D7DBD7]' },
}

function levelMeta(l: string) {
  return LEVEL_META[l] || LEVEL_META.info
}

const STATUS_META: Record<number, { text: string; cls: string }> = {
  0: { text: '待处理', cls: 'bg-[#F7F1E3] text-st-warn' },
  1: { text: '已处置', cls: 'bg-[#EDF2ED] text-st-ok' },
  2: { text: '已忽略', cls: 'bg-fill text-ink-3' },
  3: { text: '已消除', cls: 'bg-fill text-ink-4' },
}

function statusMeta(s: number) {
  return STATUS_META[s] || STATUS_META[0]
}

async function loadAlerts() {
  alertLoading.value = true
  try {
    const res = await aiApi.listAlerts(patrolStatus.value, patrolLevel.value, 100)
    alerts.value = res.list || []
    pendingCount.value = res.pendingCount || 0
    totalCount.value = res.totalCount || 0
  } catch (e) {
    toast.error(e instanceof Error ? e.message : '加载告警失败')
  } finally {
    alertLoading.value = false
  }
}

async function runPatrol() {
  if (patroling.value) return
  patroling.value = true
  try {
    const res = await aiApi.runPatrol()
    lastPatrol.value = res
    toast.success(`巡检完成：命中 ${res.findings} 条异常，新增 ${res.created} 条`)
    await loadAlerts()
  } catch (e) {
    toast.error(e instanceof Error ? e.message : '巡检失败')
  } finally {
    patroling.value = false
  }
}

/* ---- 一键处置：仍然是「生成待确认动作 → 人工确认 → 才落地」 ---- */
const actionOpen = ref(false)
const actionAlert = ref<AiAlert | null>(null)
const actionToken = ref('')
const actionConfirmText = ref('')
const actionLoading = ref(false)
const actionBusy = ref(false)
const actionResult = ref('')

async function startAction(a: AiAlert) {
  actionAlert.value = a
  actionLoading.value = true
  actionResult.value = ''
  actionOpen.value = true
  try {
    const res = await aiApi.alertAction(a.id)
    actionToken.value = res.token
    actionConfirmText.value = res.confirmText
  } catch (e) {
    actionToken.value = ''
    actionConfirmText.value = ''
    toast.error(e instanceof Error ? e.message : '生成处置动作失败')
  } finally {
    actionLoading.value = false
  }
}

async function confirmAction() {
  if (!actionAlert.value || !actionToken.value) return
  actionBusy.value = true
  try {
    const res = await aiApi.confirmAction(actionToken.value, true)
    actionResult.value = res.message || '已执行'
    await aiApi.setAlertStatus(actionAlert.value.id, 1)
    toast.success('已执行并标记为已处置')
    await loadAlerts()
  } catch (e) {
    toast.error(e instanceof Error ? e.message : '执行失败')
  } finally {
    actionBusy.value = false
  }
}

async function markAlert(a: AiAlert, status: number) {
  try {
    await aiApi.setAlertStatus(a.id, status)
    toast.success(status === 2 ? '已忽略该告警' : status === 0 ? '已重新打开' : '已更新')
    await loadAlerts()
  } catch (e) {
    toast.error(e instanceof Error ? e.message : '操作失败')
  }
}

const expanded = ref<number>(0)
function toggleExpand(id: number) {
  expanded.value = expanded.value === id ? 0 : id
}

/* ================= 经营日报 ================= */
const reports = ref<AiReport[]>([])
const reportLoading = ref(false)
const current = ref<AiReport | null>(null)
const detailLoading = ref(false)
const generating = ref(false)
const showSteps = ref(true)

const currentMetrics = computed<AiCard[]>(() => current.value?.metrics || [])
const currentCards = computed<AiCard[]>(() => current.value?.cards || [])
const currentSteps = computed<AiStep[]>(() => current.value?.steps || [])

async function loadReports(selectFirst = true) {
  reportLoading.value = true
  try {
    const res = await aiApi.listReports(40)
    reports.value = res.list || []
    if (selectFirst) {
      if (reports.value.length) await openReport(reports.value[0].id)
      else current.value = null
    }
  } catch (e) {
    toast.error(e instanceof Error ? e.message : '加载日报失败')
  } finally {
    reportLoading.value = false
  }
}

async function openReport(id: number) {
  detailLoading.value = true
  try {
    const res = await aiApi.reportDetail(id)
    current.value = res.report
  } catch (e) {
    toast.error(e instanceof Error ? e.message : '加载日报详情失败')
  } finally {
    detailLoading.value = false
  }
}

async function generate(force: boolean) {
  generating.value = true
  try {
    const res = await aiApi.generateReport({ force })
    toast.success(res.created ? '日报已生成' : '当天已有日报，已直接打开')
    await loadReports(false)
    current.value = res.report
  } catch (e) {
    toast.error(e instanceof Error ? e.message : '生成失败')
  } finally {
    generating.value = false
  }
}

onMounted(async () => {
  await Promise.all([loadAlerts(), loadReports()])
})
</script>

<template>
  <div class="space-y-4">
    <!-- 说明 + 值守状态 -->
    <div class="card flex flex-wrap items-center gap-3 px-4 py-3">
      <div class="min-w-[260px] flex-1 text-[12.5px] leading-relaxed text-ink-2">
        这是<strong class="font-medium text-ink">主动型 Agent</strong>：不用你提问，服务每
        <span class="num">30</span> 分钟自己巡检一遍经营数据、每天固定时间自己写一份日报。
        涉及改数据的处置动作仍需<strong class="font-medium text-ink">人工点确认</strong>后才会执行。
      </div>
      <div class="flex items-center gap-2">
        <span class="tag" :class="pendingCount > 0 ? 'bg-[#F7F1E3] text-st-warn' : 'bg-[#EDF2ED] text-st-ok'">
          待处理 {{ pendingCount }}
        </span>
        <span class="tag bg-fill text-ink-3">累计 {{ totalCount }}</span>
        <button class="btn btn-primary btn-md" :disabled="patroling" @click="runPatrol">
          <svg viewBox="0 0 24 24" class="mr-1 h-3.5 w-3.5" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round">
            <path d="M20 11a8 8 0 1 0-2.3 5.7M20 5v6h-6" />
          </svg>
          {{ patroling ? '巡检中…' : '立即巡检' }}
        </button>
      </div>
    </div>

    <!-- 最近一次巡检结果 -->
    <div v-if="lastPatrol" class="card animate-fade-up px-4 py-3">
      <div class="flex flex-wrap items-center gap-x-4 gap-y-1.5 text-[12.5px] text-ink-2">
        <span class="num font-medium text-ink">批次 {{ lastPatrol.batch }}</span>
        <span>扫描 <span class="num">{{ lastPatrol.scanned.orders }}</span> 单 / <span class="num">{{ lastPatrol.scanned.homestays }}</span> 套房源 / <span class="num">{{ lastPatrol.scanned.users }}</span> 位用户</span>
        <span>命中 <span class="num text-st-warn">{{ lastPatrol.findings }}</span> 条</span>
        <span>新增 <span class="num">{{ lastPatrol.created }}</span> · 更新 <span class="num">{{ lastPatrol.updated }}</span></span>
        <span>自动消除 <span class="num">{{ lastPatrol.cleared }}</span></span>
        <span class="num text-ink-3">耗时 {{ lastPatrol.costMs }}ms</span>
      </div>
      <div v-if="lastPatrol.rules.length" class="mt-2 flex flex-wrap gap-1.5">
        <span v-for="r in lastPatrol.rules" :key="r.rule" class="tag bg-fill text-ink-2">
          {{ r.title }} <span class="num ml-1 text-ink-3">×{{ r.count }}</span>
        </span>
      </div>
    </div>

    <!-- Tab -->
    <div class="flex items-center gap-1 rounded-card border border-line bg-card p-1">
      <button
        class="flex-1 rounded-[8px] px-3 py-2 text-[13px] transition-colors duration-200"
        :class="tab === 'alert' ? 'bg-ink text-white' : 'text-ink-2 hover:bg-fill-2'"
        @click="tab = 'alert'"
      >
        异常告警
        <span v-if="pendingCount" class="num ml-1">{{ pendingCount }}</span>
      </button>
      <button
        class="flex-1 rounded-[8px] px-3 py-2 text-[13px] transition-colors duration-200"
        :class="tab === 'report' ? 'bg-ink text-white' : 'text-ink-2 hover:bg-fill-2'"
        @click="tab = 'report'"
      >
        经营日报
        <span v-if="reports.length" class="num ml-1">{{ reports.length }}</span>
      </button>
    </div>

    <!-- ================= 异常告警 ================= -->
    <template v-if="tab === 'alert'">
      <div class="card flex flex-wrap items-center gap-2 p-3">
        <div class="flex flex-wrap gap-1.5">
          <button
            v-for="s in STATUS_OPTIONS"
            :key="s.value"
            class="rounded-btn px-2.5 py-1 text-[12.5px] transition-colors duration-150"
            :class="patrolStatus === s.value ? 'bg-ink text-white' : 'bg-fill-2 text-ink-2 hover:bg-fill'"
            @click="patrolStatus = s.value; loadAlerts()"
          >
            {{ s.label }}
          </button>
        </div>
        <select v-model="patrolLevel" class="select ml-auto w-[140px]" @change="loadAlerts">
          <option value="">全部等级</option>
          <option value="error">严重</option>
          <option value="warn">警告</option>
          <option value="info">提示</option>
        </select>
      </div>

      <div v-if="alertLoading" class="card space-y-2 p-4">
        <div v-for="i in 4" :key="i" class="skeleton h-16 w-full" />
      </div>

      <EmptyState
        v-else-if="!alerts.length"
        title="当前没有告警"
        desc="点上方「立即巡检」按规则扫一遍，或等定时任务自动跑。"
      />

      <div v-else class="space-y-2.5">
        <div
          v-for="a in alerts"
          :key="a.id"
          class="card overflow-hidden transition-shadow duration-200"
          :class="a.status === 0 ? 'shadow-[0_1px_2px_rgba(0,0,0,.04)]' : 'opacity-80'"
        >
          <div class="flex items-start gap-3 border-l-[3px] px-4 py-3" :class="levelMeta(a.level).bar">
            <div class="min-w-0 flex-1">
              <div class="flex flex-wrap items-center gap-2">
                <span class="tag" :class="levelMeta(a.level).cls">{{ levelMeta(a.level).text }}</span>
                <span class="tag bg-fill text-ink-2">{{ a.ruleTitle }}</span>
                <span class="ellipsis text-[13.5px] font-medium text-ink">{{ a.title }}</span>
                <span v-if="a.hitCount > 1" class="num tag bg-fill text-ink-3">命中 {{ a.hitCount }} 次</span>
              </div>
              <div class="mt-1.5 text-[12.5px] leading-relaxed text-ink-2">{{ a.suggestion }}</div>
              <div class="mt-1 flex flex-wrap items-center gap-3 text-[11.5px] text-ink-4">
                <span class="num">{{ a.relatedId }}</span>
                <span>最近命中 {{ dateTime(a.lastHitTime) }}</span>
                <span :class="statusMeta(a.status).cls" class="tag">{{ statusMeta(a.status).text }}</span>
                <button class="text-ink-3 underline-offset-2 hover:underline" @click="toggleExpand(a.id)">
                  {{ expanded === a.id ? '收起明细' : '查看明细' }}
                </button>
              </div>

              <div
                v-if="expanded === a.id"
                class="mt-2 whitespace-pre-wrap rounded-btn bg-fill-2 p-3 text-[12px] leading-relaxed text-ink-2"
              >
                {{ a.detail }}
              </div>
            </div>

            <div class="flex shrink-0 flex-col gap-1.5">
              <button
                v-if="a.status === 0 && a.actionTool"
                class="btn btn-primary btn-sm"
                @click="startAction(a)"
              >
                一键处置
              </button>
              <button v-if="a.status === 0" class="btn btn-ghost btn-sm" @click="markAlert(a, 2)">忽略</button>
              <button v-if="a.status !== 0" class="btn btn-line btn-sm" @click="markAlert(a, 0)">重开</button>
            </div>
          </div>
        </div>
      </div>
    </template>

    <!-- ================= 经营日报 ================= -->
    <template v-else>
      <div class="card flex flex-wrap items-center gap-2 p-3">
        <span class="text-[12.5px] text-ink-2">日报由服务每天自动生成（默认 <span class="num">09:00</span>），也可以手动补一份。</span>
        <div class="ml-auto flex gap-2">
          <button class="btn btn-line btn-md" :disabled="generating" @click="generate(false)">立即生成</button>
          <button class="btn btn-ghost btn-md" :disabled="generating" @click="generate(true)">强制重算当天</button>
        </div>
      </div>

      <div class="grid gap-4 lg:grid-cols-[250px_1fr]">
        <!-- 日报列表 -->
        <div class="card overflow-hidden">
          <div class="border-b border-line px-3.5 py-2.5 text-[13px] font-medium text-ink">历史日报</div>
          <div v-if="reportLoading" class="space-y-2 p-3">
            <div v-for="i in 4" :key="i" class="skeleton h-10 w-full" />
          </div>
          <EmptyState v-else-if="!reports.length" title="还没有日报" desc="点「立即生成」生成第一份。" />
          <div v-else class="max-h-[560px] overflow-y-auto">
            <button
              v-for="r in reports"
              :key="r.id"
              class="block w-full border-b border-line px-3.5 py-2.5 text-left transition-colors duration-150 last:border-0 hover:bg-fill-2"
              :class="current?.id === r.id ? 'bg-fill-2' : ''"
              @click="openReport(r.id)"
            >
              <div class="flex items-center gap-2">
                <span class="num text-[13px] font-medium text-ink">{{ r.date }}</span>
                <span class="tag bg-fill text-ink-3">{{ r.generatedBy === 'cron' ? '定时' : '手动' }}</span>
              </div>
              <div class="ellipsis mt-1 text-[11.5px] text-ink-3">{{ r.summary }}</div>
            </button>
          </div>
        </div>

        <!-- 日报详情 -->
        <div v-if="detailLoading" class="card space-y-3 p-4">
          <div class="skeleton h-6 w-[40%]" />
          <div class="skeleton h-24 w-full" />
          <div class="skeleton h-40 w-full" />
        </div>

        <div v-else-if="!current" class="card">
          <EmptyState title="选择一份日报" desc="左侧列表里挑一份查看，或先生成。" />
        </div>

        <div v-else class="space-y-3">
          <div class="card px-4 py-3.5">
            <div class="flex flex-wrap items-center gap-2">
              <h2 class="text-[15px] font-medium text-ink">{{ current.title }}</h2>
              <span class="tag bg-fill text-ink-3">{{ current.generatedBy === 'cron' ? '定时生成' : '手动生成' }}</span>
              <span v-if="current.model" class="tag bg-fill text-ink-3">模型润色 {{ current.model }}</span>
              <span v-else class="tag bg-fill text-ink-4">确定性文案</span>
              <span class="num ml-auto text-[11.5px] text-ink-4">耗时 {{ current.latencyMs }}ms · {{ dateTime(current.createTime) }}</span>
            </div>

            <div v-if="current.highlights?.length" class="mt-2.5 flex flex-wrap gap-1.5">
              <span v-for="(h, i) in current.highlights" :key="i" class="tag bg-fill-2 text-ink-2">{{ h }}</span>
            </div>
          </div>

          <!-- 指标卡片（复用对话里的卡片渲染） -->
          <AiCards v-if="currentMetrics.length" :cards="currentMetrics" />

          <!-- 正文 -->
          <div class="card px-4 py-3.5">
            <div class="whitespace-pre-wrap text-[13px] leading-relaxed text-ink-2">{{ current.content }}</div>
          </div>

          <!-- Agent 执行过程 -->
          <div v-if="currentSteps.length" class="card overflow-hidden">
            <button class="flex w-full items-center gap-2 px-4 py-3 text-left" @click="showSteps = !showSteps">
              <span class="flex h-4 w-4 items-center justify-center">
                <svg viewBox="0 0 16 16" class="h-3.5 w-3.5 text-st-ok" fill="none" stroke="currentColor" stroke-width="1.7">
                  <path d="M3 8.5l3 3 7-7" stroke-linecap="round" stroke-linejoin="round" />
                </svg>
              </span>
              <span class="flex-1 text-[12.5px] text-ink-2">
                Agent 执行过程
                <span class="text-ink-3">· 调用了 {{ currentSteps.length }} 个工具 · {{ currentSteps.map((s) => s.tool).join('、') }}</span>
              </span>
              <svg
                viewBox="0 0 16 16"
                class="h-3.5 w-3.5 text-ink-3 transition-transform duration-200"
                :class="showSteps ? 'rotate-180' : ''"
                fill="none"
                stroke="currentColor"
                stroke-width="1.6"
              >
                <path d="M4 6.5l4 4 4-4" stroke-linecap="round" stroke-linejoin="round" />
              </svg>
            </button>

            <div v-if="showSteps" class="space-y-0 border-t border-line px-4 py-3">
              <div v-for="(s, si) in currentSteps" :key="si" class="flex gap-2.5">
                <div class="flex flex-col items-center">
                  <span
                    class="mt-[5px] h-2 w-2 shrink-0 rounded-full"
                    :class="s.status === 'done' ? 'bg-st-ok' : s.status === 'fail' ? 'bg-st-err' : 'bg-ink-4'"
                  />
                  <span v-if="si < currentSteps.length - 1" class="mt-1 w-px flex-1 bg-line" />
                </div>
                <div class="min-w-0 flex-1 pb-2">
                  <div class="flex items-center gap-2 text-[12.5px] text-ink">
                    <span class="num">{{ s.tool }}</span>
                    <span v-if="s.costMs" class="num text-[11px] text-ink-4">{{ s.costMs }}ms</span>
                  </div>
                  <div v-if="s.summary" class="mt-0.5 text-[11.5px] text-ink-3">{{ s.summary }}</div>
                  <div v-if="s.error" class="mt-0.5 text-[11.5px] text-st-err">{{ s.error }}</div>
                </div>
              </div>
            </div>
          </div>

          <!-- 其余结构化卡片（趋势 / 分布） -->
          <AiCards v-if="currentCards.length > 1" :cards="currentCards.slice(1)" />
        </div>
      </div>
    </template>

    <!-- 一键处置确认 -->
    <Modal v-model="actionOpen" title="确认处置" width="480px" desc="Agent 不会自己改数据，点了确认才会真正执行。">
      <div v-if="actionAlert" class="space-y-3">
        <div class="rounded-btn bg-fill-2 p-3">
          <div class="flex items-center gap-2">
            <span class="tag" :class="levelMeta(actionAlert.level).cls">{{ levelMeta(actionAlert.level).text }}</span>
            <span class="text-[12.5px] font-medium text-ink">{{ actionAlert.title }}</span>
          </div>
          <p class="mt-1.5 text-[12.5px] leading-relaxed text-ink-2">{{ actionAlert.suggestion }}</p>
        </div>

        <div v-if="actionLoading" class="text-[12.5px] text-ink-3">正在生成待确认动作…</div>
        <div v-else-if="actionConfirmText" class="rounded-btn border border-line-2 bg-card p-3">
          <div class="text-[12px] text-ink-3">将要执行：</div>
          <div class="mt-1 text-[13px] leading-relaxed text-ink">{{ actionConfirmText }}</div>
          <div class="num mt-1.5 text-[11px] text-ink-4">工具 {{ actionAlert.actionTool }} · 令牌 5 分钟内有效且只能执行一次</div>
        </div>
        <div v-else class="text-[12.5px] text-st-err">该告警暂时无法生成处置动作。</div>

        <div v-if="actionResult" class="rounded-btn bg-[#EDF2ED] px-3 py-2 text-[12.5px] text-st-ok">{{ actionResult }}</div>
      </div>

      <template #footer>
        <button class="btn btn-line btn-md" @click="actionOpen = false">关闭</button>
        <button
          class="btn btn-primary btn-md"
          :disabled="actionBusy || actionLoading || !actionToken || !!actionResult"
          @click="confirmAction"
        >
          {{ actionBusy ? '执行中…' : '确认执行' }}
        </button>
      </template>
    </Modal>
  </div>
</template>
