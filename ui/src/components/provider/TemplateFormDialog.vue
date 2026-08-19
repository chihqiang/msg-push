<script setup lang="ts">
// 供应商模板创建/编辑弹窗
import { ref, watch } from 'vue'
import Modal from '@/components/ui/Modal.vue'
import { useToastStore } from '@/stores/toast'
import { createProviderTemplate, updateProviderTemplate } from '@/api/providers'
import type { ProviderAccount, ProviderTemplate, ProviderMeta } from '@/types'

const props = defineProps<{
  open: boolean
  template: ProviderTemplate | null
  accounts: ProviderAccount[]
  providers: ProviderMeta[]
}>()

const emit = defineEmits<{
  close: []
  saved: []
}>()

const toast = useToastStore()
const saving = ref(false)
const form = ref<{
  provider_id: number
  template_code: string
  template_name: string
  content_type: string
  template_content: string
  variables: string[]
  status?: number
}>({
  provider_id: 0,
  template_code: '',
  template_name: '',
  content_type: 'text',
  template_content: '',
  variables: [],
})

watch(
  () => props.open,
  (open) => {
    if (!open) return
    if (props.template) {
      let variables: string[] = []
      try {
        variables = props.template.variables ? JSON.parse(props.template.variables) : []
      } catch {
        variables = []
      }
      form.value = {
        provider_id: props.template.provider_id,
        template_code: props.template.template_code,
        template_name: props.template.template_name,
        content_type: props.template.content_type,
        template_content: props.template.template_content,
        variables,
        status: props.template.status,
      }
    } else {
      form.value = {
        provider_id: props.accounts[0]?.id ?? 0,
        template_code: '',
        template_name: '',
        content_type: 'text',
        template_content: '',
        variables: [],
      }
    }
  }
)

function providerName(code: string) {
  return props.providers.find((p) => p.code === code)?.name ?? code
}

async function submit() {
  if (!form.value.template_code || !form.value.template_name) {
    toast.error('请填写模板编码与名称')
    return
  }
  saving.value = true
  try {
    if (props.template) {
      await updateProviderTemplate(props.template.id, form.value)
      toast.success('模板已更新')
    } else {
      await createProviderTemplate(form.value)
      toast.success('模板已创建')
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
  <Modal :open="open" :title="template ? '编辑供应商模板' : '新建供应商模板'" width="40rem" @close="emit('close')">
    <div class="space-y-4">
      <div class="grid grid-cols-2 gap-4">
        <div>
          <label class="mb-1.5 block text-sm font-medium text-slate-700">所属服务商账号 *</label>
          <select v-model.number="form.provider_id" class="h-9 w-full rounded-md border border-input bg-card px-3 text-sm outline-none focus:border-primary">
            <option v-for="a in accounts" :key="a.id" :value="a.id">{{ a.account_name }}（{{ providerName(a.provider_code) }}）</option>
          </select>
        </div>
        <div>
          <label class="mb-1.5 block text-sm font-medium text-slate-700">模板编码 *</label>
          <input v-model="form.template_code" type="text" placeholder="服务商平台模板ID" class="h-9 w-full rounded-md border border-input bg-card px-3 text-sm outline-none focus:border-primary" />
        </div>
      </div>
      <div class="grid grid-cols-2 gap-4">
        <div>
          <label class="mb-1.5 block text-sm font-medium text-slate-700">模板名称 *</label>
          <input v-model="form.template_name" type="text" class="h-9 w-full rounded-md border border-input bg-card px-3 text-sm outline-none focus:border-primary" />
        </div>
        <div>
          <label class="mb-1.5 block text-sm font-medium text-slate-700">内容类型</label>
          <select v-model="form.content_type" class="h-9 w-full rounded-md border border-input bg-card px-3 text-sm outline-none focus:border-primary">
            <option value="text">text</option>
            <option value="html">html</option>
            <option value="markdown">markdown</option>
          </select>
        </div>
      </div>
      <div>
        <label class="mb-1.5 block text-sm font-medium text-slate-700">模板内容</label>
        <textarea v-model="form.template_content" rows="4" placeholder="如：您的验证码是 {code}" class="w-full rounded-md border border-input bg-card px-3 py-2 text-sm outline-none focus:border-primary" />
      </div>
      <div>
        <label class="mb-1.5 block text-sm font-medium text-slate-700">变量（逗号分隔）</label>
        <input
          :value="form.variables.join(',')"
          type="text"
          placeholder="code, minute"
          class="h-9 w-full rounded-md border border-input bg-card px-3 text-sm outline-none focus:border-primary"
          @input="(e: Event) => (form.variables = (e.target as HTMLInputElement).value.split(',').map((s) => s.trim()).filter(Boolean))"
        />
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
