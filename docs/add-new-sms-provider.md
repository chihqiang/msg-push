# 新增短信服务商指南

本文面向需要接入一个新短信服务商（如"某某云短信"）的开发者。按照本文步骤，你只需要改动 **3 个必改文件** 和按需实现的若干可选接口，即可让新服务商完成「发送 → 回执」全链路接入，前端管理后台表单也会自动生成，无需改动前端代码。

## 一、整体链路与改动清单

新服务商接入后，一次投递的完整链路：

```mermaid
flowchart LR
    W[Worker 消费端] -->|DefaultResolver.GetSender| S[你的发送器<br/>core/sender/xxx_sms.go]
    S -->|HTTP 调用| P[服务商 API]
    P -->|回执回调 POST /api/callback/{id}| CB[CallbackHandler]
    CB -->|GetCallbackParser| CP[回调解析器<br/>core/sender/xxx_sms.go]
    CP -->|统一结果| DB[(MySQL)]
```

### 需要改动的文件

| 文件 | 是否必改 | 说明 |
|------|:---:|------|
| `core/sender/sender.go` | ✅ 必改 | 注册服务商元信息（`registry` 清单 + `Meta` 条目），驱动管理后台表单 |
| `core/sender/xxx_sms.go` | ✅ 必改 | 新建发送器，实现 `Sender` 接口 |
| `core/sender/factory.go` | ✅ 必改 | 把发送器注册进工厂 `NewFactory()` |
| `core/sender/xxx_sms.go` | 短信建议实现 | 实现 `CallbackParser`，解析服务商回执 |
| 发送器内 | 可选 | 实现 `BatchSender` 批量发送 |
| 发送器内 | 可选 | 实现 `StatusQuerier` 主动查询回执（补单） |
| `core/sender/xxx_sms_test.go` | 建议 | 参照 `senders_test.go` 编写单测 |

> 新增一个短信服务商通常**不需要**改动 `core/pipeline/`、`handler/`、`model/`、前端代码：消费端只依赖 `sender.DefaultResolver` 与接口抽象，前端表单由 `sender.Meta.ConfigFields` 动态生成。

## 二、必改步骤

### 步骤 1：注册服务商元信息（`core/sender/sender.go`）

#### 1.1 添加服务商代码常量

在 `types.go` 顶部的常量区追加：

```go
// 服务商代码常量。
const (
    // ... 已有常量
    CodeXxxSMS = "xxx_sms" // 某某云短信
)
```

> 服务商代码是全局唯一标识，会写入 `msg_provider_accounts.provider_code`，发送器与回调解析器都通过它互相匹配，务必保证三处完全一致。

#### 1.2 在 `registry` 列表追加一条 `Meta`

参照现有的阿里/腾讯/网易短信条目，追加：

```go
{
    Code: CodeXxxSMS, Name: "某某云短信", Type: TypeSMS, SortOrder: 4,
    Description:       "某某云短信服务",
    SupportsSend:      true,
    SupportsBatchSend: true,   // 是否支持批量 API
    SupportsCallback:  true,   // 是否支持回执回调
    RequiresSignature: true,   // 短信一般强制要求签名
    Website:           "https://example.com/",
    DocsUrl:           "https://example.com/docs/sms",
    ConsoleUrl:        "https://console.example.com/",
    Tags:              []string{"国内"},
    Regions:           []string{"cn"},
    ConfigFields: []ConfigField{
        {Key: "access_key", Label: "AccessKey", Type: "text", Required: true, Placeholder: "某某云 AccessKey"},
        {Key: "secret_key", Label: "SecretKey", Type: "password", Required: true, Placeholder: "某某云 SecretKey"},
    },
},
```

`ConfigField` 各字段说明：

| 字段 | 说明 |
|------|------|
| `Key` | 配置参数 key，存进 `ProviderAccount.Config`（JSON）后，发送器用 `strVal(cfg, key)` 读取 |
| `Label` / `Placeholder` | 表单展示 |
| `Type` | `text` / `password` / `number` / `url` / `textarea` / `select` |
| `Required` | 是否必填 |
| `Options` | `select` 类型的下拉选项 |
| `DefaultValue` | 默认值 |
| `ValidationRule` / `HelpLink` | 表单校验规则 / 帮助链接（可选） |

**能力声明字段**直接影响消费端行为：

- `RequiresSignature = true`：该服务商必须配置签名，消费端会校验通道-签名映射，否则投递早期失败；
- `SupportsBatchSend = true`：批量任务会优先尝试调用批量 API；
- `SupportsCallback = true`：用于前端提示与回执功能展示。

> `SortOrder` 控制列表排序。前端 `webui/src/components/provider/AccountFormDialog.vue` 会读取 `/account/provider-config-fields/{code}`（由 `GetProviderConfigFields` handler 返回 `Meta.ConfigFields`）自动渲染表单，**无需改前端**。

### 步骤 2：实现发送器（新建 `core/sender/xxx_sms.go`）

新文件实现 `Sender` 接口：

```go
// Sender 发送器接口（core/sender/types.go 定义）。
type Sender interface {
    Send(ctx context.Context, req *SendRequest) (*SendResponse, error)
    GetProviderCode() string
}
```

参照 `aliyun_sms.go` / `tencent_sms.go`，完整骨架如下：

```go
package sender

import (
    "context"
    "net/http"
    "strings"
    "time"

    "chihqiang/msg-push/model"
)

// 某某云短信接口端点。
var xxxSmsEndpoint = "https://sms.example.com/send"

// XxxSMSSender 某某云短信发送器。
type XxxSMSSender struct{}

// GetProviderCode 返回服务商代码（必须与 registry 常量一致）。
func (s *XxxSMSSender) GetProviderCode() string {
    return CodeXxxSMS
}

// Send 发送短信。
func (s *XxxSMSSender) Send(ctx context.Context, req *SendRequest) (*SendResponse, error) {
    taskID := ""
    if req.Task != nil {
        taskID = req.Task.TaskID
    }

    // 1. 读取账号配置（JSON -> map -> 取值），配置缺失返回业务错误而非 panic
    if req.ProviderAccount == nil {
        return &SendResponse{Success: false, TaskID: taskID, ErrorMessage: "provider account missing", ErrorCode: "CONFIG_ERROR"}, nil
    }
    cfg, err := configMap(req.ProviderAccount)
    if err != nil {
        return &SendResponse{Success: false, TaskID: taskID, ErrorMessage: "invalid config: " + err.Error(), ErrorCode: "CONFIG_ERROR"}, nil
    }
    accessKey := strVal(cfg, "access_key")
    secretKey := strVal(cfg, "secret_key")
    if accessKey == "" || secretKey == "" {
        return &SendResponse{Success: false, TaskID: taskID, ErrorMessage: "access_key/secret_key required", ErrorCode: "CONFIG_ERROR"}, nil
    }

    // 2. 手机号：短信服务商统一用 smsReceiver(req) 规范化
    //    国内号（region=CN）取 11 位 national；国际号取 E.164（+国家码号码）
    receiver := smsReceiver(req)

    // 3. 签名与模板：签名在通道-签名映射里，模板在通道-模板绑定里
    signName := ""
    if req.Signature != nil {
        signName = req.Signature.SignatureCode // 与服务商平台报备一致的签名
    }
    templateCode := ""
    if req.ChannelTemplateBinding != nil && req.ChannelTemplateBinding.ProviderTemplate != nil {
        templateCode = req.ChannelTemplateBinding.ProviderTemplate.TemplateCode
    }

    // 4. 模板参数：req.MappedParams 为「供应商变量名 -> 值」映射
    //    服务商支持 JSON 参数对象时可直接 jsonDump(req.MappedParams)；
    //    若需按占位符顺序传值，用 templateKeys(templateContent) / sortedParamsFromMap(mapped) 取有序数组
    templateParam := jsonDump(req.MappedParams)

    // 5. 构造请求（按服务商要求拼装、签名、鉴权）
    body := strings.NewReader("...") // 按服务商协议组装
    httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, xxxSmsEndpoint, body)
    if err != nil {
        return &SendResponse{Success: false, TaskID: taskID, ErrorMessage: err.Error(), ErrorCode: "HTTP_ERROR"}, nil
    }
    // 设置鉴权 Header ...

    client := &http.Client{Timeout: 10 * time.Second}
    resp, err := client.Do(httpReq)
    if err != nil {
        return &SendResponse{Success: false, TaskID: taskID, ErrorMessage: err.Error(), ErrorCode: "HTTP_ERROR"}, nil
    }
    defer resp.Body.Close()
    respBody := new(strings.Builder)
    _, _ = copyBuffer(respBody, resp.Body)
    responseData := jsonDump(map[string]any{"status": resp.StatusCode, "body": respBody.String()})

    // 6. 解析响应并返回
    //    短信提交成功（进入等待回执）→ Status = "sending"，并回填服务商消息 ID（ProviderID 供回执关联）
    if success {
        return &SendResponse{
            Success:      true,
            TaskID:       taskID,
            ProviderID:   result.MsgID,                    // 服务商消息ID，回调用它关联任务
            Status:       string(model.PushTaskStatusSending), // "sending"：等待回执
            RequestData:  form.Encode(),                   // 请求快照（存 PushLog，便于排查）
            ResponseData: responseData,                    // 响应快照
        }, nil
    }
    //    失败 → Success=false + 错误码/信息，消费端会按失败规则重试/切服务商
    return &SendResponse{
        Success: false, TaskID: taskID,
        ErrorCode: result.Code, ErrorMessage: result.Message,
        RequestData: form.Encode(), ResponseData: responseData,
    }, nil
}
```

**发送器实现要点：**

| 关注点 | 约定 |
|--------|------|
| 手机号 | 短信统一调 `smsReceiver(req)`（国内 11 位 / 国际 E.164 自动规范化），不要直接拼 `Task.Receiver` |
| 模板代码 | 从 `req.ChannelTemplateBinding.ProviderTemplate.TemplateCode` 取（就是服务商平台的模板 ID） |
| 签名 | 从 `req.Signature.SignatureCode` 取 |
| 参数 | `req.MappedParams`（`map[string]string`，供应商变量名→值）；按序取值时用 `templateKeys`/`sortedParamsFromMap`（map 遍历无序，直接遍历会导致参数错位） |
| 返回状态 | **短信提交成功返回 `Status="sending"`**（等待回执，避免成功数虚高）；直接完成才返回 `"success"` |
| ProviderID | 发送成功时务必回填服务商消息 ID——回执回调按 `(provider_msg_id + receiver)` 关联任务 |
| 错误码 | 业务失败返回 `ErrorCode`/`ErrorMessage`，不要用 Go error 表达"业务拒绝"（消费端会把非 nil error 当基础设施异常处理） |
| 配置读取 | 用包内工具 `configMap` / `strVal` / `intVal` / `boolVal`，不要自己解析 `Config` JSON |
| 超时 | HTTP client 统一 `Timeout: 10 * time.Second` |

### 步骤 3：注册到发送器工厂（`core/sender/factory.go`）

在 `NewFactory()` 中追加一行：

```go
func NewFactory() *Factory {
    f := &Factory{senders: map[string]Sender{}}
    f.Register(&AliyunSMSSender{})
    f.Register(&TencentSMSSender{})
    f.Register(&NeteaseSMSSender{})
    f.Register(&XxxSMSSender{}) // 新增：某某云短信
    // ... 其余服务商
    return f
}
```

消费端通过 `sender.DefaultResolver.GetSender(providerCode)` 获取发送器；批量通过 `GetBatchSender` 获取，未实现批量时自动退回逐条。

## 三、可选步骤

### 步骤 4：实现回调解析器（`core/sender/xxx_sms.go`）

短信服务商通常有回执回调，建议实现 `CallbackParser`，否则回执只能走兜底通用解析（匹配率低）。

```go
// CallbackParser 接口。
type CallbackParser interface {
    Parse(ctx context.Context, req *CallbackRequest) (CallbackResponse, []*CallbackResult, error)
    GetProviderCode() string
}
```

```go
// ==================== 某某云短信 ====================

// XxxCallbackParser 某某云短信回执解析器。
// 假设回执格式：{"records":[{"mobile":"138****","status":"SUCCESS","errmsg":"DELIVRD","sid":"xxx","report_time":"2026-08-16 12:00:00"}]}
type XxxCallbackParser struct{}

func (p *XxxCallbackParser) GetProviderCode() string { return CodeXxxSMS }

func (p *XxxCallbackParser) Parse(ctx context.Context, req *CallbackRequest) (CallbackResponse, []*CallbackResult, error) {
    var payload struct {
        Records []struct {
            Mobile     string `json:"mobile"`
            Status     string `json:"status"`
            ErrMsg     string `json:"errmsg"`
            SID        string `json:"sid"`
            ReportTime string `json:"report_time"`
        } `json:"records"`
    }
    if err := json.Unmarshal(req.RawBody, &payload); err != nil {
        // 解析失败也返回服务商期望的成功响应，避免服务商反复推送
        return GenericCallbackOK, nil, fmt.Errorf("invalid xxx callback: %w", err)
    }

    results := make([]*CallbackResult, 0, len(payload.Records))
    for _, r := range payload.Records {
        status := model.CallbackStatusDelivered
        if r.Status != "SUCCESS" {
            status = model.CallbackStatusFailed
        }
        reportTime, _ := time.ParseInLocation("2006-01-02 15:04:05", r.ReportTime, time.Local)
        results = append(results, &CallbackResult{
            Type:         string(model.CallbackTypeReport), // report=投递回执 / upstream=上行
            ProviderID:   r.SID,                            // 必须与发送响应的 ProviderID 一致
            Status:       status,                           // delivered / failed / rejected
            ErrorCode:    r.ErrMsg,
            ErrorMessage: r.ErrMsg,
            Mobile:       r.Mobile,
            ReportTime:   reportTime,
        })
    }
    return GenericCallbackOK, results, nil
}
```

然后在文件底部 `init()` 中注册：

```go
func init() {
    RegisterCallbackParser(&AliyunCallbackParser{})
    RegisterCallbackParser(&TencentCallbackParser{})
    RegisterCallbackParser(&NeteaseCallbackParser{})
    RegisterCallbackParser(&XxxCallbackParser{}) // 新增
}
```

**回调要点：**

- 成功响应体在 `core/sender/callback.go` 有现成常量（`AliyunCallbackOK`、`TencentCallbackOK`、`NeteaseCallbackOK`、`GenericCallbackOK`），按服务商要求返回；
- `CallbackResult.ProviderID` 必须与发送响应里的 `ProviderID` 一致，回调逻辑按 `(provider_msg_id + receiver)` 尽力关联到 `PushLog`/`PushTask`；
- 回调地址：`POST /api/callback/{id}`，`{id}` 是服务商账号 ID（`msg_provider_accounts.id`），把这个地址配到服务商后台的回调 URL；
- 解析失败时也要返回服务商约定的成功响应，避免服务商重复推送。

### 步骤 5：实现批量发送（`BatchSender`）

服务商提供批量 API（一次请求发多个号码）时实现，消费端批量任务会聚合整批号码调用：

```go
type BatchSender interface {
    Sender
    BatchSend(ctx context.Context, req *BatchSendRequest) (*BatchSendResponse, error)
    SupportsBatchSend() bool
}
```

- `BatchSendRequest.Tasks` 与 `TaskParams` 一一对应（每任务的独立参数映射，为空回退共用 `MappedParams`）；
- `BatchSendResponse.Results` 必须与 `Tasks` 一一对应，否则消费端回填会错位；
- `SupportsBatchSend()` 返回 `true`；
- 批量成功但服务商只返回一个统一消息 ID 时，每个号码的最终回执仍靠回调按 `(provider_msg_id + receiver)` 尽力关联（参考阿里云实现）。

### 步骤 6：实现状态查询（`StatusQuerier`）

服务商无回执或回执延迟时，实现该接口可让 `core/scheduler/status_puller.go` **主动补单**（发送后一段时间无回执主动查服务商真实状态）：

```go
type StatusQuerier interface {
    QueryStatus(ctx context.Context, req *StatusQueryRequest) (*StatusQueryResponse, error)
    GetProviderCode() string
}
```

- `StatusQueryRequest` 含任务、服务商账号、服务商消息 ID、手机号、发送日期；
- `StatusQueryResult.Status` 返回 `delivered` / `failed` / `unknown`；
- 实现后无需注册，消费端用 `sd.(sender.StatusQuerier)` 类型断言自动发现。

## 四、编写测试

参照现有 `senders_test.go` 与 `callback_parsers_test.go`，用 `httptest` 起 mock server 断言请求头/请求体与响应解析：

```go
func TestXxxSendSuccess(t *testing.T) {
    srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // 断言鉴权头、手机号、模板参数等
        w.Write([]byte(`{"code":"OK","msg_id":"SID_001"}`))
    }))
    defer srv.Close()
    xxxSmsEndpoint = srv.URL // 覆盖包级端点变量

    pa := newTestPA(map[string]any{"access_key": "ak", "secret_key": "sk"})
    req := newSMSRequest(pa, "13800138000")
    s := &XxxSMSSender{}
    resp, err := s.Send(context.Background(), req)
    // 断言 resp.Success == true、Status == "sending"、ProviderID == "SID_001"
}
```

建议覆盖：
- 配置缺失 → `CONFIG_ERROR` 业务错误（返回 nil error）；
- 发送成功 → `Status="sending"` + `ProviderID`；
- 服务商返回业务失败码 → `Success=false` + 错误码；
- HTTP 异常 → `HTTP_ERROR`；
- 回调解析：合法回执、未知 eventType、畸形 body（仍返回 200 成功响应）。

## 五、本地验证清单

1. `go build ./... && go vet ./... && go test ./...` 通过；
2. 启动服务（`docker compose -f deploy/docker-compose.yml up -d` + `go run .`）；
3. 管理后台（webui）新增服务商账号，确认表单字段已自动生成（无需改前端）；
4. 创建服务商签名、服务商模板，并建立「通道 → 模板绑定」+「通道 → 签名映射」；
5. 用测试模式（`is_test=true`）走一遍完整链路（提交→入队→消费→模拟成功），确认状态流转与日志；
6. 关闭测试模式真实发送一条，确认服务商侧收到短信、`PushLog` 落库、`PushTask` 进入 `sending`；
7. 配置服务商后台回调 URL 为 `POST /api/callback/{id}`，用真实回执验证 `CallbackLog` 落库、任务转 `success`/`failed`、webhook 触发。

## 六、最终检查清单

- [ ] `core/sender/types.go`：`Code` 常量（`Type=TypeSMS`）
- [ ] `core/sender/sender.go`：`Meta` 条目，配置字段完整
- [ ] `core/sender/xxx_sms.go`：实现 `Sender`，`GetProviderCode()` 与常量一致
- [ ] `core/sender/factory.go`：`NewFactory()` 中 `f.Register(...)`
- [ ] 手机号用 `smsReceiver(req)` 规范化，不使用裸 `Task.Receiver`
- [ ] 发送成功回填 `ProviderID`，短信返回 `Status="sending"`
- [ ] 业务失败返回 `Success=false` + `ErrorCode`/`ErrorMessage`（nil error）
- [ ] （建议）`xxx_sms.go` 实现并注册 `CallbackParser`，`ProviderID` 与发送侧一致
- [ ] （可选）批量/状态查询按需实现
- [ ] 单测通过，真实链路按「五」验证

## 参考：现有短信服务商对照

| 服务商 | 发送器 | 批量 | 状态查询 | 回调解析器 | 签名机制 |
|--------|--------|:---:|:---:|--------|------|
| 阿里云 | `aliyun_sms.go` | ✅ | ✅ | `AliyunCallbackParser` | RPC V1 HMAC-SHA1 |
| 腾讯云 | `tencent_sms.go` | ✅ | — | `TencentCallbackParser` | TC3-HMAC-SHA256 |
| 网易云信 | `netease_sms.go` | ✅ | — | `NeteaseCallbackParser` | AppKey + Nonce + CheckSum |

## 相关文档

- [开发者指南](development.md)（核心接口定义）
- [架构设计](architecture.md)（投递/回调链路）
- [数据模型](database.md)（服务商账号/签名/模板表）
- [失败处理与重试](failure-retry.md)（失败后重试/切换供应商）
- [测试模式](testing-mode.md)（联调验证）
