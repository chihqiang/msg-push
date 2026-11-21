// 通道管理 API（含健康历史、模板绑定、签名映射）
import { get, post, put, del } from './client'
import type {
  Channel,
  ChannelTestResult,
  ChannelHealthHistory,
  Paginated,
  ChannelBinding,
  ChannelSignatureMapping,
  AvailableProviderTemplate,
  AvailableProviderSignature,
  ParamMappingItem,
} from '@/types'

export interface ChannelListParams {
  page: number
  page_size: number
  type?: string
  status?: number
  key?: string
}

export interface ChannelCreatePayload {
  code: string
  name: string
  type: string
  config?: string
  remark?: string
}

export interface ChannelUpdatePayload {
  name?: string
  status?: number
  remark?: string
}

export const listChannels = (params: ChannelListParams) =>
  get<Paginated<Channel>>('/account/channels', params as unknown as Record<string, unknown>)

export const getChannel = (id: number) => get<Channel>(`/account/channels/${id}`)

export const createChannel = (payload: ChannelCreatePayload) => post<Channel>('/account/channels', payload)

export const updateChannel = (id: number, payload: ChannelUpdatePayload) => put<Channel>(`/account/channels/${id}`, payload)

export const deleteChannel = (id: number) => del<void>(`/account/channels/${id}`)

// 通道测试发送
export const testChannel = (id: number, payload: { receiver: string; content?: string }) =>
  post<ChannelTestResult>(`/account/channels/${id}/test`, payload)

// 健康历史
export const channelHealthHistory = (id: number, params: { page: number; page_size: number }) =>
  get<Paginated<ChannelHealthHistory>>(`/account/channels/${id}/health-history`, params as unknown as Record<string, unknown>)

// ===== 通道-模板绑定 =====

export interface ChannelBindingPayload {
  provider_template_id: number
  provider_id: number
  param_mapping?: ParamMappingItem[]
  weight?: number
  priority?: number
  status?: number
  is_active?: number
  auto_disable_on_fail?: boolean
  auto_disable_threshold?: number
}

export const channelBindings = (id: number, params: { page: number; page_size: number }) =>
  get<Paginated<ChannelBinding>>(`/account/channels/${id}/bindings`, params as unknown as Record<string, unknown>)

export const channelAvailableTemplates = (id: number) =>
  get<AvailableProviderTemplate[]>(`/account/channels/${id}/available-templates`)

export const createChannelBinding = (channelId: number, payload: ChannelBindingPayload) =>
  post<ChannelBinding>(`/account/channels/${channelId}/bindings`, payload)

export const updateChannelBinding = (channelId: number, bindingId: number, payload: Partial<ChannelBindingPayload>) =>
  put<ChannelBinding>(`/account/channels/${channelId}/bindings/${bindingId}`, payload)

export const deleteChannelBinding = (channelId: number, bindingId: number) =>
  del<void>(`/account/channels/${channelId}/bindings/${bindingId}`)

// ===== 通道-签名映射 =====

export interface ChannelSignatureMappingPayload {
  signature_name: string
  provider_signature_id: number
  provider_id: number
  status?: number
}

export const channelSignatureMappings = (id: number, params: { page: number; page_size: number }) =>
  get<Paginated<ChannelSignatureMapping>>(`/account/channels/${id}/signature-mappings`, params as unknown as Record<string, unknown>)

export const channelAvailableSignatures = (id: number) =>
  get<AvailableProviderSignature[]>(`/account/channels/${id}/available-signatures`)

export const createChannelSignatureMapping = (channelId: number, payload: ChannelSignatureMappingPayload) =>
  post<ChannelSignatureMapping>(`/account/channels/${channelId}/signature-mappings`, payload)

export const updateChannelSignatureMapping = (channelId: number, mappingId: number, payload: Partial<ChannelSignatureMappingPayload>) =>
  put<ChannelSignatureMapping>(`/account/channels/${channelId}/signature-mappings/${mappingId}`, payload)

export const deleteChannelSignatureMapping = (channelId: number, mappingId: number) =>
  del<void>(`/account/channels/${channelId}/signature-mappings/${mappingId}`)
