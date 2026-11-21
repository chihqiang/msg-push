// 失败规则 API
import { get, post, put, del } from './client'
import type { FailureRule, Paginated } from '@/types'

export interface FailureRuleListParams {
  page: number
  page_size: number
  scene?: string
  status?: number
  key?: string
}

export interface FailureRuleCreatePayload {
  name: string
  scene: string
  provider_code?: string
  message_type?: string
  error_code?: string
  error_keyword?: string
  action: string
  action_config?: Record<string, unknown>
  priority?: number
  status?: number
}

export const listFailureRules = (params: FailureRuleListParams) =>
  get<Paginated<FailureRule>>('/account/failure-rules', params as unknown as Record<string, unknown>)

export const getFailureRule = (id: number) => get<FailureRule>(`/account/failure-rules/${id}`)

export const createFailureRule = (payload: FailureRuleCreatePayload) =>
  post<FailureRule>('/account/failure-rules', payload)

export const updateFailureRule = (id: number, payload: Partial<FailureRuleCreatePayload>) =>
  put<FailureRule>(`/account/failure-rules/${id}`, payload)

export const deleteFailureRule = (id: number) => del<void>(`/account/failure-rules/${id}`)

export const failureRuleOptions = () => get<unknown>('/account/failure-rules/options')

export const refreshFailureRuleCache = () => post<void>('/account/failure-rules/refresh-cache')
