# 运维与可观测

## 一、全链路追踪（request_id）

msg-push 的每次 HTTP 提交都会生成或透传一个 `request_id`（支持请求头 `X-Request-Id` 自定义），贯穿整条链路：

```
HTTP 提交 → PushTask.request_id → 队列 payload → Worker 消费
        → PushLog.request_id → 回调回填 CallbackLog.request_id
        → Webhook 通知携带 request_id
```

### 使用方式

1. **提交时自定义追踪 ID**：

```bash
curl -X POST http://127.0.0.1:8080/api/v1/messages \
  -H "X-Request-Id: my-order-12345" \
  -H "X-App-Id: test-app" -H "X-App-Secret: test-secret" \
  -H "Content-Type: application/json" \
  -d '{"channel_code":"aliyun_sms","receiver":"13800000001"}'
```

2. **响应中返回** `request_id` 字段；
3. **管理后台**「任务查询」「日志查询」按追踪 ID 搜索，即可串联整条链路（任务→推送日志→回调日志→Webhook 日志）；
4. **日志**：Worker/调度器所有日志自动携带 `request_id` 字段（通过 `httpx` context extractor 注入）。

### 落库位置

| 表 | 字段 |
|----|------|
| msg_push_tasks | request_id（提交时写入） |
| msg_push_logs | request_id（投递时写入） |
| msg_callback_logs | request_id（回调匹配回填） |
| msg_webhook_logs | request_id（通知时写入） |

四个表均建索引，可按 request_id 高效检索。

## 二、日志

### 服务日志

- 默认输出到 stdout（`config.yaml` logger 段可配文件/json）；
- 字段包含 `request_id`、`task_id`、`app_id` 等，便于 grep；
- 级别：`-1=debug, 0=info, 1=warn, 2=error`。

### 业务日志

| 类型 | 查看位置 |
|------|----------|
| 推送日志 | 管理后台「日志查询」或 `GET /api/v1/account/logs` |
| 回调日志 | `GET /api/v1/account/callbacks` |
| Webhook 日志 | `GET /api/v1/account/webhook-logs` |

## 三、健康检查

```
GET /health
# {"code":0,"msg":"ok","data":{"status":"ok"}}
```

可用作容器探针（liveness/readiness）。

## 四、监控指标点

以下环节建议接入告警：

- **队列积压**：Redis 中 `msg:send` / `msg:batch` 队列长度；
- **失败率**：推送日志 failed 占比（按 app/channel/provider 维度）；
- **回调超时**：`callback_status=timeout` 数量，说明服务商回执异常；
- **Webhook 投递失败**：`webhook_logs.status=failed`；
- **服务商健康**：熔断次数、`channel_health_history`；
- **配额水位**：`app_quota_stats` / `provider_quota_stats`。

## 五、统计分析

管理端提供：

- `GET /statistics`：综合统计（总量/成功/失败/成功率）；
- `GET /statistics/dashboard`：仪表盘；
- `GET /statistics/top-applications`：Top 应用；
- `GET /statistics/recent-activities`：最近活动。

后台 `scheduler/quota_syncer.go` 定时同步配额统计。

## 六、生产部署建议

### 基础服务

- MySQL 与 Redis 建议高可用（主从/哨兵/集群）；
- Redis 启用持久化（docker-compose 已开启 appendonly）；
- 配置项敏感信息用环境变量覆盖，JWT secret 生产必改。

### 多实例

- 服务无状态（HTTP + Worker + 调度器可水平扩展）；
- 队列消费通过 CAS 抢占与 Redis 锁保证 at-least-once 下不重复投递；
- 多实例部署时注意 `concurrency` 与队列吞吐的匹配。

### 数据备份

- MySQL 定期备份 `msg_push` 库；
- Webhook outbox 记录落库，重启不丢，可放心滚动发布。

## 相关文档

- [配置指南](configuration.md)
- [架构设计](architecture.md)
- [数据模型](database.md)
- [快速开始](quickstart.md)
