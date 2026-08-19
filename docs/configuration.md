# 配置指南

服务配置位于根目录 `config.yaml`，由 infra-go `conf` 加载，支持环境变量覆盖 `${VAR}`。

## 完整配置示例

```yaml
app:
  name: msg-push
  env: dev            # dev / test / prod

server:
  host: 0.0.0.0
  port: 8080

logger:
  level: 0            # -1=debug, 0=info, 1=warn, 2=error
  encoding: console
  output:
    - stdout
  app_name: msg-push

db:
  driver: mysql       # mysql / sqlite
  host: 127.0.0.1
  port: 3306
  username: root
  password: msgpush123456
  database: msg_push
  max_idle_conns: 10
  max_open_conns: 100

redis:
  addr: 127.0.0.1:6379
  key_prefix: msg-push

jwt:
  secret: change-me-in-prod
  issuer: msg-push
  access_token_expire: 2h
  refresh_token_expire: 168h

taskq:
  redis_addr: 127.0.0.1:6379
  concurrency: 10      # worker 并发消费数
  default_max_retry: 3
  default_timeout: 30m
  default_queue: default

account_seed:
  username: admin
  password: admin123
```

## 字段说明

### app

| 字段 | 默认 | 说明 |
|------|------|------|
| `name` | msg-push | 应用名 |
| `env` | dev | 运行环境，`options=[dev,test,prod]` |

### server

HTTP 服务监听地址与端口。

### logger

| 字段 | 说明 |
|------|------|
| `level` | 日志级别（zapcore.Level） |
| `encoding` | console / json |
| `output` | 输出目标列表（stdout / 文件路径） |
| `app_name` | 日志中的应用名 |

### db

使用 infra-go ORM 配置。

- **MySQL（默认）**：`driver: mysql` + host/port/username/password/database；
- **SQLite**：`driver: sqlite, database: msg-push.db`（开箱即用，无需 MySQL）；
- 也支持 `postgres`。

> 敏感信息建议用环境变量覆盖：`DB_HOST`、`DB_USER`、`DB_PASSWORD`、`DB_NAME` 等。

### redis

| 字段 | 说明 |
|------|------|
| `addr` | Redis 地址 |
| `key_prefix` | key 前缀（多环境隔离） |
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
| `redis_addr` | 队列使用的 Redis 地址 |
| `concurrency` | Worker 并发消费数（默认 10） |
| `default_max_retry` | 任务默认最大重试次数 |
| `default_timeout` | 任务默认超时 |
| `default_queue` | 默认队列名 |

### account_seed

管理端种子账号，启动时直接创建（数据库需为全新/已手动清理状态）：

```bash
# 建议通过环境变量覆盖初始密码
ACCOUNT_SEED_USERNAME=admin ACCOUNT_SEED_PASSWORD=your-secret go run .
```

## 环境变量覆盖

infra-go `conf` 支持用环境变量覆盖 yaml 值，命名规则为 `大写配置路径`：

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
