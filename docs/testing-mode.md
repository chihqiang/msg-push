# 测试模式

测试模式（`is_test`）让消息走**完整链路**但不**真实发送**，适合联调、验收、演示，不产生真实外部投递。

## 一、工作原理

开启测试模式后，消息依然：

- 创建任务、写入数据库；
- 入队、被 Worker 消费；
- 走选通道、模板渲染、状态流转；
- 写 `push_log`、更新批次计数、触发 Webhook；

但**不调用服务商真实发送接口**，而是模拟成功：

- 任务状态 → `success`；
- `push_log.provider_msg_id` → `TEST`；
- `push_log.provider_resp` → `{"test":true}`。

## 二、开启方式（应用级）

测试模式在**应用**上配置，应用发出的所有消息按应用配置决定是否测试：

- 管理后台「应用管理 → 新建/编辑」勾选**测试模式**，则该应用发出的**所有**消息走测试链路；
- 数据库字段：`msg_applications.is_test`；
- 管理 API：`POST/PUT /apps` 传 `"is_test": true`。

> 发送接口 `POST /messages` 与 `POST /messages/batch` **不接受** `is_test` 参数，测试与否完全由应用配置决定，业务调用方无需也无法在请求级覆盖。

## 三、内置种子应用

启动时自动创建**测试应用**（`db/migrate.go` seedDemoApp）：

| 项 | 值 |
|----|----|
| app_id | `test-app` |
| app_secret | `test-secret` |
| name | 测试应用 |
| is_test | true |

直接用它可以立即体验完整链路，无需配置任何真实服务商。

## 四、数据表现

| 表 | 字段 | 测试表现 |
|----|------|----------|
| `msg_applications` | is_test | 1 |
| `msg_push_tasks` | is_test | 1，status=success |
| `msg_push_logs` | is_test | 1，provider_msg_id=TEST |
| `msg_push_batch_tasks` | is_test | 1，success_count=总数 |

管理后台「应用管理」「任务查询」都有**模式**列（测试/正式），一眼区分。

## 五、与真实发送的区别

| 维度 | 测试模式 | 真实模式 |
|------|:---:|:---:|
| 调用服务商 API | ❌ 不调用 | ✅ 调用 |
| 需要配置供应商账号 | ❌ 不需要 | ✅ 需要 |
| 产生真实外部投递 | ❌ 不会 | ✅ 会 |
| 消耗配额 | 计入 | 计入 |
| 任务/日志/批次/Webhook | ✅ 完整 | ✅ 完整 |

> 测试模式仍然走完整的配额统计与 webhook 通知，便于联调通知链路。

## 六、使用建议

- 开发/联调环境：开启应用测试模式，避免误发真实短信/邮件；
- 生产环境：应用保持正式模式，需要联调时单独开一个测试应用验证；
- 验收演示：用 `test-app` 演示全流程，无需真实服务商凭证。

## 相关文档

- [快速开始](quickstart.md)
- [API 使用指南](api-guide.md)
- [配置指南](configuration.md)
