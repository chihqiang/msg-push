<script setup lang="ts">
import { ref } from 'vue'
import { useQuery } from '@tanstack/vue-query'
import PageToolbar from '@/components/ui/PageToolbar.vue'
import DataTable from '@/components/ui/DataTable.vue'
import { listPushLogs, listCallbacks } from '@/api/logs'
import { listWebhookLogs, listWebhookConfigs } from '@/api/webhooks'
import { listProviderAccounts } from '@/api/providers'
import type { PushLog, CallbackLog, WebhookLog } from '@/types'

type LogTab = 'push' | 'callback' | 'webhook'

const tab = ref<LogTab>('push')
const page = ref(1)
const pageSize = 20
const taskNo = ref('')
const requestId = ref('')
const status = ref('')
const startDate = ref('')
const endDate = ref('')

// 各 Tab 的状态选项
const statusOptions: Record<LogTab, { value: string; label: string; color: string }[]> = {
  push: [
    { value: 'pending', label: '待发送', color: 'bg-slate-500/10 text-slate-600' },
    { value: 'sending', label: '发送中', color: 'bg-amber-500/10 text-amber-600' },
    { value: 'success', label: '成功', color: 'bg-emerald-500/10 text-emerald-600' },
    { value: 'failed', label: '失败', color: 'bg-destructive/10 text-destructive' },
  ],
  callback: [
    { value: 'delivered', label: '已送达', color: 'bg-emerald-500/10 text-emerald-600' },
    { value: 'failed', label: '失败', color: 'bg-destructive/10 text-destructive' },
    { value: 'rejected', label: '被拒绝', color: 'bg-amber-500/10 text-amber-600' },
    { value: 'unknown', label: '未知', color: 'bg-slate-500/10 text-slate-600' },
  ],
  webhook: [
    { value: 'pending', label: '待投递', color: 'bg-slate-500/10 text-slate-600' },
    { value: 'success', label: '成功', color: 'bg-emerald-500/10 text-emerald-600' },
    { value: 'failed', label: '失败', color: 'bg-destructive/10 text-destructive' },
  ],
}

function statusMeta(t: LogTab, s: string) {
  return statusOptions[t].find((o) => o.value === s)
}

const commonParams = () => ({
  page: page.value,
  page_size: pageSize,
  task_no: taskNo.value || undefined,
  request_id: requestId.value || undefined,
  status: status.value || undefined,
  start_date: startDate.value || undefined,
  end_date: endDate.value || undefined,
})

const { data: pushData, isLoading: pushLoading } = useQuery({
  queryKey: ['push-logs', page, taskNo, requestId, status, startDate, endDate],
  queryFn: () => listPushLogs(commonParams()),
  enabled: () => tab.value === 'push', // 仅当前 Tab 生效，避免三个 Tab 同时请求
})

const { data: callbackData, isLoading: callbackLoading } = useQuery({
  queryKey: ['callback-logs', page, taskNo, requestId, status, startDate, endDate],
  queryFn: () => listCallbacks(commonParams()),
  enabled: () => tab.value === 'callback',
})

const { data: webhookData, isLoading: webhookLoading } = useQuery({
  queryKey: ['webhook-logs', page, taskNo, requestId, status, startDate, endDate],
  queryFn: () => listWebhookLogs(commonParams()),
  enabled: () => tab.value === 'webhook',
})

// 关联名称映射：服务商账号 / Webhook 配置
const { data: providerAccounts } = useQuery({
  queryKey: ['log-provider-accounts'],
  queryFn: () => listProviderAccounts({ page: 1, page_size: 100 }),
})

const { data: webhookConfigs } = useQuery({
  queryKey: ['log-webhook-configs'],
  queryFn: () => listWebhookConfigs({ page: 1, page_size: 100 }),
})

function providerAccountName(id: number) {
  return providerAccounts.value?.list?.find((a) => a.id === id)?.account_name ?? `#${id}`
}

function webhookConfigName(id: number) {
  return webhookConfigs.value?.list?.find((c) => c.id === id)?.name ?? `#${id}`
}

function switchTab(t: LogTab) {
  tab.value = t
  status.value = ''
  page.value = 1
}

function search() {
  page.value = 1
}

const pushColumns = [
  { key: 'id', label: 'ID', align: 'center' as const, width: '70px' },
  { key: 'request_id', label: '追踪ID', width: '170px' },
  { key: 'task_no', label: '任务号', width: '170px' },
  { key: 'receiver', label: '接收方' },
  { key: 'provider_account_id', label: '服务商账号', align: 'center' as const },
  { key: 'status', label: '状态', align: 'center' as const },
  { key: 'error_code', label: '错误码' },
  { key: 'cost_time', label: '耗时(ms)', align: 'right' as const },
  { key: 'created_at', label: '时间', width: '180px' },
]

const callbackColumns = [
  { key: 'id', label: 'ID', align: 'center' as const, width: '70px' },
  { key: 'type', label: '类型', align: 'center' as const, width: '80px' },
  { key: 'request_id', label: '追踪ID', width: '170px' },
  { key: 'task_no', label: '任务号', width: '170px' },
  { key: 'provider_code', label: '服务商' },
  { key: 'mobile', label: '手机号' },
  { key: 'callback_status', label: '状态', align: 'center' as const },
  { key: 'error_code', label: '错误码' },
  { key: 'created_at', label: '时间', width: '180px' },
]

const webhookColumns = [
  { key: 'id', label: 'ID', align: 'center' as const, width: '70px' },
  { key: 'request_id', label: '追踪ID', width: '170px' },
  { key: 'task_no', label: '任务号', width: '170px' },
  { key: 'event', label: '事件', align: 'center' as const, width: '90px' },
  { key: 'webhook_config_id', label: '配置', align: 'center' as const },
  { key: 'status', label: '状态', align: 'center' as const },
  { key: 'response_status', label: 'HTTP状态', align: 'center' as const },
  { key: 'retry_count', label: '重试', align: 'center' as const },
  { key: 'created_at', label: '时间', width: '180px' },
]
</script>

<template>
  <div>
    <PageToolbar title="日志查询" description="推送日志 / 服务商回调日志 / Webhook 投递日志" />

    <!-- 搜索 -->
    <div class="mb-4 flex flex-wrap items-center gap-2">
      <input v-model="taskNo" type="text" placeholder="任务编号" class="h-9 w-48 rounded-md border border-input bg-card px-3 text-sm outline-none focus:border-primary focus:ring-2 focus:ring-primary/20" />
      <input v-model="requestId" type="text" placeholder="追踪ID (request_id)" class="h-9 w-56 rounded-md border border-input bg-card px-3 text-sm outline-none focus:border-primary focus:ring-2 focus:ring-primary/20" />
      <input v-model="startDate" type="date" class="h-9 rounded-md border border-input bg-card px-3 text-sm outline-none focus:border-primary focus:ring-2 focus:ring-primary/20" />
      <input v-model="endDate" type="date" class="h-9 rounded-md border border-input bg-card px-3 text-sm outline-none focus:border-primary focus:ring-2 focus:ring-primary/20" />
      <button class="h-9 rounded-md border border-border px-4 text-sm text-muted-foreground hover:bg-muted" @click="search">查询</button>
    </div>

    <!-- Tabs -->
    <div class="mb-4 flex gap-1 border-b border-border">
      <button
        v-for="t in (['push', 'callback', 'webhook'] as LogTab[])"
        :key="t"
        class="border-b-2 px-3 py-2 text-sm transition-colors"
        :class="tab === t ? 'border-primary font-medium text-primary' : 'border-transparent text-muted-foreground hover:text-foreground'"
        @click="switchTab(t)"
      >
        {{ t === 'push' ? '推送日志' : t === 'callback' ? '回调日志' : 'Webhook 日志' }}
      </button>
    </div>

    <!-- 状态筛选 -->
    <div class="mb-4 flex items-center gap-2">
      <select v-model="status" class="h-9 rounded-md border border-input bg-card px-3 text-sm outline-none focus:border-primary" @change="page = 1">
        <option value="">全部状态</option>
        <option v-for="opt in statusOptions[tab]" :key="opt.value" :value="opt.value">{{ opt.label }}</option>
      </select>
    </div>

    <!-- 推送日志 -->
    <DataTable
      v-if="tab === 'push'"
      :columns="pushColumns"
      :data="(pushData?.list ?? []) as unknown[]"
      :loading="pushLoading"
      :total="pushData?.total ?? 0"
      :page="page"
      :page-size="pageSize"
      @update:page="(p: number) => (page = p)"
    >
      <template #cell-request_id="{ row }">
        <code class="rounded bg-muted px-1.5 py-0.5 font-mono text-xs">{{ (row as PushLog).request_id || '—' }}</code>
      </template>
      <template #cell-task_no="{ row }">
        <code class="rounded bg-muted px-1.5 py-0.5 font-mono text-xs">{{ (row as PushLog).task_no }}</code>
      </template>
      <template #cell-provider_account_id="{ row }">
        <span class="text-sm">{{ providerAccountName((row as PushLog).provider_account_id) }}</span>
      </template>
      <template #cell-status="{ row }">
        <span class="rounded-full px-2 py-0.5 text-xs font-medium" :class="statusMeta('push', (row as PushLog).status)?.color ?? 'bg-muted text-muted-foreground'">
          {{ statusMeta('push', (row as PushLog).status)?.label ?? (row as PushLog).status }}
        </span>
      </template>
    </DataTable>

    <!-- 回调日志 -->
    <DataTable
      v-else-if="tab === 'callback'"
      :columns="callbackColumns"
      :data="(callbackData?.list ?? []) as unknown[]"
      :loading="callbackLoading"
      :total="callbackData?.total ?? 0"
      :page="page"
      :page-size="pageSize"
      @update:page="(p: number) => (page = p)"
    >
      <template #cell-request_id="{ row }">
        <code class="rounded bg-muted px-1.5 py-0.5 font-mono text-xs">{{ (row as CallbackLog).request_id || '—' }}</code>
      </template>
      <template #cell-type="{ row }">
        <span class="rounded-full px-2 py-0.5 text-xs font-medium" :class="(row as CallbackLog).type === 'upstream' ? 'bg-sky-500/10 text-sky-600' : 'bg-violet-500/10 text-violet-600'">
          {{ (row as CallbackLog).type === 'upstream' ? '上行' : '回执' }}
        </span>
      </template>
      <template #cell-callback_status="{ row }">
        <span class="rounded-full px-2 py-0.5 text-xs font-medium" :class="statusMeta('callback', (row as CallbackLog).callback_status)?.color ?? 'bg-muted text-muted-foreground'">
          {{ statusMeta('callback', (row as CallbackLog).callback_status)?.label ?? (row as CallbackLog).callback_status }}
        </span>
      </template>
    </DataTable>

    <!-- Webhook 日志 -->
    <DataTable
      v-else
      :columns="webhookColumns"
      :data="(webhookData?.list ?? []) as unknown[]"
      :loading="webhookLoading"
      :total="webhookData?.total ?? 0"
      :page="page"
      :page-size="pageSize"
      @update:page="(p: number) => (page = p)"
    >
      <template #cell-request_id="{ row }">
        <code class="rounded bg-muted px-1.5 py-0.5 font-mono text-xs">{{ (row as WebhookLog).request_id || '—' }}</code>
      </template>
      <template #cell-task_no="{ row }">
        <code class="rounded bg-muted px-1.5 py-0.5 font-mono text-xs">{{ (row as WebhookLog).task_no }}</code>
      </template>
      <template #cell-event="{ row }">
        <span class="rounded bg-muted px-1.5 py-0.5 font-mono text-xs">{{ (row as WebhookLog).event }}</span>
      </template>
      <template #cell-webhook_config_id="{ row }">
        <span class="text-sm">{{ webhookConfigName((row as WebhookLog).webhook_config_id) }}</span>
      </template>
      <template #cell-status="{ row }">
        <span class="rounded-full px-2 py-0.5 text-xs font-medium" :class="statusMeta('webhook', (row as WebhookLog).status)?.color ?? 'bg-muted text-muted-foreground'">
          {{ statusMeta('webhook', (row as WebhookLog).status)?.label ?? (row as WebhookLog).status }}
        </span>
      </template>
      <template #cell-response_status="{ row }">
        <span class="text-sm">{{ (row as WebhookLog).response_status || '—' }}</span>
      </template>
      <template #cell-retry_count="{ row }">
        <span class="text-sm">{{ (row as WebhookLog).retry_count }}/{{ (row as WebhookLog).max_retries }}</span>
      </template>
    </DataTable>
  </div>
</template>
