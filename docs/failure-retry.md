# 失败处理与重试

msg-push 通过**智能路由 + 失败规则引擎 + 有限重试**三套机制保证消息投递的可靠性。

## 一、选通道（智能路由）

### 选择器 Selector

消费端投递前通过 `worker/channel.Selector` 选择服务商账号：

- 按通道 + 消息类型筛选；
- 结合签名映射；
- 排除已失败/被熔断的服务商；
- 进程内缓存节点列表（TTL 60s），配置变更/熔断后立即失效。

### 熔断

- 发送失败会报告给选择器（`ReportFailure`）；
- 连续失败触发熔断，临时禁用该服务商账号；
- 熔断后立即 `InvalidateCache`，避免 TTL 窗口内仍选中故障节点。

### 无可用供应商

选通道失败（熔断/类型不匹配/暂不可用）时，任务进入**有限指数退避重试**：

- 默认最大重试 3 次；
- 退避策略：`delay = 5s * 2^n`（上限 300s）；
- 每次重试把任务置回 `pending` 并重新入队；
- 超过上限才置 `failed`，避免消息永久丢失。

> 测试任务（`is_test=true`）选通道失败也直接模拟成功，不依赖供应商配置。

## 二、失败规则引擎

`worker/rule_engine.go` 提供可配置的失败处理规则（`msg_failure_rules` 表），命中后执行相应动作。

### 场景

| 场景 | 说明 |
|------|------|
| `send_failure` | 发送失败时 |
| `callback_failure` | 回调失败时 |

### 动作

| 动作 | 说明 |
|------|------|
| `retry` | 有限次重试（退避） |
| `switch_provider` | 切换供应商后重试 |
| `fail` | 直接置失败终态 |
| `alert` | 发送告警 webhook 通知 |

### 匹配条件

规则可按以下维度匹配（全部满足才命中）：

- `provider_code`：服务商代码（空=匹配所有）；
- `message_type`：消息类型（空=匹配所有）；
- `error_code`：错误码（逗号分隔多个）；
- `error_keyword`：错误消息关键字（模糊匹配）；
- `priority`：优先级（越大越优先）。

### 默认动作

未命中规则时按场景默认：

- `send_failure` → `retry`（有限重试）；
- `callback_failure` → `fail`。

### 重试配置

```json
{
  "max_retry": 3,
  "delay_seconds": 2,
  "backoff_rate": 2,
  "max_delay": 60
}
```

### 切换供应商配置

```json
{
  "exclude_current": true,
  "max_retry": 3
}
```

切换供应商后，当前服务商被加入任务 `exclude_provider_ids`，后续选择会避开它。

### 告警配置

```json
{
  "webhook_url": "https://example.com/alert",
  "alert_level": "critical"
}
```

### 缓存

规则在 Worker 进程内缓存，管理端修改后调用 `POST /api/v1/account/failure-rules/refresh-cache` 立即刷新（或等待缓存过期）。

## 三、发送异常处理（ActionExecutor）

发送失败后 `ActionExecutor` 依据规则执行动作：

| 情况 | 处理 |
|------|------|
| 重试/切换 | 按退避重新入队，保留 `request_id` 全链路追踪 |
| 入队失败 | 直接置 `failed`（防任务永久卡死） |
| 最终失败 | 写 `push_log`（failed）+ 触发终态 webhook |
| 发送器返回 nil | 按早期失败处理（防 panic） |

## 四、回执状态保障

短信发送成功后状态先置 `sending`（等待回执），由三路写入终态：

| 路径 | 说明 |
|------|------|
| **回调接收** | 服务商回执回调 → 关联 `push_task`/`push_log` → 更新状态 |
| **状态主动补单** | `status_puller` 定时查询服务商状态，无回执时主动补单（校验手机号匹配） |
| **超时扫描** | `sms_timeout_scanner` 超过硬超时（默认 10min）仍无回执 → 标记 `callback_status=timeout` 并转 `failed` |

> `status_puller` 的候选查询覆盖 `callback_status` 为 `timeout` 的任务，在硬超时转 failed 前确认真实状态，避免"已送达却被判失败"。

## 相关文档

- [架构设计](architecture.md)
- [数据模型](database.md)
- [Webhook 通知](webhook.md)
- [运维与可观测](operations.md)
