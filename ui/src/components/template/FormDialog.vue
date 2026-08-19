<script setup lang="ts">
// 消息模板创建/编辑弹窗
import { ref, watch } from 'vue'
import { useQuery } from '@tanstack/vue-query'
import Modal from '@/components/ui/Modal.vue'
import { useToastStore } from '@/stores/toast'
import { createTemplate, updateTemplate, type TemplateCreatePayload } from '@/api/templates'
import { listChannels } from '@/api/channels'
import type { MessageTemplate } from '@/types'

const props = defineProps<{
  open: boolean
  template: MessageTemplate | null
}>()

const emit = defineEmits<{
  close: []
  saved: []
}>()

const toast = useToastStore()
const saving = ref(false)

const { data: channelsData } = useQuery({
  queryKey: ['channels-options'],
  queryFn: () => listChannels({ page: 1, page_size: 100 }),
})

const form = ref<TemplateCreatePayload & { status?: number }>({
  code: '',
  name: '',
  channel_id: 0,
  content: '',
  signature: '',
  remark: '',
})

watch(
  () => props.open,
  (open) => {
    if (!open) return
    if (props.template) {
      form.value = {
        code: props.template.code,
        name: props.template.name,
        channel_id: props.template.channel_id,
        content: props.template.content,
        signature: props.template.signature,
        remark: props.template.remark,
        status: props.template.status,
      }
    } else {
      form.value = { code: '', name: '', channel_id: 0, content: '', signature: '', remark: '' }
    }
  }
)

async function submit() {
  if (!form.value.name || !form.value.channel_id) {
    toast.error('请填写名称并选择通道')
    return
  }
  if (!props.template && !form.value.code) {
    toast.error('请输入模板编码（创建后不可修改）')
    return
  }
  saving.value = true
  try {
    if (props.template) {
      await updateTemplate(props.template.id, {
        name: form.value.name,
        content: form.value.content,
        signature: form.value.signature,
        remark: form.value.remark,
        status: form.value.status,
      })
      toast.success('模板已更新')
    } else {
      await createTemplate(form.value)
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
  <Modal :open="open" :title="template ? '编辑模板' : '新建模板'" width="40rem" @close="emit('close')">
    <div class="space-y-4">
      <div class="grid grid-cols-2 gap-4">
        <div>
          <label class="mb-1.5 block text-sm font-medium text-slate-700">模板编码 *</label>
          <input v-model="form.code" type="text" :disabled="!!template" placeholder="唯一标识，如 sms_verify_code" class="h-9 w-full rounded-md border border-input bg-card px-3 text-sm outline-none focus:border-primary disabled:cursor-not-allowed disabled:opacity-60" />
          <p class="mt-1 text-xs text-muted-foreground">发送接口用此编码定位模板，创建后不可修改。</p>
        </div>
        <div>
          <label class="mb-1.5 block text-sm font-medium text-slate-700">模板名称 *</label>
          <input v-model="form.name" type="text" class="h-9 w-full rounded-md border border-input bg-card px-3 text-sm outline-none focus:border-primary" />
        </div>
      </div>
      <div>
        <label class="mb-1.5 block text-sm font-medium text-slate-700">所属通道 *</label>
        <select v-model.number="form.channel_id" class="h-9 w-full rounded-md border border-input bg-card px-3 text-sm outline-none focus:border-primary">
          <option :value="0" disabled>请选择通道</option>
          <option v-for="c in channelsData?.list ?? []" :key="c.id" :value="c.id">{{ c.name }}</option>
        </select>
      </div>
      <div>
        <label class="mb-1.5 block text-sm font-medium text-slate-700">模板内容 *</label>
        <textarea v-model="form.content" rows="4" placeholder="如：您的验证码是 {code}，5 分钟内有效。" class="w-full rounded-md border border-input bg-card px-3 py-2 text-sm outline-none focus:border-primary" />
        <p class="mt-1 text-xs text-muted-foreground">使用 {变量名} 占位，发送时自动提取替换</p>
      </div>
      <div>
        <label class="mb-1.5 block text-sm font-medium text-slate-700">签名 / 主题</label>
        <input v-model="form.signature" type="text" placeholder="短信签名 / 邮件主题" class="h-9 w-full rounded-md border border-input bg-card px-3 text-sm outline-none focus:border-primary" />
      </div>
      <div>
        <label class="mb-1.5 block text-sm font-medium text-slate-700">备注</label>
        <input v-model="form.remark" type="text" class="h-9 w-full rounded-md border border-input bg-card px-3 text-sm outline-none focus:border-primary" />
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
