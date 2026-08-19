<script setup lang="ts">
// 失败规则创建/编辑弹窗
import { ref, watch } from 'vue'
import Modal from '@/components/ui/Modal.vue'
import { useToastStore } from '@/stores/toast'
import { createFailureRule, updateFailureRule } from '@/api/failureRules'
import { availableProviders } from '@/api/providers'
import { useQuery } from '@tanstack/vue-query'
import type { FailureRule } from '@/types'

const props = defineProps<{
  open: boolean
  rule: FailureRule | null
}>()

const emit = defineEmits<{
  close: []
  saved: []
}>()

const toast = useToastStore()
const saving = ref(false)

const { data: providers } = useQuery({
  queryKey: ['provider-metas-rule'],
  queryFn: availableProviders,
})

const form = ref<{
  name: string
  scene: string
  provider_code: string
  message_type: string
  error_code: string
  error_keyword: string
  action: string
  priority: number
  status?: number
  action_config: Record<string, unknown>
}>({
  name: '',
  scene: 'send_failure',
  provider_code: '',
  message_type: '',
  error_code: '',
  error_keyword: '',
  action: 'retry',
  priority: 0,
  action_config: {},
})

watch(
  () => props.open,
  (open) => {
    if (!open) return
    if (props.rule) {
      let actionConfig: Record<string, unknown> = {}
      try {
        actionConfig = props.rule.action_config ? JSON.parse(props.rule.action_config) : {}
      } catch {
        actionConfig = {}
      }
      form.value = {
        name: props.rule.name,
        scene: props.rule.scene,
        provider_code: props.rule.provider_code,
        message_type: props.rule.message_type,
        error_code: props.rule.error_code,
        error_keyword: props.rule.error_keyword,
        action: props.rule.action,
        priority: props.rule.priority,
        status: props.rule.status,
        action_config: actionConfig,
      }
    } else {
      form.value = {
        name: '',
        scene: 'send_failure',
        provider_code: '',
        message_type: '',
        error_code: '',
        error_keyword: '',
        action: 'retry',
        priority: 0,
        action_config: {},
      }
    }
  }
)

async function submit() {
  if (!form.value.name || !form.value.scene || !form.value.action) {
    toast.error('请填写规则名称、场景与动作')
    return
  }
  saving.value = true
  try {
    const payload = {
      name: form.value.name,
      scene: form.value.scene,
      provider_code: form.value.provider_code || undefined,
      message_type: form.value.message_type || undefined,
      error_code: form.value.error_code || undefined,
      error_keyword: form.value.error_keyword || undefined,
      action: form.value.action,
      priority: form.value.priority,
      status: form.value.status,
      action_config: Object.keys(form.value.action_config).length ? form.value.action_config : undefined,
    }
    if (props.rule) {
      await updateFailureRule(props.rule.id, payload)
      toast.success('规则已更新')
    } else {
      await createFailureRule(payload)
      toast.success('规则已创建')
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
  <Modal :open="open" :title="rule ? '编辑规则' : '新建规则'" width="44rem" @close="emit('close')">
    <div class="space-y-4">
      <div class="grid grid-cols-3 gap-4">
        <div>
          <label class="mb-1.5 block text-sm font-medium text-slate-700">规则名称 *</label>
          <input v-model="form.name" type="text" class="h-9 w-full rounded-md border border-input bg-card px-3 text-sm outline-none focus:border-primary" />
        </div>
        <div>
          <label class="mb-1.5 block text-sm font-medium text-slate-700">场景 *</label>
          <select v-model="form.scene" class="h-9 w-full rounded-md border border-input bg-card px-3 text-sm outline-none focus:border-primary">
            <option value="send_failure">发送失败</option>
            <option value="callback_failure">回调失败</option>
          </select>
        </div>
        <div>
          <label class="mb-1.5 block text-sm font-medium text-slate-700">动作 *</label>
          <select v-model="form.action" class="h-9 w-full rounded-md border border-input bg-card px-3 text-sm outline-none focus:border-primary">
            <option value="retry">重试</option>
            <option value="switch_provider">切换供应商</option>
            <option value="fail">直接失败</option>
            <option value="alert">告警</option>
          </select>
        </div>
      </div>

      <div class="grid grid-cols-3 gap-4">
        <div>
          <label class="mb-1.5 block text-sm font-medium text-slate-700">服务商（留空=全部）</label>
          <select v-model="form.provider_code" class="h-9 w-full rounded-md border border-input bg-card px-3 text-sm outline-none focus:border-primary">
            <option value="">全部</option>
            <option v-for="p in providers ?? []" :key="p.code" :value="p.code">{{ p.name }}</option>
          </select>
        </div>
        <div>
          <label class="mb-1.5 block text-sm font-medium text-slate-700">消息类型（留空=全部）</label>
          <select v-model="form.message_type" class="h-9 w-full rounded-md border border-input bg-card px-3 text-sm outline-none focus:border-primary">
            <option value="">全部</option>
            <option value="sms">短信</option>
            <option value="email">邮件</option>
            <option value="wecom">企业微信</option>
            <option value="dingtalk">钉钉</option>
          </select>
        </div>
        <div>
          <label class="mb-1.5 block text-sm font-medium text-slate-700">优先级（越大越优先）</label>
          <input v-model.number="form.priority" type="number" class="h-9 w-full rounded-md border border-input bg-card px-3 text-sm outline-none focus:border-primary" />
        </div>
      </div>

      <div>
        <label class="mb-1.5 block text-sm font-medium text-slate-700">错误码（逗号分隔，留空=全部）</label>
        <input v-model="form.error_code" type="text" placeholder="如 ISV_SMS_SIGNATURE_ILLEGAL,isv.SMS_SIGNATURE_ILLEGAL" class="h-9 w-full rounded-md border border-input bg-card px-3 text-sm outline-none focus:border-primary" />
      </div>
      <div>
        <label class="mb-1.5 block text-sm font-medium text-slate-700">错误关键字（模糊匹配，逗号分隔）</label>
        <input v-model="form.error_keyword" type="text" placeholder="如 blacklist, 签名" class="h-9 w-full rounded-md border border-input bg-card px-3 text-sm outline-none focus:border-primary" />
      </div>

      <!-- 动作配置 -->
      <div v-if="form.action === 'retry'" class="rounded-lg border border-border p-4">
        <p class="mb-2 text-xs font-medium text-muted-foreground">重试配置（可选）</p>
        <div class="grid grid-cols-3 gap-4">
          <div>
            <label class="mb-1.5 block text-sm font-medium text-slate-700">最大重试</label>
            <input v-model.number="form.action_config.max_retry" type="number" min="0" placeholder="3" class="h-9 w-full rounded-md border border-input bg-card px-3 text-sm outline-none focus:border-primary" />
          </div>
          <div>
            <label class="mb-1.5 block text-sm font-medium text-slate-700">延迟(秒)</label>
            <input v-model.number="form.action_config.delay_seconds" type="number" min="0" placeholder="5" class="h-9 w-full rounded-md border border-input bg-card px-3 text-sm outline-none focus:border-primary" />
          </div>
          <div>
            <label class="mb-1.5 block text-sm font-medium text-slate-700">退避倍率</label>
            <input v-model.number="form.action_config.backoff_rate" type="number" min="1" placeholder="2" class="h-9 w-full rounded-md border border-input bg-card px-3 text-sm outline-none focus:border-primary" />
          </div>
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
