// 服务商管理 API（账号/签名/供应商模板）
import { get, post, put, del } from './client'
import type {
  ProviderAccount,
  ProviderMeta,
  ProviderSignature,
  ProviderTemplate,
  Paginated,
} from '@/types'

// ===== 服务商元信息 =====

export const availableProviders = () => get<ProviderMeta[]>('/account/provider-accounts/available')

export const providerConfigFields = (code: string) => get<unknown[]>(`/account/provider-config-fields/${code}`)

// ===== 服务商账号 =====

export interface ProviderAccountListParams {
  page: number
  page_size: number
  provider_type?: string
  status?: number
  key?: string
}

export interface ProviderAccountCreatePayload {
  provider_code: string
  account_name: string
  config: Record<string, unknown>
  remark?: string
}

export interface ProviderAccountUpdatePayload {
  account_name?: string
  config?: Record<string, unknown>
  status?: number
  remark?: string
}

export const listProviderAccounts = (params: ProviderAccountListParams) =>
  get<Paginated<ProviderAccount>>('/account/provider-accounts', params as unknown as Record<string, unknown>)

export const getProviderAccount = (id: number) => get<ProviderAccount>(`/account/provider-accounts/${id}`)

export const createProviderAccount = (payload: ProviderAccountCreatePayload) =>
  post<ProviderAccount>('/account/provider-accounts', payload)

export const updateProviderAccount = (id: number, payload: ProviderAccountUpdatePayload) =>
  put<ProviderAccount>(`/account/provider-accounts/${id}`, payload)

export const deleteProviderAccount = (id: number) => del<void>(`/account/provider-accounts/${id}`)

// ===== 服务商签名 =====

export interface ProviderSignatureCreatePayload {
  provider_account_id: number
  signature_code: string
  signature_name: string
}

export const listProviderSignatures = (params: { page: number; page_size: number; provider_account_id?: number; key?: string }) =>
  get<Paginated<ProviderSignature>>('/account/provider-signatures', params as unknown as Record<string, unknown>)

export const createProviderSignature = (payload: ProviderSignatureCreatePayload) =>
  post<ProviderSignature>('/account/provider-signatures', payload)

export const updateProviderSignature = (id: number, payload: Partial<ProviderSignatureCreatePayload> & { status?: number }) =>
  put<ProviderSignature>(`/account/provider-signatures/${id}`, payload)

export const deleteProviderSignature = (id: number) => del<void>(`/account/provider-signatures/${id}`)

export const signaturesByProvider = (providerId: number) =>
  get<ProviderSignature[]>(`/account/provider-accounts/${providerId}/signatures`)

// ===== 供应商模板 =====

export interface ProviderTemplateCreatePayload {
  provider_id: number
  template_code: string
  template_name: string
  content_type?: string
  template_content?: string
  variables?: string[]
}

export const listProviderTemplates = (params: { page: number; page_size: number; provider_id?: number; key?: string }) =>
  get<Paginated<ProviderTemplate>>('/account/provider-templates', params as unknown as Record<string, unknown>)

export const createProviderTemplate = (payload: ProviderTemplateCreatePayload) =>
  post<ProviderTemplate>('/account/provider-templates', payload)

export const updateProviderTemplate = (id: number, payload: Partial<ProviderTemplateCreatePayload> & { status?: number }) =>
  put<ProviderTemplate>(`/account/provider-templates/${id}`, payload)

export const deleteProviderTemplate = (id: number) => del<void>(`/account/provider-templates/${id}`)

export const templatesByProvider = (providerId: number) =>
  get<ProviderTemplate[]>(`/account/provider-accounts/${providerId}/templates`)
