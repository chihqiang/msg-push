<script setup lang="ts">
// 通道-模板绑定 创建/编辑弹窗
import { ref, watch, computed } from 'vue'
import { useQuery } from '@tanstack/vue-query'
import { Plus, Trash2 } from '@lucide/vue'
import Modal from '@/components/ui/Modal.vue'
import { useToastStore } from '@/stores/toast'
import { channelAvailableTemplates, createChannelBinding, updateChannelBinding } from '@/api/channels'
import type { Channel, ChannelBinding, ParamMappingItem, AvailableProviderTemplate } from '@/types'

const props = defineProps<{
  open: boolean
  channel: Channel | null
  binding: ChannelBinding | null
}>()

const emit = defineEmits<{
  close: []
  saved: []
}>()

const toast = useToastStore()
const saving = ref(false)

const { data: templates } = useQuery({
  queryKey: ['channel-available-templates', () => props.channel?.id],
  queryFn: () => channelAvailableTemplates(props.channel!.id),
  enabled: () => !!props.open && !!props.channel,
})

const form = ref<{
  provider_template_id: number
  provider_id: number
  param_mapping: ParamMappingItem[]
  weight: number
  priority: number
  status: number
  is_active: number
  auto_disable_on_fail: boolean
  auto_disable_threshold: number
}>({
  provider_template_id: 0,
  provider_id: 0,
  param_mapping: [],
  weight: 10,
  priority: 100,
  status: 1,
  is_active: 1,
  auto_disable_on_fail: false,
  auto_disable_threshold: 5,
})

function emptyMapping(): ParamMappingItem {
  return { type: 'mapping', provider_var: '', system_var: '', value: '' }
}

function resetForm() {
  form.value = {
    provider_template_id: 0,
    provider_id: 0,
    param_mapping: [],
    weight: 10,
    priority: 100,
    status: 1,
    is_active: 1,
    auto_disable_on_fail: false,
    auto_disable_threshold: 5,
  }
}

watch(
  () => props.open,
  (open) => {
    if (!open) return
    if (props.binding) {
      form.value = {
        provider_template_id: props.binding.provider_template_id,
        provider_id: props.binding.provider_id,
        param_mapping: (props.binding.param_mapping ?? []).map((m) => ({ ...m })),
        weight: props.binding.weight,
        priority: props.binding.priority,
        status: props.binding.status,
        is_active: props.binding.is_active,
        auto_disable_on_fail: props.binding.auto_disable_on_fail,
        auto_disable_threshold: props.binding.auto_disable_threshold,
      }
    } else {
      resetForm()
    }
  }
)

const selectedTemplate = computed<AvailableProviderTemplate | undefined>(() =>
  (templates.value ?? []).find((t) => t.id === form.value.provider_template_id)
)

function onTemplateChange() {
  const tpl = selectedTemplate.value
  if (!tpl) {
    form.value.provider_id = 0
    form.value.param_mapping = []
    return
  }
  // 自动带出服务商账号 ID 与变量映射行
  form.value.provider_id = tpl.provider_id
  const vars = tpl.variables ?? []
  form.value.param_mapping = vars.map((v) => ({ type: 'mapping', provider_var: v, system_var: '', value: '' }))
  if (form.value.param_mapping.length === 0) form.value.param_mapping.push(emptyMapping())
}

function addMapping() {
  form.value.param_mapping.push(emptyMapping())
}

function removeMapping(index: number) {
  form.value.param_mapping.splice(index, 1)
}

async function submit() {
  if (!form.value.provider_template_id || !form.value.provider_id) {
    toast.error('请选择供应商模板')
    return
  }
  saving.value = true
  try {
    const payload = {
      provider_template_id: form.value.provider_template_id,
      provider_id: form.value.provider_id,
      param_mapping: form.value.param_mapping.filter((m) => m.provider_var.trim() || m.system_var.trim() || m.value.trim()),
      weight: form.value.weight,
      priority: form.value.priority,
      status: form.value.status,
      is_active: form.value.is_active,
      auto_disable_on_fail: form.value.auto_disable_on_fail,
      auto_disable_threshold: form.value.auto_disable_threshold,
    }
    if (props.binding) {
      await updateChannelBinding(props.channel!.id, props.binding.id, payload)
      toast.success('绑定已更新')
    } else {
      await createChannelBinding(props.channel!.id, payload)
      toast.success('绑定已创建')
    }
    emit('saved')
  } catch (e) {
    toast.error(e instanceof Error ? e.message : '保存失败')
  } finally {
    saving.value = false
  }
}
</script>

<template>
  <Modal :open="open" :title="binding ? '编辑绑定' : '新增绑定'" width="44rem" @close="emit('close')">
    <div class="space-y-4">
      <div>
        <label class="mb-1.5 block text-sm font-medium text-slate-700">供应商模板 *</label>
        <select v-model.number="form.provider_template_id" class="h-9 w-full rounded-md border border-input bg-card px-3 text-sm outline-none focus:border-primary" @change="onTemplateChange">
          <option :value="0" disabled>请选择供应商模板</option>
          <option v-for="t in templates ?? []" :key="t.id" :value="t.id">
            {{ t.template_name }}（{{ t.provider_code }} / {{ t.provider_type }}）
          </option>
        </select>
        <p v-if="selectedTemplate" class="mt-1.5 truncate text-xs text-muted-foreground">模板内容：{{ selectedTemplate.template_content || '—' }}</p>
      </div>

      <div class="grid grid-cols-3 gap-4">
        <div>
          <label class="mb-1.5 block text-sm font-medium text-slate-700">权重（1-100）</label>
          <input v-model.number="form.weight" type="number" min="1" max="100" class="h-9 w-full rounded-md border border-input bg-card px-3 text-sm outline-none focus:border-primary" />
        </div>
        <div>
          <label class="mb-1.5 block text-sm font-medium text-slate-700">优先级（越小越优先）</label>
          <input v-model.number="form.priority" type="number" min="0" max="1000" class="h-9 w-full rounded-md border border-input bg-card px-3 text-sm outline-none focus:border-primary" />
        </div>
        <div>
          <label class="mb-1.5 block text-sm font-medium text-slate-700">状态</label>
          <select v-model.number="form.status" class="h-9 w-full rounded-md border border-input bg-card px-3 text-sm outline-none focus:border-primary">
            <option :value="1">启用</option>
            <option :value="0">禁用</option>
          </select>
        </div>
      </div>

      <!-- 参数映射 -->
      <div class="rounded-lg border border-border p-4">
        <div class="mb-2 flex items-center justify-between">
          <p class="text-xs font-medium text-muted-foreground">参数映射（将系统变量/固定值映射到供应商模板变量）</p>
          <button class="inline-flex items-center gap-1 rounded-md border border-border px-2 py-1 text-xs text-muted-foreground hover:bg-muted" @click="addMapping">
            <Plus class="h-3.5 w-3.5" />
            添加映射
          </button>
        </div>
        <div v-if="form.param_mapping.length" class="space-y-2">
          <div v-for="(m, i) in form.param_mapping" :key="i" class="flex items-center gap-2">
            <select v-model="m.type" class="h-8 w-28 rounded-md border border-input bg-card px-2 text-xs outline-none focus:border-primary">
              <option value="mapping">系统变量</option>
              <option value="fixed">固定值</option>
            </select>
            <input v-model="m.provider_var" type="text" placeholder="供应商变量" class="h-8 w-40 rounded-md border border-input bg-card px-2 text-xs outline-none focus:border-primary" />
            <input
              v-if="m.type === 'mapping'"
              v-model="m.system_var"
              type="text"
              placeholder="系统变量（如 receiver）"
              class="h-8 flex-1 rounded-md border border-input bg-card px-2 text-xs outline-none focus:border-primary"
            />
            <input
              v-else
              v-model="m.value"
              type="text"
              placeholder="固定值"
              class="h-8 flex-1 rounded-md border border-input bg-card px-2 text-xs outline-none focus:border-primary"
            />
            <button class="rounded-md p-1 text-destructive/70 hover:bg-destructive/10 hover:text-destructive" @click="removeMapping(i)">
              <Trash2 class="h-4 w-4" />
            </button>
          </div>
        </div>
        <p v-else class="text-xs text-muted-foreground">选择模板后自动生成变量映射，也可手动添加</p>
      </div>

      <!-- 自动禁用 -->
      <div class="rounded-lg border border-border p-4">
        <div class="flex items-center gap-2">
          <input id="auto-disable" v-model="form.auto_disable_on_fail" type="checkbox" class="h-4 w-4 rounded border-input text-primary focus:ring-primary" />
          <label for="auto-disable" class="text-sm font-medium text-slate-700">连续失败自动禁用该绑定</label>
        </div>
        <div v-if="form.auto_disable_on_fail" class="mt-3">
          <label class="mb-1.5 block text-sm font-medium text-slate-700">失败阈值（1-100 次）</label>
          <input v-model.number="form.auto_disable_threshold" type="number" min="1" max="100" class="h-9 w-full rounded-md border border-input bg-card px-3 text-sm outline-none focus:border-primary" />
        </div>
      </div>
    </div>
    <template #footer>
      <div class="flex justify-end gap-2.5">
        <button class="h-9 rounded-md border border-border px-4 text-sm text-muted-foreground hover:bg-muted" @click="emit('close')">取消</button>
        <button class="h-9 rounded-md bg-primary px-4 text-sm font-medium text-white hover:bg-primary-hover disabled:opacity-50" :disabled="saving" @click="submit">
          {{ saving ? '保存中…' : '保存' }}
        </button>
      </div>
    </template>
  </Modal>
</template>
