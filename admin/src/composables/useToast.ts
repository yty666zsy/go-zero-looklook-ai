import { ref } from 'vue'

export interface Toast {
  id: number
  type: 'success' | 'error' | 'info'
  text: string
}

const toasts = ref<Toast[]>([])
let seq = 0

export function useToast() {
  function push(type: Toast['type'], text: string, duration = 2400) {
    const id = ++seq
    toasts.value = [...toasts.value, { id, type, text }]
    window.setTimeout(() => {
      toasts.value = toasts.value.filter((t) => t.id !== id)
    }, duration)
  }

  return {
    toasts,
    success: (t: string) => push('success', t),
    error: (t: string) => push('error', t),
    info: (t: string) => push('info', t),
    dismiss: (id: number) => {
      toasts.value = toasts.value.filter((t) => t.id !== id)
    },
  }
}
