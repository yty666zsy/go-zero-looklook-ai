import { ref } from 'vue'

export interface ToastItem {
  id: number
  text: string
  type: 'info' | 'success' | 'error'
}

const items = ref<ToastItem[]>([])
let seed = 0

function push(text: string, type: ToastItem['type'] = 'info', duration = 2000) {
  const id = ++seed
  items.value.push({ id, text, type })
  window.setTimeout(() => {
    items.value = items.value.filter((i) => i.id !== id)
  }, duration)
}

export function useToast() {
  return {
    items,
    info: (t: string) => push(t, 'info'),
    success: (t: string) => push(t, 'success'),
    error: (t: string) => push(t, 'error', 2600),
  }
}
