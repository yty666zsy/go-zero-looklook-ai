<script setup lang="ts">
import { computed, ref } from 'vue'

interface Point {
  label: string
  value: number
}

const props = withDefaults(
  defineProps<{
    points: Point[]
    /** 第二组数据（例如金额），用浅色虚线画 */
    series2?: Point[]
    height?: number
    unit?: string
  }>(),
  { height: 200, unit: '', series2: () => [] },
)

const W = 720
const H = 220
const PAD = { top: 18, right: 18, bottom: 26, left: 40 }

const hoverIndex = ref<number | null>(null)

const max = computed(() => {
  const all = [...props.points.map((p) => p.value), ...props.series2.map((p) => p.value), 1]
  return Math.max(...all)
})

const innerW = computed(() => W - PAD.left - PAD.right)
const innerH = computed(() => H - PAD.top - PAD.bottom)

function x(i: number, n: number) {
  if (n <= 1) return PAD.left + innerW.value / 2
  return PAD.left + (innerW.value * i) / (n - 1)
}

function y(v: number) {
  return PAD.top + innerH.value * (1 - v / max.value)
}

function buildPath(points: Point[]) {
  if (points.length === 0) return ''
  return points.map((p, i) => `${i === 0 ? 'M' : 'L'}${x(i, points.length).toFixed(1)},${y(p.value).toFixed(1)}`).join(' ')
}

const line1 = computed(() => buildPath(props.points))
const area1 = computed(() => {
  if (props.points.length === 0) return ''
  const first = x(0, props.points.length)
  const last = x(props.points.length - 1, props.points.length)
  const base = PAD.top + innerH.value
  return `${line1.value} L${last.toFixed(1)},${base} L${first.toFixed(1)},${base} Z`
})
const line2 = computed(() => (props.series2.length ? buildPath(props.series2) : ''))

const ticks = computed(() => [0, 0.5, 1].map((r) => ({ y: PAD.top + innerH.value * (1 - r), v: max.value * r })))

const active = computed(() => (hoverIndex.value === null ? null : props.points[hoverIndex.value] ?? null))

function onMove(e: MouseEvent) {
  const svg = e.currentTarget as SVGSVGElement
  const rect = svg.getBoundingClientRect()
  const px = ((e.clientX - rect.left) / rect.width) * W
  const n = props.points.length
  if (n <= 1) {
    hoverIndex.value = 0
    return
  }
  const ratio = (px - PAD.left) / innerW.value
  const i = Math.round(ratio * (n - 1))
  hoverIndex.value = Math.max(0, Math.min(n - 1, i))
}

function fmt(v: number) {
  const s = (Math.round(v * 100) / 100).toString()
  const [int, dec] = s.split('.')
  return int.replace(/\B(?=(\d{3})+(?!\d))/g, ',') + (dec ? `.${dec}` : '')
}
</script>

<template>
  <div class="relative">
    <svg
      :viewBox="`0 0 ${W} ${H}`"
      class="w-full select-none"
      :style="{ height: `${height}px` }"
      @mousemove="onMove"
      @mouseleave="hoverIndex = null"
    >
      <defs>
        <linearGradient id="areaFill" x1="0" y1="0" x2="0" y2="1">
          <stop offset="0%" stop-color="#22262B" stop-opacity="0.10" />
          <stop offset="100%" stop-color="#22262B" stop-opacity="0" />
        </linearGradient>
      </defs>

      <!-- 横向参考线 -->
      <g>
        <line
          v-for="t in ticks"
          :key="t.y"
          :x1="PAD.left"
          :x2="W - PAD.right"
          :y1="t.y"
          :y2="t.y"
          stroke="#E9EBEE"
          stroke-width="1"
          stroke-dasharray="3 5"
        />
        <text
          v-for="t in ticks"
          :key="`l-${t.y}`"
          :x="PAD.left - 8"
          :y="t.y + 3.5"
          text-anchor="end"
          fill="#9AA0A8"
          font-size="10"
        >
          {{ fmt(t.v) }}
        </text>
      </g>

      <path v-if="area1" :d="area1" fill="url(#areaFill)" />
      <path
        :d="line2"
        fill="none"
        stroke="#C7CCD2"
        stroke-width="1.6"
        stroke-dasharray="4 4"
        stroke-linecap="round"
        v-if="line2"
      />
      <path
        :d="line1"
        fill="none"
        stroke="#22262B"
        stroke-width="2"
        stroke-linecap="round"
        stroke-linejoin="round"
        class="chart-line"
        pathLength="1"
      />

      <!-- 数据点 -->
      <circle
        v-for="(p, i) in points"
        :key="`d-${i}`"
        :cx="x(i, points.length)"
        :cy="y(p.value)"
        :r="hoverIndex === i ? 4.5 : 2.6"
        fill="#FFFFFF"
        stroke="#22262B"
        stroke-width="1.8"
        class="transition-all duration-150"
      />

      <!-- 悬浮竖线 -->
      <line
        v-if="hoverIndex !== null"
        :x1="x(hoverIndex, points.length)"
        :x2="x(hoverIndex, points.length)"
        :y1="PAD.top"
        :y2="PAD.top + innerH"
        stroke="#C7CCD2"
        stroke-width="1"
      />

      <!-- X 轴标签 -->
      <text
        v-for="(p, i) in points"
        :key="`t-${i}`"
        :x="x(i, points.length)"
        :y="H - 8"
        text-anchor="middle"
        font-size="10"
        :fill="hoverIndex === i ? '#5A6068' : '#9AA0A8'"
      >
        {{ p.label }}
      </text>
    </svg>

    <!-- 悬浮提示 -->
    <transition name="dialog">
      <div
        v-if="active"
        class="pointer-events-none absolute top-2 rounded-btn border border-line bg-card/95 px-3 py-2 text-[12px] shadow-float backdrop-blur"
        :style="{ left: `calc(${(x(hoverIndex!, points.length) / W) * 100}% - 52px)` }"
      >
        <div class="text-ink-3">{{ active.label }}</div>
        <div class="num mt-0.5 font-medium text-ink">{{ fmt(active.value) }}{{ unit }}</div>
        <div v-if="series2[hoverIndex!]" class="num text-ink-3">
          金额 ¥{{ fmt(series2[hoverIndex!].value) }}
        </div>
      </div>
    </transition>
  </div>
</template>

<style scoped>
.chart-line {
  stroke-dasharray: 1;
  stroke-dashoffset: 0;
  animation: dash 0.9s cubic-bezier(0.16, 1, 0.3, 1) both;
}
@keyframes dash {
  from {
    stroke-dashoffset: 1;
  }
  to {
    stroke-dashoffset: 0;
  }
}
</style>
