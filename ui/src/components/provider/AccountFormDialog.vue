<script setup lang="ts">
// 服务商账号创建/编辑弹窗（含动态配置字段）
import { computed, ref, watch } from 'vue'
import { Eye, EyeOff } from '@lucide/vue'
import Modal from '@/components/ui/Modal.vue'
import { useToastStore } from '@/stores/toast'
import { createProviderAccount, updateProviderAccount } from '@/api/providers'
import type { ProviderAccount, ProviderMeta, ConfigField } from '@/types'

const props = defineProps<{
  open: boolean
  account: ProviderAccount | null
  providers: ProviderMeta[]
  defaultProviderCode: string
}>()

const emit = defineEmits<{
  close: []
  saved: []
}>()

const toast = useToastStore()
const saving = ref(false)
const form = ref<{
  provider_code: string
  account_name: string
  config: Record<string, unknown>
  remark: string
  status?: number
}>({ provider_code: '', account_name: '', config: {}, remark: '' })

watch(
  () => props.open,
  (open) => {
    if (!open) return
    if (props.account) {
      const config = props.account.config ?? {}
      form.value = {
        provider_code: props.account.provider_code,
        account_name: props.account.account_name,
        config,
        remark: props.account.remark,
        status: props.account.status,
      }
    } else {
      form.value = { provider_code: props.defaultProviderCode, account_name: '', config: {}, remark: '' }
    }
  }
)

const currentConfigFields = computed<ConfigField[]>(() => {
  const meta = props.providers.find((p) => p.code === form.value.provider_code)
  return meta?.config_fields ?? []
})

const selectedProvider = computed<ProviderMeta | undefined>(() =>
  props.providers.find((p) => p.code === form.value.provider_code)
)

// 密码字段显隐（key -> 是否明文）
const showPassword = ref<Record<string, boolean>>({})

function togglePassword(key: string) {
  showPassword.value[key] = !showPassword.value[key]
}

function fieldType(field: ConfigField): string {
  if (field.type !== 'password') return field.type
  return showPassword.value[field.key] ? 'text' : 'password'
}

function fieldValue(key: string): unknown {
  return (form.value.config ?? {})[key]
}

function setFieldValue(key: string, value: unknown) {
  if (!form.value.config) form.value.config = {}
  form.value.config[key] = value
}

async function submit() {
  if (!form.value.account_name || !form.value.provider_code) {
    toast.error('请填写账号名称并选择服务商')
    return
  }
  saving.value = true
  try {
    if (props.account) {
      await updateProviderAccount(props.account.id, {
        account_name: form.value.account_name,
        config: form.value.config,
        status: form.value.status,
        remark: form.value.remark,
      })
      toast.success('服务商账号已更新')
    } else {
      await createProviderAccount(form.value)
      toast.success('服务商账号已创建')
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
  <Modal :open="open" :title="account ? '编辑服务商账号' : '新建服务商账号'" width="42rem" @close="emit('close')">
    <div class="space-y-5">
      <!-- 基础信息 -->
      <section class="rounded-lg border border-border bg-muted/30 p-4">
        <h4 class="mb-3 flex items-center gap-1.5 text-xs font-semibold uppercase tracking-wide text-muted-foreground">
          <span class="inline-block h-1 w-1 rounded-full bg-primary" /> 基础信息
        </h4>
        <div class="grid grid-cols-2 gap-4">
          <div>
            <label class="mb-1.5 block text-sm font-medium text-foreground">服务商 <span class="text-destructive">*</span></label>
            <select
              v-model="form.provider_code"
              :disabled="!!account"
              class="h-9 w-full rounded-md border border-input bg-card px-3 text-sm text-foreground outline-none transition-colors placeholder:text-muted-foreground focus:border-primary focus:ring-2 focus:ring-primary/20 disabled:cursor-not-allowed disabled:opacity-60"
            >
              <option v-for="p in providers" :key="p.code" :value="p.code">{{ p.name }}</option>
            </select>
            <p v-if="selectedProvider?.description && !account" class="mt-1.5 text-xs text-muted-foreground">{{ selectedProvider.description }}</p>
          </div>
          <div>
            <label class="mb-1.5 block text-sm font-medium text-foreground">账号名称 <span class="text-destructive">*</span></label>
            <input
              v-model="form.account_name"
              type="text"
              class="h-9 w-full rounded-md border border-input bg-card px-3 text-sm text-foreground outline-none transition-colors placeholder:text-muted-foreground focus:border-primary focus:ring-2 focus:ring-primary/20"
            />
          </div>
        </div>
      </section>

      <!-- 服务商配置 -->
      <section v-if="currentConfigFields.length" class="rounded-lg border border-border bg-muted/30 p-4">
        <h4 class="mb-1 flex items-center gap-1.5 text-xs font-semibold uppercase tracking-wide text-muted-foreground">
          <span class="inline-block h-1 w-1 rounded-full bg-primary" /> 服务商配置
        </h4>
        <p class="mb-3 text-xs text-muted-foreground">填写 {{ selectedProvider?.name ?? '服务商' }} 的接入参数，保存后将整合为 JSON 存储于账号配置。</p>
        <div class="grid grid-cols-2 gap-4">
          <div v-for="field in currentConfigFields" :key="field.key">
            <label class="mb-1.5 block text-sm font-medium text-foreground">
              {{ field.label }}
              <span v-if="field.required" class="text-destructive">*</span>
            </label>
            <select
              v-if="field.type === 'select'"
              :value="(fieldValue(field.key) as string) ?? field.default_value ?? ''"
              class="h-9 w-full rounded-md border border-input bg-card px-3 text-sm text-foreground outline-none transition-colors focus:border-primary focus:ring-2 focus:ring-primary/20"
              @change="setFieldValue(field.key, ($event.target as HTMLSelectElement).value)"
            >
              <option value="">请选择</option>
              <option v-for="opt in field.options ?? []" :key="opt.value" :value="opt.value">{{ opt.label }}</option>
            </select>
            <div v-else-if="field.type === 'password'" class="relative">
              <input
                :type="fieldType(field)"
                :value="(fieldValue(field.key) as string) ?? field.default_value ?? ''"
                :placeholder="field.placeholder"
                autocomplete="new-password"
                class="h-9 w-full rounded-md border border-input bg-card pr-9 pl-3 text-sm text-foreground outline-none transition-colors placeholder:text-muted-foreground focus:border-primary focus:ring-2 focus:ring-primary/20"
                @input="setFieldValue(field.key, ($event.target as HTMLInputElement).value)"
              />
              <button
                type="button"
                class="absolute inset-y-0 right-0 flex items-center pr-2.5 text-muted-foreground transition-colors hover:text-foreground"
                @click="togglePassword(field.key)"
              >
                <EyeOff v-if="showPassword[field.key]" class="h-4 w-4" />
                <Eye v-else class="h-4 w-4" />
              </button>
            </div>
            <input
              v-else
              :type="field.type === 'number' ? 'number' : 'text'"
              :value="(fieldValue(field.key) as string) ?? field.default_value ?? ''"
              :placeholder="field.placeholder"
              class="h-9 w-full rounded-md border border-input bg-card px-3 text-sm text-foreground outline-none transition-colors placeholder:text-muted-foreground focus:border-primary focus:ring-2 focus:ring-primary/20"
              @input="setFieldValue(field.key, ($event.target as HTMLInputElement).value)"
            />
            <p v-if="field.description" class="mt-1.5 text-xs text-muted-foreground">{{ field.description }}</p>
          </div>
        </div>
      </section>

      <!-- 其他 -->
      <section class="rounded-lg border border-border bg-muted/30 p-4">
        <h4 class="mb-3 flex items-center gap-1.5 text-xs font-semibold uppercase tracking-wide text-muted-foreground">
          <span class="inline-block h-1 w-1 rounded-full bg-primary" /> 其他
        </h4>
        <div>
          <label class="mb-1.5 block text-sm font-medium text-foreground">备注</label>
          <input
            v-model="form.remark"
            type="text"
            placeholder="选填，如：主账号 / 备用通道"
            class="h-9 w-full rounded-md border border-input bg-card px-3 text-sm text-foreground outline-none transition-colors placeholder:text-muted-foreground focus:border-primary focus:ring-2 focus:ring-primary/20"
          />
        </div>
      </section>
    </div>
    <template #footer>
      <div class="flex justify-end gap-2.5">
        <button class="h-9 rounded-md border border-border px-4 text-sm text-muted-foreground transition-colors hover:bg-muted hover:text-foreground" @click="emit('close')">取消</button>
        <button class="inline-flex h-9 items-center gap-1.5 rounded-md bg-primary px-4 text-sm font-medium text-white transition-colors hover:bg-primary-hover disabled:opacity-50" :disabled="saving" @click="submit">
          {{ saving ? '保存中…' : '保存' }}
        </button>
      </div>
    </template>
  </Modal>
</template>
