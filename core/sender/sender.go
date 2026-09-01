package sender

// ==================== 服务商元信息 ====================

// FieldOption 下拉选项。
type FieldOption struct {
	Value string `json:"value"` // 选项值
	Label string `json:"label"` // 选项显示名称
}

// ConfigField 配置字段定义。
type ConfigField struct {
	Key            string        `json:"key"`             // 参数key
	Label          string        `json:"label"`           // 显示名称
	Description    string        `json:"description"`     // 说明
	Type           string        `json:"type"`            // 类型 text/password/number/url/textarea/select
	Required       bool          `json:"required"`        // 是否必填
	Example        string        `json:"example"`         // 示例值
	Placeholder    string        `json:"placeholder"`     // 占位符
	ValidationRule string        `json:"validation_rule"` // 验证规则
	HelpLink       string        `json:"help_link"`       // 帮助文档链接
	DefaultValue   string        `json:"default_value"`   // 默认值
	Options        []FieldOption `json:"options"`         // 下拉选项(select)
}

// Meta 服务商元信息。
type Meta struct {
	Code         string        `json:"code"`          // 服务商代码（唯一标识）
	Name         string        `json:"name"`          // 服务商名称
	Type         string        `json:"type"`          // 消息类型 sms/email/wechat_work/dingtalk
	Description  string        `json:"description"`   // 服务商描述
	ConfigFields []ConfigField `json:"config_fields"` // 配置参数定义

	// 能力声明
	SupportsSend      bool `json:"supports_send"`       // 是否支持单条发送
	SupportsBatchSend bool `json:"supports_batch_send"` // 是否支持批量发送
	SupportsCallback  bool `json:"supports_callback"`   // 是否支持回调
	RequiresSignature bool `json:"requires_signature"`  // 是否必须配置签名

	// 扩展信息
	Website    string   `json:"website"`     // 官网地址
	Icon       string   `json:"icon"`        // 图标URL
	DocsUrl    string   `json:"docs_url"`    // API文档链接
	ConsoleUrl string   `json:"console_url"` // 管理控制台链接
	SortOrder  int      `json:"sort_order"`  // 排序权重
	Tags       []string `json:"tags"`        // 标签
	Regions    []string `json:"regions"`     // 支持区域
	Deprecated bool     `json:"deprecated"`  // 是否已弃用
}

// providerMetas 内置服务商元信息清单。
var providerMetas = []*Meta{
	{
		Code: CodeAliyunSMS, Name: "阿里云短信", Type: TypeSMS, SortOrder: 1,
		Description:       "阿里云短信服务，支持国内/国际短信",
		SupportsSend:      true,
		SupportsBatchSend: true,
		SupportsCallback:  true,
		RequiresSignature: true,
		ConsoleUrl:        "https://dysms.console.aliyun.com/",
		Tags:              []string{"国内", "国际"},
		Regions:           []string{"cn", "global"},
		ConfigFields: []ConfigField{
			{Key: "access_key_id", Label: "AccessKey ID", Type: "text", Required: true, Placeholder: "阿里云 AccessKey ID"},
			{Key: "access_key_secret", Label: "AccessKey Secret", Type: "password", Required: true, Placeholder: "阿里云 AccessKey Secret"},
		},
	},
	{
		Code: CodeTencentSMS, Name: "腾讯云短信", Type: TypeSMS, SortOrder: 2,
		Description:       "腾讯云短信服务",
		SupportsSend:      true,
		SupportsBatchSend: true,
		SupportsCallback:  true,
		RequiresSignature: true,
		ConsoleUrl:        "https://console.cloud.tencent.com/smsv2",
		Tags:              []string{"国内", "国际"},
		Regions:           []string{"cn"},
		ConfigFields: []ConfigField{
			{Key: "secret_id", Label: "SecretId", Type: "text", Required: true, Placeholder: "腾讯云 SecretId"},
			{Key: "secret_key", Label: "SecretKey", Type: "password", Required: true, Placeholder: "腾讯云 SecretKey"},
			{Key: "sdk_app_id", Label: "SdkAppId", Type: "text", Required: true, Placeholder: "短信应用 SdkAppId"},
		},
	},
	{
		Code: CodeNeteaseSMS, Name: "网易云信短信", Type: TypeSMS, SortOrder: 3,
		Description:       "网易云信短信服务",
		SupportsSend:      true,
		SupportsBatchSend: true,
		SupportsCallback:  true,
		RequiresSignature: true,
		ConsoleUrl:        "https://yunxin.163.com/",
		Tags:              []string{"国内"},
		Regions:           []string{"cn"},
		ConfigFields: []ConfigField{
			{Key: "app_key", Label: "AppKey", Type: "text", Required: true, Placeholder: "网易云信 AppKey"},
			{Key: "app_secret", Label: "AppSecret", Type: "password", Required: true, Placeholder: "网易云信 AppSecret"},
			{Key: "send_type", Label: "发送模式", Type: "select", Required: false, DefaultValue: "template", Options: []FieldOption{
				{Value: "template", Label: "模板短信"},
				{Value: "code", Label: "验证码短信"},
			}},
		},
	},
	{
		Code: CodeSMTP, Name: "SMTP邮件", Type: TypeEmail, SortOrder: 4,
		Description:       "SMTP 邮件发送",
		SupportsSend:      true,
		SupportsBatchSend: false,
		SupportsCallback:  false,
		RequiresSignature: false,
		ConfigFields: []ConfigField{
			{Key: "host", Label: "SMTP主机", Type: "text", Required: true, Placeholder: "如 smtp.example.com"},
			{Key: "port", Label: "端口", Type: "number", Required: true, DefaultValue: "587", Placeholder: "465/587/25"},
			{Key: "username", Label: "用户名", Type: "text", Required: true, Placeholder: "SMTP 账号"},
			{Key: "password", Label: "密码", Type: "password", Required: true, Placeholder: "SMTP 密码/授权码"},
			{Key: "from", Label: "发件人", Type: "text", Required: true, Placeholder: "发件人邮箱"},
			{Key: "from_name", Label: "发件人名称", Type: "text", Required: false, Placeholder: "如 验证码中心"},
			{Key: "encryption", Label: "加密方式", Type: "select", Required: false, DefaultValue: "starttls", Options: []FieldOption{
				{Value: "ssl", Label: "SSL/TLS"},
				{Value: "starttls", Label: "STARTTLS"},
				{Value: "none", Label: "无"},
			}},
		},
	},
	{
		Code: CodeWeChatWork, Name: "企业微信", Type: TypeWeChatWork, SortOrder: 5,
		Description:       "企业微信应用消息",
		SupportsSend:      true,
		SupportsBatchSend: false,
		SupportsCallback:  false,
		RequiresSignature: false,
		ConsoleUrl:        "https://work.weixin.qq.com/wework_admin/frame",
		Tags:              []string{"企业微信"},
		Regions:           []string{"cn"},
		ConfigFields: []ConfigField{
			{Key: "corp_id", Label: "企业ID", Type: "text", Required: true, Placeholder: "企业微信 CorpId"},
			{Key: "agent_id", Label: "应用AgentId", Type: "number", Required: true, Placeholder: "应用 AgentId"},
			{Key: "secret", Label: "应用Secret", Type: "password", Required: true, Placeholder: "应用 Secret"},
		},
	},
	{
		Code: CodeDingTalk, Name: "钉钉", Type: TypeDingTalk, SortOrder: 6,
		Description:       "钉钉工作通知",
		SupportsSend:      true,
		SupportsBatchSend: false,
		SupportsCallback:  false,
		RequiresSignature: false,
		ConsoleUrl:        "https://open.dingtalk.com/",
		Tags:              []string{"钉钉"},
		Regions:           []string{"cn"},
		ConfigFields: []ConfigField{
			{Key: "app_key", Label: "AppKey", Type: "text", Required: true, Placeholder: "钉钉开放平台 AppKey"},
			{Key: "app_secret", Label: "AppSecret", Type: "password", Required: true, Placeholder: "钉钉开放平台 AppSecret"},
			{Key: "agent_id", Label: "AgentId", Type: "number", Required: true, Placeholder: "应用 AgentId"},
		},
	},
	{
		Code: CodeWeChatWorkRobot, Name: "企业微信群机器人", Type: TypeWeChatWork, SortOrder: 7,
		Description:       "企业微信群机器人(webhook)",
		SupportsSend:      true,
		SupportsBatchSend: false,
		SupportsCallback:  false,
		RequiresSignature: false,
		ConsoleUrl:        "https://work.weixin.qq.com/api/doc/90000/90136/91770",
		Tags:              []string{"企业微信", "机器人"},
		Regions:           []string{"cn"},
		ConfigFields: []ConfigField{
			{Key: "webhook_url", Label: "Webhook地址", Type: "url", Required: true, Placeholder: "群机器人 Webhook URL"},
			{Key: "msg_type", Label: "消息类型", Type: "select", Required: false, DefaultValue: "text", Options: []FieldOption{
				{Value: "text", Label: "文本"},
				{Value: "markdown", Label: "Markdown"},
			}},
		},
	},
	{
		Code: CodeDingTalkRobot, Name: "钉钉群机器人", Type: TypeDingTalk, SortOrder: 8,
		Description:       "钉钉群机器人(webhook)",
		SupportsSend:      true,
		SupportsBatchSend: false,
		SupportsCallback:  false,
		RequiresSignature: false,
		ConsoleUrl:        "https://open.dingtalk.com/document/robots/custom-robot-access",
		Tags:              []string{"钉钉", "机器人"},
		Regions:           []string{"cn"},
		ConfigFields: []ConfigField{
			{Key: "webhook_url", Label: "Webhook地址", Type: "url", Required: true, Placeholder: "群机器人 Webhook URL"},
			{Key: "secret", Label: "加签密钥", Type: "password", Required: false, Placeholder: "加签方式密钥(可选)"},
			{Key: "msg_type", Label: "消息类型", Type: "select", Required: false, DefaultValue: "text", Options: []FieldOption{
				{Value: "text", Label: "文本"},
				{Value: "markdown", Label: "Markdown"},
			}},
		},
	},
}

// GetAll 返回所有服务商元信息。
func GetAll() []*Meta {
	out := make([]*Meta, 0, len(providerMetas))
	for _, m := range providerMetas {
		cp := *m
		out = append(out, &cp)
	}
	return out
}

// GetByType 按消息类型返回服务商元信息。
func GetByType(msgType string) []*Meta {
	var out []*Meta
	for _, m := range providerMetas {
		if m.Type == msgType {
			cp := *m
			out = append(out, &cp)
		}
	}
	return out
}

// GetByCode 按代码返回服务商元信息。
func GetByCode(code string) (*Meta, bool) {
	for _, m := range providerMetas {
		if m.Code == code {
			cp := *m
			return &cp, true
		}
	}
	return nil, false
}
