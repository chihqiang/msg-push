<script setup lang="ts">
import { ref } from 'vue'
import { useQuery, useQueryClient } from '@tanstack/vue-query'
import { Plus, Pencil, Trash2, HeartPulse } from '@lucide/vue'
import PageToolbar from '@/components/ui/PageToolbar.vue'
import DataTable from '@/components/ui/DataTable.vue'
import ConfirmDialog from '@/components/ui/ConfirmDialog.vue'
import StatusBadge from '@/components/ui/StatusBadge.vue'
import FormDialog from '@/components/channel/FormDialog.vue'
import DetailDialog from '@/components/channel/DetailDialog.vue'
import { useToastStore } from '@/stores/toast'
import { listChannels, deleteChannel } from '@/api/channels'
import type { Channel } from '@/types'

const toast = useToastStore()
const queryClient = useQueryClient()

const page = ref(1)
const pageSize = 20
const keyword = ref('')

const { data, isLoading } = useQuery({
  queryKey: ['channels', page, keyword],
  queryFn: () => listChannels({ page: page.value, page_size: pageSize, key: keyword.value || undefined }),
})

// 通道类型标签
const typeLabel: Record<string, string> = {
  sms: '短信',
  email: '邮件',
  wecom: '企业微信',
  dingtalk: '钉钉',
}

const typeColor: Record<string, string> = {
  sms: 'bg-cyan-500/10 text-cyan-600',
  email: 'bg-blue-500/10 text-blue-600',
  wecom: 'bg-emerald-500/10 text-emerald-600',
  dingtalk: 'bg-violet-500/10 text-violet-600',
}

// 弹窗状态
const formOpen = ref(false)
const editing = ref<Channel | null>(null)
const detailOpen = ref(false)
const detailChannel = ref<Channel | null>(null)

const deleteTarget = ref<Channel | null>(null)
const deleting = ref(false)

function openCreate() {
  editing.value = null
  formOpen.value = true
}

function openEdit(ch: Channel) {
  editing.value = ch
  formOpen.value = true
}

function openDetail(ch: Channel) {
  detailChannel.value = ch
  detailOpen.value = true
}

function onSaved() {
  formOpen.value = false
  queryClient.invalidateQueries({ queryKey: ['channels'] })
}

async function confirmDelete() {
  if (!deleteTarget.value) return
  deleting.value = true
  try {
    await deleteChannel(deleteTarget.value.id)
    toast.success('通道已删除')
    queryClient.invalidateQueries({ queryKey: ['channels'] })
  } catch (e) {
    toast.error(e instanceof Error ? e.message : '删除失败')
  } finally {
    deleting.value = false
    deleteTarget.value = null
  }
}

const columns = [
  { key: 'id', label: 'ID', align: 'center' as const, width: '70px' },
  { key: 'code', label: '编码' },
  { key: 'name', label: '通道名称' },
  { key: 'type', label: '类型', align: 'center' as const },
  { key: 'status', label: '状态', align: 'center' as const },
  { key: 'created_at', label: '创建时间', width: '180px' },
  { key: 'actions', label: '操作', align: 'center' as const, width: '200px' },
]
</script>

<template>
  <div>
    <PageToolbar title="通道管理" description="管理发送通道（短信/邮件/企微/钉钉）">
      <button
        class="inline-flex items-center gap-1.5 rounded-md bg-primary px-3.5 py-2 text-sm font-medium text-white transition-colors hover:bg-primary-hover"
        @click="openCreate"
      >
        <Plus class="h-4 w-4" />
        新建通道
      </button>
    </PageToolbar>

    <!-- 搜索 -->
    <div class="mb-4 flex flex-wrap items-center gap-2">
      <input
        v-model="keyword"
        type="text"
        placeholder="搜索编码 / 名称"
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
      <template #cell-type="{ row }">
        <span class="rounded-full px-2 py-0.5 text-xs font-medium" :class="typeColor[(row as Channel).type] ?? 'bg-muted text-muted-foreground'">
          {{ typeLabel[(row as Channel).type] ?? (row as Channel).type }}
        </span>
      </template>
      <template #cell-status="{ row }">
        <StatusBadge :value="(row as Channel).status === 1" />
      </template>
      <template #cell-actions="{ row }">
        <div class="flex items-center justify-center gap-1">
          <button class="rounded-md p-1.5 text-muted-foreground hover:bg-muted hover:text-foreground" title="详情" @click="openDetail(row as Channel)">
            <HeartPulse class="h-4 w-4" />
          </button>
          <button class="rounded-md p-1.5 text-muted-foreground hover:bg-muted hover:text-foreground" title="编辑" @click="openEdit(row as Channel)">
            <Pencil class="h-4 w-4" />
          </button>
          <button class="rounded-md p-1.5 text-destructive/70 hover:bg-destructive/10 hover:text-destructive" title="删除" @click="deleteTarget = row as Channel">
            <Trash2 class="h-4 w-4" />
          </button>
        </div>
      </template>
    </DataTable>

    <!-- 弹窗组件 -->
    <FormDialog :open="formOpen" :channel="editing" @close="formOpen = false" @saved="onSaved" />
    <DetailDialog :open="detailOpen" :channel="detailChannel" @close="detailOpen = false" />

    <ConfirmDialog
      :open="!!deleteTarget"
      title="删除通道"
      :message="`确定删除通道「${deleteTarget?.name}」吗？`"
      danger
      :loading="deleting"
      @confirm="confirmDelete"
      @cancel="deleteTarget = null"
    />
  </div>
</template>
