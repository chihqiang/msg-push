<script setup lang="ts">
// 通道创建/编辑弹窗
import { ref, watch } from 'vue'
import Modal from '@/components/ui/Modal.vue'
import { useToastStore } from '@/stores/toast'
import { createChannel, updateChannel, type ChannelCreatePayload } from '@/api/channels'
import type { Channel } from '@/types'

const props = defineProps<{
  open: boolean
  channel: Channel | null
}>()

const emit = defineEmits<{
  close: []
  saved: []
}>()

const toast = useToastStore()
const saving = ref(false)
const form = ref<ChannelCreatePayload & { status?: number }>({ code: '', name: '', type: 'sms', remark: '' })

watch(
  () => props.open,
  (open) => {
    if (!open) return
    if (props.channel) {
      form.value = {
        code: props.channel.code,
        name: props.channel.name,
        type: props.channel.type,
        remark: props.channel.remark,
        status: props.channel.status,
      }
    } else {
      form.value = { code: '', name: '', type: 'sms', remark: '' }
    }
  }
)

async function submit() {
  if (!form.value.name) {
    toast.error('请输入通道名称')
    return
  }
  if (!props.channel && !form.value.code) {
    toast.error('请输入通道编码（创建后不可修改）')
    return
  }
  saving.value = true
  try {
    if (props.channel) {
      await updateChannel(props.channel.id, { name: form.value.name, status: form.value.status, remark: form.value.remark })
      toast.success('通道已更新')
    } else {
      await createChannel(form.value)
      toast.success('通道已创建')
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
  <Modal :open="open" :title="channel ? '编辑通道' : '新建通道'" @close="emit('close')">
    <div class="space-y-4">
      <div class="grid grid-cols-2 gap-4">
        <div>
          <label class="mb-1.5 block text-sm font-medium text-slate-700">通道编码 *</label>
          <input v-model="form.code" type="text" :disabled="!!channel" placeholder="唯一标识，如 aliyun_sms" class="h-9 w-full rounded-md border border-input bg-card px-3 text-sm outline-none focus:border-primary disabled:cursor-not-allowed disabled:opacity-60" />
          <p class="mt-1 text-xs text-muted-foreground">发送接口用此编码定位通道，创建后不可修改。</p>
        </div>
        <div>
          <label class="mb-1.5 block text-sm font-medium text-slate-700">通道名称 *</label>
          <input v-model="form.name" type="text" class="h-9 w-full rounded-md border border-input bg-card px-3 text-sm outline-none focus:border-primary" />
        </div>
      </div>
      <div v-if="!channel">
        <label class="mb-1.5 block text-sm font-medium text-slate-700">类型 *</label>
        <select v-model="form.type" class="h-9 w-full rounded-md border border-input bg-card px-3 text-sm outline-none focus:border-primary">
          <option value="sms">短信</option>
          <option value="email">邮件</option>
          <option value="wecom">企业微信</option>
          <option value="dingtalk">钉钉</option>
        </select>
      </div>
      <div>
        <label class="mb-1.5 block text-sm font-medium text-slate-700">备注</label>
        <textarea v-model="form.remark" rows="2" class="w-full rounded-md border border-input bg-card px-3 py-2 text-sm outline-none focus:border-primary" />
      </div>
      <p v-if="!channel" class="text-xs text-muted-foreground">
        实际发送配置取自服务商账号（在「服务商管理」中配置），通道仅作类型与归属标识。
      </p>
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
