<script setup lang="ts">
import { onMounted, ref, watch } from 'vue'
import { useRoute } from 'vue-router'
import { useQuery, useQueryClient } from '@tanstack/vue-query'
import { Plus, Pencil, Trash2 } from '@lucide/vue'
import PageToolbar from '@/components/ui/PageToolbar.vue'
import DataTable from '@/components/ui/DataTable.vue'
import ConfirmDialog from '@/components/ui/ConfirmDialog.vue'
import StatusBadge from '@/components/ui/StatusBadge.vue'
import FormDialog from '@/components/template/FormDialog.vue'
import { useToastStore } from '@/stores/toast'
import { listTemplates, deleteTemplate } from '@/api/templates'
import { listChannels } from '@/api/channels'
import type { MessageTemplate } from '@/types'

const toast = useToastStore()
const queryClient = useQueryClient()
const route = useRoute()

// 新手引导跳转：?action=create 自动打开新建模板弹窗
// 不能用 immediate（此时 formOpen 等 ref 尚未初始化会 TDZ 报错），用 onMounted + watch
watch(
  () => route.query,
  (q) => {
    if (q.action === 'create') openCreate()
  }
)
onMounted(() => {
  if (route.query.action === 'create') openCreate()
})

const page = ref(1)
const pageSize = 20
const keyword = ref('')

const { data, isLoading } = useQuery({
  queryKey: ['templates', page, keyword],
  queryFn: () => listTemplates({ page: page.value, page_size: pageSize, key: keyword.value || undefined }),
})

// 通道下拉
const { data: channelsData } = useQuery({
  queryKey: ['channels-options'],
  queryFn: () => listChannels({ page: 1, page_size: 100 }),
})

const formOpen = ref(false)
const editing = ref<MessageTemplate | null>(null)

function openCreate() {
  editing.value = null
  formOpen.value = true
}

function openEdit(t: MessageTemplate) {
  editing.value = t
  formOpen.value = true
}

function onSaved() {
  formOpen.value = false
  queryClient.invalidateQueries({ queryKey: ['templates'] })
}

const deleteTarget = ref<MessageTemplate | null>(null)
const deleting = ref(false)

async function confirmDelete() {
  if (!deleteTarget.value) return
  deleting.value = true
  try {
    await deleteTemplate(deleteTarget.value.id)
    toast.success('模板已删除')
    queryClient.invalidateQueries({ queryKey: ['templates'] })
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
  { key: 'name', label: '模板名称' },
  { key: 'channel_id', label: '通道', align: 'center' as const },
  { key: 'content', label: '内容' },
  { key: 'status', label: '状态', align: 'center' as const },
  { key: 'created_at', label: '创建时间', width: '180px' },
  { key: 'actions', label: '操作', align: 'center' as const, width: '120px' },
]

function channelName(id: number) {
  return channelsData.value?.list?.find((c) => c.id === id)?.name ?? `#${id}`
}
</script>

<template>
  <div>
    <PageToolbar title="模板管理" description="消息模板，支持 {var} 占位符">
      <button
        class="inline-flex items-center gap-1.5 rounded-md bg-primary px-3.5 py-2 text-sm font-medium text-white transition-colors hover:bg-primary-hover"
        @click="openCreate"
      >
        <Plus class="h-4 w-4" />
        新建模板
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
      <template #cell-channel_id="{ row }">
        <span class="text-sm">{{ channelName((row as MessageTemplate).channel_id) }}</span>
      </template>
      <template #cell-content="{ row }">
        <span class="line-clamp-1 max-w-md text-muted-foreground">{{ (row as MessageTemplate).content }}</span>
      </template>
      <template #cell-status="{ row }">
        <StatusBadge :value="(row as MessageTemplate).status === 1" />
      </template>
      <template #cell-actions="{ row }">
        <div class="flex items-center justify-center gap-1">
          <button class="rounded-md p-1.5 text-muted-foreground hover:bg-muted hover:text-foreground" @click="openEdit(row as MessageTemplate)">
            <Pencil class="h-4 w-4" />
          </button>
          <button class="rounded-md p-1.5 text-destructive/70 hover:bg-destructive/10 hover:text-destructive" @click="deleteTarget = row as MessageTemplate">
            <Trash2 class="h-4 w-4" />
          </button>
        </div>
      </template>
    </DataTable>

    <!-- 弹窗组件 -->
    <FormDialog :open="formOpen" :template="editing" @close="formOpen = false" @saved="onSaved" />

    <ConfirmDialog
      :open="!!deleteTarget"
      title="删除模板"
      :message="`确定删除模板「${deleteTarget?.name}」吗？`"
      danger
      :loading="deleting"
      @confirm="confirmDelete"
      @cancel="deleteTarget = null"
    />
  </div>
</template>
