<script setup lang="ts">
import { ref, watch } from 'vue'
import BottomSheet from './BottomSheet.vue'
import Stepper from './Stepper.vue'

const props = defineProps<{
  modelValue: boolean
  minPrice: number
  maxPrice: number
  peopleNum: number
}>()

const emit = defineEmits<{
  'update:modelValue': [boolean]
  change: [payload: { minPrice: number; maxPrice: number; peopleNum: number }]
}>()

const PRESETS = [
  { label: '不限', min: 0, max: 0 },
  { label: '¥300 以下', min: 0, max: 300 },
  { label: '¥300-600', min: 300, max: 600 },
  { label: '¥600-1000', min: 600, max: 1000 },
  { label: '¥1000 以上', min: 1000, max: 0 },
]

const priceIdx = ref(0)
const people = ref(props.peopleNum)

watch(
  () => props.modelValue,
  (v) => {
    if (!v) return
    people.value = props.peopleNum
    const idx = PRESETS.findIndex((p) => p.min === props.minPrice && p.max === props.maxPrice)
    priceIdx.value = idx >= 0 ? idx : 0
  },
)

function reset() {
  priceIdx.value = 0
  people.value = 2
}

function confirm() {
  const p = PRESETS[priceIdx.value]
  emit('change', { minPrice: p.min, maxPrice: p.max, peopleNum: people.value })
  emit('update:modelValue', false)
}
</script>

<template>
  <BottomSheet :model-value="modelValue" title="筛选" @update:model-value="$emit('update:modelValue', $event)">
    <div class="px-5 pb-5">
      <section class="pt-4">
        <h4 class="mb-3 text-[13px] font-medium text-ink">价格区间</h4>
        <div class="flex flex-wrap gap-2">
          <button
            v-for="(p, i) in PRESETS"
            :key="p.label"
            class="chip"
            :class="{ 'chip-active': priceIdx === i }"
            @click="priceIdx = i"
          >
            {{ p.label }}
          </button>
        </div>
      </section>

      <section class="mt-6">
        <h4 class="mb-3 text-[13px] font-medium text-ink">入住人数</h4>
        <div class="flex items-center justify-between rounded-card bg-fill px-4 py-3">
          <span class="text-[13px] text-ink-2">需要住下</span>
          <Stepper v-model="people" :min="1" :max="10" suffix=" 人" />
        </div>
      </section>
    </div>

    <div class="sticky bottom-0 flex gap-3 border-t border-line bg-card px-4 py-3" style="padding-bottom: calc(12px + var(--safe-bottom))">
      <button class="btn-line h-11 flex-1 text-[15px]" @click="reset">重置</button>
      <button class="btn-primary h-11 flex-[1.6] text-[15px]" @click="confirm">查看结果</button>
    </div>
  </BottomSheet>
</template>
