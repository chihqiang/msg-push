// msg-push 类型定义（对齐后端 DTO）

// ===== 认证 =====

export interface LoginResponse {
  access_token: string
  refresh_token: string
  // 过期时间（Unix 秒时间戳）
  expires_at: number
}

// ===== 账号（个人中心） =====

export interface AccountProfile {
  id: number
  username: string
  name: string
  status: number
  created_at: string
}

// ===== 通用分页 =====

export interface Paginated<T> {
  list: T[]
  total: number
  page: number
  page_size: number
}

// ===== 应用 =====

export interface App {
  id: number
  app_id: string
  name: string
  status: number
  is_test: boolean
  daily_quota: number
  rate_limit: number
  remark: string
  created_at: string
}

export interface AppWithSecret extends App {
  secret?: string
}

export interface AppQuotaUsage {
  daily_quota: number
  today_used: number
  remaining: number
  usage_percentage: number
}

// ===== 通道 =====

export interface Channel {
  id: number
  code: string
  name: string
  type: string // sms/email/wecom/dingtalk
  config: string
  status: number
  remark: string
  created_at: string
}

export interface ChannelTestResult {
  task_id: string
  status: string
}

export interface ChannelHealthHistory {
  id: number
  provider_channel_id: number
  check_time: string
  status: string // healthy/unhealthy
  response_time: number
  error_count: number
  success_rate: number
  is_available: number
}

// ===== 消息模板 =====

export interface MessageTemplate {
  id: number
  code: string
  channel_id: number
  name: string
  content: string
  signature: string
  status: number
  remark: string
  created_at: string
}

// ===== 服务商账号 =====

export interface ProviderAccount {
  id: number
  account_code: string
  account_name: string
  provider_code: string
  provider_name: string
  provider_type: string
  config: Record<string, unknown>
  status: number
  remark: string
  created_at: string
  updated_at: string
}

export interface ProviderMeta {
  code: string
  name: string
  type: string
  description: string
  supports_send: boolean
  supports_batch_send: boolean
  supports_callback: boolean
  requires_signature: boolean
  sort_order: number
  tags: string[]
  regions: string[]
  config_fields: ConfigField[]
}

export interface ConfigField {
  key: string
  label: string
  type: string
  required: boolean
  placeholder?: string
  default_value?: string
  description?: string
  help_link?: string
  options?: { value: string; label: string }[]
}

// ===== 服务商签名 =====

export interface ProviderSignature {
  id: number
  provider_account_id: number
  provider_code: string
  provider_type: string
  signature_code: string
  signature_name: string
  status: number
  remark: string
  created_at: string
  updated_at: string
}

// ===== 供应商模板 =====

export interface ProviderTemplate {
  id: number
  provider_id: number
  template_code: string
  template_name: string
  content_type: string
  template_content: string
  variables: string[]
  status: number
  remark: string
  created_at: string
  updated_at: string
}

// ===== 通道-模板绑定 =====

export interface ChannelBinding {
  id: number
  channel_id: number
  provider_template_id: number
  provider_template_name: string
  provider_id: number
  provider_name: string
  provider_type: string
  param_mapping: ParamMappingItem[]
  weight: number
  priority: number
  status: number
  is_active: number
  auto_disable_on_fail: boolean
  auto_disable_threshold: number
  created_at: string
}

export interface ParamMappingItem {
  type: string // fixed/mapping
  provider_var: string
  system_var: string
  value: string
}

// 可用供应商模板（绑定下拉）
export interface AvailableProviderTemplate {
  id: number
  template_code: string
  template_name: string
  template_content: string
  variables: string[]
  provider_id: number
  provider_code: string
  provider_type: string
  status: number
}

// ===== 通道-签名映射 =====

export interface ChannelSignatureMapping {
  id: number
  channel_id: number
  signature_name: string
  provider_signature_id: number
  signature_code: string
  provider_id: number
  provider_name: string
  provider_type: string
  status: number
  created_at: string
}

// 可用供应商签名（映射下拉）
export interface AvailableProviderSignature {
  id: number
  signature_code: string
  signature_name: string
  provider_id: number
  provider_code: string
  provider_type: string
  status: number
}

// ===== 失败规则 =====

export interface FailureRule {
  id: number
  name: string
  scene: string // send_failure/callback_failure
  provider_code: string
  message_type: string
  error_code: string
  error_keyword: string
  action: string // retry/switch_provider/fail/alert
  action_config: string
  priority: number
  status: number
  created_at: string
}

// ===== 任务 =====

export interface PushTask {
  id: number
  task_id: string
  request_id: string
  app_id: number
  batch_id: string
  channel_id: number
  template_id: number
  receiver: string
  params: string
  signature: string
  is_test: boolean
  status: string // pending/sending/success/failed
  error_msg: string
  scheduled_at: string | null
  sent_at: string | null
  created_at: string
  updated_at: string
  channel_name?: string
}

export interface PushBatchTask {
  id: number
  app_id: number
  batch_id: string
  channel_id: number
  template_id: number
  total_count: number
  success_count: number
  failed_count: number
  pending_count: number
  is_test: boolean
  status: string // processing/completed/failed
  completion_rate: number
  created_at: string
  updated_at: string
}

// ===== 日志 =====

export interface PushLog {
  id: number
  request_id: string
  task_id: number
  task_no: string
  app_id: number
  provider_account_id: number
  provider_msg_id: string
  receiver: string
  status: string
  provider_resp: string
  error_code: string
  error_msg: string
  cost_time: number
  created_at: string
}

export interface CallbackLog {
  id: number
  request_id: string
  type: string // report/upstream
  task_no: string
  app_id: number
  provider_code: string
  provider_account_id: number
  provider_id: string
  mobile: string
  content: string
  callback_status: string
  error_code: string
  error_message: string
  raw_data: string
  created_at: string
}

// ===== Webhook =====

export interface WebhookConfig {
  id: number
  name: string
  app_id: number
  webhook_url: string
  events: string // 逗号分隔，如 "success,failed"
  status: number
  retry_count: number
  timeout: number
  description: string
  created_at: string
  updated_at: string
}

export interface WebhookLog {
  id: number
  request_id: string
  task_no: string
  app_id: number
  webhook_config_id: number
  webhook_url: string
  event: string
  request_data: string
  response_status: number
  response_data: string
  status: string // pending/processing/success/failed
  error_message: string
  retry_count: number
  max_retries: number
  timeout_seconds: number
  created_at: string
}

// ===== 统计分析 =====

export interface StatisticsCounts {
  total_count: number
  success_count: number
  failure_count: number
  pending_count: number
  processing_count: number
  sent_count: number
  in_progress_count: number
  completed_success_rate: number | null
}

export interface StatisticsSummary extends StatisticsCounts {
  success_rate: string
}

export interface DailyStatistics extends StatisticsSummary {
  date: string
}

export interface StatisticsResponse {
  period: {
    start_date: string
    end_date: string
    timezone: string
    granularity: string
  }
  summary: StatisticsSummary
  daily: DailyStatistics[]
  message_type_distribution: (StatisticsCounts & { message_type: string })[]
  top_applications: (StatisticsCounts & {
    id: number
    app_id: string
    app_name: string
  })[]
  top_channels: (StatisticsCounts & {
    channel_id: number
    channel_name: string
    channel_type: string
  })[]
}

export interface DashboardResponse {
  total_applications: number
  active_applications: number
  total_channels: number
  active_channels: number
  total_providers: number
  active_providers: number
  today_push_count: number
  today_success_count: number
  today_failed_count: number
  today_in_progress_count: number
  today_success_rate: string
  today_completed_success_rate: number | null
  total_push_count: number
}

export interface TopApplicationResponse {
  id: number
  app_id: string
  app_name: string
  push_count: number
  success_count: number
  success_rate: string
}

export interface RecentActivityResponse {
  id: number
  description: string
  app_name: string
  created_at: string
}
