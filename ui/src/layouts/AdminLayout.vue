<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { MessageSquareText, LogOut, Menu, X, ChevronDown, UserCog, Lightbulb } from '@lucide/vue'
import { useAuthStore } from '@/stores/auth'
import ProfileDialog from '@/components/account/ProfileDialog.vue'
import OnboardingWizard from '@/components/onboarding/OnboardingWizard.vue'
import { getMenuGroups } from '@/router'

const route = useRoute()
const router = useRouter()
const auth = useAuthStore()

const menuGroups = computed(() => getMenuGroups())
const activePath = computed(() => route.path)

const pageTitle = computed(() => {
  const meta = route.meta.title as string | undefined
  if (meta) return meta
  return ''
})

const sidebarOpen = ref(false)
const displayName = computed(() => auth.username || '管理员')
const avatarChar = computed(() => (displayName.value || 'A').charAt(0).toUpperCase())

// 用户下拉
const userMenuOpen = ref(false)
const userMenuRef = ref<HTMLElement | null>(null)

// 个人中心
const profileOpen = ref(false)

function openProfile() {
  userMenuOpen.value = false
  profileOpen.value = true
}

// 新手引导
const guideOpen = ref(false)

function toggleUserMenu() {
  userMenuOpen.value = !userMenuOpen.value
}

function handleLogout() {
  userMenuOpen.value = false
  auth.logout()
  router.push('/auth')
}

function onDocumentClick(e: MouseEvent) {
  if (userMenuRef.value && !userMenuRef.value.contains(e.target as Node)) {
    userMenuOpen.value = false
  }
}

onMounted(() => document.addEventListener('click', onDocumentClick))
onBeforeUnmount(() => document.removeEventListener('click', onDocumentClick))
</script>

<template>
  <div class="flex h-screen overflow-hidden bg-background">
    <!-- 移动端遮罩 -->
    <div
      v-if="sidebarOpen"
      class="fixed inset-0 z-40 bg-black/40 lg:hidden"
      @click="sidebarOpen = false"
    />

    <!-- 侧边栏 -->
    <aside
      class="fixed inset-y-0 left-0 z-50 flex w-60 flex-col bg-sidebar text-sidebar-foreground transition-transform lg:static lg:translate-x-0"
      :class="sidebarOpen ? 'translate-x-0' : '-translate-x-full'"
    >
      <!-- Logo -->
      <div class="flex h-16 items-center gap-2.5 border-b border-sidebar-border px-5">
        <div class="flex h-9 w-9 items-center justify-center rounded-lg bg-primary text-primary-foreground">
          <MessageSquareText class="h-5 w-5" />
        </div>
        <div class="leading-tight">
          <div class="text-sm font-semibold text-slate-900">msg-push</div>
          <div class="text-[11px] text-slate-500">消息推送控制台</div>
        </div>
        <button class="ml-auto text-slate-500 hover:text-slate-900 lg:hidden" @click="sidebarOpen = false">
          <X class="h-5 w-5" />
        </button>
      </div>

      <!-- 菜单 -->
      <nav class="flex-1 space-y-5 overflow-y-auto px-3 py-4">
        <div v-for="group in menuGroups" :key="group.label">
          <div class="px-3 pb-1.5 text-[11px] font-medium uppercase tracking-wider text-slate-500">
            {{ group.label }}
          </div>
          <div class="space-y-0.5">
            <RouterLink
              v-for="item in group.items"
              :key="item.path"
              :to="item.path"
              class="group flex items-center gap-2.5 rounded-md px-3 py-2 text-sm transition-colors"
              :class="
                activePath === item.path
                  ? 'bg-sidebar-active font-medium text-sidebar-active-foreground'
                  : 'text-sidebar-foreground hover:bg-sidebar-muted hover:text-slate-900'
              "
              @click="sidebarOpen = false"
            >
              <component :is="item.icon" class="h-4 w-4 shrink-0" />
              {{ item.label }}
            </RouterLink>
          </div>
        </div>
      </nav>

      <!-- 底部版本 -->
      <div class="border-t border-sidebar-border px-5 py-3 text-[11px] text-slate-500">v0.1.0</div>
    </aside>

    <!-- 右侧主区 -->
    <div class="flex min-w-0 flex-1 flex-col">
      <!-- 顶部栏 -->
      <header class="flex h-16 shrink-0 items-center gap-4 border-b border-border bg-card px-4 sm:px-6">
        <button class="rounded-md p-2 text-slate-500 hover:bg-muted lg:hidden" @click="sidebarOpen = true">
          <Menu class="h-5 w-5" />
        </button>
        <h1 class="text-base font-semibold text-foreground">{{ pageTitle }}</h1>

        <div class="ml-auto flex items-center gap-3">
          <!-- 新手引导按钮 -->
          <button
            class="inline-flex items-center gap-1.5 rounded-md border border-border px-2.5 py-1.5 text-xs font-medium text-muted-foreground transition-colors hover:bg-muted hover:text-foreground"
            title="新手引导：按消息类型配置完整链路"
            @click="guideOpen = true"
          >
            <Lightbulb class="h-3.5 w-3.5" />
            <span class="hidden sm:inline">新手引导</span>
          </button>
          <!-- 用户下拉 -->
          <div ref="userMenuRef" class="relative">
            <button
              class="flex items-center gap-2 rounded-md px-2 py-1.5 transition-colors hover:bg-muted"
              @click.stop="toggleUserMenu"
            >
              <div class="flex h-8 w-8 items-center justify-center rounded-full bg-primary text-sm font-medium text-primary-foreground">
                {{ avatarChar }}
              </div>
              <div class="hidden text-left leading-tight sm:block">
                <div class="text-sm font-medium text-foreground">{{ displayName }}</div>
                <div class="text-[11px] text-muted-foreground">管理员</div>
              </div>
              <ChevronDown
                class="h-4 w-4 text-muted-foreground transition-transform duration-200"
                :class="userMenuOpen ? 'rotate-180' : ''"
              />
            </button>

            <Transition
              enter-active-class="transition duration-100 ease-out"
              enter-from-class="translate-y-1 opacity-0"
              enter-to-class="translate-y-0 opacity-100"
              leave-active-class="transition duration-75 ease-in"
              leave-from-class="translate-y-0 opacity-100"
              leave-to-class="translate-y-1 opacity-0"
            >
              <div
                v-if="userMenuOpen"
                class="absolute right-0 top-full z-50 mt-2 w-56 overflow-hidden rounded-lg border border-border bg-card shadow-lg"
              >
                <div class="flex items-center gap-3 px-4 py-3">
                  <div class="flex h-9 w-9 items-center justify-center rounded-full bg-primary text-sm font-medium text-primary-foreground">
                    {{ avatarChar }}
                  </div>
                  <div class="min-w-0 leading-tight">
                    <div class="truncate text-sm font-medium text-foreground">{{ displayName }}</div>
                    <div class="truncate text-xs text-muted-foreground">管理员</div>
                  </div>
                </div>
                <div class="h-px bg-border" />
                <div class="p-1.5">
                  <button
                    class="flex w-full items-center gap-2.5 rounded-md px-2.5 py-2 text-sm text-foreground transition-colors hover:bg-muted"
                    @click="openProfile"
                  >
                    <UserCog class="h-4 w-4 text-muted-foreground" />
                    个人中心
                  </button>
                  <button
                    class="flex w-full items-center gap-2.5 rounded-md px-2.5 py-2 text-sm text-destructive transition-colors hover:bg-destructive/10"
                    @click="handleLogout"
                  >
                    <LogOut class="h-4 w-4" />
                    退出登录
                  </button>
                </div>
              </div>
            </Transition>
          </div>
        </div>
      </header>

      <!-- 内容区 -->
      <main class="flex-1 overflow-y-auto p-4 sm:p-6 lg:p-8">
        <RouterView />
      </main>
    </div>

    <!-- 个人中心弹窗 -->
    <ProfileDialog :open="profileOpen" @close="profileOpen = false" />

    <!-- 新手引导向导 -->
    <OnboardingWizard :open="guideOpen" @close="guideOpen = false" />
  </div>
</template>
