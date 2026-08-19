<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { MessageSquareText, Loader2 } from '@lucide/vue'
import { useAuthStore } from '@/stores/auth'
import { useToastStore } from '@/stores/toast'

const router = useRouter()
const auth = useAuthStore()
const toast = useToastStore()

const username = ref('')
const password = ref('')
const loading = ref(false)

async function onSubmit() {
  if (!username.value || !password.value) {
    toast.error('请输入用户名和密码')
    return
  }
  loading.value = true
  try {
    await auth.login(username.value, password.value)
    toast.success('登录成功')
    router.push('/dashboard')
  } catch (e) {
    const message = e instanceof Error ? e.message : '登录失败'
    toast.error(message)
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="relative flex min-h-screen items-center justify-center overflow-hidden bg-linear-to-br from-slate-50 via-white to-cyan-50/60 p-4">
    <!-- 浅色网格纹理 -->
    <div class="tech-grid-light pointer-events-none absolute inset-0" />

    <!-- 柔和光斑 -->
    <div class="pointer-events-none absolute inset-0">
      <div class="absolute -left-40 -top-40 h-112 w-md rounded-full bg-cyan-400/20 blur-3xl" />
      <div class="absolute -bottom-40 -right-40 h-112 w-md rounded-full bg-blue-400/20 blur-3xl" />
    </div>

    <div class="relative w-full max-w-md">
      <!-- 顶部品牌 -->
      <div class="mb-8 flex flex-col items-center gap-3">
        <div class="relative flex h-14 w-14 items-center justify-center rounded-2xl bg-linear-to-br from-cyan-500 to-blue-600 text-white shadow-[0_8px_30px_-6px_rgba(6,182,212,0.5)]">
          <MessageSquareText class="h-7 w-7" />
          <div class="absolute inset-0 rounded-2xl ring-1 ring-cyan-400/50" />
        </div>
        <div class="text-center">
          <h1 class="text-2xl font-bold tracking-wide text-slate-900">msg-push</h1>
          <p class="mt-1 text-sm tracking-[0.2em] text-cyan-600/70">消息推送控制台</p>
        </div>
      </div>

      <div class="relative overflow-hidden rounded-2xl border border-slate-200/80 bg-white/70 p-6 shadow-[0_8px_40px_-12px_rgba(15,23,42,0.15)] backdrop-blur-xl">
        <div class="absolute inset-x-0 top-0 mx-auto h-0.5 w-3/5 rounded-full bg-linear-to-r from-transparent via-cyan-400/80 to-transparent" />

        <form class="space-y-4" @submit.prevent="onSubmit">
          <div>
            <label class="mb-1.5 block text-sm font-medium text-slate-700">用户名</label>
            <input
              v-model="username"
              type="text"
              autocomplete="username"
              placeholder="请输入用户名"
              class="h-10 w-full rounded-lg border border-slate-200 bg-white/80 px-3.5 text-sm text-slate-900 outline-none transition-colors placeholder:text-slate-400 focus:border-cyan-400 focus:ring-2 focus:ring-cyan-400/20"
            />
          </div>
          <div>
            <label class="mb-1.5 block text-sm font-medium text-slate-700">密码</label>
            <input
              v-model="password"
              type="password"
              autocomplete="current-password"
              placeholder="请输入密码"
              class="h-10 w-full rounded-lg border border-slate-200 bg-white/80 px-3.5 text-sm text-slate-900 outline-none transition-colors placeholder:text-slate-400 focus:border-cyan-400 focus:ring-2 focus:ring-cyan-400/20"
            />
          </div>

          <button
            type="submit"
            :disabled="loading"
            class="flex h-10 w-full items-center justify-center gap-2 rounded-lg bg-linear-to-r from-cyan-500 to-blue-600 text-sm font-medium text-white shadow-[0_4px_14px_-4px_rgba(6,182,212,0.6)] transition-all hover:from-cyan-600 hover:to-blue-700 disabled:opacity-60"
          >
            <Loader2 v-if="loading" class="h-4 w-4 animate-spin" />
            {{ loading ? '登录中…' : '登 录' }}
          </button>
        </form>

        <p class="mt-4 text-center text-xs text-slate-400">默认账号 admin / admin123</p>
      </div>
    </div>
  </div>
</template>
