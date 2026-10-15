<script setup lang="ts">
withDefaults(defineProps<{ score?: number; size?: number }>(), { score: 0, size: 12 })
</script>

<template>
  <span class="inline-flex items-center gap-[2px]" :style="{ height: `${size}px` }">
    <svg
      v-for="i in 5"
      :key="i"
      :width="size"
      :height="size"
      viewBox="0 0 24 24"
      :fill="i <= Math.round(score) ? '#C8A96A' : '#DFE2E6'"
    >
      <path d="M12 3.5l2.6 5.3 5.9.9-4.2 4.1 1 5.8L12 16.9 6.7 19.6l1-5.8L3.5 9.7l5.9-.9z" />
    </svg>
  </span>
</template>
