<script setup lang="ts">
// 个人中心弹窗：显示当前账号信息 + 修改密码
import { computed, ref, watch } from 'vue'
import { useRouter } from 'vue-router'
import { useQuery } from '@tanstack/vue-query'
import { UserRound, ShieldCheck, KeyRound, CalendarDays, Save } from '@lucide/vue'
import Modal from '@/components/ui/Modal.vue'
import StatusBadge from '@/components/ui/StatusBadge.vue'
import { useToastStore } from '@/stores/toast'
import { useAuthStore } from '@/stores/auth'
import { getAccountProfile, changePassword } from '@/api/account'

const props = defineProps<{
  open: boolean
}>()

const emit = defineEmits<{
  close: []
}>()

const router = useRouter()
const toast = useToastStore()
const auth = useAuthStore()

// 当前账号资料（打开弹窗时自动拉取）
const { data: profile, isFetching } = useQuery({
  queryKey: ['account-profile'],
  queryFn: getAccountProfile,
  enabled: () => props.open,
})

// 修改密码表单
const oldPassword = ref('')
const newPassword = ref('')
const confirmPassword = ref('')
const saving = ref(false)

function resetForm() {
  oldPassword.value = ''
  newPassword.value = ''
  confirmPassword.value = ''
}

watch(
  () => props.open,
  (open) => {
    if (!open) return
    resetForm()
  }
)

// 响应式计算头像首字符：profile 为异步加载，必须用 computed 才能随数据更新
const avatarChar = computed(() => (profile.value?.name || profile.value?.username || 'A').charAt(0).toUpperCase())

async function onSubmit() {
  if (!oldPassword.value || !newPassword.value) {
    toast.error('请填写旧密码与新密码')
    return
  }
  if (newPassword.value.length < 6) {
    toast.error('新密码长度至少 6 位')
    return
  }
  if (newPassword.value !== confirmPassword.value) {
    toast.error('两次输入的新密码不一致')
    return
  }
  saving.value = true
  try {
    await changePassword({ old_password: oldPassword.value, new_password: newPassword.value })
    toast.success('密码修改成功，请重新登录')
    resetForm()
    emit('close')
    // 密码已变更：清空登录态并跳转登录页，强制用新密码重新登录
    auth.logout()
    router.push('/auth')
  } catch (e) {
    toast.error(e instanceof Error ? e.message : '修改失败')
  } finally {
    saving.value = false
  }
}
</script>

<template>
  <Modal :open="open" title="个人中心" width="32rem" @close="emit('close')">
    <div v-if="isFetching" class="py-10 text-center text-sm text-muted-foreground">加载中…</div>

    <!-- 账号信息 -->
    <div v-else class="space-y-5">
      <div class="flex items-center gap-4">
        <div class="flex h-14 w-14 items-center justify-center rounded-full bg-primary text-xl font-semibold text-primary-foreground">
          {{ avatarChar }}
        </div>
        <div class="leading-tight">
          <div class="text-lg font-semibold text-foreground">{{ profile?.name || profile?.username }}</div>
          <div class="text-sm text-muted-foreground">{{ profile?.username }}</div>
        </div>
      </div>

      <div class="grid grid-cols-2 gap-3">
        <div class="flex items-center gap-2 rounded-lg border border-border px-3 py-2.5">
          <UserRound class="h-4 w-4 text-muted-foreground" />
          <div class="leading-tight">
            <div class="text-[11px] text-muted-foreground">账号 ID</div>
            <div class="text-sm font-medium text-foreground">#{{ profile?.id }}</div>
          </div>
        </div>
        <div class="flex items-center gap-2 rounded-lg border border-border px-3 py-2.5">
          <ShieldCheck class="h-4 w-4 text-muted-foreground" />
          <div class="leading-tight">
            <div class="text-[11px] text-muted-foreground">状态</div>
            <div class="text-sm font-medium text-foreground">
              <StatusBadge :value="(profile?.status ?? 0) === 1" active-text="启用" inactive-text="禁用" />
            </div>
          </div>
        </div>
        <div class="col-span-2 flex items-center gap-2 rounded-lg border border-border px-3 py-2.5">
          <CalendarDays class="h-4 w-4 text-muted-foreground" />
          <div class="leading-tight">
            <div class="text-[11px] text-muted-foreground">创建时间</div>
            <div class="text-sm font-medium text-foreground">{{ profile?.created_at || '—' }}</div>
          </div>
        </div>
      </div>

      <!-- 修改密码 -->
      <div class="rounded-lg border border-border p-4">
        <div class="mb-3 flex items-center gap-2">
          <KeyRound class="h-4 w-4 text-muted-foreground" />
          <span class="text-sm font-medium text-foreground">修改密码</span>
        </div>
        <div class="space-y-3">
          <div>
            <label class="mb-1.5 block text-sm font-medium text-slate-700">旧密码 *</label>
            <input v-model="oldPassword" type="password" placeholder="请输入旧密码" class="h-9 w-full rounded-md border border-input bg-card px-3 text-sm outline-none focus:border-primary" />
          </div>
          <div>
            <label class="mb-1.5 block text-sm font-medium text-slate-700">新密码 *</label>
            <input v-model="newPassword" type="password" placeholder="至少 6 位" class="h-9 w-full rounded-md border border-input bg-card px-3 text-sm outline-none focus:border-primary" />
          </div>
          <div>
            <label class="mb-1.5 block text-sm font-medium text-slate-700">确认新密码 *</label>
            <input v-model="confirmPassword" type="password" placeholder="再次输入新密码" class="h-9 w-full rounded-md border border-input bg-card px-3 text-sm outline-none focus:border-primary" />
          </div>
        </div>
      </div>
    </div>

    <template #footer>
      <div class="flex justify-end gap-2.5">
        <button class="h-9 rounded-md border border-border px-4 text-sm text-muted-foreground hover:bg-muted" @click="emit('close')">取消</button>
        <button class="inline-flex h-9 items-center gap-1.5 rounded-md bg-primary px-4 text-sm font-medium text-white hover:bg-primary-hover disabled:opacity-50" :disabled="saving" @click="onSubmit">
          <Save class="h-4 w-4" />
          {{ saving ? '保存中…' : '保存' }}
        </button>
      </div>
    </template>
  </Modal>
</template>
