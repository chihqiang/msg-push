// 统计分析 API
import { get } from './client'
import type { StatisticsResponse, DashboardResponse, TopApplicationResponse, RecentActivityResponse } from '@/types'

export const getStatistics = (params?: { start_date?: string; end_date?: string; app_id?: number; channel_id?: number; message_type?: string }) =>
  get<StatisticsResponse>('/account/statistics', params as unknown as Record<string, unknown>)

export const getDashboard = () => get<DashboardResponse>('/account/statistics/dashboard')

export const getTopApplications = () => get<TopApplicationResponse[]>('/account/statistics/top-applications')

export const getRecentActivities = (params?: { limit?: number }) =>
  get<RecentActivityResponse[]>('/account/statistics/recent-activities', params as unknown as Record<string, unknown>)
