// Toast 通知状态
import { defineStore } from 'pinia'
import { ref } from 'vue'

export interface ToastItem {
  id: number
  type: 'success' | 'error' | 'info'
  message: string
}

let nextId = 1

export const useToastStore = defineStore('toast', () => {
  const toasts = ref<ToastItem[]>([])

  function push(type: ToastItem['type'], message: string) {
    const id = nextId++
    toasts.value.push({ id, type, message })
    // 3 秒后自动移除
    setTimeout(() => remove(id), 3000)
  }

  function remove(id: number) {
    toasts.value = toasts.value.filter((t) => t.id !== id)
  }

  function success(message: string) {
    push('success', message)
  }
  function error(message: string) {
    push('error', message)
  }
  function info(message: string) {
    push('info', message)
  }

  return { toasts, push, remove, success, error, info }
})
