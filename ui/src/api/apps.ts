// 应用管理 API（管理端，前缀 /account）
import { get, post, put, del } from './client'
import type { App, AppWithSecret, AppQuotaUsage, Paginated } from '@/types'

export interface AppListParams {
  page: number
  page_size: number
  key?: string
  status?: number
}

export interface AppCreatePayload {
  name: string
  remark?: string
  is_test?: boolean
  daily_quota?: number
  rate_limit?: number
}

export interface AppUpdatePayload {
  name?: string
  status?: number
  remark?: string
  is_test?: boolean
  daily_quota?: number
  rate_limit?: number
}

export const listApps = (params: AppListParams) =>
  get<Paginated<App>>('/account/apps', params as unknown as Record<string, unknown>)

export const getApp = (id: number) => get<App>(`/account/apps/${id}`)

export const createApp = (payload: AppCreatePayload) => post<AppWithSecret>('/account/apps', payload)

export const updateApp = (id: number, payload: AppUpdatePayload) => put<App>(`/account/apps/${id}`, payload)

export const deleteApp = (id: number) => del<void>(`/account/apps/${id}`)

export const resetAppSecret = (id: number) => post<AppWithSecret>(`/account/apps/${id}/reset-secret`)

export const getAppQuotaUsage = (id: number) => get<AppQuotaUsage>(`/account/apps/${id}/quota-usage`)
