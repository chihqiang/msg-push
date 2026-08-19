<script setup lang="ts">
// 批量任务下钻弹窗：展示某批次的全部推送任务
import { useQuery } from '@tanstack/vue-query'
import Modal from '@/components/ui/Modal.vue'
import DataTable from '@/components/ui/DataTable.vue'
import { tasksByBatch } from '@/api/tasks'
import type { PushBatchTask, PushTask } from '@/types'

const props = defineProps<{
  open: boolean
  batch: PushBatchTask | null
}>()

const emit = defineEmits<{
  close: []
}>()

const { data } = useQuery({
  queryKey: ['batch-drill', () => props.batch?.batch_id],
  queryFn: () => tasksByBatch(props.batch!.batch_id, { page: 1, page_size: 50 }),
  enabled: () => !!props.open && !!props.batch,
})

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

const columns = [
  { key: 'task_id', label: '任务ID' },
  { key: 'receiver', label: '接收方' },
  { key: 'status', label: '状态', align: 'center' as const },
  { key: 'error_msg', label: '错误信息' },
]
</script>

<template>
  <Modal :open="open" :title="`批次任务：${batch?.batch_id ?? ''}`" width="48rem" @close="emit('close')">
    <DataTable
      :columns="columns"
      :data="(data?.list ?? []) as unknown[]"
      :loading="false"
      :total="data?.total ?? 0"
      empty-text="该批次暂无任务"
    >
      <template #cell-task_id="{ row }">
        <code class="rounded bg-muted px-1.5 py-0.5 font-mono text-xs">{{ (row as PushTask).task_id }}</code>
      </template>
      <template #cell-status="{ row }">
        <span class="rounded-full px-2 py-0.5 text-xs font-medium" :class="statusColor[(row as PushTask).status] ?? 'bg-muted text-muted-foreground'">
          {{ statusLabel[(row as PushTask).status] ?? (row as PushTask).status }}
        </span>
      </template>
    </DataTable>
  </Modal>
</template>
