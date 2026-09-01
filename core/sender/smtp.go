package sender

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/smtp"
	"strings"
	"time"

	"chihqiang/msg-push/model"
)

// SMTPSender SMTP 邮件发送器。
type SMTPSender struct{}

// smtpConfig SMTP 配置。
type smtpConfig struct {
	Host       string `json:"host"`
	Port       int    `json:"port"`
	Username   string `json:"username"`
	Password   string `json:"password"`
	From       string `json:"from"`
	FromName   string `json:"from_name"`  // 发件人显示名称（可选）
	Encryption string `json:"encryption"` // none/starttls/ssl
}

// GetProviderCode 返回服务商代码。
func (s *SMTPSender) GetProviderCode() string {
	return CodeSMTP
}

// Send 发送邮件。
func (s *SMTPSender) Send(ctx context.Context, req *SendRequest) (*SendResponse, error) {
	taskID := ""
	if req.Task != nil {
		taskID = req.Task.TaskID
	}
	if req.ProviderAccount == nil {
		return &SendResponse{Success: false, TaskID: taskID, ErrorMessage: "provider account missing", ErrorCode: "CONFIG_ERROR"}, nil
	}

	var cfg smtpConfig
	if err := req.ProviderAccount.GetConfigInto(&cfg); err != nil {
		return &SendResponse{Success: false, TaskID: taskID, ErrorMessage: "invalid smtp config: " + err.Error(), ErrorCode: "CONFIG_ERROR"}, nil
	}
	if cfg.Host == "" {
		return &SendResponse{Success: false, TaskID: taskID, ErrorMessage: "smtp host required", ErrorCode: "CONFIG_ERROR"}, nil
	}
	if cfg.Port == 0 {
		cfg.Port = 587
	}

	subject := "通知"
	if req.Task != nil {
		subject = normalizeEmailSubject(req.Task.Signature)
	}
	contentType := "text/plain; charset=UTF-8"
	if req.ChannelTemplateBinding != nil && req.ChannelTemplateBinding.ProviderTemplate != nil {
		if req.ChannelTemplateBinding.ProviderTemplate.ContentType == "html" {
			contentType = "text/html; charset=UTF-8"
		}
	}

	receiver := ""
	if req.Task != nil {
		receiver = req.Task.Receiver
	}

	// From 头部支持显示名称（"名称 <邮箱>"），MAIL FROM 命令仍用纯邮箱地址
	fromHeader := cfg.From
	if cfg.FromName != "" {
		fromHeader = fmt.Sprintf("%s <%s>", cfg.FromName, cfg.From)
	}

	msg := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\nContent-Type: %s\r\n\r\n%s",
		fromHeader, receiver, subject, contentType, req.RenderedContent)

	addr := net.JoinHostPort(cfg.Host, fmt.Sprintf("%d", cfg.Port))
	if err := sendMail(ctx, addr, cfg, msg); err != nil {
		return &SendResponse{
			Success: false, TaskID: taskID, ErrorCode: "SMTP_ERROR",
			ErrorMessage: err.Error(),
			RequestData:  jsonDump(map[string]any{"to": receiver, "subject": subject}),
		}, nil
	}

	return &SendResponse{
		Success:      true,
		TaskID:       taskID,
		ProviderID:   "smtp_" + taskID,
		Status:       string(model.PushTaskStatusSuccess),
		RequestData:  jsonDump(map[string]any{"to": receiver, "subject": subject}),
		ResponseData: "{}",
	}, nil
}

// normalizeEmailSubject 规范化邮件主题（拒绝换行）。
func normalizeEmailSubject(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return "通知"
	}
	if strings.ContainsAny(s, "\r\n") {
		return "通知"
	}
	return s
}

// sendMail 按加密方式发送邮件。
func sendMail(ctx context.Context, addr string, cfg smtpConfig, msg string) error {
	encryption := strings.ToLower(cfg.Encryption)
	if encryption == "" {
		encryption = "starttls"
	}

	tlsConfig := &tls.Config{ServerName: cfg.Host, InsecureSkipVerify: false}
	var conn net.Conn
	var err error

	switch encryption {
	case "ssl":
		dialer := &net.Dialer{Timeout: 30 * time.Second}
		conn, err = tls.DialWithDialer(dialer, "tcp", addr, tlsConfig)
	case "none":
		dialer := &net.Dialer{Timeout: 30 * time.Second}
		conn, err = dialer.DialContext(ctx, "tcp", addr)
	default: // starttls
		dialer := &net.Dialer{Timeout: 30 * time.Second}
		conn, err = dialer.DialContext(ctx, "tcp", addr)
	}
	if err != nil {
		return err
	}
	defer conn.Close()

	client, err := smtp.NewClient(conn, cfg.Host)
	if err != nil {
		return err
	}
	defer client.Close()

	if encryption == "starttls" {
		if ok, _ := client.Extension("STARTTLS"); ok {
			if err := client.StartTLS(tlsConfig); err != nil {
				return err
			}
		}
	}
	if cfg.Username != "" {
		auth := smtp.PlainAuth("", cfg.Username, cfg.Password, cfg.Host)
		if err := client.Auth(auth); err != nil {
			return err
		}
	}
	if err := client.Mail(cfg.From); err != nil {
		return err
	}
	rcpt := extractTo(msg)
	if err := client.Rcpt(rcpt); err != nil {
		return err
	}
	w, err := client.Data()
	if err != nil {
		return err
	}
	if _, err := w.Write([]byte(msg)); err != nil {
		return err
	}
	if err := w.Close(); err != nil {
		return err
	}
	return client.Quit()
}

// extractTo 从消息中提取收件人。
func extractTo(msg string) string {
	for _, line := range strings.Split(msg, "\r\n") {
		if strings.HasPrefix(strings.ToLower(line), "to:") {
			return strings.TrimSpace(line[3:])
		}
	}
	return ""
}
