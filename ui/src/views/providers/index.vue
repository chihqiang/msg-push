<script setup lang="ts">
import { computed, ref } from 'vue'
import { useQuery, useQueryClient } from '@tanstack/vue-query'
import { Plus, Pencil, Trash2, ClipboardCopy } from '@lucide/vue'
import PageToolbar from '@/components/ui/PageToolbar.vue'
import DataTable from '@/components/ui/DataTable.vue'
import ConfirmDialog from '@/components/ui/ConfirmDialog.vue'
import StatusBadge from '@/components/ui/StatusBadge.vue'
import AccountFormDialog from '@/components/provider/AccountFormDialog.vue'
import CallbackDialog from '@/components/provider/CallbackDialog.vue'
import SignatureFormDialog from '@/components/provider/SignatureFormDialog.vue'
import TemplateFormDialog from '@/components/provider/TemplateFormDialog.vue'
import { useToastStore } from '@/stores/toast'
import {
  availableProviders,
  listProviderAccounts,
  deleteProviderAccount,
  listProviderSignatures,
  deleteProviderSignature,
  listProviderTemplates,
  deleteProviderTemplate,
} from '@/api/providers'
import type { ProviderAccount, ProviderMeta, ProviderSignature, ProviderTemplate } from '@/types'

const toast = useToastStore()
const queryClient = useQueryClient()

const tab = ref<'accounts' | 'signatures' | 'templates'>('accounts')

// 服务商元信息
const { data: providers } = useQuery({
  queryKey: ['provider-metas'],
  queryFn: availableProviders,
})

const providerMetaMap = computed(() => {
  const m = new Map<string, ProviderMeta>()
  for (const p of providers.value ?? []) m.set(p.code, p)
  return m
})

function providerName(code: string) {
  return providerMetaMap.value.get(code)?.name ?? code
}

// 服务商账号名称映射（模板列表的 provider_id → 账号名称）
function accountName(id: number) {
  return accountsData.value?.list?.find((a) => a.id === id)?.account_name ?? `#${id}`
}

// ===== 服务商账号 =====
const accountPage = ref(1)
const pageSize = 20
const accountKeyword = ref('')

const { data: accountsData, isLoading: accountsLoading } = useQuery({
  queryKey: ['provider-accounts', accountPage, accountKeyword],
  queryFn: () => listProviderAccounts({ page: accountPage.value, page_size: pageSize, key: accountKeyword.value || undefined }),
})

const accountFormOpen = ref(false)
const accountEditing = ref<ProviderAccount | null>(null)
const accountDelete = ref<ProviderAccount | null>(null)
const accountDeleting = ref(false)

// 回调地址弹窗
const callbackAccount = ref<ProviderAccount | null>(null)
const callbackOpen = computed(() => !!callbackAccount.value)

function openCallbackDialog(a: ProviderAccount) {
  callbackAccount.value = a
}

function closeCallbackDialog() {
  callbackAccount.value = null
}

function openCreateAccount() {
  accountEditing.value = null
  accountFormOpen.value = true
}

function openEditAccount(a: ProviderAccount) {
  accountEditing.value = a
  accountFormOpen.value = true
}

function onAccountSaved() {
  accountFormOpen.value = false
  queryClient.invalidateQueries({ queryKey: ['provider-accounts'] })
}

async function confirmDeleteAccount() {
  if (!accountDelete.value) return
  accountDeleting.value = true
  try {
    await deleteProviderAccount(accountDelete.value.id)
    toast.success('已删除')
    queryClient.invalidateQueries({ queryKey: ['provider-accounts'] })
  } catch (e) {
    toast.error(e instanceof Error ? e.message : '删除失败')
  } finally {
    accountDeleting.value = false
    accountDelete.value = null
  }
}

// ===== 服务商签名 =====
const sigPage = ref(1)
const sigKeyword = ref('')
const { data: sigsData, isLoading: sigsLoading } = useQuery({
  queryKey: ['provider-signatures', sigPage, sigKeyword],
  queryFn: () => listProviderSignatures({ page: sigPage.value, page_size: pageSize, key: sigKeyword.value || undefined }),
})

const sigFormOpen = ref(false)
const sigEditing = ref<ProviderSignature | null>(null)
const sigDelete = ref<ProviderSignature | null>(null)
const sigDeleting = ref(false)

function openCreateSig() {
  sigEditing.value = null
  sigFormOpen.value = true
}

function openEditSig(s: ProviderSignature) {
  sigEditing.value = s
  sigFormOpen.value = true
}

function onSigSaved() {
  sigFormOpen.value = false
  queryClient.invalidateQueries({ queryKey: ['provider-signatures'] })
}

async function confirmDeleteSig() {
  if (!sigDelete.value) return
  sigDeleting.value = true
  try {
    await deleteProviderSignature(sigDelete.value.id)
    toast.success('已删除')
    queryClient.invalidateQueries({ queryKey: ['provider-signatures'] })
  } catch (e) {
    toast.error(e instanceof Error ? e.message : '删除失败')
  } finally {
    sigDeleting.value = false
    sigDelete.value = null
  }
}

// ===== 供应商模板 =====
const tplPage = ref(1)
const tplKeyword = ref('')
const { data: tplsData, isLoading: tplsLoading } = useQuery({
  queryKey: ['provider-templates', tplPage, tplKeyword],
  queryFn: () => listProviderTemplates({ page: tplPage.value, page_size: pageSize, key: tplKeyword.value || undefined }),
})

const tplFormOpen = ref(false)
const tplEditing = ref<ProviderTemplate | null>(null)
const tplDelete = ref<ProviderTemplate | null>(null)
const tplDeleting = ref(false)

function openCreateTpl() {
  tplEditing.value = null
  tplFormOpen.value = true
}

function openEditTpl(t: ProviderTemplate) {
  tplEditing.value = t
  tplFormOpen.value = true
}

function onTplSaved() {
  tplFormOpen.value = false
  queryClient.invalidateQueries({ queryKey: ['provider-templates'] })
}

async function confirmDeleteTpl() {
  if (!tplDelete.value) return
  tplDeleting.value = true
  try {
    await deleteProviderTemplate(tplDelete.value.id)
    toast.success('已删除')
    queryClient.invalidateQueries({ queryKey: ['provider-templates'] })
  } catch (e) {
    toast.error(e instanceof Error ? e.message : '删除失败')
  } finally {
    tplDeleting.value = false
    tplDelete.value = null
  }
}

// 列定义
const accountColumns = [
  { key: 'id', label: 'ID', align: 'center' as const, width: '70px' },
  { key: 'account_name', label: '账号名称' },
  { key: 'provider_code', label: '服务商', align: 'center' as const },
  { key: 'status', label: '状态', align: 'center' as const },
  { key: 'created_at', label: '创建时间', width: '180px' },
  { key: 'actions', label: '操作', align: 'center' as const, width: '160px' },
]

const sigColumns = [
  { key: 'id', label: 'ID', align: 'center' as const, width: '70px' },
  { key: 'signature_name', label: '签名名称' },
  { key: 'signature_code', label: '签名编码' },
  { key: 'provider_code', label: '服务商' },
  { key: 'status', label: '状态', align: 'center' as const },
  { key: 'actions', label: '操作', align: 'center' as const, width: '120px' },
]

const tplColumns = [
  { key: 'id', label: 'ID', align: 'center' as const, width: '70px' },
  { key: 'template_name', label: '模板名称' },
  { key: 'template_code', label: '模板编码' },
  { key: 'provider_id', label: '服务商账号', align: 'center' as const },
  { key: 'content_type', label: '类型', align: 'center' as const },
  { key: 'status', label: '状态', align: 'center' as const },
  { key: 'actions', label: '操作', align: 'center' as const, width: '120px' },
]
</script>

<template>
  <div>
    <PageToolbar title="服务商管理" description="管理服务商账号、签名与供应商模板">
      <button
        class="inline-flex items-center gap-1.5 rounded-md bg-primary px-3.5 py-2 text-sm font-medium text-white transition-colors hover:bg-primary-hover"
        @click="tab === 'accounts' ? openCreateAccount() : tab === 'signatures' ? openCreateSig() : openCreateTpl()"
      >
        <Plus class="h-4 w-4" />
        新建{{ tab === 'accounts' ? '账号' : tab === 'signatures' ? '签名' : '模板' }}
      </button>
    </PageToolbar>

    <!-- Tabs -->
    <div class="mb-4 flex gap-1 border-b border-border">
      <button
        class="border-b-2 px-3 py-2 text-sm transition-colors"
        :class="tab === 'accounts' ? 'border-primary font-medium text-primary' : 'border-transparent text-muted-foreground hover:text-foreground'"
        @click="tab = 'accounts'"
      >
        服务商账号
      </button>
      <button
        class="border-b-2 px-3 py-2 text-sm transition-colors"
        :class="tab === 'signatures' ? 'border-primary font-medium text-primary' : 'border-transparent text-muted-foreground hover:text-foreground'"
        @click="tab = 'signatures'"
      >
        服务商签名
      </button>
      <button
        class="border-b-2 px-3 py-2 text-sm transition-colors"
        :class="tab === 'templates' ? 'border-primary font-medium text-primary' : 'border-transparent text-muted-foreground hover:text-foreground'"
        @click="tab = 'templates'"
      >
        供应商模板
      </button>
    </div>

    <!-- 搜索 -->
    <div class="mb-4 flex flex-wrap items-center gap-2">
      <input
        v-if="tab === 'accounts'"
        v-model="accountKeyword"
        type="text"
        placeholder="搜索账号名称 / 服务商"
        class="h-9 w-64 rounded-md border border-input bg-card px-3 text-sm outline-none focus:border-primary focus:ring-2 focus:ring-primary/20"
        @keyup.enter="accountPage = 1"
      />
      <input
        v-else-if="tab === 'signatures'"
        v-model="sigKeyword"
        type="text"
        placeholder="搜索签名名称 / 编码"
        class="h-9 w-64 rounded-md border border-input bg-card px-3 text-sm outline-none focus:border-primary focus:ring-2 focus:ring-primary/20"
        @keyup.enter="sigPage = 1"
      />
      <input
        v-else
        v-model="tplKeyword"
        type="text"
        placeholder="搜索模板名称 / 编码"
        class="h-9 w-64 rounded-md border border-input bg-card px-3 text-sm outline-none focus:border-primary focus:ring-2 focus:ring-primary/20"
        @keyup.enter="tplPage = 1"
      />
      <button
        class="h-9 rounded-md border border-border px-4 text-sm text-muted-foreground hover:bg-muted"
        @click="tab === 'accounts' ? (accountPage = 1) : tab === 'signatures' ? (sigPage = 1) : (tplPage = 1)"
      >
        搜索
      </button>
    </div>

    <!-- 账号列表 -->
    <DataTable
      v-if="tab === 'accounts'"
      :columns="accountColumns"
      :data="(accountsData?.list ?? []) as unknown[]"
      :loading="accountsLoading"
      :total="accountsData?.total ?? 0"
      :page="accountPage"
      :page-size="pageSize"
      @update:page="(p: number) => (accountPage = p)"
    >
      <template #cell-provider_code="{ row }">
        <span class="text-sm">{{ providerName((row as ProviderAccount).provider_code) }}</span>
      </template>
      <template #cell-status="{ row }">
        <StatusBadge :value="(row as ProviderAccount).status === 1" />
      </template>
      <template #cell-actions="{ row }">
        <div class="flex items-center justify-center gap-1">
          <button
            class="rounded-md p-1.5 text-muted-foreground hover:bg-muted hover:text-foreground"
            title="查看回调地址"
            @click="openCallbackDialog(row as ProviderAccount)"
          >
            <ClipboardCopy class="h-4 w-4" />
          </button>
          <button class="rounded-md p-1.5 text-muted-foreground hover:bg-muted hover:text-foreground" @click="openEditAccount(row as ProviderAccount)">
            <Pencil class="h-4 w-4" />
          </button>
          <button class="rounded-md p-1.5 text-destructive/70 hover:bg-destructive/10 hover:text-destructive" @click="accountDelete = row as ProviderAccount">
            <Trash2 class="h-4 w-4" />
          </button>
        </div>
      </template>
    </DataTable>

    <!-- 签名列表 -->
    <DataTable
      v-else-if="tab === 'signatures'"
      :columns="sigColumns"
      :data="(sigsData?.list ?? []) as unknown[]"
      :loading="sigsLoading"
      :total="sigsData?.total ?? 0"
      :page="sigPage"
      :page-size="pageSize"
      @update:page="(p: number) => (sigPage = p)"
    >
      <template #cell-status="{ row }">
        <StatusBadge :value="(row as ProviderSignature).status === 1" />
      </template>
      <template #cell-provider_code="{ row }">
        <span class="text-sm">{{ providerName((row as ProviderSignature).provider_code) }}</span>
      </template>
      <template #cell-actions="{ row }">
        <div class="flex items-center justify-center gap-1">
          <button class="rounded-md p-1.5 text-muted-foreground hover:bg-muted hover:text-foreground" @click="openEditSig(row as ProviderSignature)">
            <Pencil class="h-4 w-4" />
          </button>
          <button class="rounded-md p-1.5 text-destructive/70 hover:bg-destructive/10 hover:text-destructive" @click="sigDelete = row as ProviderSignature">
            <Trash2 class="h-4 w-4" />
          </button>
        </div>
      </template>
    </DataTable>

    <!-- 模板列表 -->
    <DataTable
      v-else
      :columns="tplColumns"
      :data="(tplsData?.list ?? []) as unknown[]"
      :loading="tplsLoading"
      :total="tplsData?.total ?? 0"
      :page="tplPage"
      :page-size="pageSize"
      @update:page="(p: number) => (tplPage = p)"
    >
      <template #cell-provider_id="{ row }">
        <span class="text-sm">{{ accountName((row as ProviderTemplate).provider_id) }}</span>
      </template>
      <template #cell-content_type="{ row }">
        <span class="rounded-full bg-muted px-2 py-0.5 text-xs text-muted-foreground">{{ (row as ProviderTemplate).content_type }}</span>
      </template>
      <template #cell-status="{ row }">
        <StatusBadge :value="(row as ProviderTemplate).status === 1" />
      </template>
      <template #cell-actions="{ row }">
        <div class="flex items-center justify-center gap-1">
          <button class="rounded-md p-1.5 text-muted-foreground hover:bg-muted hover:text-foreground" @click="openEditTpl(row as ProviderTemplate)">
            <Pencil class="h-4 w-4" />
          </button>
          <button class="rounded-md p-1.5 text-destructive/70 hover:bg-destructive/10 hover:text-destructive" @click="tplDelete = row as ProviderTemplate">
            <Trash2 class="h-4 w-4" />
          </button>
        </div>
      </template>
    </DataTable>

    <!-- 弹窗组件 -->
    <AccountFormDialog
      :open="accountFormOpen"
      :account="accountEditing"
      :providers="providers ?? []"
      :default-provider-code="providers?.[0]?.code ?? ''"
      @close="accountFormOpen = false"
      @saved="onAccountSaved"
    />
    <SignatureFormDialog
      :open="sigFormOpen"
      :signature="sigEditing"
      :accounts="accountsData?.list ?? []"
      :providers="providers ?? []"
      @close="sigFormOpen = false"
      @saved="onSigSaved"
    />
    <TemplateFormDialog
      :open="tplFormOpen"
      :template="tplEditing"
      :accounts="accountsData?.list ?? []"
      :providers="providers ?? []"
      @close="tplFormOpen = false"
      @saved="onTplSaved"
    />

    <!-- 回调地址弹窗 -->
    <CallbackDialog :open="callbackOpen" :account="callbackAccount" :providers="providers ?? []" @close="closeCallbackDialog" />

    <!-- 删除确认 -->
    <ConfirmDialog
      :open="!!accountDelete || !!sigDelete || !!tplDelete"
      title="删除确认"
      :message="accountDelete ? `确定删除服务商账号「${accountDelete.account_name}」吗？` : sigDelete ? `确定删除签名「${sigDelete.signature_name}」吗？` : `确定删除模板「${tplDelete?.template_name}」吗？`"
      danger
      :loading="accountDeleting || sigDeleting || tplDeleting"
      @confirm="accountDelete ? confirmDeleteAccount() : sigDelete ? confirmDeleteSig() : confirmDeleteTpl()"
      @cancel="accountDelete = null; sigDelete = null; tplDelete = null"
    />
  </div>
</template>
