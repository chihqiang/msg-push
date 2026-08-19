# 快速开始

本指南带你从零启动 msg-push 并完成第一次消息发送。

## 环境要求

| 依赖 | 版本 | 说明 |
|------|------|------|
| Go | 1.25+ | 运行后端 |
| Docker | 任意 | 提供 MySQL + Redis |
| Node.js | 18+ | 仅运行管理后台时需要 |

## 第一步：启动基础服务

项目根目录提供 `deploy/docker-compose.yml`，一键启动 MySQL 与 Redis：

```bash
docker compose -f deploy/docker-compose.yml up -d

# 查看状态
docker compose -f deploy/docker-compose.yml ps
```

连接信息（对应 `config.yaml` 默认值）：

- MySQL：`127.0.0.1:3306`，用户 `root`，密码 `msgpush123456`，库名 `msg_push`
- Redis：`127.0.0.1:6379`，无密码

> 不使用 Docker 时也可直接改用 SQLite：把 `config.yaml` 的 `db` 段改为
> `driver: sqlite, database: msg-push.db` 即可开箱即用（无需 Redis 以外依赖）。

## 第二步：启动后端

```bash
go mod download
go run .
```

启动过程会自动：

1. 加载 `config.yaml`；
2. 连接 MySQL / Redis；
3. 执行 `db.Migrate` 自动建表；
4. 填充种子数据：种子应用、管理账号（通道表为空，需在管理后台自行创建）。

> 数据库需为全新或已手动清理的状态（清理由运维手动完成，代码不负责清表），否则种子数据会与旧数据冲突。

健康检查：

```bash
curl http://127.0.0.1:8080/health
# {"code":0,"msg":"ok","data":{"status":"ok"}}
```

## 第三步：发送第一条消息

后端默认内置一个**测试应用**：

- `app_id`: `test-app`
- `app_secret`: `test-secret`
- 模式：**测试模式**（`is_test=true`，走完整链路但不真实发送，模拟成功）

```bash
curl -X POST http://127.0.0.1:8080/api/v1/messages \
  -H "X-App-Id: test-app" \
  -H "X-App-Secret: test-secret" \
  -H "Content-Type: application/json" \
  -d '{"channel_code":"aliyun_sms","receiver":"13800000001","template_params":{"code":"1234"}}'
```

响应示例：

```json
{
  "code": 0,
  "msg": "ok",
  "data": {
    "task_id": "task_xxxx",
    "status": "pending",
    "is_test": true,
    "created_at": "2026-08-15T10:00:00+08:00"
  },
  "request_id": "a1b2c3d4-..."
}
```

等待几秒后查询任务状态，测试任务会变为 `success`：

```bash
curl http://127.0.0.1:8080/api/v1/tasks/task_xxxx \
  -H "X-App-Id: test-app" -H "X-App-Secret: test-secret"
```

## 第四步：启动管理后台（可选）

`webui/` 是独立的 git 子模块（Vue 3 管理控制台）：

```bash
cd webui
npm install
npm run dev
```

访问 `http://localhost:5173`，默认账号 `admin / admin123`。

## 下一步

- 了解配置项：[配置指南](configuration.md)
- 了解发送 API 与认证：[API 使用指南](api-guide.md)
- 理解架构与链路：[架构设计](architecture.md)
- 配置真实服务商并发送真实消息：[测试模式](testing-mode.md) 与 [配置指南](configuration.md)
