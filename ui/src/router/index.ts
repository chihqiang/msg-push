// 路由：认证守卫 + 侧边栏菜单生成
import { createRouter, createWebHistory } from 'vue-router'
import type { RouteRecordRaw, RouteMeta } from 'vue-router'
import {
  LayoutDashboard,
  Blocks,
  Share2,
  FileText,
  Server,
  ShieldAlert,
  Send,
  ScrollText,
  Webhook,
} from '@lucide/vue'
import { useAuthStore } from '@/stores/auth'

declare module 'vue-router' {
  interface RouteMeta {
    title?: string
    icon?: unknown
    menuGroup?: string
    public?: boolean
    requiresAuth?: boolean
  }
}

// 管理页路由（含侧边栏菜单元数据）
const adminRoutes: RouteRecordRaw[] = [
  {
    path: '/dashboard',
    component: () => import('@/views/dashboard/index.vue'),
    meta: { title: '数据总览', icon: LayoutDashboard, menuGroup: '概览' },
  },
  {
    path: '/apps',
    component: () => import('@/views/apps/index.vue'),
    meta: { title: '应用管理', icon: Blocks, menuGroup: '消息管理' },
  },
  {
    path: '/channels',
    component: () => import('@/views/channels/index.vue'),
    meta: { title: '通道管理', icon: Share2, menuGroup: '消息管理' },
  },
  {
    path: '/templates',
    component: () => import('@/views/templates/index.vue'),
    meta: { title: '模板管理', icon: FileText, menuGroup: '消息管理' },
  },
  {
    path: '/providers',
    component: () => import('@/views/providers/index.vue'),
    meta: { title: '服务商管理', icon: Server, menuGroup: '资源管理' },
  },
  {
    path: '/failure-rules',
    component: () => import('@/views/failure-rules/index.vue'),
    meta: { title: '失败规则', icon: ShieldAlert, menuGroup: '资源管理' },
  },
  {
    path: '/webhooks',
    component: () => import('@/views/webhooks/index.vue'),
    meta: { title: 'Webhook 配置', icon: Webhook, menuGroup: '资源管理' },
  },
  {
    path: '/tasks',
    component: () => import('@/views/tasks/index.vue'),
    meta: { title: '任务查询', icon: Send, menuGroup: '运营监控' },
  },
  {
    path: '/logs',
    component: () => import('@/views/logs/index.vue'),
    meta: { title: '日志查询', icon: ScrollText, menuGroup: '运营监控' },
  },
]

const routes: RouteRecordRaw[] = [
  {
    path: '/auth',
    component: () => import('@/views/auth/index.vue'),
    meta: { title: '登录', public: true },
  },
  {
    path: '/',
    component: () => import('@/layouts/AdminLayout.vue'),
    meta: { requiresAuth: true },
    redirect: '/dashboard',
    children: adminRoutes,
  },
  {
    path: '/:pathMatch(.*)*',
    redirect: '/dashboard',
  },
]

const router = createRouter({
  history: createWebHistory(),
  routes,
})

// 认证守卫：登录态处理 + 未登录跳登录页
router.beforeEach((to) => {
  const auth = useAuthStore()

  // 清理已过期的登录态：cookie 里有 token 但已过期时清空，避免残留过期登录态
  if (auth.accessToken && !auth.isLoggedIn) {
    auth.logout()
  }

  if (to.meta.public) return true
  if (to.meta.requiresAuth && !auth.isLoggedIn) {
    return { path: '/auth', query: { redirect: to.fullPath } }
  }
  if (auth.isLoggedIn && to.path === '/auth') {
    return { path: '/dashboard' }
  }
  return true
})

// ===== 侧边栏菜单生成 =====

export interface MenuItem {
  path: string
  label: string
  icon: unknown
  action?: string
}

export interface MenuGroup {
  label: string
  items: MenuItem[]
}

export function getMenuGroups(): MenuGroup[] {
  const byGroup = new Map<string, MenuItem[]>()
  for (const r of adminRoutes) {
    const meta = r.meta as RouteMeta
    if (!meta?.menuGroup || !meta.icon) continue
    const items = byGroup.get(meta.menuGroup) ?? []
    items.push({ path: r.path, label: meta.title ?? r.path, icon: meta.icon })
    byGroup.set(meta.menuGroup, items)
  }
  return Array.from(byGroup.entries()).map(([label, items]) => ({ label, items }))
}

export default router
