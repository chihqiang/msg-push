// 消息模板 API
import { get, post, put, del } from './client'
import type { MessageTemplate, Paginated } from '@/types'

export interface TemplateListParams {
  page: number
  page_size: number
  channel_id?: number
  status?: number
  key?: string
}

export interface TemplateCreatePayload {
  code: string
  channel_id: number
  name: string
  content: string
  signature?: string
  remark?: string
}

export const listTemplates = (params: TemplateListParams) =>
  get<Paginated<MessageTemplate>>('/account/templates', params as unknown as Record<string, unknown>)

export const getTemplate = (id: number) => get<MessageTemplate>(`/account/templates/${id}`)

export const createTemplate = (payload: TemplateCreatePayload) => post<MessageTemplate>('/account/templates', payload)

export const updateTemplate = (id: number, payload: Partial<TemplateCreatePayload> & { status?: number }) =>
  put<MessageTemplate>(`/account/templates/${id}`, payload)

export const deleteTemplate = (id: number) => del<void>(`/account/templates/${id}`)
