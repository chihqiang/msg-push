// Webhook 配置与日志 API
import { get, post, put, del } from './client'
import type { WebhookConfig, WebhookLog, Paginated } from '@/types'

export interface WebhookConfigCreatePayload {
  name: string
  app_id?: number
  webhook_url: string
  secret?: string
  events?: string // 逗号分隔
  status?: number
  retry_count?: number
  timeout?: number
  description?: string
}

export const listWebhookConfigs = (params: { page: number; page_size: number; status?: number; key?: string }) =>
  get<Paginated<WebhookConfig>>('/account/webhook-configs', params as unknown as Record<string, unknown>)

export const getWebhookConfig = (id: number) => get<WebhookConfig>(`/account/webhook-configs/${id}`)

export const createWebhookConfig = (payload: WebhookConfigCreatePayload) =>
  post<WebhookConfig>('/account/webhook-configs', payload)

export const updateWebhookConfig = (id: number, payload: Partial<WebhookConfigCreatePayload>) =>
  put<WebhookConfig>(`/account/webhook-configs/${id}`, payload)

export const deleteWebhookConfig = (id: number) => del<void>(`/account/webhook-configs/${id}`)

// Webhook 日志
export interface WebhookLogListParams {
  page: number
  page_size: number
  status?: string
  task_no?: string
  request_id?: string
}

export const listWebhookLogs = (params: WebhookLogListParams) =>
  get<Paginated<WebhookLog>>('/account/webhook-logs', params as unknown as Record<string, unknown>)

export const webhookLogsByTask = (taskId: string) => get<WebhookLog[]>(`/account/webhook-logs/task/${taskId}`)
