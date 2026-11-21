<script setup lang="ts">
import { computed, ref } from 'vue'
import { useQuery } from '@tanstack/vue-query'
import PageToolbar from '@/components/ui/PageToolbar.vue'
import DataTable from '@/components/ui/DataTable.vue'
import BatchDrillDialog from '@/components/task/BatchDrillDialog.vue'
import { listTasks, listBatchTasks } from '@/api/tasks'
import type { PushTask, PushBatchTask } from '@/types'

const tab = ref<'tasks' | 'batches'>('tasks')

const page = ref(1)
const pageSize = 20
const status = ref('')
const taskNo = ref('')
const requestId = ref('')

// 状态选项：任务与批量语义不同，按当前 Tab 切换
const statusOptions = computed<{ value: string; label: string }[]>(() =>
  tab.value === 'batches'
    ? [
        { value: 'processing', label: '处理中' },
        { value: 'completed', label: '已完成' },
        { value: 'failed', label: '失败' },
      ]
    : [
        { value: 'pending', label: '待发送' },
        { value: 'sending', label: '发送中' },
        { value: 'success', label: '成功' },
        { value: 'failed', label: '失败' },
      ]
)

// 切换 Tab：重置页码与状态过滤，避免共用 page/status 造成残留
function switchTab(t: 'tasks' | 'batches') {
  if (tab.value === t) return
  tab.value = t
  status.value = ''
  page.value = 1
}

const { data: tasksData, isLoading: tasksLoading } = useQuery({
  queryKey: ['tasks', page, status, taskNo, requestId],
  queryFn: () =>
    listTasks({
      page: page.value,
      page_size: pageSize,
      status: status.value || undefined,
      task_no: taskNo.value || undefined,
      request_id: requestId.value || undefined,
    }),
})

const { data: batchesData, isLoading: batchesLoading } = useQuery({
  queryKey: ['batch-tasks', page, status],
  queryFn: () =>
    listBatchTasks({
      page: page.value,
      page_size: pageSize,
      status: status.value || undefined,
    }),
})

// 批次下钻
const drillOpen = ref(false)
const drillBatch = ref<PushBatchTask | null>(null)

function openDrill(b: PushBatchTask) {
  drillBatch.value = b
  drillOpen.value = true
}

const statusLabel: Record<string, string> = {
  pending: '待发送',
  sending: '发送中',
  success: '成功',
  failed: '失败',
}

const statusColor: Record<string, string> = {
  pending: 'bg-slate-500/10 text-slate-600',
  sending: 'bg-amber-500/10 text-amber-600',
  success: 'bg-emerald-500/10 text-emerald-600',
  failed: 'bg-destructive/10 text-destructive',
}

const batchStatusLabel: Record<string, string> = {
  processing: '处理中',
  completed: '已完成',
  failed: '失败',
}

const taskColumns = [
  { key: 'id', label: 'ID', align: 'center' as const, width: '70px' },
  { key: 'task_id', label: '任务ID', width: '170px' },
  { key: 'request_id', label: '追踪ID', width: '170px' },
  { key: 'receiver', label: '接收方' },
  { key: 'channel_name', label: '通道' },
  { key: 'is_test', label: '模式', align: 'center' as const, width: '80px' },
  { key: 'status', label: '状态', align: 'center' as const },
  { key: 'created_at', label: '创建时间', width: '180px' },
]

const batchColumns = [
  { key: 'id', label: 'ID', align: 'center' as const, width: '70px' },
  { key: 'batch_id', label: '批次ID', width: '180px' },
  { key: 'total_count', label: '总数', align: 'center' as const },
  { key: 'success_count', label: '成功', align: 'center' as const },
  { key: 'failed_count', label: '失败', align: 'center' as const },
  { key: 'pending_count', label: '待处理', align: 'center' as const },
  { key: 'completion_rate', label: '完成率', align: 'center' as const },
  { key: 'status', label: '状态', align: 'center' as const },
  { key: 'created_at', label: '创建时间', width: '180px' },
  { key: 'actions', label: '操作', align: 'center' as const, width: '80px' },
]
</script>

<template>
  <div>
    <PageToolbar title="任务查询" description="查看推送任务与批量任务" />

    <!-- 搜索 -->
    <div class="mb-4 flex flex-wrap items-center gap-2">
      <!-- 任务 Tab：按任务编号/追踪ID 搜索；批量 Tab 不支持这两项，隐藏避免误导 -->
      <template v-if="tab === 'tasks'">
        <input v-model="taskNo" type="text" placeholder="任务编号" class="h-9 w-48 rounded-md border border-input bg-card px-3 text-sm outline-none focus:border-primary focus:ring-2 focus:ring-primary/20" />
        <input v-model="requestId" type="text" placeholder="追踪ID (request_id)" class="h-9 w-56 rounded-md border border-input bg-card px-3 text-sm outline-none focus:border-primary focus:ring-2 focus:ring-primary/20" />
      </template>
      <button class="h-9 rounded-md border border-border px-4 text-sm text-muted-foreground hover:bg-muted" @click="page = 1">查询</button>
    </div>

    <!-- Tabs -->
    <div class="mb-4 flex gap-1 border-b border-border">
      <button
        class="border-b-2 px-3 py-2 text-sm transition-colors"
        :class="tab === 'tasks' ? 'border-primary font-medium text-primary' : 'border-transparent text-muted-foreground hover:text-foreground'"
        @click="switchTab('tasks')"
      >
        推送任务
      </button>
      <button
        class="border-b-2 px-3 py-2 text-sm transition-colors"
        :class="tab === 'batches' ? 'border-primary font-medium text-primary' : 'border-transparent text-muted-foreground hover:text-foreground'"
        @click="switchTab('batches')"
      >
        批量任务
      </button>
    </div>

    <!-- 状态筛选 -->
    <div class="mb-4 flex items-center gap-2">
      <select v-model="status" class="h-9 rounded-md border border-input bg-card px-3 text-sm outline-none focus:border-primary">
        <option value="">全部状态</option>
        <option v-for="opt in statusOptions" :key="opt.value" :value="opt.value">{{ opt.label }}</option>
      </select>
    </div>

    <!-- 任务列表 -->
    <DataTable
      v-if="tab === 'tasks'"
      :columns="taskColumns"
      :data="(tasksData?.list ?? []) as unknown[]"
      :loading="tasksLoading"
      :total="tasksData?.total ?? 0"
      :page="page"
      :page-size="pageSize"
      @update:page="(p: number) => (page = p)"
    >
      <template #cell-task_id="{ row }">
        <code class="rounded bg-muted px-1.5 py-0.5 font-mono text-xs">{{ (row as PushTask).task_id }}</code>
      </template>
      <template #cell-request_id="{ row }">
        <code class="rounded bg-muted px-1.5 py-0.5 font-mono text-xs">{{ (row as PushTask).request_id || '—' }}</code>
      </template>
      <template #cell-is_test="{ row }">
        <span v-if="(row as PushTask).is_test" class="rounded-full bg-amber-500/10 px-2 py-0.5 text-xs font-medium text-amber-600">测试</span>
        <span v-else class="rounded-full bg-emerald-500/10 px-2 py-0.5 text-xs font-medium text-emerald-600">正式</span>
      </template>
      <template #cell-channel_name="{ row }">
        <span class="text-sm">{{ (row as PushTask).channel_name ?? '—' }}</span>
      </template>
      <template #cell-status="{ row }">
        <span class="rounded-full px-2 py-0.5 text-xs font-medium" :class="statusColor[(row as PushTask).status] ?? 'bg-muted text-muted-foreground'">
          {{ statusLabel[(row as PushTask).status] ?? (row as PushTask).status }}
        </span>
      </template>
    </DataTable>

    <!-- 批量任务列表 -->
    <DataTable
      v-else
      :columns="batchColumns"
      :data="(batchesData?.list ?? []) as unknown[]"
      :loading="batchesLoading"
      :total="batchesData?.total ?? 0"
      :page="page"
      :page-size="pageSize"
      @update:page="(p: number) => (page = p)"
    >
      <template #cell-batch_id="{ row }">
        <code class="rounded bg-muted px-1.5 py-0.5 font-mono text-xs">{{ (row as PushBatchTask).batch_id }}</code>
      </template>
      <template #cell-completion_rate="{ row }">
        <span class="text-sm">{{ (row as PushBatchTask).completion_rate.toFixed(1) }}%</span>
      </template>
      <template #cell-status="{ row }">
        <span class="rounded-full px-2 py-0.5 text-xs font-medium" :class="(row as PushBatchTask).status === 'completed' ? 'bg-emerald-500/10 text-emerald-600' : (row as PushBatchTask).status === 'failed' ? 'bg-destructive/10 text-destructive' : 'bg-amber-500/10 text-amber-600'">
          {{ batchStatusLabel[(row as PushBatchTask).status] ?? (row as PushBatchTask).status }}
        </span>
      </template>
      <template #cell-actions="{ row }">
        <button class="rounded-md px-2 py-1 text-xs text-primary hover:bg-primary/10" @click="openDrill(row as PushBatchTask)">下钻</button>
      </template>
    </DataTable>

    <!-- 批次下钻弹窗组件 -->
    <BatchDrillDialog :open="drillOpen" :batch="drillBatch" @close="drillOpen = false" />
  </div>
</template>
