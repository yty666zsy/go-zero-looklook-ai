<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import * as aiApi from '@/api/ai'
import type { KnowledgeItem } from '@/api/ai'
import { useToast } from '@/composables/useToast'
import Modal from '@/components/Modal.vue'
import EmptyState from '@/components/EmptyState.vue'
import { dateTime } from '@/utils/format'

const toast = useToast()

const SCENES = [
  { value: '', label: '全部场景' },
  { value: 'all', label: '通用（全场景）' },
  { value: 'c', label: '仅客户端' },
  { value: 'admin', label: '仅管理端' },
]

const sceneText: Record<string, string> = { all: '通用', c: '客户端', admin: '管理端' }
const sceneCls: Record<string, string> = {
  all: 'bg-fill text-ink-2',
  c: 'bg-[#EDF2ED] text-st-ok',
  admin: 'bg-[#F7F1E3] text-st-warn',
}

const list = ref<KnowledgeItem[]>([])
const loading = ref(false)
const keyword = ref('')
const scene = ref('')

const filtered = computed(() => list.value)

async function load() {
  loading.value = true
  try {
    const res = await aiApi.listKnowledge(keyword.value.trim(), scene.value)
    list.value = res.list || []
  } catch (e) {
    toast.error(e instanceof Error ? e.message : '加载失败')
  } finally {
    loading.value = false
  }
}

onMounted(load)

/* ---------------- 编辑 / 新增 ---------------- */
const formOpen = ref(false)
const saving = ref(false)
const form = reactive({
  id: 0,
  scene: 'c',
  category: '',
  title: '',
  content: '',
  keywords: '',
  enabled: 1,
})

function openCreate() {
  Object.assign(form, { id: 0, scene: 'c', category: '', title: '', content: '', keywords: '', enabled: 1 })
  formOpen.value = true
}

function openEdit(k: KnowledgeItem) {
  Object.assign(form, {
    id: k.id,
    scene: k.scene || 'c',
    category: k.category || '',
    title: k.title,
    content: k.content,
    keywords: k.keywords || '',
    enabled: k.enabled,
  })
  formOpen.value = true
}

async function submit() {
  if (!form.title.trim() || !form.content.trim()) {
    toast.error('标题和内容都需要填写')
    return
  }
  saving.value = true
  try {
    await aiApi.saveKnowledge({
      id: form.id || undefined,
      scene: form.scene,
      category: form.category.trim(),
      title: form.title.trim(),
      content: form.content.trim(),
      keywords: form.keywords.trim(),
      enabled: form.enabled,
    })
    toast.success(form.id ? '已保存，索引已热更新' : '已新增，索引已热更新')
    formOpen.value = false
    await load()
  } catch (e) {
    toast.error(e instanceof Error ? e.message : '保存失败')
  } finally {
    saving.value = false
  }
}

/* ---------------- 启停 / 删除 ---------------- */
const toggling = ref<number>(0)

async function toggleEnabled(k: KnowledgeItem) {
  toggling.value = k.id
  try {
    await aiApi.saveKnowledge({
      id: k.id,
      scene: k.scene,
      category: k.category,
      title: k.title,
      content: k.content,
      keywords: k.keywords,
      enabled: k.enabled === 1 ? 0 : 1,
    })
    toast.success(k.enabled === 1 ? '已停用，该条目不再被检索' : '已启用')
    await load()
  } catch (e) {
    toast.error(e instanceof Error ? e.message : '操作失败')
  } finally {
    toggling.value = 0
  }
}

const delTarget = ref<KnowledgeItem | null>(null)
const deleting = ref(false)

async function doDelete() {
  const t = delTarget.value
  if (!t) return
  deleting.value = true
  try {
    await aiApi.deleteKnowledge(t.id)
    toast.success('已删除')
    delTarget.value = null
    await load()
  } catch (e) {
    toast.error(e instanceof Error ? e.message : '删除失败')
  } finally {
    deleting.value = false
  }
}

/* ---------------- 预览 ---------------- */
const previewTarget = ref<KnowledgeItem | null>(null)
</script>

<template>
  <div class="space-y-4">
    <!-- 说明条 -->
    <div class="card flex flex-wrap items-center gap-3 px-4 py-3">
      <div class="flex-1 text-[12.5px] leading-relaxed text-ink-2">
        这里是 AI 助手（客户端 + 管理端）共用的知识库。保存后<strong class="font-medium text-ink">立即生效</strong>（内存索引热更新，无需重启服务）；
        用「场景」控制可见范围，避免运营内部口径泄露给 C 端用户。
      </div>
      <button class="btn btn-primary btn-sm" @click="openCreate">
        <svg viewBox="0 0 24 24" class="h-3.5 w-3.5" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round">
          <path d="M12 5v14M5 12h14" />
        </svg>
        新增知识条目
      </button>
    </div>

    <!-- 过滤 -->
    <div class="card flex flex-wrap items-center gap-2 p-3">
      <input
        v-model="keyword"
        class="input w-[260px]"
        placeholder="搜索标题 / 内容 / 关键词"
        @keydown.enter="load"
      />
      <select v-model="scene" class="select w-[160px]" @change="load">
        <option v-for="s in SCENES" :key="s.value" :value="s.value">{{ s.label }}</option>
      </select>
      <button class="btn btn-line btn-md" @click="load">查询</button>
      <button
        class="btn btn-ghost btn-md"
        @click="keyword = ''; scene = ''; load()"
      >
        重置
      </button>
      <span class="ml-auto text-[12.5px] text-ink-3">共 {{ filtered.length }} 条</span>
    </div>

    <!-- 列表 -->
    <div class="card overflow-hidden">
      <div v-if="loading" class="space-y-2 p-4">
        <div v-for="i in 5" :key="i" class="skeleton h-9 w-full" />
      </div>

      <EmptyState v-else-if="!filtered.length" title="没有匹配的知识条目" desc="换个关键词，或新增一条。" />

      <div v-else class="table-wrap">
        <table class="tbl">
          <thead>
            <tr>
              <th class="w-[64px]">ID</th>
              <th class="w-[96px]">场景</th>
              <th class="w-[110px]">分类</th>
              <th>标题 / 内容</th>
              <th class="w-[170px]">关键词</th>
              <th class="w-[92px]">状态</th>
              <th class="w-[150px]">更新时间</th>
              <th class="w-[190px] text-right">操作</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="k in filtered" :key="k.id">
              <td class="num text-ink-3">{{ k.id }}</td>
              <td>
                <span class="tag" :class="sceneCls[k.scene] || 'bg-fill text-ink-2'">
                  {{ sceneText[k.scene] || k.scene }}
                </span>
              </td>
              <td class="text-[12.5px]">{{ k.category || '-' }}</td>
              <td>
                <button class="block max-w-[520px] text-left" @click="previewTarget = k">
                  <div class="ellipsis font-medium text-ink">{{ k.title }}</div>
                  <div class="ellipsis text-[12px] text-ink-3">{{ k.content }}</div>
                </button>
              </td>
              <td class="ellipsis text-[12px] text-ink-3">{{ k.keywords || '-' }}</td>
              <td>
                <button
                  class="tag transition-colors duration-150"
                  :class="k.enabled === 1 ? 'bg-[#EDF2ED] text-st-ok' : 'bg-fill text-st-off'"
                  :disabled="toggling === k.id"
                  @click="toggleEnabled(k)"
                >
                  {{ k.enabled === 1 ? '已启用' : '已停用' }}
                </button>
              </td>
              <td class="num text-[12px] text-ink-3">{{ dateTime(k.updateTime) }}</td>
              <td>
                <div class="flex justify-end gap-1.5">
                  <button class="btn btn-line btn-sm" @click="openEdit(k)">编辑</button>
                  <button class="btn btn-danger btn-sm" @click="delTarget = k">删除</button>
                </div>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <!-- 编辑弹窗 -->
    <Modal
      v-model="formOpen"
      :title="form.id ? `编辑知识条目 #${form.id}` : '新增知识条目'"
      desc="场景决定该条知识对谁可见；关键词用于提升检索命中率（逗号分隔）。"
      width="560px"
    >
      <div class="space-y-3.5">
        <div class="grid grid-cols-2 gap-3">
          <label class="block">
            <span class="mb-1.5 block text-[12.5px] text-ink-2">场景</span>
            <select v-model="form.scene" class="select">
              <option value="all">通用（客户端 + 管理端）</option>
              <option value="c">仅客户端</option>
              <option value="admin">仅管理端</option>
            </select>
          </label>
          <label class="block">
            <span class="mb-1.5 block text-[12.5px] text-ink-2">分类</span>
            <input v-model="form.category" class="input" placeholder="如 政策 / 运营 / 产品" />
          </label>
        </div>

        <label class="block">
          <span class="mb-1.5 block text-[12.5px] text-ink-2">标题</span>
          <input v-model="form.title" class="input" placeholder="一句话说明这条知识" />
        </label>

        <label class="block">
          <span class="mb-1.5 block text-[12.5px] text-ink-2">内容</span>
          <textarea
            v-model="form.content"
            rows="7"
            class="input h-auto resize-y py-2 leading-relaxed"
            placeholder="写清楚规则细节，AI 会据此回答用户"
          />
        </label>

        <label class="block">
          <span class="mb-1.5 block text-[12.5px] text-ink-2">关键词</span>
          <input v-model="form.keywords" class="input" placeholder="取消,退款,押金" />
        </label>

        <label class="flex items-center gap-2">
          <input type="checkbox" :checked="form.enabled === 1" @change="form.enabled = form.enabled === 1 ? 0 : 1" />
          <span class="text-[12.5px] text-ink-2">启用（停用后不会被检索到）</span>
        </label>
      </div>

      <template #footer>
        <button class="btn btn-line btn-md" @click="formOpen = false">取消</button>
        <button class="btn btn-primary btn-md" :disabled="saving" @click="submit">
          {{ saving ? '保存中…' : '保存' }}
        </button>
      </template>
    </Modal>

    <!-- 删除确认 -->
    <Modal
      :model-value="!!delTarget"
      title="删除知识条目"
      width="400px"
      @update:model-value="delTarget = null"
    >
      <p class="text-[13px] leading-relaxed text-ink-2">
        确定删除「{{ delTarget?.title }}」吗？删除后 AI 将不再引用这条内容，该操作不可撤销。
      </p>
      <template #footer>
        <button class="btn btn-line btn-md" @click="delTarget = null">取消</button>
        <button class="btn btn-danger btn-md" :disabled="deleting" @click="doDelete">
          {{ deleting ? '删除中…' : '确认删除' }}
        </button>
      </template>
    </Modal>

    <!-- 预览 -->
    <Modal
      :model-value="!!previewTarget"
      :title="previewTarget?.title"
      width="560px"
      @update:model-value="previewTarget = null"
    >
      <div class="space-y-3">
        <div class="flex flex-wrap items-center gap-2">
          <span class="tag" :class="sceneCls[previewTarget?.scene || ''] || 'bg-fill text-ink-2'">
            {{ sceneText[previewTarget?.scene || ''] || previewTarget?.scene }}
          </span>
          <span class="tag bg-fill text-ink-2">{{ previewTarget?.category || '未分类' }}</span>
          <span class="tag bg-fill text-ink-3">#{{ previewTarget?.id }}</span>
          <span class="ml-auto text-[12px] text-ink-3">更新于 {{ dateTime(previewTarget?.updateTime || 0) }}</span>
        </div>
        <p class="whitespace-pre-wrap rounded-btn bg-fill-2 p-3 text-[13px] leading-relaxed text-ink-2">
          {{ previewTarget?.content }}
        </p>
        <div v-if="previewTarget?.keywords" class="flex flex-wrap gap-1.5">
          <span v-for="k in previewTarget.keywords.split(',')" :key="k" class="tag bg-fill text-ink-3">{{ k }}</span>
        </div>
      </div>
    </Modal>
  </div>
</template>
