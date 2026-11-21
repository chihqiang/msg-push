// API 客户端：统一响应解包、Bearer token、401 自动刷新（单飞）
import axios from 'axios'
import type { AxiosInstance, AxiosRequestConfig, InternalAxiosRequestConfig } from 'axios'
import { useAuthStore } from '@/stores/auth'
import router from '@/router'

// 后端统一响应包装
export interface ApiResponse<T = unknown> {
  code: number
  msg: string
  data: T
  request_id?: string
}

// 后端业务码
export const CodeOK = 0
export const CodeUnauthorized = 401
export const CodeForbidden = 403

// 后端 API 基础地址：统一相对路径 /api/v1（开发环境经 Vite 代理转发，生产同域部署）
const client: AxiosInstance = axios.create({
  baseURL: '/api/v1',
  timeout: 15000,
})

// 进行中的刷新令牌请求（单飞：并发 401 只触发一次刷新）
let refreshPromise: Promise<boolean> | null = null

function refreshOnce(): Promise<boolean> {
  const auth = useAuthStore()
  if (!refreshPromise) {
    refreshPromise = auth.tryRefresh().finally(() => {
      refreshPromise = null
    })
  }
  return refreshPromise
}

// 请求拦截器：自动附加 Bearer token
client.interceptors.request.use((config: InternalAxiosRequestConfig) => {
  const auth = useAuthStore()
  if (auth.accessToken) {
    config.headers.Authorization = `Bearer ${auth.accessToken}`
  }
  return config
})

// 响应拦截器：解包统一响应 / 处理 401
client.interceptors.response.use(
  (response) => {
    const body = response.data as ApiResponse
    if (body && typeof body.code === 'number' && body.code !== CodeOK) {
      return Promise.reject(new ApiError(body.code, body.msg))
    }
    return response
  },
  async (error) => {
    const status = error.response?.status
    const body = error.response?.data as ApiResponse | undefined

    if (status === CodeUnauthorized) {
      const auth = useAuthStore()
      const retried = (error.config as { _retried?: boolean } | undefined)?._retried
      const ok = !retried && (await refreshOnce())
      if (ok) {
        const config = error.config as AxiosRequestConfig & { _retried?: boolean }
        config._retried = true
        return client(config)
      }
      auth.logout()
      router.push('/auth')
      return Promise.reject(new ApiError(CodeUnauthorized, body?.msg || '未登录或登录已过期'))
    }

    return Promise.reject(new ApiError(body?.code ?? status ?? -1, body?.msg || error.message || '请求失败'))
  }
)

// 业务错误类型
export class ApiError extends Error {
  code: number
  constructor(code: number, message: string) {
    super(message)
    this.code = code
    this.name = 'ApiError'
  }
}

// 封装请求方法：直接返回 data（解包统一响应）
export async function request<T>(config: AxiosRequestConfig): Promise<T> {
  const response = await client.request<ApiResponse<T>>(config)
  return (response.data as ApiResponse<T>).data
}

export const get = <T>(url: string, params?: Record<string, unknown>) =>
  request<T>({ method: 'GET', url, params })

export const post = <T>(url: string, data?: unknown) => request<T>({ method: 'POST', url, data })

export const put = <T>(url: string, data?: unknown) => request<T>({ method: 'PUT', url, data })

export const del = <T>(url: string, data?: unknown) => request<T>({ method: 'DELETE', url, data })

export default client
