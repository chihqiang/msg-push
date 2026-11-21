# 配置指南

服务配置位于根目录 `config.yaml`，由 ten-go `conf` 加载，支持环境变量覆盖 `${VAR}`。

## 完整配置示例

`rest.Config` 以嵌入方式并入 `config.Config`，因此 HTTP 服务相关键（`host` / `port` /
`logger` / `middlewares` / `health`）与应用信息（`name` / `env`）都位于配置顶层，
不需要额外的 `server:` / `app:` 层级。

```yaml
host: 0.0.0.0
port: 8080
health:
  enabled: true       # 内置 GET /health（默认关闭，需在此显式开启）
  path: /health       # 可省略，默认 /health

name: msg-push
env: dev              # dev / test / prod

logger:
  level: info         # debug / info / error / severe
  mode: console       # console / file / volume
  encoding: plain     # plain（可读文本）/ json
  serviceName: msg-push

db:
  driver: mysql       # mysql / sqlite
  host: 127.0.0.1
  port: 3306
  username: root
  password: msgpush123456
  database: msg_push
  maxIdleConns: 10
  maxOpenConns: 100

redis:
  addr: 127.0.0.1:6379
  keyPrefix: msg-push

jwt:
  secret: change-me-in-prod
  issuer: msg-push
  access_token_expire: 2h
  refresh_token_expire: 168h

taskq:
  redisAddr: 127.0.0.1:6379
  concurrency: 10      # worker 并发消费数
  defaultMaxRetry: 3
  defaultTimeout: 30m
  defaultQueue: default

account_seed:
  username: admin
  password: admin123
```

> 键名匹配规则：ten-go `conf` 按字段名**不区分大小写**匹配，多词字段请使用驼峰键
> （如 `keyPrefix`、`redisAddr`、`maxIdleConns`）。`redis_addr` 这类下划线写法不会生效，
> 字段会回退到库内置默认值。本项目自有配置（`jwt`、`scheduler`、`account_seed`）
> 通过结构体 tag 显式声明了键名，仍可使用下划线写法。

## 字段说明

### 应用信息（顶层）

| 字段 | 默认 | 说明 |
|------|------|------|
| `name` | msg-push | 应用名（同时用于访问日志 `service` 字段） |
| `env` | dev | 运行环境，`options=[dev,test,prod]` |

### HTTP 服务（嵌入的 rest.Config）

顶层 `host` / `port` 与内置中间件、健康检查：

| 字段 | 说明 |
|------|------|
| `host` / `port` | 监听地址与端口 |
| `middlewares.recovery` | panic 恢复，默认启用 |
| `middlewares.requestID` | 生成/透传 `X-Request-Id`，默认启用 |
| `middlewares.log` | 访问日志，默认启用；可配置 `serviceName`、`skipPaths`、`trustedProxies` |
| `middlewares.maxBytes` | 请求体上限，默认启用（1MB） |
| `health` | 内置健康检查路由，`enabled=true` 时注册 `GET /health` |

其余可用中间件（`trace` / `auth` / `rateLimit` / `breaker` / `timeout` / `signature` 等）默认关闭，
按需在顶层 `middlewares` 下开启。

### logger

| 字段 | 说明 |
|------|------|
| `level` | 日志级别：debug / info / error / severe |
| `mode` | 输出模式：console（控制台）/ file（文件）/ volume（按服务名分目录） |
| `encoding` | 输出格式：plain / json |
| `serviceName` | 服务名（日志固定字段） |
| `path` | 文件模式下的日志目录，默认 `logs` |
| `rotation` | 轮转方式：daily（按天）/ size（按大小） |
| `maxSize` / `maxBackups` / `keepDays` | 文件轮转与保留策略 |

### db

使用 ten-go ORM 配置。

- **MySQL（默认）**：`driver: mysql` + host/port/username/password/database；
- **SQLite**：`driver: sqlite, database: msg-push.db`（开箱即用，无需 MySQL）；
- 也支持 `postgres`。

> 敏感信息建议用环境变量覆盖：`DB_HOST`、`DB_USER`、`DB_PASSWORD`、`DB_NAME` 等。

### redis

| 字段 | 说明 |
|------|------|
| `addr` | Redis 地址 |
| `keyPrefix` | key 前缀（多环境隔离） |
| `password` | 可选，密码 |
| `db` | 可选，选库 |

### jwt

管理后台账号认证（JWT）。

| 字段 | 说明 |
|------|------|
| `secret` | JWT 签名密钥，**生产必须修改** |
| `issuer` | 签发者 |
| `access_token_expire` | 访问令牌有效期（默认 2h） |
| `refresh_token_expire` | 刷新令牌有效期（默认 168h） |

### taskq

异步任务队列（Redis asynq）。

| 字段 | 说明 |
|------|------|
| `redisAddr` | 队列使用的 Redis 地址 |
| `concurrency` | Worker 并发消费数（默认 10） |
| `defaultMaxRetry` | 任务默认最大重试次数 |
| `defaultTimeout` | 任务默认超时 |
| `defaultQueue` | 默认队列名 |

### account_seed

管理端种子账号，启动时直接创建（数据库需为全新/已手动清理状态）：

```bash
# 建议通过环境变量覆盖初始密码
ACCOUNT_SEED_USERNAME=admin ACCOUNT_SEED_PASSWORD=your-secret go run .
```

## 环境变量覆盖

ten-go `conf` 支持用环境变量覆盖 yaml 值，命名规则为 `大写配置路径`：

```bash
export DB_PASSWORD='prod-password'
export JWT_SECRET='prod-secret'
export SERVER_PORT=9090
go run .
```

## 相关文档

- [快速开始](quickstart.md)
- [架构设计](architecture.md)
- [运维与可观测](operations.md)
