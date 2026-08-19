// 日志查询 API（推送日志/回调日志/Webhook 日志）
import { get } from './client'
import type { PushLog, CallbackLog, Paginated } from '@/types'

export interface LogListParams {
  page: number
  page_size: number
  task_no?: string
  request_id?: string
  app_id?: number
  provider_account_id?: number
  status?: string
  start_date?: string
  end_date?: string
}

export const listPushLogs = (params: LogListParams) =>
  get<Paginated<PushLog>>('/account/logs', params as unknown as Record<string, unknown>)

export const pushLogsByTask = (taskId: string) => get<PushLog[]>(`/account/logs/task/${taskId}`)

export const listCallbacks = (params: LogListParams) =>
  get<Paginated<CallbackLog>>('/account/callbacks', params as unknown as Record<string, unknown>)

export const callbacksByTask = (taskId: string) => get<CallbackLog[]>(`/account/callbacks/task/${taskId}`)
