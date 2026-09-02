<script setup lang="ts">
import { onMounted, ref, watch } from 'vue'
import { useRoute } from 'vue-router'
import { useQuery, useQueryClient } from '@tanstack/vue-query'
import { Plus, KeyRound, Pencil, Trash2 } from '@lucide/vue'
import PageToolbar from '@/components/ui/PageToolbar.vue'
import DataTable from '@/components/ui/DataTable.vue'
import ConfirmDialog from '@/components/ui/ConfirmDialog.vue'
import StatusBadge from '@/components/ui/StatusBadge.vue'
import FormDialog from '@/components/app/FormDialog.vue'
import SecretDialog from '@/components/app/SecretDialog.vue'
import QuotaDialog from '@/components/app/QuotaDialog.vue'
import { useToastStore } from '@/stores/toast'
import { listApps, deleteApp, resetAppSecret, getAppQuotaUsage } from '@/api/apps'
import type { App, AppQuotaUsage, AppWithSecret } from '@/types'

const toast = useToastStore()
const queryClient = useQueryClient()
const route = useRoute()

// 新手引导跳转：?action=create 自动打开新建应用弹窗
// 注意：不能用 immediate（此时 formOpen 等 ref 尚未初始化会 TDZ 报错），
// 改用 onMounted 处理首次进入，watch 处理已挂载后的 query 变化。
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
  queryKey: ['apps', page, keyword],
  queryFn: () => listApps({ page: page.value, page_size: pageSize, key: keyword.value || undefined }),
})

// 弹窗状态
const formOpen = ref(false)
const editing = ref<App | null>(null)
const secretOpen = ref(false)
const secretData = ref<{ app_id: string; secret: string } | null>(null)
const quotaOpen = ref(false)
const quotaData = ref<AppQuotaUsage | null>(null)

const deleteTarget = ref<App | null>(null)
const deleting = ref(false)

function openCreate() {
  editing.value = null
  formOpen.value = true
}

function openEdit(app: App) {
  editing.value = app
  formOpen.value = true
}

// 保存成功：新建时展示密钥
function onSaved(resp: AppWithSecret | null) {
  formOpen.value = false
  if (resp?.secret) {
    secretData.value = { app_id: resp.app_id, secret: resp.secret }
    secretOpen.value = true
  }
  queryClient.invalidateQueries({ queryKey: ['apps'] })
}

async function confirmDelete() {
  if (!deleteTarget.value) return
  deleting.value = true
  try {
    await deleteApp(deleteTarget.value.id)
    toast.success('应用已删除')
    queryClient.invalidateQueries({ queryKey: ['apps'] })
  } catch (e) {
    toast.error(e instanceof Error ? e.message : '删除失败')
  } finally {
    deleting.value = false
    deleteTarget.value = null
  }
}

async function onResetSecret(app: App) {
  try {
    const resp = await resetAppSecret(app.id)
    secretData.value = { app_id: resp.app_id, secret: resp.secret ?? '' }
    secretOpen.value = true
  } catch (e) {
    toast.error(e instanceof Error ? e.message : '重置失败')
  }
}

async function onViewQuota(app: App) {
  try {
    quotaData.value = await getAppQuotaUsage(app.id)
    quotaOpen.value = true
  } catch (e) {
    toast.error(e instanceof Error ? e.message : '查询失败')
  }
}

const columns = [
  { key: 'id', label: 'ID', align: 'center' as const, width: '70px' },
  { key: 'name', label: '应用名称' },
  { key: 'app_id', label: 'AppID' },
  { key: 'is_test', label: '模式', align: 'center' as const, width: '80px' },
  { key: 'daily_quota', label: '每日配额', align: 'center' as const },
  { key: 'rate_limit', label: '限速 QPS', align: 'center' as const },
  { key: 'status', label: '状态', align: 'center' as const },
  { key: 'created_at', label: '创建时间', width: '180px' },
  { key: 'actions', label: '操作', align: 'center' as const, width: '240px' },
]
</script>

<template>
  <div>
    <PageToolbar title="应用管理" description="管理调用消息推送 API 的应用及配额">
      <button
        class="inline-flex items-center gap-1.5 rounded-md bg-primary px-3.5 py-2 text-sm font-medium text-white transition-colors hover:bg-primary-hover"
        @click="openCreate"
      >
        <Plus class="h-4 w-4" />
        新建应用
      </button>
    </PageToolbar>

    <div class="mb-4 flex gap-2">
      <input
        v-model="keyword"
        type="text"
        placeholder="搜索应用名称 / AppID"
        class="h-9 w-64 rounded-md border border-input bg-card px-3 text-sm outline-none focus:border-primary focus:ring-2 focus:ring-primary/20"
        @keyup.enter="page = 1"
      />
      <button class="h-9 rounded-md border border-border px-4 text-sm text-muted-foreground hover:bg-muted" @click="page = 1">
        搜索
      </button>
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
      <template #cell-status="{ row }">
        <StatusBadge :value="(row as App).status === 1" />
      </template>
      <template #cell-is_test="{ row }">
        <span v-if="(row as App).is_test" class="rounded-full bg-amber-500/10 px-2 py-0.5 text-xs font-medium text-amber-600">测试</span>
        <span v-else class="rounded-full bg-emerald-500/10 px-2 py-0.5 text-xs font-medium text-emerald-600">正式</span>
      </template>
      <template #cell-app_id="{ row }">
        <code class="rounded bg-muted px-1.5 py-0.5 font-mono text-xs">{{ (row as App).app_id }}</code>
      </template>
      <template #cell-daily_quota="{ row }">
        <span>{{ (row as App).daily_quota === 0 ? '不限' : (row as App).daily_quota }}</span>
      </template>
      <template #cell-actions="{ row }">
        <div class="flex items-center justify-center gap-1.5">
          <button class="rounded-md px-2 py-1 text-xs text-muted-foreground hover:bg-muted hover:text-foreground" title="查看配额" @click="onViewQuota(row as App)">
            配额
          </button>
          <button class="rounded-md p-1.5 text-muted-foreground hover:bg-muted hover:text-foreground" title="重置密钥" @click="onResetSecret(row as App)">
            <KeyRound class="h-4 w-4" />
          </button>
          <button class="rounded-md p-1.5 text-muted-foreground hover:bg-muted hover:text-foreground" title="编辑" @click="openEdit(row as App)">
            <Pencil class="h-4 w-4" />
          </button>
          <button class="rounded-md p-1.5 text-destructive/70 hover:bg-destructive/10 hover:text-destructive" title="删除" @click="deleteTarget = row as App">
            <Trash2 class="h-4 w-4" />
          </button>
        </div>
      </template>
    </DataTable>

    <!-- 弹窗组件 -->
    <FormDialog :open="formOpen" :app="editing" @close="formOpen = false" @saved="onSaved" />
    <SecretDialog :open="secretOpen" :app-id="secretData?.app_id ?? ''" :secret="secretData?.secret ?? ''" @close="secretOpen = false" />
    <QuotaDialog :open="quotaOpen" :data="quotaData" @close="quotaOpen = false" />

    <ConfirmDialog
      :open="!!deleteTarget"
      title="删除应用"
      :message="`确定删除应用「${deleteTarget?.name}」吗？此操作不可恢复。`"
      danger
      :loading="deleting"
      @confirm="confirmDelete"
      @cancel="deleteTarget = null"
    />
  </div>
</template>
