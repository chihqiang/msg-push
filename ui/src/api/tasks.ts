// 任务/批量任务查询 API
import { get } from './client'
import type { PushTask, PushBatchTask, Paginated } from '@/types'

export interface TaskListParams {
  page: number
  page_size: number
  task_no?: string
  request_id?: string
  batch_id?: string
  app_id?: number
  channel_id?: number
  receiver?: string
  status?: string
  start_date?: string
  end_date?: string
}

export interface BatchTaskListParams {
  page: number
  page_size: number
  batch_id?: string
  app_id?: number
  status?: string
  start_date?: string
  end_date?: string
}

export const listTasks = (params: TaskListParams) =>
  get<Paginated<PushTask>>('/account/push-tasks', params as unknown as Record<string, unknown>)

export const getTask = (id: number) => get<PushTask>(`/account/push-tasks/${id}`)

export const getTaskByNo = (taskNo: string) => get<PushTask>(`/account/push-tasks/no/${taskNo}`)

export const listBatchTasks = (params: BatchTaskListParams) =>
  get<Paginated<PushBatchTask>>('/account/batch-tasks', params as unknown as Record<string, unknown>)

export const getBatchTask = (id: number) => get<PushBatchTask>(`/account/batch-tasks/${id}`)

export const tasksByBatch = (batchId: string, params: { page: number; page_size: number }) =>
  get<Paginated<PushTask>>(`/account/batch-tasks/batch/${batchId}/tasks`, params as unknown as Record<string, unknown>)
