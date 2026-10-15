<script setup lang="ts">
withDefaults(defineProps<{ count?: number; layout?: 'row' | 'grid' }>(), { count: 4, layout: 'row' })
</script>

<template>
  <div :class="layout === 'grid' ? 'grid grid-cols-2 gap-3' : 'flex flex-col gap-3'">
    <div v-for="i in count" :key="i" class="card overflow-hidden">
      <div class="skeleton h-0 w-full" style="padding-bottom: 62%" />
      <div class="space-y-2 p-3">
        <div class="skeleton h-4 w-3/4" />
        <div class="skeleton h-3 w-1/2" />
        <div class="flex justify-between pt-1">
          <div class="skeleton h-3 w-16" />
          <div class="skeleton h-4 w-14" />
        </div>
      </div>
    </div>
  </div>
</template>
