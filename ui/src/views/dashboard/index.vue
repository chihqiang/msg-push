<script setup lang="ts">
import { computed } from 'vue'
import { useQuery } from '@tanstack/vue-query'
import { useRouter } from 'vue-router'
import { Loader2, Send, CheckCircle2, XCircle, Hourglass, Blocks, Share2, Server, TrendingUp } from '@lucide/vue'
import PageToolbar from '@/components/ui/PageToolbar.vue'
import DataTable from '@/components/ui/DataTable.vue'
import { getDashboard, getStatistics, getTopApplications, getRecentActivities } from '@/api/statistics'

const router = useRouter()

// Dashboard 概览
const { data: dashboard, isLoading } = useQuery({
  queryKey: ['dashboard'],
  queryFn: getDashboard,
})

// 统计（近 30 天趋势 + Top）
const { data: stats } = useQuery({
  queryKey: ['statistics'],
  queryFn: () => getStatistics({}),
})

const { data: topApps } = useQuery({
  queryKey: ['top-applications'],
  queryFn: getTopApplications,
})

const { data: recent } = useQuery({
  queryKey: ['recent-activities'],
  queryFn: () => getRecentActivities({ limit: 10 }),
})

// 顶部统计卡片
const cards = computed(() => {
  const d = dashboard.value
  return [
    {
      label: '今日发送',
      value: d?.today_push_count ?? 0,
      sub: `累计 ${d?.total_push_count ?? 0}`,
      icon: Send,
      color: 'text-cyan-500 bg-cyan-500/10',
    },
    {
      label: '今日成功',
      value: d?.today_success_count ?? 0,
      sub: `成功率 ${d?.today_success_rate ?? '-'}`,
      icon: CheckCircle2,
      color: 'text-emerald-500 bg-emerald-500/10',
    },
    {
      label: '今日失败',
      value: d?.today_failed_count ?? 0,
      sub: '需关注',
      icon: XCircle,
      color: 'text-destructive bg-destructive/10',
    },
    {
      label: '进行中',
      value: d?.today_in_progress_count ?? 0,
      sub: '待回执/待发送',
      icon: Hourglass,
      color: 'text-amber-500 bg-amber-500/10',
    },
    {
      label: '应用数',
      value: d?.total_applications ?? 0,
      sub: `启用 ${d?.active_applications ?? 0}`,
      icon: Blocks,
      color: 'text-blue-500 bg-blue-500/10',
    },
    {
      label: '通道数',
      value: d?.total_channels ?? 0,
      sub: `启用 ${d?.active_channels ?? 0}`,
      icon: Share2,
      color: 'text-violet-500 bg-violet-500/10',
    },
    {
      label: '服务商',
      value: d?.total_providers ?? 0,
      sub: `启用 ${d?.active_providers ?? 0}`,
      icon: Server,
      color: 'text-orange-500 bg-orange-500/10',
    },
  ]
})

// 近 30 天趋势（简化：日柱状条）
const maxDaily = computed(() => {
  const daily = stats.value?.daily ?? []
  let max = 1
  for (const d of daily) {
    if (d.total_count > max) max = d.total_count
  }
  return max
})

const trendData = computed(() => {
  const daily = stats.value?.daily ?? []
  // 取最近 14 天
  return daily.slice(-14)
})

// 最近动态表格列
const recentColumns = [
  { key: 'created_at', label: '时间', width: '170px' },
  { key: 'app_name', label: '应用' },
  { key: 'description', label: '描述' },
]

function goTo(path: string) {
  router.push(path)
}
</script>

<template>
  <div>
    <PageToolbar title="数据总览" description="消息推送核心指标与近期动态">
      <RouterLink
        to="/tasks"
        class="inline-flex items-center gap-1.5 rounded-md bg-primary px-3.5 py-2 text-sm font-medium text-white transition-colors hover:bg-primary-hover"
      >
        <TrendingUp class="h-4 w-4" />
        查看任务
      </RouterLink>
    </PageToolbar>

    <!-- 统计卡片 -->
    <div v-if="isLoading" class="flex justify-center py-16">
      <Loader2 class="h-6 w-6 animate-spin text-muted-foreground" />
    </div>

    <div v-else class="grid grid-cols-2 gap-4 sm:grid-cols-3 xl:grid-cols-7">
      <div
        v-for="card in cards"
        :key="card.label"
        class="card cursor-pointer p-4 transition-shadow hover:shadow-md"
        @click="goTo(card.label === '应用数' ? '/apps' : card.label === '通道数' ? '/channels' : card.label === '服务商' ? '/providers' : '/tasks')"
      >
        <div class="flex items-center justify-between">
          <div class="flex h-9 w-9 items-center justify-center rounded-lg" :class="card.color">
            <component :is="card.icon" class="h-4.5 w-4.5" />
          </div>
        </div>
        <div class="mt-3 text-2xl font-semibold text-foreground">{{ card.value }}</div>
        <div class="text-sm text-muted-foreground">{{ card.label }}</div>
        <div class="mt-0.5 text-xs text-muted-foreground/80">{{ card.sub }}</div>
      </div>
    </div>

    <!-- 趋势 + 最近动态 -->
    <div class="mt-6 grid gap-4 lg:grid-cols-3">
      <!-- 近 14 天趋势 -->
      <div class="card p-5 lg:col-span-2">
        <h3 class="mb-4 text-base font-semibold text-foreground">近 14 天推送趋势</h3>
        <div v-if="trendData.length === 0" class="py-10 text-center text-sm text-muted-foreground">暂无数据</div>
        <div v-else class="flex h-48 items-end gap-1.5">
          <div v-for="day in trendData" :key="day.date" class="group flex flex-1 flex-col items-center">
            <div
              class="flex w-full flex-col items-center justify-end rounded-t-sm transition-all"
              style="min-height: 4px"
            >
              <div
                class="w-full rounded-t-sm bg-gradient-to-t from-cyan-500 to-blue-400 transition-all group-hover:from-cyan-600"
                :style="{
                  height: `${Math.max(4, Math.round((day.total_count / maxDaily) * 140))}px`,
                }"
                :title="`${day.date}: ${day.total_count} 条`"
              />
            </div>
            <div class="mt-1 text-[10px] text-muted-foreground">{{ day.date.slice(5) }}</div>
          </div>
        </div>
      </div>

      <!-- 最近动态 -->
      <div class="card p-5 lg:col-span-1">
        <h3 class="mb-3 text-base font-semibold text-foreground">最近动态</h3>
        <DataTable :columns="recentColumns" :data="recent ?? []" :loading="false" :page-size="10" />
      </div>
    </div>

    <!-- Top 应用 -->
    <div class="mt-4 card p-5">
      <h3 class="mb-3 text-base font-semibold text-foreground">热门应用</h3>
      <div v-if="!topApps || topApps.length === 0" class="py-6 text-center text-sm text-muted-foreground">
        暂无应用数据
      </div>
      <div v-else class="grid gap-3 sm:grid-cols-2 lg:grid-cols-3">
        <div v-for="app in topApps" :key="app.id" class="rounded-lg border border-border p-4">
          <div class="flex items-center justify-between">
            <div class="min-w-0">
              <div class="truncate text-sm font-medium text-foreground">{{ app.app_name }}</div>
              <div class="text-xs text-muted-foreground">{{ app.app_id }}</div>
            </div>
          </div>
          <div class="mt-3 flex items-center gap-4 text-sm">
            <span class="text-muted-foreground">
              发送 <span class="font-medium text-foreground">{{ app.push_count }}</span>
            </span>
            <span class="text-muted-foreground">
              成功 <span class="font-medium text-emerald-600">{{ app.success_count }}</span>
            </span>
            <span class="ml-auto rounded-full bg-cyan-500/10 px-2 py-0.5 text-xs text-cyan-600">
              {{ app.success_rate }}
            </span>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
