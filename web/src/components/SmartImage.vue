<script setup lang="ts">
import { computed, ref, watch } from 'vue'

const props = withDefaults(
  defineProps<{
    src?: string
    alt?: string
    ratio?: number
    rounded?: string
  }>(),
  { src: '', alt: '', ratio: 0.66, rounded: 'inherit' },
)

const loaded = ref(false)
const failed = ref(false)

watch(
  () => props.src,
  () => {
    loaded.value = false
    failed.value = false
  },
)

const style = computed(() => ({
  paddingBottom: `${props.ratio * 100}%`,
  borderRadius: props.rounded,
}))
</script>

<template>
  <div class="relative w-full overflow-hidden bg-fill" :style="style">
    <!-- 载入中的骨架 -->
    <div v-if="!loaded && !failed" class="skeleton absolute inset-0" />

    <!-- 失败兜底：一块安静的灰色渐变，不用花哨图片 -->
    <div
      v-if="failed"
      class="absolute inset-0 flex items-center justify-center text-ink-4 text-[12px] tracking-widest"
      style="background: linear-gradient(135deg, #eef0f2 0%, #e2e5e9 100%)"
    >
      暂无图片
    </div>

    <img
      v-if="src && !failed"
      :src="src"
      :alt="alt"
      loading="lazy"
      decoding="async"
      class="absolute inset-0 h-full w-full object-cover transition-all duration-500 ease-out"
      :class="loaded ? 'opacity-100 scale-100' : 'opacity-0 scale-[1.03]'"
      @load="loaded = true"
      @error="failed = true"
    />
  </div>
</template>
