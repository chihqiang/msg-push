<script setup lang="ts">
// 通道详情弹窗：健康历史 / 模板绑定 / 签名映射 / 测试发送
import { ref, watch } from 'vue'
import { useQuery } from '@tanstack/vue-query'
import { Send, HeartPulse, Cable, FileSignature, Trash2, Pencil, Plus } from '@lucide/vue'
import Modal from '@/components/ui/Modal.vue'
import DataTable from '@/components/ui/DataTable.vue'
import StatusBadge from '@/components/ui/StatusBadge.vue'
import ConfirmDialog from '@/components/ui/ConfirmDialog.vue'
import BindingFormDialog from '@/components/channel/BindingFormDialog.vue'
import SignatureMappingFormDialog from '@/components/channel/SignatureMappingFormDialog.vue'
import { useToastStore } from '@/stores/toast'
import {
  testChannel,
  channelHealthHistory,
  channelBindings,
  deleteChannelBinding,
  channelSignatureMappings,
  deleteChannelSignatureMapping,
} from '@/api/channels'
import type { Channel, ChannelBinding, ChannelSignatureMapping, ChannelHealthHistory } from '@/types'

const props = defineProps<{
  open: boolean
  channel: Channel | null
}>()

const emit = defineEmits<{
  close: []
}>()

const toast = useToastStore()
const detailTab = ref<'health' | 'bindings' | 'signatures' | 'test'>('health')

// 各 tab 按需加载
const { data: healthData, refetch: refetchHealth } = useQuery({
  queryKey: ['channel-health', () => props.channel?.id],
  queryFn: () => channelHealthHistory(props.channel!.id, { page: 1, page_size: 20 }),
  enabled: false,
})

const { data: bindingsData, refetch: refetchBindings } = useQuery({
  queryKey: ['channel-bindings', () => props.channel?.id],
  queryFn: () => channelBindings(props.channel!.id, { page: 1, page_size: 50 }),
  enabled: false,
})

const { data: sigMappingsData, refetch: refetchSigs } = useQuery({
  queryKey: ['channel-sig-mappings', () => props.channel?.id],
  queryFn: () => channelSignatureMappings(props.channel!.id, { page: 1, page_size: 50 }),
  enabled: false,
})

// 打开时刷新数据
watch(
  () => props.open,
  (open) => {
    if (!open || !props.channel) return
    detailTab.value = 'health'
    setTimeout(() => {
      refetchHealth()
      refetchBindings()
      refetchSigs()
    }, 50)
  }
)

// 测试发送
const testReceiver = ref('')
const testContent = ref('')
const testing = ref(false)

async function onTest() {
  if (!props.channel) return
  if (!testReceiver.value) {
    toast.error('请输入接收方')
    return
  }
  testing.value = true
  try {
    const result = await testChannel(props.channel.id, {
      receiver: testReceiver.value,
      content: testContent.value || undefined,
    })
    toast.success(result.status === 'success' ? '测试发送成功' : `测试发送失败（状态：${result.status}）`)
  } catch (e) {
    toast.error(e instanceof Error ? e.message : '测试失败')
  } finally {
    testing.value = false
  }
}

// ===== 删除（二次确认） =====
const deleteTarget = ref<{ kind: 'binding' | 'mapping'; id: number; name: string } | null>(null)
const deleting = ref(false)

async function confirmDelete() {
  if (!deleteTarget.value || !props.channel) return
  deleting.value = true
  try {
    if (deleteTarget.value.kind === 'binding') {
      await deleteChannelBinding(props.channel.id, deleteTarget.value.id)
      toast.success('绑定已删除')
      refetchBindings()
    } else {
      await deleteChannelSignatureMapping(props.channel.id, deleteTarget.value.id)
      toast.success('映射已删除')
      refetchSigs()
    }
  } catch (e) {
    toast.error(e instanceof Error ? e.message : '删除失败')
  } finally {
    deleting.value = false
    deleteTarget.value = null
  }
}

// ===== 绑定新增/编辑 =====
const bindingFormOpen = ref(false)
const bindingEditing = ref<ChannelBinding | null>(null)

function openBindingCreate() {
  bindingEditing.value = null
  bindingFormOpen.value = true
}

function openBindingEdit(binding: ChannelBinding) {
  bindingEditing.value = binding
  bindingFormOpen.value = true
}

function onBindingSaved() {
  bindingFormOpen.value = false
  refetchBindings()
}

// ===== 签名映射新增/编辑 =====
const sigFormOpen = ref(false)
const sigMappingEditing = ref<ChannelSignatureMapping | null>(null)

function openSigCreate() {
  sigMappingEditing.value = null
  sigFormOpen.value = true
}

function openSigEdit(m: ChannelSignatureMapping) {
  sigMappingEditing.value = m
  sigFormOpen.value = true
}

function onSigSaved() {
  sigFormOpen.value = false
  refetchSigs()
}

const healthColumns = [
  { key: 'check_time', label: '检查时间', width: '180px' },
  { key: 'provider_channel_id', label: '服务商通道ID', align: 'center' as const },
  { key: 'status', label: '状态', align: 'center' as const },
  { key: 'response_time', label: '耗时(ms)', align: 'right' as const },
  { key: 'error_count', label: '错误数', align: 'center' as const },
  { key: 'success_rate', label: '成功率', align: 'center' as const },
]

const bindingColumns = [
  { key: 'provider_template_name', label: '供应商模板' },
  { key: 'provider_name', label: '服务商' },
  { key: 'weight', label: '权重', align: 'center' as const },
  { key: 'priority', label: '优先级', align: 'center' as const },
  { key: 'is_active', label: '激活', align: 'center' as const },
  { key: 'actions', label: '操作', align: 'center' as const, width: '100px' },
]

const sigColumns = [
  { key: 'signature_name', label: '签名别名' },
  { key: 'provider_name', label: '服务商' },
  { key: 'status', label: '状态', align: 'center' as const },
  { key: 'actions', label: '操作', align: 'center' as const, width: '100px' },
]

const tabs = [
  { key: 'health', label: '健康历史', icon: HeartPulse },
  { key: 'bindings', label: '模板绑定', icon: Cable },
  { key: 'signatures', label: '签名映射', icon: FileSignature },
  { key: 'test', label: '测试发送', icon: Send },
] as const
</script>

<template>
  <Modal :open="open" :title="`通道详情：${channel?.name ?? ''}`" width="48rem" @close="emit('close')">
    <!-- Tabs -->
    <div class="mb-4 flex gap-1 border-b border-border">
      <button
        v-for="tab in tabs"
        :key="tab.key"
        class="flex items-center gap-1.5 border-b-2 px-3 py-2 text-sm transition-colors"
        :class="detailTab === tab.key ? 'border-primary font-medium text-primary' : 'border-transparent text-muted-foreground hover:text-foreground'"
        @click="detailTab = tab.key"
      >
        <component :is="tab.icon" class="h-4 w-4" />
        {{ tab.label }}
      </button>
    </div>

    <!-- 健康历史 -->
    <div v-if="detailTab === 'health'">
      <DataTable :columns="healthColumns" :data="(healthData?.list ?? []) as unknown[]" :loading="false" :total="healthData?.total ?? 0">
        <template #cell-status="{ row }">
          <span class="rounded-full px-2 py-0.5 text-xs font-medium" :class="(row as ChannelHealthHistory).status === 'healthy' ? 'bg-emerald-500/10 text-emerald-600' : 'bg-destructive/10 text-destructive'">
            {{ (row as ChannelHealthHistory).status === 'healthy' ? '健康' : '不健康' }}
          </span>
        </template>
        <template #cell-success_rate="{ row }">
          <span class="text-sm">{{ (row as ChannelHealthHistory).success_rate }}%</span>
        </template>
      </DataTable>
    </div>

    <!-- 模板绑定 -->
    <div v-else-if="detailTab === 'bindings'">
      <div class="mb-3 flex items-center justify-between">
        <p class="text-xs text-muted-foreground">同一通道可绑定多个供应商模板，按优先级+权重分配流量</p>
        <button
          class="inline-flex items-center gap-1.5 rounded-md bg-primary px-3 py-1.5 text-sm font-medium text-white transition-colors hover:bg-primary-hover"
          @click="openBindingCreate"
        >
          <Plus class="h-4 w-4" />
          新增绑定
        </button>
      </div>
      <DataTable
        :columns="bindingColumns"
        :data="(bindingsData?.list ?? []) as unknown[]"
        :loading="false"
        :total="bindingsData?.total ?? 0"
        empty-text="暂无绑定（请先在服务商管理创建供应商模板）"
      >
        <template #cell-is_active="{ row }">
          <StatusBadge :value="(row as ChannelBinding).is_active === 1" active-text="激活" inactive-text="停用" />
        </template>
        <template #cell-actions="{ row }">
          <div class="flex items-center justify-center gap-1">
            <button class="rounded-md p-1.5 text-muted-foreground hover:bg-muted hover:text-foreground" @click="openBindingEdit(row as ChannelBinding)">
              <Pencil class="h-4 w-4" />
            </button>
            <button class="rounded-md p-1.5 text-destructive/70 hover:bg-destructive/10 hover:text-destructive" @click="deleteTarget = { kind: 'binding', id: (row as ChannelBinding).id, name: (row as ChannelBinding).provider_template_name }">
              <Trash2 class="h-4 w-4" />
            </button>
          </div>
        </template>
      </DataTable>
    </div>

    <!-- 签名映射 -->
    <div v-else-if="detailTab === 'signatures'">
      <div class="mb-3 flex items-center justify-between">
        <p class="text-xs text-muted-foreground">将自定义签名名称映射到供应商签名</p>
        <button
          class="inline-flex items-center gap-1.5 rounded-md bg-primary px-3 py-1.5 text-sm font-medium text-white transition-colors hover:bg-primary-hover"
          @click="openSigCreate"
        >
          <Plus class="h-4 w-4" />
          新增映射
        </button>
      </div>
      <DataTable
        :columns="sigColumns"
        :data="(sigMappingsData?.list ?? []) as unknown[]"
        :loading="false"
        :total="sigMappingsData?.total ?? 0"
        empty-text="暂无签名映射"
      >
        <template #cell-status="{ row }">
          <StatusBadge :value="(row as ChannelSignatureMapping).status === 1" />
        </template>
        <template #cell-actions="{ row }">
          <div class="flex items-center justify-center gap-1">
            <button class="rounded-md p-1.5 text-muted-foreground hover:bg-muted hover:text-foreground" @click="openSigEdit(row as ChannelSignatureMapping)">
              <Pencil class="h-4 w-4" />
            </button>
            <button class="rounded-md p-1.5 text-destructive/70 hover:bg-destructive/10 hover:text-destructive" @click="deleteTarget = { kind: 'mapping', id: (row as ChannelSignatureMapping).id, name: (row as ChannelSignatureMapping).signature_name }">
              <Trash2 class="h-4 w-4" />
            </button>
          </div>
        </template>
      </DataTable>
    </div>

    <!-- 测试发送 -->
    <div v-else class="space-y-4">
      <div>
        <label class="mb-1.5 block text-sm font-medium text-slate-700">接收方 *</label>
        <input v-model="testReceiver" type="text" placeholder="手机号 / 邮箱 / 企微成员ID" class="h-9 w-full rounded-md border border-input bg-card px-3 text-sm outline-none focus:border-primary" />
      </div>
      <div>
        <label class="mb-1.5 block text-sm font-medium text-slate-700">内容（可留空走模板）</label>
        <textarea v-model="testContent" rows="3" class="w-full rounded-md border border-input bg-card px-3 py-2 text-sm outline-none focus:border-primary" />
      </div>
      <button class="inline-flex items-center gap-1.5 rounded-md bg-primary px-4 py-2 text-sm font-medium text-white hover:bg-primary-hover disabled:opacity-50" :disabled="testing" @click="onTest">
        <Send class="h-4 w-4" />
        {{ testing ? '发送中…' : '发送测试' }}
      </button>
    </div>

    <!-- 绑定/映射 新增编辑弹窗 -->
    <BindingFormDialog :open="bindingFormOpen" :channel="channel" :binding="bindingEditing" @close="bindingFormOpen = false" @saved="onBindingSaved" />
    <SignatureMappingFormDialog :open="sigFormOpen" :channel="channel" :mapping="sigMappingEditing" @close="sigFormOpen = false" @saved="onSigSaved" />

    <!-- 删除确认 -->
    <ConfirmDialog
      :open="!!deleteTarget"
      title="删除确认"
      :message="deleteTarget?.kind === 'binding' ? `确定删除绑定「${deleteTarget?.name}」吗？` : `确定删除签名映射「${deleteTarget?.name}」吗？`"
      danger
      :loading="deleting"
      @confirm="confirmDelete"
      @cancel="deleteTarget = null"
    />
  </Modal>
</template>
