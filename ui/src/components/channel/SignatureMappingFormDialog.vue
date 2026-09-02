<script setup lang="ts">
// 通道-签名映射 创建/编辑弹窗
import { ref, watch, computed } from 'vue'
import { useQuery } from '@tanstack/vue-query'
import Modal from '@/components/ui/Modal.vue'
import { useToastStore } from '@/stores/toast'
import { channelAvailableSignatures, createChannelSignatureMapping, updateChannelSignatureMapping } from '@/api/channels'
import type { Channel, ChannelSignatureMapping, AvailableProviderSignature } from '@/types'

const props = defineProps<{
  open: boolean
  channel: Channel | null
  mapping: ChannelSignatureMapping | null
}>()

const emit = defineEmits<{
  close: []
  saved: []
}>()

const toast = useToastStore()
const saving = ref(false)

const { data: signatures } = useQuery({
  queryKey: ['channel-available-signatures', props.channel?.id],
  queryFn: () => channelAvailableSignatures(props.channel!.id),
  enabled: () => !!props.open && !!props.channel,
})

const form = ref<{
  signature_name: string
  provider_signature_id: number
  provider_id: number
  status: number
}>({
  signature_name: '',
  provider_signature_id: 0,
  provider_id: 0,
  status: 1,
})

function resetForm() {
  form.value = { signature_name: '', provider_signature_id: 0, provider_id: 0, status: 1 }
}

watch(
  () => props.open,
  (open) => {
    if (!open) return
    if (props.mapping) {
      form.value = {
        signature_name: props.mapping.signature_name,
        provider_signature_id: props.mapping.provider_signature_id,
        provider_id: props.mapping.provider_id,
        status: props.mapping.status,
      }
    } else {
      resetForm()
    }
  }
)

const selectedSignature = computed<AvailableProviderSignature | undefined>(() =>
  (signatures.value ?? []).find((s) => s.id === form.value.provider_signature_id)
)

function onSignatureChange() {
  const sig = selectedSignature.value
  if (!sig) {
    form.value.provider_id = 0
    return
  }
  // 自动带出服务商账号 ID 与签名名
  form.value.provider_id = sig.provider_id
  if (!form.value.signature_name) {
    form.value.signature_name = sig.signature_name
  }
}

async function submit() {
  if (!form.value.signature_name.trim()) {
    toast.error('请填写签名名称')
    return
  }
  if (!form.value.provider_signature_id || !form.value.provider_id) {
    toast.error('请选择供应商签名')
    return
  }
  saving.value = true
  try {
    const payload = {
      signature_name: form.value.signature_name.trim(),
      provider_signature_id: form.value.provider_signature_id,
      provider_id: form.value.provider_id,
      status: form.value.status,
    }
    if (props.mapping) {
      await updateChannelSignatureMapping(props.channel!.id, props.mapping.id, payload)
      toast.success('签名映射已更新')
    } else {
      await createChannelSignatureMapping(props.channel!.id, payload)
      toast.success('签名映射已创建')
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
  <Modal :open="open" :title="mapping ? '编辑签名映射' : '新增签名映射'" width="36rem" @close="emit('close')">
    <div class="space-y-4">
      <div>
        <label class="mb-1.5 block text-sm font-medium text-slate-700">签名名称 *</label>
        <input v-model="form.signature_name" type="text" placeholder="自定义签名，如【公司】验证码" class="h-9 w-full rounded-md border border-input bg-card px-3 text-sm outline-none focus:border-primary" />
      </div>
      <div>
        <label class="mb-1.5 block text-sm font-medium text-slate-700">供应商签名 *</label>
        <select v-model.number="form.provider_signature_id" class="h-9 w-full rounded-md border border-input bg-card px-3 text-sm outline-none focus:border-primary" @change="onSignatureChange">
          <option :value="0" disabled>请选择供应商签名</option>
          <option v-for="s in signatures ?? []" :key="s.id" :value="s.id">
            {{ s.signature_name }}（{{ s.provider_code }} / {{ s.provider_type }}）
          </option>
        </select>
      </div>
      <div>
        <label class="mb-1.5 block text-sm font-medium text-slate-700">状态</label>
        <select v-model.number="form.status" class="h-9 w-full rounded-md border border-input bg-card px-3 text-sm outline-none focus:border-primary">
          <option :value="1">启用</option>
          <option :value="0">禁用</option>
        </select>
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
