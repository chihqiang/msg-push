<script setup lang="ts">
// 服务商回调地址展示弹窗：查看完整回调地址并复制到剪贴板
import { computed } from 'vue'
import Modal from '@/components/ui/Modal.vue'
import { useToastStore } from '@/stores/toast'
import type { ProviderAccount, ProviderMeta } from '@/types'

const props = defineProps<{
  open: boolean
  account: ProviderAccount | null
  providers: ProviderMeta[]
}>()

const emit = defineEmits<{
  close: []
}>()

const toast = useToastStore()

// 回调地址：VITE_API_PROXY_TARGET（后端地址）+ /api/callback/{provider_account_id}
function callbackUrl(id: number): string {
  const base = (import.meta.env.VITE_API_PROXY_TARGET || 'http://127.0.0.1:8080').replace(/\/+$/, '')
  return `${base}/api/callback/${id}`
}

const providerName = computed(() => {
  const code = props.account?.provider_code
  if (!code) return ''
  return props.providers.find((p) => p.code === code)?.name ?? code
})

function copyCallbackUrl() {
  if (!props.account) return
  navigator.clipboard.writeText(callbackUrl(props.account.id))
  toast.success('回调地址已复制')
}
</script>

<template>
  <Modal :open="open" :title="`回调地址：${account?.account_name ?? ''}`" width="36rem" @close="emit('close')">
    <div class="space-y-4">
      <div class="rounded-lg bg-primary/5 p-3 text-sm text-primary">
        将以下地址配置到服务商控制台的「回执推送 URL」，短信回执将回调到 msg-push。
      </div>
      <div>
        <label class="mb-1.5 block text-sm font-medium text-slate-700">服务商账号</label>
        <code class="block rounded-md bg-muted px-3 py-2 text-sm">
          {{ account ? `${account.account_name}（${providerName}，ID ${account.id}）` : '' }}
        </code>
      </div>
      <div>
        <label class="mb-1.5 block text-sm font-medium text-slate-700">回调地址</label>
        <div class="flex gap-2">
          <code class="flex-1 break-all rounded-md bg-muted px-3 py-2 font-mono text-xs">{{ account ? callbackUrl(account.id) : '' }}</code>
          <button class="h-9 shrink-0 rounded-md border border-border px-3 text-sm text-muted-foreground hover:bg-muted" @click="copyCallbackUrl">复制</button>
        </div>
      </div>
    </div>
    <template #footer>
      <div class="flex justify-end">
        <button class="h-9 rounded-md bg-primary px-4 text-sm font-medium text-white hover:bg-primary-hover" @click="emit('close')">关闭</button>
      </div>
    </template>
  </Modal>
</template>
