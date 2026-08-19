# msg-push

统一高可用消息投递服务，一套 API 对接短信、邮件、企业微信、钉钉，提供模板管理、异步排队、智能路由、故障切换、回执与投递统计全能力。

## 特性

- **多通道**：短信（阿里云 / 腾讯云 / 网易云信）、邮件（SMTP）、企业微信（应用消息 / 群机器人）、钉钉（工作通知 / 群机器人）
- **模板管理**：业务模板 → 通道绑定 → 供应商模板渲染，变量自动映射
- **异步投递**：基于 Redis (asynq) 任务队列，提交与投递解耦，支持定时发送
- **智能路由**：按通道/签名选择服务商，故障自动切换、熔断降级、有限指数退避重试
- **批量能力**：支持服务商批量 API 聚合发送，不支持时自动退回逐条
- **回执闭环**：回调接收 → 回执关联 → 状态主动补单 → 超时扫描，状态全链路可追踪
- **可观测**：`request_id` 全链路追踪（提交→投递→回执→webhook），日志自动携带
- **测试模式**：`is_test` 走完整链路但不真实发送（模拟成功），适合联调
- **Webhook 通知**：outbox 模式异步投递，签名校验 + 重试退避
- **管理后台**：Web 控制台（应用/通道/模板/服务商/日志/统计）

## 技术栈

- Go 1.25 + infra-go（HTTP / ORM / Redis / 任务队列 / JWT / 日志）
- MySQL（或 SQLite） + Redis
- Vue 3 + Vite + Tailwind（独立 webui 子模块）

## 快速开始

```bash
# 1. 启动基础服务（MySQL + Redis）
docker compose -f deploy/docker-compose.yml up -d

# 2. 启动后端（默认 :8080，自动建表与种子数据）
go run .

# 3. 启动管理后台（可选，http://localhost:5173）
cd webui && npm install && npm run dev
```

## 文档

- [快速开始](docs/quickstart.md)
- [配置与发送指南](docs/setup-and-send.md)
- [配置指南](docs/configuration.md)
- [API 使用指南](docs/api-guide.md)
- [架构设计](docs/architecture.md)
- [数据模型](docs/database.md)
- [失败处理与重试](docs/failure-retry.md)
- [短信回调地址](docs/sms-callback.md)
- [Webhook 通知](docs/webhook.md)
- [测试模式](docs/testing-mode.md)
- [运维与可观测](docs/operations.md)
- [开发者指南](docs/development.md)
- [新增短信服务商](docs/add-new-sms-provider.md)
