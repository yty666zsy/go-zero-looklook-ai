import { defineStore } from 'pinia'
import { computed, ref } from 'vue'
import { todayStart } from '@/utils/format'

const DAY = 86400

/** 搜索条件在首页/列表页之间共享 */
export const useSearchStore = defineStore('search', () => {
  const keyword = ref('')
  const city = ref('上海')
  const startTs = ref(todayStart() + DAY)
  const endTs = ref(todayStart() + DAY * 2)
  const peopleNum = ref(2)
  const minPrice = ref(0)
  const maxPrice = ref(0)
  const sortBy = ref('default')

  const nights = computed(() => Math.max(1, Math.round((endTs.value - startTs.value) / DAY)))

  function setRange(start: number, end: number) {
    startTs.value = start
    endTs.value = end
  }

  function reset() {
    keyword.value = ''
    minPrice.value = 0
    maxPrice.value = 0
    peopleNum.value = 2
    sortBy.value = 'default'
    startTs.value = todayStart() + DAY
    endTs.value = todayStart() + DAY * 2
  }

  return {
    keyword,
    city,
    startTs,
    endTs,
    peopleNum,
    minPrice,
    maxPrice,
    sortBy,
    nights,
    setRange,
    reset,
  }
})
