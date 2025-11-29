package smtpclient

import (
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"net"
	"net/smtp"
	"strings"
	"time"
)

// Message 邮件消息
type Message struct {
	From    string   // 发件人
	To      []string // 收件人列表
	Subject string   // 主题
	Body    string   // 正文（纯文本）
	HTML    string   // HTML正文（可选）
}

// Client SMTP客户端
type Client struct {
	domain string // 本地域名
}

// NewClient 创建SMTP客户端
func NewClient(domain string) *Client {
	return &Client{domain: domain}
}

// Send 发送邮件（通过查找MX记录直接发送）
func (c *Client) Send(msg Message) error {
	// 验证发件人地址
	if msg.From == "" {
		return fmt.Errorf("发件人地址不能为空")
	}

	// 构建邮件内容
	body := c.buildMessage(msg)

	// 按收件人域名分组发送
	recipientsByDomain := make(map[string][]string)
	for _, to := range msg.To {
		domain := extractDomain(to)
		if domain == "" {
			return fmt.Errorf("无效的收件人地址: %s", to)
		}
		recipientsByDomain[domain] = append(recipientsByDomain[domain], to)
	}

	// 向每个域名发送邮件
	var lastErr error
	for domain, recipients := range recipientsByDomain {
		if err := c.sendToDomain(domain, msg.From, recipients, body); err != nil {
			// 简化错误提示
			errMsg := strings.ToLower(err.Error())

			// 检测常见的拒绝原因
			if strings.Contains(errMsg, "not authorized") ||
				strings.Contains(errMsg, "blocked") ||
				strings.Contains(errMsg, "spamhaus") ||
				strings.Contains(errMsg, "relay") {
				// 国际邮箱拒绝
				if domain == "gmail.com" || domain == "googlemail.com" {
					lastErr = fmt.Errorf("Gmail 拒绝，建议用 QQ/163")
				} else if domain == "outlook.com" || domain == "hotmail.com" || domain == "live.com" {
					lastErr = fmt.Errorf("Outlook 拒绝，建议用 QQ/163")
				} else {
					lastErr = fmt.Errorf("对方拒绝接收（IP 被限制）")
				}
			} else if strings.Contains(errMsg, "连接失败") || strings.Contains(errMsg, "timeout") {
				lastErr = fmt.Errorf("连接超时，请稍后重试")
			} else if strings.Contains(errMsg, "无效") {
				lastErr = fmt.Errorf("收件人地址无效")
			} else {
				// 其他错误，提供通用提示
				lastErr = fmt.Errorf("发送失败，建议用 QQ/163")
			}
			// 继续尝试其他域名
		}
	}

	return lastErr
}

// sendToDomain 向指定域名发送邮件
func (c *Client) sendToDomain(domain string, from string, to []string, body string) error {
	// 查找MX记录
	mxRecords, err := net.LookupMX(domain)
	if err != nil || len(mxRecords) == 0 {
		// 如果没有MX记录，尝试使用A记录
		return c.sendToHost(domain+":25", from, to, body)
	}

	// 尝试每个MX记录（按优先级排序）
	var lastErr error
	for _, mx := range mxRecords {
		host := strings.TrimSuffix(mx.Host, ".")
		if err := c.sendToHost(host+":25", from, to, body); err == nil {
			return nil
		} else {
			lastErr = err // 只保留最后一个错误
		}
	}

	return lastErr
}

// sendToHost 向指定主机发送邮件
func (c *Client) sendToHost(addr string, from string, to []string, body string) error {
	// 连接到SMTP服务器
	client, err := smtp.Dial(addr)
	if err != nil {
		return fmt.Errorf("连接失败 %s: %w", addr, err)
	}
	defer client.Close()

	// 发送HELO命令
	if err = client.Hello(c.domain); err != nil {
		return fmt.Errorf("HELO失败 (domain=%s): %w", c.domain, err)
	}

	// 尝试升级到 TLS（如果服务器支持）
	if ok, _ := client.Extension("STARTTLS"); ok {
		tlsConfig := &tls.Config{
			ServerName:         strings.Split(addr, ":")[0],
			InsecureSkipVerify: false, // 验证证书
		}
		if err = client.StartTLS(tlsConfig); err != nil {
			// TLS 失败不致命，继续尝试明文发送
			// 但记录日志便于调试
			// log.Printf("STARTTLS 失败，继续使用明文: %v", err)
		}
	}

	// 设置发件人
	if err = client.Mail(from); err != nil {
		return fmt.Errorf("MAIL FROM失败 (from=%s): %w", from, err)
	}

	// 设置收件人
	for _, recipient := range to {
		if err = client.Rcpt(recipient); err != nil {
			return fmt.Errorf("RCPT TO失败 (to=%s): %w", recipient, err)
		}
	}

	// 发送邮件内容
	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("DATA命令失败: %w", err)
	}

	_, err = w.Write([]byte(body))
	if err != nil {
		w.Close()
		return fmt.Errorf("写入邮件内容失败: %w", err)
	}

	if err = w.Close(); err != nil {
		return fmt.Errorf("关闭DATA失败: %w", err)
	}

	// 退出
	return client.Quit()
}

// extractDomain 从邮件地址中提取域名
func extractDomain(email string) string {
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return ""
	} 
	return parts[1]
}

// buildMessage 构建邮件内容
func (c *Client) buildMessage(msg Message) string {
	var sb strings.Builder

	// 邮件头
	sb.WriteString(fmt.Sprintf("From: %s\r\n", msg.From))
	sb.WriteString(fmt.Sprintf("To: %s\r\n", strings.Join(msg.To, ", ")))
	sb.WriteString(fmt.Sprintf("Subject: %s\r\n", encodeSubject(msg.Subject)))
	sb.WriteString(fmt.Sprintf("Date: %s\r\n", time.Now().Format(time.RFC1123Z)))
	sb.WriteString("MIME-Version: 1.0\r\n")

	// 如果有HTML内容，使用multipart
	if msg.HTML != "" {
		boundary := fmt.Sprintf("boundary_%d", time.Now().Unix())
		sb.WriteString(fmt.Sprintf("Content-Type: multipart/alternative; boundary=\"%s\"\r\n", boundary))
		sb.WriteString("\r\n")

		// 纯文本部分
		if msg.Body != "" {
			sb.WriteString(fmt.Sprintf("--%s\r\n", boundary))
			sb.WriteString("Content-Type: text/plain; charset=UTF-8\r\n")
			sb.WriteString("Content-Transfer-Encoding: quoted-printable\r\n")
			sb.WriteString("\r\n")
			sb.WriteString(msg.Body)
			sb.WriteString("\r\n")
		}

		// HTML部分
		sb.WriteString(fmt.Sprintf("--%s\r\n", boundary))
		sb.WriteString("Content-Type: text/html; charset=UTF-8\r\n")
		sb.WriteString("Content-Transfer-Encoding: quoted-printable\r\n")
		sb.WriteString("\r\n")
		sb.WriteString(msg.HTML)
		sb.WriteString("\r\n")

		sb.WriteString(fmt.Sprintf("--%s--\r\n", boundary))
	} else {
		// 只有纯文本
		sb.WriteString("Content-Type: text/plain; charset=UTF-8\r\n")
		sb.WriteString("Content-Transfer-Encoding: quoted-printable\r\n")
		sb.WriteString("\r\n")
		sb.WriteString(msg.Body)
	}

	return sb.String()
}

// encodeSubject 编码邮件主题（支持中文）
func encodeSubject(subject string) string {
	// 检查是否包含非ASCII字符
	needEncode := false
	for _, r := range subject {
		if r > 127 {
			needEncode = true
			break
		}
	}

	if !needEncode {
		return subject
	}

	// 使用Base64编码
	encoded := base64.StdEncoding.EncodeToString([]byte(subject))
	return fmt.Sprintf("=?UTF-8?B?%s?=", encoded)
}
