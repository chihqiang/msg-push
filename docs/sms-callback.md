# 短信服务商回调地址

短信发出后，最终是否送达（成功 / 失败 / 被拒）需要**服务商主动回调**通知 msg-push。本文说明回调地址的格式、如何在服务商控制台配置、各服务商回执格式差异，以及验证与排查方法。

> 回执是短信任务**终态闭环**的三路写入之一：**回执回调 / 状态主动补单 / 超时扫描**。回调配好后，短信发送、回执、Webhook 通知全链路可追踪。

## 一、回调地址格式

msg-push 提供统一的回执接收地址，**无需认证**（公开路由，供服务商调用）：

```
POST /api/callback/{provider_account_id}
GET  /api/callback/{provider_account_id}
```

- `{provider_account_id}`：**服务商账号的主键 ID**（不是通道、也不是服务商编码）；
- POST 与 GET 均支持（部分服务商用 GET 验证地址或推送）；
- 服务商以 **POST + JSON/表单** 推送回执为主。

### 完整地址示例

假设服务商账号 ID 为 `1`，服务部署在 `https://push.example.com`：

```
https://push.example.com/api/callback/1
```

## 二、如何获取服务商账号 ID

服务商账号 ID 即「服务商管理 → 服务商账号」列表中该账号的 **ID**（数据库 `msg_provider_accounts.id` 主键）。

获取方式：

1. **管理后台**：服务商管理列表每行对应一个账号，其 ID 在列表/详情中可见；
2. **接口查询**：

```bash
# 登录获取 token
TOKEN=$(curl -s -X POST http://127.0.0.1:8080/api/v1/account/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin123"}' | python3 -c "import sys,json;print(json.load(sys.stdin)['data']['access_token'])")

# 列出服务商账号，响应 data.list[].id 即回调地址所需 ID
curl -s http://127.0.0.1:8080/api/v1/account/provider-accounts \
  -H "Authorization: Bearer $TOKEN" | python3 -m json.tool
```

> 若更换了服务商账号（删除重建），主键 ID 会变化，需同步更新服务商控制台的回调地址。

## 三、在服务商控制台配置

将回调地址（含 `/api/callback/{id}`）填入服务商短信平台的回调配置项（一般叫「上行/下行回执推送 URL」「状态报告回调地址」）。

### 各服务商回执格式

msg-push 内置了阿里云 / 腾讯云 / 网易云信三个短信服务商的**定制解析器**（`sender/callback_parsers.go`），自动按各自格式解析：

| 服务商 | 回执格式 | 成功响应（服务商期望） |
|--------|----------|------------------------|
| 阿里云短信 | JSON 数组 `[{"phone_number":"138****","success":true,"err_code":"DELIVRD","biz_id":"...",...}]` | `{"code":0,"msg":"接收成功"}` |
| 腾讯云短信 | JSON 数组 `[{"mobile":"138****","report_status":"SUCCESS","sid":"...",...}]` | `{"result":0,"errmsg":"OK"}` |
| 网易云信 | `{"eventType":"11","objects":[...]}`（11=下行回执，12=上行短信） | `{"code":200,"msg":"success"}` |
| 其他服务商 | 走兜底通用解析（匹配率较低，建议为定制服务商实现 `CallbackParser`） | `{"code":0,"message":"ok"}` |

### 配置要点

1. **回调地址**填完整 URL，路径必须带 `/{provider_account_id}`；
2. **推送方式**用 POST，Content-Type 按服务商默认（JSON / 表单均可，msg-push 都会解析）；
3. **签名/鉴权**：msg-push 的回调地址是公开的，不校验服务商签名（回执本身是弱敏感数据）；如需更高安全性，可在网关层（Nginx/WAF）限制来源 IP；
4. 解析失败时 msg-push 也会返回服务商期望的成功响应，**避免服务商反复重推**。

## 四、回执处理链路

服务商回调到达后，msg-push 依次执行：

```mermaid
flowchart LR
    P[服务商回调<br/>POST /api/callback/{id}] --> A[CallbackHandler]
    A --> B[按 provider_code 选解析器<br/>或通用兜底解析]
    B --> C[落库 CallbackLog]
    C --> D[按 provider_msg_id + receiver<br/>关联 PushLog / PushTask]
    D --> E[更新任务终态<br/>success / failed]
    D --> F[触发 Webhook 通知<br/>delivered / failed]
```

- 回执按 `(provider_msg_id + receiver)` **尽力关联**任务（批量场景同一 BizId 多条记录时按手机号精确定位）；
- 回填 `callback_status`（success/failed/timeout）、`callback_time`，同步 `push_task.status`；
- 触发 Webhook 事件（`delivered` / `failed`），见 [Webhook 通知](webhook.md)。

## 五、验证与排查

### 用 curl 模拟回执（联调）

```bash
# 模拟阿里云短信回执（替换 {id} 为服务商账号 ID）
curl -X POST http://127.0.0.1:8080/api/callback/{id} \
  -H "Content-Type: application/json" \
  -d '[{"phone_number":"13800000001","success":true,"err_code":"DELIVRD","err_msg":"","biz_id":"12345^67890","report_time":"2026-08-16 12:00:00"}]'
```

### 查看处理结果

1. **回调日志**：管理后台「日志查询 → 回调」Tab，或 `GET /api/v1/account/callbacks`；
2. **任务状态**：`GET /api/v1/tasks/{task_id}`，短信任务从 `sending` 变 `success`/`failed`，`callback_status` 非空；
3. **Webhook 日志**：管理后台「Webhook 配置」→ 查看投递记录。

### 常见问题

| 现象 | 原因 / 解决 |
|------|------|
| 任务一直 `sending`，无回执 | ① 服务商回调 URL 未配置或填错；② 回调地址里的账号 ID 与发送时选中的服务商账号不一致 |
| 回调日志有记录但任务状态未更新 | 回执 `provider_msg_id` / 手机号无法匹配到 `push_log`（如批量场景回执未带手机号）；可在日志查询按 request_id 排查 |
| 服务商反复推送同一条回执 | 解析失败或响应码不对——检查返回的服务商期望响应体是否匹配（见上表） |
| 服务商无回调能力 | 可依赖**状态主动补单**（StatusPuller 主动查询）+ **超时扫描**（SMSTimeoutScanner 兜底判失败）双保险，无需回调也能终态闭环 |

## 六、接入新的短信服务商

如需为自定义/其他短信服务商接入回执，请参考 [新增短信服务商指南](add-new-sms-provider.md) 的「实现回调解析器」章节：实现 `CallbackParser` 并在 `sender/callback_parsers.go` 注册，即可获得与阿里/腾讯/网易同等的定制解析能力。

## 相关文档

- [API 使用指南](api-guide.md)（回调接收接口定义）
- [配置与发送指南](setup-and-send.md)（完整配置流程）
- [新增短信服务商](add-new-sms-provider.md)（回调解析器实现）
- [Webhook 通知](webhook.md)（回执触发的事件通知）
- [失败处理与重试](failure-retry.md)（状态主动补单 / 超时扫描）
