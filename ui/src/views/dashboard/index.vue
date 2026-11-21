<script setup lang="ts">
import { computed } from 'vue'
import { useQuery } from '@tanstack/vue-query'
import { useRouter } from 'vue-router'
import { Loader2, Send, CheckCircle2, XCircle, Hourglass, Blocks, Share2, Server, TrendingUp } from '@lucide/vue'
import PageToolbar from '@/components/ui/PageToolbar.vue'
import DataTable from '@/components/ui/DataTable.vue'
import { getDashboard, getStatistics, getRecentActivities } from '@/api/statistics'

const router = useRouter()

// Dashboard 概览
const { data: dashboard, isLoading } = useQuery({
  queryKey: ['dashboard'],
  queryFn: getDashboard,
})

// 趋势图独立请求近 14 天（柱高基准基于展示数据自身）
const trendDays = 14
const today = new Date()
const trendStart = new Date(today)
trendStart.setDate(today.getDate() - (trendDays - 1))
const fmtDate = (d: Date) =>
  `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}-${String(d.getDate()).padStart(2, '0')}`
const { data: trendStats } = useQuery({
  queryKey: ['statistics-trend', trendDays],
  queryFn: () => getStatistics({ start_date: fmtDate(trendStart), end_date: fmtDate(today) }),
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

// 近 14 天趋势：柱高基准基于展示数据自身最大值
const trendData = computed(() => trendStats.value?.daily ?? [])

const maxDaily = computed(() => {
  let max = 1
  for (const d of trendData.value) {
    if (d.total_count > max) max = d.total_count
  }
  return max
})

// 柱高比例：小数据（最大日 < 100）用固定基准，避免全部顶满看不出差异
const barHeight = (total: number) => {
  const max = maxDaily.value
  const base = max < 100 ? 100 : max
  return `${Math.max(4, Math.round((total / base) * 140))}px`
}

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

    <!-- 近 14 天趋势（整行） -->
    <div class="mt-6 card p-5">
      <h3 class="mb-4 text-base font-semibold text-foreground">近 14 天推送趋势</h3>
      <div v-if="trendData.length === 0" class="py-10 text-center text-sm text-muted-foreground">暂无数据</div>
      <div v-else class="flex h-48 items-end gap-1.5">
        <div v-for="day in trendData" :key="day.date" class="group flex flex-1 flex-col items-center">
          <div class="flex w-full flex-col items-center justify-end rounded-t-sm transition-all" style="min-height: 4px">
            <div
              class="w-full rounded-t-sm bg-gradient-to-t from-cyan-500 to-blue-400 transition-all group-hover:from-cyan-600"
              :style="{ height: barHeight(day.total_count) }"
              :title="`${day.date}: ${day.total_count} 条`"
            />
          </div>
          <div class="mt-1 text-[10px] text-muted-foreground">{{ day.date.slice(5) }}</div>
        </div>
      </div>
    </div>

    <!-- 最近动态（整行，位于趋势下方） -->
    <div class="mt-4 card p-5">
      <h3 class="mb-3 text-base font-semibold text-foreground">最近动态</h3>
      <DataTable :columns="recentColumns" :data="recent ?? []" :loading="false" :page-size="10" />
    </div>
  </div>
</template>
