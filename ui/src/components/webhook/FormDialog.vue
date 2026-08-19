<script setup lang="ts">
// Webhook 配置创建/编辑弹窗
import { ref, watch } from 'vue'
import { useQuery } from '@tanstack/vue-query'
import Modal from '@/components/ui/Modal.vue'
import { useToastStore } from '@/stores/toast'
import { createWebhookConfig, updateWebhookConfig } from '@/api/webhooks'
import { listApps } from '@/api/apps'
import type { WebhookConfig } from '@/types'

const props = defineProps<{
  open: boolean
  config: WebhookConfig | null
}>()

const emit = defineEmits<{
  close: []
  saved: []
}>()

const toast = useToastStore()
const saving = ref(false)

const { data: appsData } = useQuery({
  queryKey: ['apps-options'],
  queryFn: () => listApps({ page: 1, page_size: 100 }),
})

const form = ref<{
  name: string
  webhook_url: string
  app_id: number
  events: string
  secret: string
  description: string
  retry_count: number
  timeout: number
  status?: number
}>({
  name: '',
  webhook_url: '',
  app_id: 0,
  events: 'success,failed',
  secret: '',
  description: '',
  retry_count: 3,
  timeout: 10,
})

watch(
  () => props.open,
  (open) => {
    if (!open) return
    if (props.config) {
      form.value = {
        name: props.config.name,
        webhook_url: props.config.webhook_url,
        app_id: props.config.app_id,
        events: props.config.events,
        secret: '',
        description: props.config.description,
        retry_count: props.config.retry_count,
        timeout: props.config.timeout,
        status: props.config.status,
      }
    } else {
      form.value = { name: '', webhook_url: '', app_id: 0, events: 'success,failed', secret: '', description: '', retry_count: 3, timeout: 10 }
    }
  }
)

async function submit() {
  if (!form.value.name || !form.value.webhook_url) {
    toast.error('请填写名称与回调地址')
    return
  }
  saving.value = true
  try {
    const payload = {
      name: form.value.name,
      webhook_url: form.value.webhook_url,
      app_id: form.value.app_id || undefined,
      events: form.value.events,
      secret: form.value.secret || undefined,
      description: form.value.description || undefined,
      retry_count: form.value.retry_count,
      timeout: form.value.timeout,
      status: form.value.status,
    }
    if (props.config) {
      await updateWebhookConfig(props.config.id, payload)
      toast.success('配置已更新')
    } else {
      await createWebhookConfig(payload)
      toast.success('配置已创建')
    }
    emit('saved')
  } catch (e) {
    toast.error(e instanceof Error ? e.message : '保存失败')
  } finally {
    saving.value = false
  }
}
</script>

<template>
  <Modal :open="open" :title="config ? '编辑 Webhook 配置' : '新建 Webhook 配置'" width="40rem" @close="emit('close')">
    <div class="space-y-4">
      <div class="grid grid-cols-2 gap-4">
        <div>
          <label class="mb-1.5 block text-sm font-medium text-slate-700">名称 *</label>
          <input v-model="form.name" type="text" class="h-9 w-full rounded-md border border-input bg-card px-3 text-sm outline-none focus:border-primary" />
        </div>
        <div>
          <label class="mb-1.5 block text-sm font-medium text-slate-700">应用（留空=全部）</label>
          <select v-model.number="form.app_id" class="h-9 w-full rounded-md border border-input bg-card px-3 text-sm outline-none focus:border-primary">
            <option :value="0">全部应用</option>
            <option v-for="a in appsData?.list ?? []" :key="a.id" :value="a.id">{{ a.name }}</option>
          </select>
        </div>
      </div>
      <div>
        <label class="mb-1.5 block text-sm font-medium text-slate-700">回调地址 *</label>
        <input v-model="form.webhook_url" type="url" placeholder="https://example.com/hook" class="h-9 w-full rounded-md border border-input bg-card px-3 text-sm outline-none focus:border-primary" />
      </div>
      <div class="grid grid-cols-2 gap-4">
        <div>
          <label class="mb-1.5 block text-sm font-medium text-slate-700">事件（逗号分隔）</label>
          <input v-model="form.events" type="text" placeholder="success,failed,delivered,upstream" class="h-9 w-full rounded-md border border-input bg-card px-3 text-sm outline-none focus:border-primary" />
        </div>
        <div>
          <label class="mb-1.5 block text-sm font-medium text-slate-700">签名密钥（可选）</label>
          <input v-model="form.secret" type="text" placeholder="用于 HMAC 验签" class="h-9 w-full rounded-md border border-input bg-card px-3 text-sm outline-none focus:border-primary" />
        </div>
      </div>
      <div class="grid grid-cols-2 gap-4">
        <div>
          <label class="mb-1.5 block text-sm font-medium text-slate-700">重试次数</label>
          <input v-model.number="form.retry_count" type="number" min="0" max="10" class="h-9 w-full rounded-md border border-input bg-card px-3 text-sm outline-none focus:border-primary" />
        </div>
        <div>
          <label class="mb-1.5 block text-sm font-medium text-slate-700">超时（秒）</label>
          <input v-model.number="form.timeout" type="number" min="1" max="60" class="h-9 w-full rounded-md border border-input bg-card px-3 text-sm outline-none focus:border-primary" />
        </div>
      </div>
      <div>
        <label class="mb-1.5 block text-sm font-medium text-slate-700">描述</label>
        <input v-model="form.description" type="text" class="h-9 w-full rounded-md border border-input bg-card px-3 text-sm outline-none focus:border-primary" />
      </div>
    </div>
    <template #footer>
      <div class="flex justify-end gap-2.5">
        <button class="h-9 rounded-md border border-border px-4 text-sm text-muted-foreground hover:bg-muted" @click="emit('close')">取消</button>
        <button class="h-9 rounded-md bg-primary px-4 text-sm font-medium text-white hover:bg-primary-hover disabled:opacity-50" :disabled="saving" @click="submit">
          {{ saving ? '保存中…' : '保存' }}
        </button>
      </div>
    </template>
  </Modal>
</template>
