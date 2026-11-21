<script setup lang="ts">
import { ref } from 'vue'
import { useQuery, useQueryClient } from '@tanstack/vue-query'
import { Plus, Pencil, Trash2, ScrollText } from '@lucide/vue'
import PageToolbar from '@/components/ui/PageToolbar.vue'
import DataTable from '@/components/ui/DataTable.vue'
import ConfirmDialog from '@/components/ui/ConfirmDialog.vue'
import StatusBadge from '@/components/ui/StatusBadge.vue'
import FormDialog from '@/components/webhook/FormDialog.vue'
import { useToastStore } from '@/stores/toast'
import { listWebhookConfigs, deleteWebhookConfig } from '@/api/webhooks'
import { listApps } from '@/api/apps'
import type { WebhookConfig } from '@/types'

const toast = useToastStore()
const queryClient = useQueryClient()

const page = ref(1)
const pageSize = 20
const keyword = ref('')

const { data, isLoading } = useQuery({
  queryKey: ['webhook-configs', page, keyword],
  queryFn: () => listWebhookConfigs({ page: page.value, page_size: pageSize, key: keyword.value || undefined }),
})

const { data: appsData } = useQuery({
  queryKey: ['apps-options'],
  queryFn: () => listApps({ page: 1, page_size: 100 }),
})

const eventLabels: Record<string, string> = {
  success: '发送成功',
  failed: '发送失败',
  delivered: '已送达',
  upstream: '上行短信',
}

const formOpen = ref(false)
const editing = ref<WebhookConfig | null>(null)

function openCreate() {
  editing.value = null
  formOpen.value = true
}

function openEdit(w: WebhookConfig) {
  editing.value = w
  formOpen.value = true
}

function onSaved() {
  formOpen.value = false
  queryClient.invalidateQueries({ queryKey: ['webhook-configs'] })
}

const deleteTarget = ref<WebhookConfig | null>(null)
const deleting = ref(false)

async function confirmDelete() {
  if (!deleteTarget.value) return
  deleting.value = true
  try {
    await deleteWebhookConfig(deleteTarget.value.id)
    toast.success('已删除')
    queryClient.invalidateQueries({ queryKey: ['webhook-configs'] })
  } catch (e) {
    toast.error(e instanceof Error ? e.message : '删除失败')
  } finally {
    deleting.value = false
    deleteTarget.value = null
  }
}

function appName(id: number) {
  if (id === 0) return '全部应用'
  return appsData.value?.list?.find((a) => a.id === id)?.name ?? `#${id}`
}

function formatEvents(events: string) {
  if (!events) return '—'
  return events
    .split(',')
    .filter(Boolean)
    .map((e) => eventLabels[e] ?? e)
    .join('、')
}

const columns = [
  { key: 'id', label: 'ID', align: 'center' as const, width: '70px' },
  { key: 'name', label: '名称' },
  { key: 'app_id', label: '应用', align: 'center' as const },
  { key: 'webhook_url', label: '回调地址' },
  { key: 'events', label: '事件', align: 'center' as const },
  { key: 'retry_count', label: '重试次数', align: 'center' as const },
  { key: 'status', label: '状态', align: 'center' as const },
  { key: 'actions', label: '操作', align: 'center' as const, width: '120px' },
]
</script>

<template>
  <div>
    <PageToolbar title="Webhook 配置" description="消息状态变化时回调通知第三方系统">
      <RouterLink
        to="/logs?tab=webhook"
        class="inline-flex items-center gap-1.5 rounded-md border border-border px-3.5 py-2 text-sm text-muted-foreground transition-colors hover:bg-muted"
      >
        <ScrollText class="h-4 w-4" />
        Webhook 日志
      </RouterLink>
      <button
        class="inline-flex items-center gap-1.5 rounded-md bg-primary px-3.5 py-2 text-sm font-medium text-white transition-colors hover:bg-primary-hover"
        @click="openCreate"
      >
        <Plus class="h-4 w-4" />
        新建配置
      </button>
    </PageToolbar>

    <!-- 搜索 -->
    <div class="mb-4 flex flex-wrap items-center gap-2">
      <input
        v-model="keyword"
        type="text"
        placeholder="搜索名称 / 回调地址"
        class="h-9 w-64 rounded-md border border-input bg-card px-3 text-sm outline-none focus:border-primary focus:ring-2 focus:ring-primary/20"
        @keyup.enter="page = 1"
      />
      <button class="h-9 rounded-md border border-border px-4 text-sm text-muted-foreground hover:bg-muted" @click="page = 1">搜索</button>
    </div>

    <DataTable
      :columns="columns"
      :data="(data?.list ?? []) as unknown[]"
      :loading="isLoading"
      :total="data?.total ?? 0"
      :page="page"
      :page-size="pageSize"
      @update:page="(p: number) => (page = p)"
    >
      <template #cell-app_id="{ row }">
        <span class="text-sm">{{ appName((row as WebhookConfig).app_id) }}</span>
      </template>
      <template #cell-webhook_url="{ row }">
        <span class="line-clamp-1 max-w-xs font-mono text-xs text-muted-foreground">{{ (row as WebhookConfig).webhook_url }}</span>
      </template>
      <template #cell-events="{ row }">
        <span class="text-sm">{{ formatEvents((row as WebhookConfig).events) }}</span>
      </template>
      <template #cell-status="{ row }">
        <StatusBadge :value="(row as WebhookConfig).status === 1" />
      </template>
      <template #cell-actions="{ row }">
        <div class="flex items-center justify-center gap-1">
          <button class="rounded-md p-1.5 text-muted-foreground hover:bg-muted hover:text-foreground" @click="openEdit(row as WebhookConfig)">
            <Pencil class="h-4 w-4" />
          </button>
          <button class="rounded-md p-1.5 text-destructive/70 hover:bg-destructive/10 hover:text-destructive" @click="deleteTarget = row as WebhookConfig">
            <Trash2 class="h-4 w-4" />
          </button>
        </div>
      </template>
    </DataTable>

    <!-- 弹窗组件 -->
    <FormDialog :open="formOpen" :config="editing" @close="formOpen = false" @saved="onSaved" />

    <ConfirmDialog
      :open="!!deleteTarget"
      title="删除配置"
      :message="`确定删除 Webhook 配置「${deleteTarget?.name}」吗？`"
      danger
      :loading="deleting"
      @confirm="confirmDelete"
      @cancel="deleteTarget = null"
    />
  </div>
</template>
