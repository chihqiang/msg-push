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
- Vue 3 + Vite + Tailwind

## 快速开始

```bash
# 1. 启动基础服务（MySQL + Redis）
docker compose -f deploy/docker-compose.yml up -d

# 2. 启动管理后台（可选，http://localhost:5173）
cd ui && npm install && npm run build

# 3. 启动后端（默认 :8080，自动建表与种子数据）
go run .
```

## Docker 构建与部署

项目根目录提供 `Dockerfile`（多阶段构建：`node:22-alpine` 构建前端 → `golang:1.25-alpine` 编译 Go → `alpine:3.20` 精简运行），前端产物经 `go:embed` 打入二进制，一键产出可运行镜像。

### 构建镜像

```bash
docker build -t zhiqiangwang/app:msg-push .
```

> 前端产物 `ui/dist` 被 git 忽略，构建时由 `node` 阶段实时生成；修改前端后重新构建即可。

### 方式一：docker compose 一键启动（推荐）

`deploy/docker-compose.full.yml` 提供 MySQL + Redis + msg-push 完整编排：

```bash
docker compose -f deploy/docker-compose.full.yml up -d --build

# 查看状态
docker compose -f deploy/docker-compose.full.yml ps

# 查看日志
docker compose -f deploy/docker-compose.full.yml logs -f msg-push

# 停止
docker compose -f deploy/docker-compose.full.yml down
```

启动完成后访问管理后台：http://localhost:8080 （默认账号 `admin` / `admin123`）。

首次部署建议通过环境变量覆盖默认密码：

```bash
JWT_SECRET='your-strong-secret' \
ACCOUNT_SEED_PASSWORD='your-admin-password' \
MYSQL_ROOT_PASSWORD='your-db-password' \
docker compose -f deploy/docker-compose.full.yml up -d --build
```

### 方式二：单独运行 msg-push 容器

配合已有的 MySQL / Redis（如 `deploy/docker-compose.yml` 启动的基础服务）：

```bash
# 1. 启动基础服务
docker compose -f deploy/docker-compose.yml up -d

# 2. 运行 msg-push，通过环境变量连接基础服务
docker run -d --name msg-push \
  -p 8080:8080 \
  -e DB_HOST=127.0.0.1 \
  -e DB_PASSWORD=msgpush123456 \
  -e REDIS_ADDR=127.0.0.1:6379 \
  -e JWT_SECRET=change-me-in-prod \
  zhiqiangwang/app:msg-push
```

> 容器内默认 `DB_HOST=mysql`、`REDIS_ADDR=redis:6379`（compose 服务名）；宿主机直连需改为 `127.0.0.1`。

### 容器环境变量

容器内置配置 `/app/config.yaml`，所有连接信息通过环境变量覆盖（`${VAR}` 占位符），常用变量如下：

| 变量 | 默认值 | 说明 |
|------|--------|------|
| `APP_ENV` | `prod` | 运行环境（dev/test/prod） |
| `SERVER_PORT` | `8080` | HTTP 监听端口 |
| `DB_HOST` | `mysql` | MySQL 地址（compose 服务名） |
| `DB_PORT` | `3306` | MySQL 端口 |
| `DB_USER` / `DB_PASSWORD` | `root` / `msgpush123456` | MySQL 账号密码 |
| `DB_NAME` | `msg_push` | 数据库名 |
| `REDIS_ADDR` | `redis:6379` | Redis 地址 |
| `TASKQ_REDIS_ADDR` | `redis:6379` | 任务队列 Redis 地址 |
| `JWT_SECRET` | `change-me-in-prod` | JWT 密钥（生产必须修改） |
| `ACCOUNT_SEED_USERNAME` / `ACCOUNT_SEED_PASSWORD` | `admin` / `admin123` | 首次建库的种子账号 |

其他说明：

- **健康检查**：`GET /health`（镜像已内置 `HEALTHCHECK`，compose 亦会检查依赖就绪）
- **日志**：默认输出到 stdout，`docker compose logs -f msg-push` 查看
- **SQLite 模式**：如需容器内使用 SQLite，可将 `deploy/docker-config.yaml` 的 `db.driver` 改为 `sqlite` 并挂载数据卷

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
