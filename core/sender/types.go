// Package sender 提供消息发送器层：
//   - 定义 Sender / Resolver / BatchSender / StatusQuerier / CallbackParser 接口
//   - 常用请求/响应结构体、服务商元信息注册表、发送器工厂与回调解析器注册表
//   - 公共方法与工具函数（配置读取、HTTP、JSON、加密、字符串、模板）
//
// 各服务商实现按"一个服务商一个文件"组织（aliyun_sms.go / tencent_sms.go /
// netease_sms.go / smtp.go / wechat_work.go / dingtalk.go / wechat_work_robot.go /
// dingtalk_robot.go），供消费端按 provider_code 解析调用。
package sender

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"chihqiang/msg-push/model"
)

// ==================== 常量 ====================

// 服务商代码常量。
const (
	CodeAliyunSMS       = "aliyun_sms"        // 阿里云短信
	CodeTencentSMS      = "tencent_sms"       // 腾讯云短信
	CodeNeteaseSMS      = "netease_sms"       // 网易云信短信
	CodeSMTP            = "smtp"              // SMTP邮件
	CodeWeChatWork      = "wechat_work"       // 企业微信（应用消息）
	CodeDingTalk        = "dingtalk"          // 钉钉（工作通知）
	CodeWeChatWorkRobot = "wechat_work_robot" // 企业微信群机器人
	CodeDingTalkRobot   = "dingtalk_robot"    // 钉钉群机器人
)

// 消息类型常量。
const (
	TypeSMS        = "sms"
	TypeEmail      = "email"
	TypeWeChatWork = "wechat_work"
	TypeDingTalk   = "dingtalk"
)

// ==================== 发送接口与结构体 ====================

// SendRequest 发送请求。
type SendRequest struct {
	Task                   *model.PushTask
	ProviderAccount        *model.ProviderAccount
	ChannelTemplateBinding *model.ChannelTemplateBinding
	Signature              *model.ProviderSignature
	MappedParams           map[string]string // 供应商变量名到值的映射
	RenderedContent        string            // 供应商模板渲染后的内容

	// 手机号预解析字段（仅 SMS 类型，worker 填充）
	PhoneRegion         string
	PhoneCountryCode    string
	PhoneNationalNumber string
	PhoneE164           string
}

// smsReceiver 选择发送用手机号（短信服务商通用）。
// 规则：国内号（region=CN）用 11 位 national；国际号用 E.164（+<国家码><号码>）；
// 均未解析出时回退到原始 receiver。
func smsReceiver(req *SendRequest) string {
	if req == nil {
		return ""
	}
	if req.PhoneRegion == "CN" && req.PhoneNationalNumber != "" {
		return req.PhoneNationalNumber
	}
	if req.PhoneE164 != "" {
		return req.PhoneE164
	}
	if req.Task != nil {
		return req.Task.Receiver
	}
	return ""
}

// SendResponse 发送响应。
type SendResponse struct {
	Success      bool
	ProviderID   string // 服务商消息ID
	ErrorCode    string
	ErrorMessage string
	TaskID       string
	Status       string // "sending"(等待回执) 或 "success"(直接完成)
	RequestData  string // 请求参数(JSON)
	ResponseData string // 响应数据(JSON)
}

// Sender 发送器接口。
type Sender interface {
	Send(ctx context.Context, req *SendRequest) (*SendResponse, error)
	GetProviderCode() string
}

// Resolver 发送器解析器。
type Resolver interface {
	GetSender(providerCode string) (Sender, error)
	GetBatchSender(providerCode string) (BatchSender, error)
}

// StatusQueryRequest 状态查询请求（主动回执查询）。
type StatusQueryRequest struct {
	Task            *model.PushTask
	ProviderAccount *model.ProviderAccount
	ProviderMsgID   string    // 服务商消息ID（发送响应中的 ProviderID）
	PhoneNumber     string    // 接收手机号
	SendDate        time.Time // 发送日期（按天查询）
}

// StatusQueryResult 单条状态查询结果。
type StatusQueryResult struct {
	ProviderMsgID string
	PhoneNumber   string
	Status        string // delivered / failed / unknown
	ErrorCode     string
	ErrorMessage  string
}

// StatusQueryResponse 状态查询响应。
type StatusQueryResponse struct {
	Results []*StatusQueryResult
}

// StatusQuerier 状态查询器接口（短信类服务商可选实现，用于主动补单回执）。
type StatusQuerier interface {
	QueryStatus(ctx context.Context, req *StatusQueryRequest) (*StatusQueryResponse, error)
	GetProviderCode() string
}

// BatchSendRequest 批量发送请求（批量发送）。
// 所有任务共用一个通道绑定/签名/服务商账号；TaskParams 与 Tasks 一一对应，
// 存放每个任务解析后的供应商变量映射（为空项回退到 MappedParams）。
type BatchSendRequest struct {
	Tasks                  []*model.PushTask
	ProviderAccount        *model.ProviderAccount
	ChannelTemplateBinding *model.ChannelTemplateBinding
	Signature              *model.ProviderSignature
	MappedParams           map[string]string   // 共用映射参数（无独立参数时使用）
	TaskParams             []map[string]string // 每任务独立映射参数（可空，与 Tasks 对齐）
	RenderedContent        string              // 供应商模板渲染后的内容
}

// BatchSendResponse 批量发送响应，Results 与 Tasks 一一对应。
type BatchSendResponse struct {
	Results []*SendResponse
}

// BatchSender 批量发送器接口（可选实现，用于批量消息聚合调用服务商批量 API）。
type BatchSender interface {
	Sender
	// BatchSend 批量发送消息（一次 API 调用发送多个号码）。
	BatchSend(ctx context.Context, req *BatchSendRequest) (*BatchSendResponse, error)
	// SupportsBatchSend 是否支持批量发送。
	SupportsBatchSend() bool
}

// batchTaskParams 获取批量请求第 i 个任务的模板参数映射。
// 优先使用每任务独立参数（TaskParams[i]），否则回退到共用 MappedParams。
func batchTaskParams(req *BatchSendRequest, i int) map[string]string {
	if req != nil && req.TaskParams != nil && i < len(req.TaskParams) && req.TaskParams[i] != nil {
		return req.TaskParams[i]
	}
	if req != nil {
		return req.MappedParams
	}
	return nil
}

// ==================== 配置读取公共方法 ====================

// configMap 反序列化服务商配置。
func configMap(pa *model.ProviderAccount) (map[string]any, error) {
	if pa == nil {
		return map[string]any{}, nil
	}
	return pa.GetConfig()
}

// strVal 读取字符串配置。
func strVal(cfg map[string]any, key string) string {
	if cfg == nil {
		return ""
	}
	switch v := cfg[key].(type) {
	case string:
		return v
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64)
	case int:
		return strconv.Itoa(v)
	}
	return ""
}

// intVal 读取整型配置（兼容字符串/数字）。
func intVal(cfg map[string]any, key string) int {
	if cfg == nil {
		return 0
	}
	switch v := cfg[key].(type) {
	case float64:
		return int(v)
	case int:
		return v
	case string:
		n, _ := strconv.Atoi(v)
		return n
	}
	return 0
}

// boolVal 读取布尔配置。
func boolVal(cfg map[string]any, key string) bool {
	if cfg == nil {
		return false
	}
	switch v := cfg[key].(type) {
	case bool:
		return v
	case string:
		return strings.EqualFold(v, "true") || v == "1"
	case float64:
		return v != 0
	}
	return false
}

// ==================== 工具函数：HTTP ====================

// senderTransport 共享 HTTP Transport（连接池），供各发送器复用。
// 复用 TCP 连接减少握手开销；整体超时由每次请求的 ctx 或 client.Timeout 控制。
var senderTransport = &http.Transport{
	MaxIdleConns:        100,
	MaxIdleConnsPerHost: 16,
	IdleConnTimeout:     90 * time.Second,
	TLSHandshakeTimeout: 10 * time.Second,
	DialContext:         (&net.Dialer{Timeout: 10 * time.Second}).DialContext,
}

// httpPost HTTP POST 辅助（JSON 请求体）。
func httpPost(ctx context.Context, url string, body []byte, timeout time.Duration) ([]byte, int, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, 0, err
	}
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{Timeout: timeout, Transport: senderTransport}
	resp, err := client.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()
	buf := new(bytes.Buffer)
	_, _ = buf.ReadFrom(resp.Body)
	return buf.Bytes(), resp.StatusCode, nil
}

// copyBuffer 从 reader 拷贝到 writer。
func copyBuffer(w io.Writer, r io.Reader) (int64, error) {
	return io.Copy(w, r)
}

// ==================== 工具函数：JSON ====================

// jsonDump 序列化调试数据。
func jsonDump(v any) string {
	b, err := json.Marshal(v)
	if err != nil {
		return "{}"
	}
	return string(b)
}

// rawToAny 将 json.RawMessage 转为 any（宽松解析字符串/数字）。
func rawToAny(raw json.RawMessage) any {
	if len(raw) == 0 {
		return nil
	}
	var v any
	if err := json.Unmarshal(raw, &v); err != nil {
		return string(raw)
	}
	return v
}

// ==================== 工具函数：加密/随机 ====================

// randomHex 生成 n 字节随机数的 hex 字符串。
func randomHex(n int) string {
	b := make([]byte, n)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

// sha1Sum 计算 SHA1 摘要。
func sha1Sum(b []byte) []byte {
	h := sha1.New()
	_, _ = h.Write(b)
	return h.Sum(nil)
}

// ==================== 工具函数：字符串 ====================

// cnMobileRe 中国大陆手机号正则。
var cnMobileRe = regexp.MustCompile(`^1[3-9]\d{9}$`)

// isAllDigits 是否全数字。
func isAllDigits(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

// truncateUTF8Bytes 按 UTF-8 字节安全截断（不切断多字节字符）。
func truncateUTF8Bytes(s string, maxBytes int) string {
	if maxBytes <= 0 {
		return ""
	}
	if len(s) <= maxBytes {
		return s
	}
	b := []byte(s)
	cut := b[:maxBytes]
	// 回退到最后一个完整 UTF-8 字符边界
	for len(cut) > 0 && !utf8.Valid(cut) {
		cut = cut[:len(cut)-1]
	}
	return string(cut)
}

// ==================== 工具函数：模板 ====================

// templateVarRe 模板占位符正则。
var templateVarRe = regexp.MustCompile(`\{([a-zA-Z0-9_]+)\}`)

// templateKeys 提取模板占位符顺序。
func templateKeys(content string) []string {
	matches := templateVarRe.FindAllStringSubmatch(content, -1)
	keys := make([]string, 0, len(matches))
	for _, m := range matches {
		if len(m) > 1 {
			keys = append(keys, m[1])
		}
	}
	return keys
}

// sortedParamsFromMap 无模板占位符时的稳定回退：按 key 字典序取值。
// 避免直接遍历 map 产生随机顺序（Go map 迭代无序，会导致短信模板参数内容错位且不稳定）。
func sortedParamsFromMap(mapped map[string]string) []string {
	if len(mapped) == 0 {
		return nil
	}
	keys := make([]string, 0, len(mapped))
	for k := range mapped {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	vals := make([]string, 0, len(keys))
	for _, k := range keys {
		vals = append(vals, mapped[k])
	}
	return vals
}
