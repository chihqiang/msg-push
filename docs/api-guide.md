# API 使用指南

msg-push 提供两类 API：

1. **应用侧 API**（`/api/v1/...`）：供业务系统调用，发送消息、查询任务，使用**应用鉴权**；
2. **管理侧 API**（`/api/v1/account/...`）：供管理后台使用，管理应用/通道/模板/服务商/日志/统计，使用 **JWT 账号鉴权**。

## 一、应用侧 API

### 认证方式

请求携带应用凭证，二选一：

**方式 A：兼容模式（简单）** —— 明文密钥：

```
X-App-Id: test-app
X-App-Secret: test-secret
```

**方式 B：HMAC-SHA256 签名（推荐生产使用）**

| Header | 说明 |
|--------|------|
| `X-App-Id` | 应用 ID |
| `X-Signature` | HMAC 签名 |
| `X-Timestamp` | Unix 秒级时间戳（±5 分钟防重放） |
| `X-Nonce` | 随机串（防重放） |

签名算法：

```text
签名串 = method + path + sortedParams + timestamp + nonce
signature = HEX( HMAC-SHA256(签名串, app_secret) )
```

- `method`：如 `POST`；
- `path`：如 `/api/v1/messages`；
- `sortedParams`：请求体 JSON 按 key 字典序排序后序列化（无 body 为空串）。

> 注意：签名串中 `method`、`path`、`timestamp`、`nonce` 直接拼接，`sortedParams` 夹在 path 与 timestamp 之间。

### 公共说明

- 统一响应结构：`{"code":0,"msg":"ok","data":...,"request_id":"..."}`；
- 每次请求都会生成/透传 `request_id`（可传 `X-Request-Id` 请求头自定义）；
- 应用侧接口挂载了中间件：应用鉴权 → 配额校验 → 限流。

### 发送单条消息

```
POST /api/v1/messages
```

请求体：

```json
{
  "channel_code": "aliyun_sms",
  "template_code": "",
  "receiver": "13800000001",
  "template_params": { "code": "1234" },
  "signature_name": "",
  "scheduled_at": null
}
```

| 字段 | 必填 | 说明 |
|------|:---:|------|
| `channel_code` | | 通道编码（对应通道的 `code`，如 `aliyun_sms`）。与 `template_code` 至少传一个：只传模板时自动定位模板所属通道 |
| `template_code` | | 模板编码（唯一标识）。与 `channel_code` 至少传一个：空且传了 `channel_code` 表示不套模板 |
| `receiver` | ✅ | 接收方（手机号 / 邮箱 / 用户ID） |
| `template_params` | | 模板参数（key-value） |
| `signature_name` | | 签名别名（走签名映射） |
| `scheduled_at` | | 定时发送时间（ISO8601） |

> `channel_code` 与 `template_code` 至少传一个：只传 `template_code` 时按模板反查其所属通道；两者都不传报错 `channel_code or template_code required`。

> 是否测试模式（`is_test`）**由应用配置决定**，发送请求不再传该字段：应用为测试应用则模拟成功，否则真实发送。

响应：

```json
{
  "code": 0,
  "msg": "ok",
  "data": {
    "task_id": "task_xxx",
    "status": "pending",
    "is_test": false,
    "created_at": "2026-08-15T10:00:00+08:00"
  }
}
```

### 批量发送

```
POST /api/v1/messages/batch
```

```json
{
  "channel_code": "aliyun_sms",
  "receivers": ["13800000001", "13800000002"],
  "template_params": { "code": "1234" }
}
```

响应：返回 `batch_id`、总数与初始成功/失败数（最终状态由消费端异步流转）。

### 查询任务

```
GET /api/v1/tasks/{task_id}
```

返回任务详情（状态、错误、时间、is_test、request_id 等）。

### 回调接收（公开）

服务商回执回调统一地址，无需认证：

```
POST /api/callback/{provider_account_id}
GET  /api/callback/{provider_account_id}
```

回调 URL 配置到服务商控制台后，回执会：
1. 落 `callback_log`；
2. 匹配关联 `push_task` / `push_log`，更新回执状态；
3. 触发 webhook 事件通知。

## 二、管理侧 API

管理侧接口前缀 `/api/v1/account`。除登录/刷新外均需 JWT：请求头 `Authorization: Bearer <access_token>`。

### 认证

```
POST /api/v1/account/auth/login
POST /api/v1/account/auth/refresh
```

登录请求：`{ "username": "admin", "password": "admin123" }`，返回 `access_token` / `refresh_token` / `expires_at`。

### 应用管理

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/apps` | 创建应用（返回明文密钥仅此一次） |
| GET | `/apps` | 应用列表 |
| GET | `/apps/{id}` | 应用详情 |
| PUT | `/apps/{id}` | 更新应用（含 `is_test` 测试模式） |
| DELETE | `/apps/{id}` | 删除应用 |
| POST | `/apps/{id}/reset-secret` | 重置密钥 |
| GET | `/apps/{id}/quota-usage` | 配额使用情况 |

### 通道管理

| 方法 | 路径 | 说明 |
|------|------|------|
| POST/GET | `/channels` | 创建 / 列表 |
| GET/PUT/DELETE | `/channels/{id}` | 详情 / 更新 / 删除 |
| POST | `/channels/{id}/test` | 通道测试发送 |
| GET | `/channels/{id}/health-history` | 健康历史 |
| GET/POST | `/channels/{id}/bindings` | 通道-模板绑定 |
| GET/POST | `/channels/{id}/signature-mappings` | 通道-签名映射 |
| GET | `/channels/{id}/available-templates` | 可选模板 |
| GET | `/channels/{id}/available-signatures` | 可选签名 |

### 模板管理

| 方法 | 路径 | 说明 |
|------|------|------|
| POST/GET | `/templates` | 创建 / 列表 |
| GET/PUT/DELETE | `/templates/{id}` | 详情 / 更新 / 删除 |

### 服务商管理

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/provider-accounts/available` | 可用服务商列表 |
| GET | `/provider-config-fields/{provider_code}` | 服务商配置字段定义 |
| POST/GET | `/provider-accounts` | 创建 / 列表 |
| GET/PUT/DELETE | `/provider-accounts/{id}` | 详情 / 更新 / 删除 |
| POST/GET | `/provider-signatures` | 服务商签名管理 |
| POST/GET | `/provider-templates` | 供应商模板管理 |
| GET | `/provider-accounts/{id}/signatures` | 某账号可用签名 |
| GET | `/provider-accounts/{id}/templates` | 某账号可用模板 |

### 失败规则

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/failure-rules/options` | 可选项 |
| POST | `/failure-rules/refresh-cache` | 刷新缓存 |
| CRUD | `/failure-rules` | 失败规则管理 |

### Webhook 配置与日志

| 方法 | 路径 | 说明 |
|------|------|------|
| CRUD | `/webhook-configs` | Webhook 配置管理 |
| GET | `/webhook-logs` | Webhook 投递日志 |
| GET | `/webhook-logs/task/{task_id}` | 按任务号查日志 |

### 日志查询

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/logs` | 推送日志（支持 task_no / request_id / status / 时间过滤） |
| GET | `/logs/task/{task_id}` | 按任务号查推送日志 |
| GET | `/callbacks` | 回调日志（支持 request_id 等过滤） |
| GET | `/callbacks/task/{task_id}` | 按任务号查回调日志 |

### 任务查询

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/push-tasks` | 任务列表（支持 task_no / request_id / status 过滤） |
| GET | `/push-tasks/{id}` | 任务详情 |
| GET | `/push-tasks/no/{task_no}` | 按任务号查 |
| GET | `/batch-tasks` | 批次列表 |
| GET | `/batch-tasks/batch/{batch_id}/tasks` | 批次下任务 |

### 统计分析

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/statistics` | 综合统计 |
| GET | `/statistics/dashboard` | 仪表盘数据 |
| GET | `/statistics/top-applications` | Top 应用 |
| GET | `/statistics/recent-activities` | 最近活动 |

## 相关文档

- [快速开始](quickstart.md)
- [Webhook 通知](webhook.md)
- [测试模式](testing-mode.md)
- [数据模型](database.md)
