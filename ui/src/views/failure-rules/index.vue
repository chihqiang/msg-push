<script setup lang="ts">
import { ref } from 'vue'
import { useQuery, useQueryClient } from '@tanstack/vue-query'
import { Plus, Pencil, Trash2, RefreshCw } from '@lucide/vue'
import PageToolbar from '@/components/ui/PageToolbar.vue'
import DataTable from '@/components/ui/DataTable.vue'
import ConfirmDialog from '@/components/ui/ConfirmDialog.vue'
import StatusBadge from '@/components/ui/StatusBadge.vue'
import FormDialog from '@/components/failure-rule/FormDialog.vue'
import { useToastStore } from '@/stores/toast'
import { listFailureRules, deleteFailureRule, refreshFailureRuleCache } from '@/api/failureRules'
import type { FailureRule } from '@/types'

const toast = useToastStore()
const queryClient = useQueryClient()

const page = ref(1)
const pageSize = 20
const keyword = ref('')

const { data, isLoading } = useQuery({
  queryKey: ['failure-rules', page, keyword],
  queryFn: () => listFailureRules({ page: page.value, page_size: pageSize, key: keyword.value || undefined }),
})

const sceneLabel: Record<string, string> = {
  send_failure: '发送失败',
  callback_failure: '回调失败',
}

const actionLabel: Record<string, string> = {
  retry: '重试',
  switch_provider: '切换供应商',
  fail: '直接失败',
  alert: '告警',
}

const actionColor: Record<string, string> = {
  retry: 'bg-cyan-500/10 text-cyan-600',
  switch_provider: 'bg-violet-500/10 text-violet-600',
  fail: 'bg-destructive/10 text-destructive',
  alert: 'bg-amber-500/10 text-amber-600',
}

const formOpen = ref(false)
const editing = ref<FailureRule | null>(null)

function openCreate() {
  editing.value = null
  formOpen.value = true
}

function openEdit(rule: FailureRule) {
  editing.value = rule
  formOpen.value = true
}

function onSaved() {
  formOpen.value = false
  queryClient.invalidateQueries({ queryKey: ['failure-rules'] })
}

const deleteTarget = ref<FailureRule | null>(null)
const deleting = ref(false)

async function confirmDelete() {
  if (!deleteTarget.value) return
  deleting.value = true
  try {
    await deleteFailureRule(deleteTarget.value.id)
    toast.success('已删除')
    queryClient.invalidateQueries({ queryKey: ['failure-rules'] })
  } catch (e) {
    toast.error(e instanceof Error ? e.message : '删除失败')
  } finally {
    deleting.value = false
    deleteTarget.value = null
  }
}

async function onRefreshCache() {
  try {
    await refreshFailureRuleCache()
    toast.success('缓存已刷新')
  } catch (e) {
    toast.error(e instanceof Error ? e.message : '刷新失败')
  }
}

const columns = [
  { key: 'id', label: 'ID', align: 'center' as const, width: '70px' },
  { key: 'name', label: '规则名称' },
  { key: 'scene', label: '场景', align: 'center' as const },
  { key: 'action', label: '动作', align: 'center' as const },
  { key: 'provider_code', label: '服务商', align: 'center' as const },
  { key: 'message_type', label: '消息类型', align: 'center' as const },
  { key: 'priority', label: '优先级', align: 'center' as const },
  { key: 'status', label: '状态', align: 'center' as const },
  { key: 'actions', label: '操作', align: 'center' as const, width: '120px' },
]
</script>

<template>
  <div>
    <PageToolbar title="失败规则" description="发送/回调失败时的处理策略（重试/切供应商/告警/失败）">
      <button
        class="inline-flex items-center gap-1.5 rounded-md border border-border px-3.5 py-2 text-sm text-muted-foreground transition-colors hover:bg-muted"
        @click="onRefreshCache"
      >
        <RefreshCw class="h-4 w-4" />
        刷新缓存
      </button>
      <button
        class="inline-flex items-center gap-1.5 rounded-md bg-primary px-3.5 py-2 text-sm font-medium text-white transition-colors hover:bg-primary-hover"
        @click="openCreate"
      >
        <Plus class="h-4 w-4" />
        新建规则
      </button>
    </PageToolbar>

    <!-- 搜索 -->
    <div class="mb-4 flex flex-wrap items-center gap-2">
      <input
        v-model="keyword"
        type="text"
        placeholder="搜索规则名称"
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
      <template #cell-scene="{ row }">
        <span class="text-sm">{{ sceneLabel[(row as FailureRule).scene] ?? (row as FailureRule).scene }}</span>
      </template>
      <template #cell-action="{ row }">
        <span class="rounded-full px-2 py-0.5 text-xs font-medium" :class="actionColor[(row as FailureRule).action] ?? 'bg-muted text-muted-foreground'">
          {{ actionLabel[(row as FailureRule).action] ?? (row as FailureRule).action }}
        </span>
      </template>
      <template #cell-provider_code="{ row }">
        <span class="text-sm">{{ (row as FailureRule).provider_code || '全部' }}</span>
      </template>
      <template #cell-message_type="{ row }">
        <span class="text-sm">{{ (row as FailureRule).message_type || '全部' }}</span>
      </template>
      <template #cell-status="{ row }">
        <StatusBadge :value="(row as FailureRule).status === 1" />
      </template>
      <template #cell-actions="{ row }">
        <div class="flex items-center justify-center gap-1">
          <button class="rounded-md p-1.5 text-muted-foreground hover:bg-muted hover:text-foreground" @click="openEdit(row as FailureRule)">
            <Pencil class="h-4 w-4" />
          </button>
          <button class="rounded-md p-1.5 text-destructive/70 hover:bg-destructive/10 hover:text-destructive" @click="deleteTarget = row as FailureRule">
            <Trash2 class="h-4 w-4" />
          </button>
        </div>
      </template>
    </DataTable>

    <!-- 弹窗组件 -->
    <FormDialog :open="formOpen" :rule="editing" @close="formOpen = false" @saved="onSaved" />

    <ConfirmDialog
      :open="!!deleteTarget"
      title="删除规则"
      :message="`确定删除规则「${deleteTarget?.name}」吗？`"
      danger
      :loading="deleting"
      @confirm="confirmDelete"
      @cancel="deleteTarget = null"
    />
  </div>
</template>
