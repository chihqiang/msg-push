<script setup lang="ts">
// 服务商签名创建/编辑弹窗
import { ref, watch } from 'vue'
import Modal from '@/components/ui/Modal.vue'
import { useToastStore } from '@/stores/toast'
import { createProviderSignature, updateProviderSignature } from '@/api/providers'
import type { ProviderAccount, ProviderSignature, ProviderMeta } from '@/types'

const props = defineProps<{
  open: boolean
  signature: ProviderSignature | null
  accounts: ProviderAccount[]
  providers: ProviderMeta[]
}>()

const emit = defineEmits<{
  close: []
  saved: []
}>()

const toast = useToastStore()
const saving = ref(false)
const form = ref<{ provider_account_id: number; signature_code: string; signature_name: string; status?: number }>({
  provider_account_id: 0,
  signature_code: '',
  signature_name: '',
})

watch(
  () => props.open,
  (open) => {
    if (!open) return
    if (props.signature) {
      form.value = {
        provider_account_id: props.signature.provider_account_id,
        signature_code: props.signature.signature_code,
        signature_name: props.signature.signature_name,
        status: props.signature.status,
      }
    } else {
      form.value = { provider_account_id: props.accounts[0]?.id ?? 0, signature_code: '', signature_name: '' }
    }
  }
)

function providerName(code: string) {
  return props.providers.find((p) => p.code === code)?.name ?? code
}

async function submit() {
  if (!form.value.signature_code || !form.value.signature_name) {
    toast.error('请填写签名编码与名称')
    return
  }
  saving.value = true
  try {
    if (props.signature) {
      await updateProviderSignature(props.signature.id, form.value)
      toast.success('签名已更新')
    } else {
      await createProviderSignature(form.value)
      toast.success('签名已创建')
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
  <Modal :open="open" :title="signature ? '编辑签名' : '新建签名'" @close="emit('close')">
    <div class="space-y-4">
      <div>
        <label class="mb-1.5 block text-sm font-medium text-slate-700">所属服务商账号 *</label>
        <select v-model.number="form.provider_account_id" class="h-9 w-full rounded-md border border-input bg-card px-3 text-sm outline-none focus:border-primary">
          <option v-for="a in accounts" :key="a.id" :value="a.id">{{ a.account_name }}（{{ providerName(a.provider_code) }}）</option>
        </select>
      </div>
      <div>
        <label class="mb-1.5 block text-sm font-medium text-slate-700">签名名称 *</label>
        <input v-model="form.signature_name" type="text" placeholder="如：XX科技" class="h-9 w-full rounded-md border border-input bg-card px-3 text-sm outline-none focus:border-primary" />
      </div>
      <div>
        <label class="mb-1.5 block text-sm font-medium text-slate-700">签名编码 *</label>
        <input v-model="form.signature_code" type="text" placeholder="服务商平台报备的签名" class="h-9 w-full rounded-md border border-input bg-card px-3 text-sm outline-none focus:border-primary" />
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
