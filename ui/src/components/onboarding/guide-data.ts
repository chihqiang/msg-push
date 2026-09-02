// 新手引导数据：按消息类型定义完整配置链路
// 每条链路由若干步骤组成，每步对应一个配置页面 + 具体操作指引
import type { Component } from 'vue'
import { MessageSquare, Mail, Building2, Bell, FlaskConical } from '@lucide/vue'

export type MessageType = 'sms' | 'email' | 'wecom' | 'dingtalk' | 'test'

export interface GuideStep {
  title: string
  desc: string
  /** 跳转目标页面（route path） */
  target?: string
  /** 可选：提示在此页面需完成的交互（如「新建」按钮位置） */
  action?: string
  /** 跳转时附加的 query 参数（页面据此自动打开新建弹窗等） */
  query?: Record<string, string>
}

/** 发送请求示例（结构化，组件美化渲染为终端样式） */
export interface SendSample {
  method: string
  path: string
  /** 请求头（如 X-App-Id） */
  headers: string[]
  /** 请求体 */
  body: Record<string, unknown>
}

export interface GuideConfig {
  type: MessageType
  label: string
  desc: string
  icon: Component
  color: string // tailwind 文字颜色类
  /** 该类型是否需要签名 */
  requiresSignature: boolean
  steps: GuideStep[]
  /** 测试模式快捷验证提示 */
  testTip?: string
  /** 发送验证请求示例 */
  sendSample?: SendSample
}

const sendPath = '/tasks'

export const guideConfigs: GuideConfig[] = [
  {
    type: 'sms',
    label: '短信',
    desc: '阿里云 / 腾讯云 / 网易云信等短信服务商',
    icon: MessageSquare,
    color: 'text-cyan-600 bg-cyan-500/10',
    requiresSignature: true,
    steps: [
      { title: '创建应用', desc: '在「应用管理」新建应用，获得 app_id 与 app_secret，发送时用于鉴权。', target: '/apps', action: '点击「新建应用」', query: { action: 'create' } },
      { title: '创建通道', desc: '在「通道管理」创建短信通道，类型选择「短信」，编码如 sms_aliyun。', target: '/channels', action: '点击「新建通道」', query: { action: 'create' } },
      { title: '创建业务模板', desc: '在「模板管理」创建业务模板，绑定短信通道，内容用 {变量} 占位，如「验证码 {code}」。', target: '/templates', action: '点击「新建模板」', query: { action: 'create' } },
      { title: '创建服务商账号', desc: '在「服务商管理 → 服务商账号」新建，选择短信服务商（阿里云/腾讯云/网易云信），填入密钥。', target: '/providers', action: '点击「新建账号」', query: { action: 'create', tab: 'accounts' } },
      { title: '创建服务商签名', desc: '短信需要签名。在「服务商管理 → 服务商签名」创建，签名编码需与平台报备一致。', target: '/providers', action: '点击「新建签名」', query: { action: 'create', tab: 'signatures' } },
      { title: '创建供应商模板', desc: '在「服务商管理 → 供应商模板」创建，模板编码填服务商平台报备的模板ID，并声明变量。', target: '/providers', action: '点击「新建模板」', query: { action: 'create', tab: 'templates' } },
      { title: '通道-模板绑定', desc: '在「通道管理」打开通道详情，绑定供应商模板与服务商账号，设置权重/优先级。', target: '/channels', action: '通道详情 → 绑定 Tab → 新建绑定' },
      { title: '通道-签名映射', desc: '在「通道管理」打开通道详情，把业务签名名映射到服务商签名。', target: '/channels', action: '通道详情 → 签名映射 Tab → 新建映射' },
      { title: '发送验证', desc: '以上全部完成后，调用发送接口即可真实下发短信。', target: sendPath },
    ],
    testTip: '想先快速验证链路？创建「测试应用」(is_test=true)，无需真实服务商即可模拟发送成功。',
    sendSample: {
      method: 'POST',
      path: '/api/v1/messages',
      headers: ['X-App-Id: <你的 app_id>', 'X-App-Secret: <你的 app_secret>', 'Content-Type: application/json'],
      body: {
        channel_code: 'sms_aliyun',
        template_code: 'verify_code',
        receiver: '13800000001',
        template_params: { code: '123456' },
        signature_name: '你的签名',
      },
    },
  },
  {
    type: 'email',
    label: '邮件',
    desc: 'SMTP 邮件服务',
    icon: Mail,
    color: 'text-blue-600 bg-blue-500/10',
    requiresSignature: false,
    steps: [
      { title: '创建应用', desc: '在「应用管理」新建应用，获得 app_id 与 app_secret。', target: '/apps', action: '点击「新建应用」', query: { action: 'create' } },
      { title: '创建通道', desc: '在「通道管理」创建邮件通道，类型选择「邮件」，编码如 email_smtp。', target: '/channels', action: '点击「新建通道」', query: { action: 'create' } },
      { title: '创建业务模板', desc: '在「模板管理」创建业务模板，绑定邮件通道，内容可用 {变量} 占位。', target: '/templates', action: '点击「新建模板」', query: { action: 'create' } },
      { title: '创建服务商账号', desc: '在「服务商管理 → 服务商账号」新建，选择 SMTP 邮件，配置主机/端口/账号/密码/发件人。', target: '/providers', action: '点击「新建账号」', query: { action: 'create', tab: 'accounts' } },
      { title: '创建供应商模板', desc: '在「服务商管理 → 供应商模板」创建，邮件模板内容类型可选 text/html，声明变量。', target: '/providers', action: '点击「新建模板」', query: { action: 'create', tab: 'templates' } },
      { title: '通道-模板绑定', desc: '在「通道管理」打开通道详情，绑定供应商模板与服务商账号。', target: '/channels', action: '通道详情 → 绑定 Tab → 新建绑定' },
      { title: '发送验证', desc: '邮件无需签名映射，配置完成即可发送。', target: sendPath },
    ],
    testTip: '想先快速验证链路？创建「测试应用」(is_test=true)，无需真实 SMTP 即可模拟发送成功。',
    sendSample: {
      method: 'POST',
      path: '/api/v1/messages',
      headers: ['X-App-Id: <你的 app_id>', 'X-App-Secret: <你的 app_secret>', 'Content-Type: application/json'],
      body: {
        channel_code: 'email_smtp',
        template_code: 'notice',
        receiver: 'user@example.com',
        template_params: { name: '张三' },
      },
    },
  },
  {
    type: 'wecom',
    label: '企业微信',
    desc: '企业微信应用消息 / 群机器人',
    icon: Building2,
    color: 'text-emerald-600 bg-emerald-500/10',
    requiresSignature: false,
    steps: [
      { title: '创建应用', desc: '在「应用管理」新建应用，获得 app_id 与 app_secret。', target: '/apps', action: '点击「新建应用」', query: { action: 'create' } },
      { title: '创建通道', desc: '在「通道管理」创建企业微信通道，类型选择「企业微信」，编码如 wecom_app。', target: '/channels', action: '点击「新建通道」', query: { action: 'create' } },
      { title: '创建业务模板', desc: '在「模板管理」创建业务模板，绑定企业微信通道。', target: '/templates', action: '点击「新建模板」', query: { action: 'create' } },
      { title: '创建服务商账号', desc: '在「服务商管理 → 服务商账号」新建，选择「企业微信」或「企业微信群机器人」，配置 CorpId/AgentId/Secret 或 Webhook 地址。', target: '/providers', action: '点击「新建账号」', query: { action: 'create', tab: 'accounts' } },
      { title: '创建供应商模板', desc: '在「服务商管理 → 供应商模板」创建，内容类型可选 text/markdown。', target: '/providers', action: '点击「新建模板」', query: { action: 'create', tab: 'templates' } },
      { title: '通道-模板绑定', desc: '在「通道管理」打开通道详情，绑定供应商模板与服务商账号。', target: '/channels', action: '通道详情 → 绑定 Tab → 新建绑定' },
      { title: '发送验证', desc: '企业微信无需签名映射，配置完成即可发送。', target: sendPath },
    ],
    testTip: '想先快速验证链路？创建「测试应用」(is_test=true) 即可模拟发送成功。',
    sendSample: {
      method: 'POST',
      path: '/api/v1/messages',
      headers: ['X-App-Id: <你的 app_id>', 'X-App-Secret: <你的 app_secret>', 'Content-Type: application/json'],
      body: {
        channel_code: 'wecom_app',
        template_code: 'alert',
        receiver: 'userid_or_@all',
        template_params: { msg: '磁盘告警' },
      },
    },
  },
  {
    type: 'dingtalk',
    label: '钉钉',
    desc: '钉钉工作通知 / 群机器人',
    icon: Bell,
    color: 'text-violet-600 bg-violet-500/10',
    requiresSignature: false,
    steps: [
      { title: '创建应用', desc: '在「应用管理」新建应用，获得 app_id 与 app_secret。', target: '/apps', action: '点击「新建应用」', query: { action: 'create' } },
      { title: '创建通道', desc: '在「通道管理」创建钉钉通道，类型选择「钉钉」，编码如 dingtalk_app。', target: '/channels', action: '点击「新建通道」', query: { action: 'create' } },
      { title: '创建业务模板', desc: '在「模板管理」创建业务模板，绑定钉钉通道。', target: '/templates', action: '点击「新建模板」', query: { action: 'create' } },
      { title: '创建服务商账号', desc: '在「服务商管理 → 服务商账号」新建，选择「钉钉」或「钉钉群机器人」，配置 AppKey/AppSecret/AgentId 或 Webhook 地址。', target: '/providers', action: '点击「新建账号」', query: { action: 'create', tab: 'accounts' } },
      { title: '创建供应商模板', desc: '在「服务商管理 → 供应商模板」创建，内容类型可选 text/markdown。', target: '/providers', action: '点击「新建模板」', query: { action: 'create', tab: 'templates' } },
      { title: '通道-模板绑定', desc: '在「通道管理」打开通道详情，绑定供应商模板与服务商账号。', target: '/channels', action: '通道详情 → 绑定 Tab → 新建绑定' },
      { title: '发送验证', desc: '钉钉无需签名映射，配置完成即可发送。', target: sendPath },
    ],
    testTip: '想先快速验证链路？创建「测试应用」(is_test=true) 即可模拟发送成功。',
    sendSample: {
      method: 'POST',
      path: '/api/v1/messages',
      headers: ['X-App-Id: <你的 app_id>', 'X-App-Secret: <你的 app_secret>', 'Content-Type: application/json'],
      body: {
        channel_code: 'dingtalk_app',
        template_code: 'notice',
        receiver: 'userid',
        template_params: { title: '通知' },
      },
    },
  },
]

/** 测试模式快捷引导（无真实配置，走完整链路但不真实发送） */
export const testGuide: GuideConfig = {
  type: 'test',
  label: '测试模式',
  desc: '创建测试应用，消息走完整链路但不真实发送，用于联调验证',
  icon: FlaskConical,
  color: 'text-amber-600 bg-amber-500/10',
  requiresSignature: false,
  steps: [
    { title: '创建测试应用', desc: '在「应用管理」新建应用，勾选「测试模式」(is_test=true)。', target: '/apps', action: '点击「新建应用」→ 开启测试模式', query: { action: 'create' } },
    { title: '创建通道', desc: '在「通道管理」创建任意类型通道。', target: '/channels', action: '点击「新建通道」', query: { action: 'create' } },
    { title: '创建业务模板', desc: '在「模板管理」创建业务模板，绑定该通道。', target: '/templates', action: '点击「新建模板」', query: { action: 'create' } },
    { title: '发送验证', desc: '使用测试应用的 app_id/app_secret 调用发送接口，消息会被模拟为发送成功，但不会真实投递。', target: sendPath, action: '任务查询中查看 success 状态' },
  ],
  testTip: '种子数据已内置测试应用 test-app / test-secret，可直接使用。',
  sendSample: {
    method: 'POST',
    path: '/api/v1/messages',
    headers: ['X-App-Id: test-app', 'X-App-Secret: test-secret', 'Content-Type: application/json'],
    body: {
      channel_code: '<通道编码>',
      template_code: '<模板编码>',
      receiver: '13800000001',
    },
  },
}
