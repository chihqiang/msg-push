# 开发者指南

面向希望二次开发或贡献代码的开发者。

## 一、本地开发环境

```bash
# 1. 基础服务
docker compose -f deploy/docker-compose.yml up -d

# 2. 后端
go mod download
go run .        # 或 go run . -config config.yaml

# 3. 前端（webui 子模块）
cd webui && npm install && npm run dev
```

### 常用命令

```bash
# 编译 + 静态检查 + 测试
go build ./...
go vet ./...
go test ./...

# 前端类型检查
cd webui && npm run typecheck

# 前端构建
npm run build
```

## 二、代码结构速览

```
main.go              # 入口：配置/日志/依赖/迁移/服务编排
route/               # 路由注册（应用侧 + 管理侧 + 回调）
handler/             # HTTP 层：绑定、响应、错误
logic/               # 业务逻辑：落库、入队、查询
middleware/          # 应用鉴权(HMAC)/账号JWT/配额/限流
model/               # GORM 模型（19 张表）
dto/                 # 请求/响应结构
core/pipeline/       # 消息管道：队列契约+生产者/消费者、选择器、渲染器、规则引擎、终态
core/scheduler/      # 后台调度：配额同步、超时扫描、状态补单、Webhook outbox 投递
core/sender/         # 服务商发送器（8 个）+ 服务商元信息注册表
core/app/            # 后台服务装配：NewServiceGroup 统一启停
svc/                 # 依赖装配
db/                  # 迁移 + 种子
deploy/              # docker-compose
webui/               # 管理后台（子模块）
```

## 三、核心接口

### 发送器 Sender

```go
type Sender interface {
    Send(ctx context.Context, req *SendRequest) (*SendResponse, error)
    GetProviderCode() string
}
```

- `SendRequest` 含任务、服务商账号、通道绑定、签名、映射参数、渲染内容、手机号预解析；
- `SendResponse` 含成功标志、服务商消息 ID、状态（`sending`/`success`）、错误码、请求/响应数据。

### 批量发送器 BatchSender

```go
type BatchSender interface {
    Sender
    BatchSend(ctx context.Context, req *BatchSendRequest) (*BatchSendResponse, error)
    SupportsBatchSend() bool
}
```

- `BatchSendRequest.Tasks` 与 `TaskParams` 一一对应；
- `BatchSendResponse.Results` 与 `Tasks` 一一对应；
- 不支持批量时返回 false，消费端自动退回逐条。

### 状态查询器 StatusQuerier

```go
type StatusQuerier interface {
    QueryStatus(ctx context.Context, req *StatusQueryRequest) (*StatusQueryResponse, error)
    GetProviderCode() string
}
```

短信类服务商可选实现，供 `status_puller` 主动补单回执。

### 回调解析器 CallbackParser

服务商定制回执解析（`core/sender/xxx_sms.go` 内实现 `CallbackParser`），将原始回执解析为统一结果：

```go
type CallbackParser interface {
    Parse(ctx context.Context, req *CallbackRequest) (*CallbackResponse, []*CallbackResult, error)
}
```

## 四、如何新增一个服务商

以新增"某某短信"为例（**完整、可照着做的步骤见 [新增短信服务商指南](add-new-sms-provider.md)**，以下为摘要）：

### 1. 注册服务商元信息（core/sender/sender.go）

在 `providerMetas` 列表追加一条 `Meta`：

```go
{
    Code: "xxx_sms", Name: "某某短信", Type: TypeSMS, SortOrder: 4,
    SupportsSend: true, SupportsBatchSend: false, SupportsCallback: true, RequiresSignature: true,
    ConfigFields: []ConfigField{ ... }, // 配置字段定义（前端表单自动生成）
}
```

### 2. 实现发送器（core/sender/xxx_sms.go）

```go
type XxxSMSSender struct{}

func (s *XxxSMSSender) GetProviderCode() string { return "xxx_sms" }

func (s *XxxSMSSender) Send(ctx context.Context, req *SendRequest) (*SendResponse, error) {
    // 1. 解析账号配置（req.ProviderAccount.Config）
    // 2. 选手机号（短信用 smsReceiver(req)，国内/国际自动规范化）
    // 3. 签名/模板处理
    // 4. 调服务商 HTTP API
    // 5. 返回 SendResponse（短信提交成功返回 Status="sending"）
}
```

### 3. 注册到工厂（core/sender/factory.go）

```go
f.Register(&XxxSMSSender{})
```

### 4.（可选）实现批量 / 状态查询 / 回调解析

按需实现 `BatchSender` / `StatusQuerier`，并在 `core/sender/xxx_sms.go` 内实现 `CallbackParser`（注册表见 `callback.go`）。

### 5. 测试

在管理后台配置服务商账号，创建通道-模板绑定，用测试模式先验证链路，再真实发送验证。

## 五、新增后台调度器

1. 在 `core/scheduler/` 新建调度器（参考 `quota_syncer.go`）；
2. 在 `core/app/app.go` 的 `NewServiceGroup` 中 `sg.Add(...)` 纳入统一启停（组件需实现无返回值的 `Start()`/`Stop()`）。

## 六、新增数据表

1. 在 `model/` 新建模型（注意 `TableName()` 返回 `msg_` 前缀表名）；
2. 在 `db/migrate.go` 的 `Migrate` 列表追加；
3. 重启即自动建表（数据库清理由运维手动完成）。

## 七、编码约定

- 分层清晰：handler 不写业务逻辑，logic 不写 SQL 细节，统一经 `svc.ServiceContext` 访问依赖；
- 错误处理：业务错误返回给调用方，基础设施错误记日志；
- 日志：用 `logger.Infof/Errorf`，需要带上下文用 `InfoCtx/ErrorCtx`（自动含 request_id）；
- 新增字段：优先用 `bool/int8` 等简单类型，指针类型表示"未传"时用 `*bool`；
- 全链路追踪：任何新增的异步/重试路径都要透传 `request_id`。

## 相关文档

- [架构设计](architecture.md)
- [数据模型](database.md)
- [失败处理与重试](failure-retry.md)
- [运维与可观测](operations.md)
