<script setup lang="ts">
// 应用密钥展示弹窗（创建/重置后仅此一次可见）
import Modal from '@/components/ui/Modal.vue'
import { useToastStore } from '@/stores/toast'

const props = defineProps<{
  open: boolean
  appId: string
  secret: string
}>()

const emit = defineEmits<{
  close: []
}>()

const toast = useToastStore()

function copySecret() {
  if (!props.secret) return
  navigator.clipboard.writeText(props.secret)
  toast.success('已复制到剪贴板')
}
</script>

<template>
  <Modal :open="open" title="应用密钥" @close="emit('close')">
    <div class="space-y-4">
      <div class="rounded-lg bg-amber-50 p-3 text-sm text-amber-700">
        请立即保存密钥，关闭后不再显示（仅创建/重置时返回一次）。
      </div>
      <div>
        <label class="mb-1.5 block text-sm font-medium text-slate-700">AppID</label>
        <code class="block rounded-md bg-muted px-3 py-2 font-mono text-sm">{{ appId }}</code>
      </div>
      <div>
        <label class="mb-1.5 block text-sm font-medium text-slate-700">AppSecret</label>
        <div class="flex gap-2">
          <code class="flex-1 break-all rounded-md bg-muted px-3 py-2 font-mono text-xs">{{ secret }}</code>
          <button class="h-9 shrink-0 rounded-md border border-border px-3 text-sm text-muted-foreground hover:bg-muted" @click="copySecret">复制</button>
        </div>
      </div>
    </div>
    <template #footer>
      <div class="flex justify-end">
        <button class="h-9 rounded-md bg-primary px-4 text-sm font-medium text-white hover:bg-primary-hover" @click="emit('close')">我已保存</button>
      </div>
    </template>
  </Modal>
</template>
