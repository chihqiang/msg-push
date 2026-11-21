<script setup lang="ts">
// 应用创建/编辑弹窗
import { ref, watch } from 'vue'
import Modal from '@/components/ui/Modal.vue'
import { useToastStore } from '@/stores/toast'
import { createApp, updateApp, type AppCreatePayload } from '@/api/apps'
import type { App, AppWithSecret } from '@/types'

const props = defineProps<{
  open: boolean
  app: App | null // null=新建
}>()

const emit = defineEmits<{
  close: []
  saved: [resp: AppWithSecret | null]
}>()

const toast = useToastStore()

const saving = ref(false)
const form = ref<AppCreatePayload & { status?: number }>({
  name: '',
  remark: '',
  is_test: false,
  daily_quota: 0,
  rate_limit: 100,
})

// 打开时重置表单
watch(
  () => props.open,
  (open) => {
    if (!open) return
    if (props.app) {
      form.value = {
        name: props.app.name,
        remark: props.app.remark,
        is_test: props.app.is_test,
        daily_quota: props.app.daily_quota,
        rate_limit: props.app.rate_limit,
        status: props.app.status,
      }
    } else {
      form.value = { name: '', remark: '', is_test: false, daily_quota: 0, rate_limit: 100 }
    }
  }
)

async function submit() {
  if (!form.value.name) {
    toast.error('请输入应用名称')
    return
  }
  saving.value = true
  try {
    if (props.app) {
      await updateApp(props.app.id, {
        name: form.value.name,
        remark: form.value.remark,
        is_test: form.value.is_test,
        daily_quota: form.value.daily_quota,
        rate_limit: form.value.rate_limit,
        status: form.value.status,
      })
      toast.success('应用已更新')
      emit('saved', null)
    } else {
      const resp = await createApp(form.value)
      toast.success('应用已创建')
      // 创建时返回明文密钥，仅此一次
      emit('saved', resp)
    }
  } catch (e) {
    toast.error(e instanceof Error ? e.message : '保存失败')
  } finally {
    saving.value = false
  }
}
</script>

<template>
  <Modal :open="open" :title="app ? '编辑应用' : '新建应用'" @close="emit('close')">
    <div class="space-y-4">
      <div>
        <label class="mb-1.5 block text-sm font-medium text-slate-700">应用名称 *</label>
        <input v-model="form.name" type="text" placeholder="如：订单通知" class="h-9 w-full rounded-md border border-input bg-card px-3 text-sm outline-none focus:border-primary" />
      </div>
      <div class="grid grid-cols-2 gap-4">
        <div>
          <label class="mb-1.5 block text-sm font-medium text-slate-700">每日配额（0=不限）</label>
          <input v-model.number="form.daily_quota" type="number" min="0" class="h-9 w-full rounded-md border border-input bg-card px-3 text-sm outline-none focus:border-primary" />
        </div>
        <div>
          <label class="mb-1.5 block text-sm font-medium text-slate-700">限速 QPS</label>
          <input v-model.number="form.rate_limit" type="number" min="1" class="h-9 w-full rounded-md border border-input bg-card px-3 text-sm outline-none focus:border-primary" />
        </div>
      </div>
      <div>
        <label class="mb-1.5 block text-sm font-medium text-slate-700">备注</label>
        <textarea v-model="form.remark" rows="2" class="w-full rounded-md border border-input bg-card px-3 py-2 text-sm outline-none focus:border-primary" />
      </div>
      <div class="flex items-center justify-between rounded-md border border-border bg-muted/40 px-3 py-2.5">
        <div>
          <div class="text-sm font-medium text-slate-700">测试模式</div>
          <div class="text-xs text-muted-foreground">开启后该应用发送的消息走完整链路但不真实发送（模拟成功），适合联调</div>
        </div>
        <label class="relative inline-flex cursor-pointer items-center">
          <input v-model="form.is_test" type="checkbox" class="peer sr-only" />
          <div class="h-6 w-11 rounded-full bg-muted-foreground/30 transition-colors peer-checked:bg-primary peer-checked:after:translate-x-5 after:absolute after:left-0.5 after:top-0.5 after:h-5 after:w-5 after:rounded-full after:bg-white after:shadow after:transition-transform"></div>
        </label>
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
