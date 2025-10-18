package httpapi

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime"
	"net/http"
	stdmail "net/mail"
	"strings"

	"temp_mail/internal/smtpclient"
	"temp_mail/internal/storage"
)

func NewMux(store storage.Store, domain string, smtpClient *smtpclient.Client) http.Handler {
	mux := http.NewServeMux()
	// API
	mux.HandleFunc("/api/address", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			local := r.URL.Query().Get("local")
			local = sanitizeLocal(local)
			created := store.CreateAddress(local)
			writeJSON(w, map[string]interface{}{
				"address": fmt.Sprintf("%s@%s", created, domain),
				"local":   created,
				"ttl":     int(store.TTL().Minutes()),
			})
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/api/messages/", func(w http.ResponseWriter, r *http.Request) {
		// /api/messages/{local} or /api/messages/{local}/{id}
		path := strings.TrimPrefix(r.URL.Path, "/api/messages/")
		parts := strings.Split(path, "/")
		if parts[0] == "" {
			http.NotFound(w, r)
			return
		}
		local := sanitizeLocal(parts[0])
		if len(parts) == 1 {
			msgs := store.List(local)
			writeJSON(w, msgs)
			return
		}
		id := parts[1]
		msg, ok := store.Get(local, id)
		if !ok {
			http.NotFound(w, r)
			return
		}
		switch r.URL.Query().Get("format") {
		case "raw":
			w.Header().Set("Content-Type", "message/rfc822")
			_, _ = w.Write(msg.Raw)
		default:
			writeJSON(w, msg)
		}
	})

	// 发送邮件API
	mux.HandleFunc("/api/send", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		// 解析请求体
		var req struct {
			From    string   `json:"from"`    // 发件人本地部分（如 "test"），将拼接域名
			To      []string `json:"to"`      // 收件人列表（完整邮箱地址）
			Subject string   `json:"subject"` // 邮件主题
			Body    string   `json:"body"`    // 邮件正文
			HTML    string   `json:"html"`    // HTML正文（可选）
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			writeJSON(w, map[string]interface{}{
				"error": "无效的请求格式",
			})
			return
		}

		// 验证必填字段
		if req.From == "" {
			w.WriteHeader(http.StatusBadRequest)
			writeJSON(w, map[string]interface{}{
				"error": "发件人不能为空（请输入您创建的邮箱名称）",
			})
			return
		}

		if len(req.To) == 0 {
			w.WriteHeader(http.StatusBadRequest)
			writeJSON(w, map[string]interface{}{
				"error": "收件人不能为空",
			})
			return
		}

		if req.Subject == "" {
			w.WriteHeader(http.StatusBadRequest)
			writeJSON(w, map[string]interface{}{
				"error": "主题不能为空",
			})
			return
		}

		if req.Body == "" && req.HTML == "" {
			w.WriteHeader(http.StatusBadRequest)
			writeJSON(w, map[string]interface{}{
				"error": "邮件内容不能为空",
			})
			return
		}

		// 验证发件人邮箱是否存在
		fromLocal := sanitizeLocal(req.From)
		if fromLocal == "" {
			w.WriteHeader(http.StatusBadRequest)
			writeJSON(w, map[string]interface{}{
				"error": "发件人不能为空",
			})
			return
		}

		// 如果地址不存在，自动创建
		if !store.AddressExists(fromLocal) {
			store.CreateAddress(fromLocal)
			log.Printf("自动创建发件邮箱: %s@%s", fromLocal, domain)
		}

		// 构造完整的发件人地址
		fromAddr := fmt.Sprintf("%s@%s", fromLocal, domain)

		// 发送邮件
		msg := smtpclient.Message{
			From:    fromAddr,
			To:      req.To,
			Subject: req.Subject,
			Body:    req.Body,
			HTML:    req.HTML,
		}

		if err := smtpClient.Send(msg); err != nil {
			log.Printf("发送邮件失败 (from=%s): %v", fromAddr, err)
			w.WriteHeader(http.StatusInternalServerError)
			writeJSON(w, map[string]interface{}{
				"error": fmt.Sprintf("发送失败: %v", err),
			})
			return
		}

		log.Printf("邮件已发送: from=%s, to=%v, subject=%s", fromAddr, req.To, req.Subject)
		writeJSON(w, map[string]interface{}{
			"success": true,
			"message": "邮件已发送",
			"from":    fromAddr,
		})
	})

	// UI
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(indexHTML))
	})

	// Message detail page
	mux.HandleFunc("/view/", func(w http.ResponseWriter, r *http.Request) {
		// /view/{local}/{id}
		path := strings.TrimPrefix(r.URL.Path, "/view/")
		parts := strings.Split(path, "/")
		if len(parts) < 2 {
			http.NotFound(w, r)
			return
		}
		local := sanitizeLocal(parts[0])
		id := parts[1]

		msg, ok := store.Get(local, id)
		if !ok {
			http.NotFound(w, r)
			return
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(renderMessageDetailPage(msg, local, domain)))
	})

	return mux
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("writeJSON: %v", err)
	}
}

func sanitizeLocal(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	s = strings.Trim(s, "@ ")
	if i := strings.IndexByte(s, '@'); i > 0 {
		s = s[:i]
	}
	if i := strings.IndexByte(s, '+'); i > 0 {
		s = s[:i]
	}
	if s == "" {
		return s
	}
	// allow [a-z0-9-.]
	b := strings.Builder{}
	for _, ch := range s {
		if (ch >= 'a' && ch <= 'z') || (ch >= '0' && ch <= '9') || ch == '-' || ch == '.' || ch == '_' {
			b.WriteRune(ch)
		}
	}
	return b.String()
}

func renderMessageDetailPage(msg storage.Message, local, domain string) string {
	// Parse email content from raw bytes
	htmlContent, textContent, _ := parseEmailContent(msg.Raw)

	// Choose which content to display
	bodyHTML := ""
	if htmlContent != "" {
		bodyHTML = fmt.Sprintf(`<iframe id="html-frame" srcdoc="%s" sandbox="allow-same-origin"></iframe>`, escapeHTMLAttr(htmlContent))
	} else if textContent != "" {
		bodyHTML = fmt.Sprintf(`<pre class="text-content">%s</pre>`, escapeHTML(textContent))
	} else {
		bodyHTML = `<div class="no-content">无邮件正文</div>`
	}

	timeStr := msg.CreatedAt.Format("2006-01-02 15:04:05")

	return fmt.Sprintf(messageDetailTemplate,
		escapeHTML(msg.Subject),
		local,
		escapeHTML(msg.Subject),
		escapeHTML(msg.From),
		timeStr,
		bodyHTML,
		local,
		local,
		msg.ID,
	)
}

func parseEmailContent(raw []byte) (htmlContent, textContent string, headers map[string]string) {
	headers = make(map[string]string)

	if len(raw) == 0 {
		return "", "", headers
	}

	// Parse email using net/mail
	m, err := stdmail.ReadMessage(bytes.NewReader(raw))
	if err != nil {
		textContent = string(raw)
		return "", textContent, headers
	}

	// Extract and decode headers using mime.WordDecoder
	dec := new(mime.WordDecoder)
	for k := range m.Header {
		rawValue := m.Header.Get(k)
		// Decode MIME encoded-words (=?charset?encoding?text?=)
		if decoded, err := dec.DecodeHeader(rawValue); err == nil {
			headers[k] = decoded
		} else {
			headers[k] = rawValue
		}
	}

	// Read body
	body, err := io.ReadAll(m.Body)
	if err != nil {
		return "", "", headers
	}

	// Check Content-Transfer-Encoding and decode body
	contentTransferEncoding := strings.ToLower(m.Header.Get("Content-Transfer-Encoding"))
	decodedBody := body

	switch contentTransferEncoding {
	case "base64":
		if decoded, err := base64.StdEncoding.DecodeString(string(body)); err == nil {
			decodedBody = decoded
		}
	case "quoted-printable":
		// Quoted-printable decoding
		decodedBody = decodeQuotedPrintable(body)
	}

	bodyStr := string(decodedBody)

	// Check if it's multipart
	contentType := m.Header.Get("Content-Type")
	if strings.Contains(contentType, "multipart") {
		// Simple multipart parsing with encoding support
		htmlContent, textContent = parseMultipart(bodyStr, contentType)
	} else if strings.Contains(contentType, "text/html") {
		htmlContent = bodyStr
	} else {
		textContent = bodyStr
	}

	return htmlContent, textContent, headers
}

func parseMultipart(body, contentType string) (html, text string) {
	// Extract boundary
	boundary := ""
	if idx := strings.Index(contentType, "boundary="); idx >= 0 {
		boundary = contentType[idx+9:]
		boundary = strings.Trim(boundary, `"`)
		if idx2 := strings.IndexAny(boundary, "; \t\r\n"); idx2 >= 0 {
			boundary = boundary[:idx2]
		}
	}

	if boundary == "" {
		return "", body
	}

	parts := strings.Split(body, "--"+boundary)
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" || part == "--" {
			continue
		}

		// Split headers and body
		splitIdx := strings.Index(part, "\r\n\r\n")
		if splitIdx < 0 {
			splitIdx = strings.Index(part, "\n\n")
			if splitIdx < 0 {
				continue
			}
			splitIdx += 2
		} else {
			splitIdx += 4
		}

		headers := part[:splitIdx]
		content := part[splitIdx:]

		// Decode content based on Content-Transfer-Encoding
		if strings.Contains(headers, "base64") {
			content = strings.ReplaceAll(content, "\r\n", "")
			content = strings.ReplaceAll(content, "\n", "")
			if decoded, err := base64.StdEncoding.DecodeString(content); err == nil {
				content = string(decoded)
			}
		} else if strings.Contains(headers, "quoted-printable") {
			content = string(decodeQuotedPrintable([]byte(content)))
		}

		if strings.Contains(headers, "text/html") {
			html = content
		} else if strings.Contains(headers, "text/plain") && text == "" {
			text = content
		}
	}

	return html, text
}

// decodeQuotedPrintable decodes quoted-printable encoded text
func decodeQuotedPrintable(data []byte) []byte {
	var result bytes.Buffer
	i := 0
	for i < len(data) {
		if data[i] == '=' {
			if i+2 < len(data) {
				// Soft line break
				if data[i+1] == '\r' && data[i+2] == '\n' {
					i += 3
					continue
				}
				if data[i+1] == '\n' {
					i += 2
					continue
				}
				// Hex encoding
				if i+2 < len(data) {
					hex := string(data[i+1 : i+3])
					var b byte
					if _, err := fmt.Sscanf(hex, "%02X", &b); err == nil {
						result.WriteByte(b)
						i += 3
						continue
					}
				}
			}
			// Invalid encoding, keep the =
			result.WriteByte(data[i])
			i++
		} else {
			result.WriteByte(data[i])
			i++
		}
	}
	return result.Bytes()
}

func escapeHTML(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, "\"", "&quot;")
	s = strings.ReplaceAll(s, "'", "&#39;")
	return s
}

func escapeHTMLAttr(s string) string {
	s = escapeHTML(s)
	s = strings.ReplaceAll(s, "\n", "&#10;")
	s = strings.ReplaceAll(s, "\r", "&#13;")
	return s
}

const indexHTML = `<!doctype html>
<html lang="zh-CN">
<head>
  <meta charset="utf-8" />
  <meta name="viewport" content="width=device-width, initial-scale=1" />
  <title>临时邮箱 - Temp Mail</title>
  <link rel="preconnect" href="https://fonts.googleapis.com">
  <link rel="preconnect" href="https://fonts.gstatic.com" crossorigin>
  <link href="https://fonts.googleapis.com/css2?family=Inter:wght@400;500;600;700&display=swap" rel="stylesheet">
  <style>
    * {
      margin: 0;
      padding: 0;
      box-sizing: border-box;
    }
    
    body {
      font-family: 'Inter', -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
      background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
      min-height: 100vh;
      padding: 1rem 1rem;
      color: #2d3748;
    }
    
    .container {
      max-width: 700px;
      margin: 0 auto;
      animation: fadeIn 0.6s ease-out;
    }
    
    @keyframes fadeIn {
      from { opacity: 0; transform: translateY(20px); }
      to { opacity: 1; transform: translateY(0); }
    }
    
    .header {
      text-align: center;
      margin-bottom: 1rem;
      color: white;
    }
    
    .header h1 {
      font-size: 1.5rem;
      font-weight: 700;
      margin-bottom: 0.2rem;
      text-shadow: 0 2px 10px rgba(0,0,0,0.1);
    }
    
    .header p {
      font-size: 0.8rem;
      opacity: 0.95;
      font-weight: 400;
    }
    
    .card {
      background: white;
      border-radius: 12px;
      box-shadow: 0 10px 30px rgba(0,0,0,0.3);
      padding: 1rem;
      margin-bottom: 1rem;
      transition: transform 0.3s ease, box-shadow 0.3s ease;
    }
    
    .card:hover {
      transform: translateY(-2px);
      box-shadow: 0 15px 40px rgba(0,0,0,0.35);
    }
    
    .create-section {
      display: flex;
      gap: 0.75rem;
      align-items: stretch;
      margin-bottom: 1rem;
      flex-wrap: wrap;
    }
    
    .input-wrapper {
      flex: 1;
      min-width: 180px;
      position: relative;
    }
    
    .input-wrapper input {
      width: 100%;
      padding: 0.6rem 0.85rem;
      font-size: 0.85rem;
      border: 2px solid #e2e8f0;
      border-radius: 8px;
      transition: all 0.3s ease;
      font-family: inherit;
      background: #f8fafc;
    }
    
    .input-wrapper input:focus {
      outline: none;
      border-color: #667eea;
      background: white;
      box-shadow: 0 0 0 3px rgba(102, 126, 234, 0.1);
    }
    
    .btn {
      padding: 0.6rem 1.2rem;
      font-size: 0.85rem;
      font-weight: 600;
      border: none;
      border-radius: 8px;
      cursor: pointer;
      transition: all 0.3s ease;
      font-family: inherit;
      white-space: nowrap;
    }
    
    .btn-primary {
      background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
      color: white;
      box-shadow: 0 4px 15px rgba(102, 126, 234, 0.4);
    }
    
    .btn-primary:hover {
      transform: translateY(-2px);
      box-shadow: 0 6px 20px rgba(102, 126, 234, 0.5);
    }
    
    .btn-primary:active {
      transform: translateY(0);
    }
    
    .address-display {
      background: linear-gradient(135deg, #f6f8fb 0%, #e9ecef 100%);
      padding: 0.75rem;
      border-radius: 8px;
      border: 2px dashed #cbd5e0;
      display: flex;
      align-items: center;
      justify-content: space-between;
      gap: 0.75rem;
      flex-wrap: wrap;
    }
    
    .address-label {
      font-size: 0.65rem;
      color: #718096;
      font-weight: 600;
      text-transform: uppercase;
      letter-spacing: 0.5px;
      margin-bottom: 0.2rem;
    }
    
    .address-value {
      font-size: 0.9rem;
      font-weight: 600;
      color: #2d3748;
      font-family: 'Courier New', monospace;
      background: white;
      padding: 0.35rem 0.65rem;
      border-radius: 6px;
      border: 1px solid #e2e8f0;
      flex: 1;
      min-width: 200px;
      word-break: break-all;
    }
    
    .address-value.empty {
      color: #a0aec0;
      font-style: italic;
    }
    
    .btn-copy {
      background: #48bb78;
      color: white;
      padding: 0.5rem 1rem;
      border-radius: 6px;
      border: none;
      font-weight: 600;
      cursor: pointer;
      transition: all 0.3s ease;
      font-size: 0.85rem;
    }
    
    .btn-copy:hover {
      background: #38a169;
      transform: scale(1.05);
    }
    
    .messages-header {
      display: flex;
      justify-content: space-between;
      align-items: center;
      margin-bottom: 0.75rem;
      flex-wrap: wrap;
      gap: 0.6rem;
    }
    
    .messages-header h2 {
      font-size: 1rem;
      font-weight: 700;
      color: #2d3748;
    }
    
    .badge {
      background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
      color: white;
      padding: 0.25rem 0.6rem;
      border-radius: 20px;
      font-size: 0.7rem;
      font-weight: 600;
    }
    
    .tabs {
      display: flex;
      gap: 0.5rem;
      margin-bottom: 1.5rem;
      border-bottom: 2px solid #e2e8f0;
    }
    
    .tab {
      padding: 0.75rem 1.5rem;
      background: none;
      border: none;
      border-bottom: 3px solid transparent;
      color: #718096;
      font-weight: 600;
      font-size: 1rem;
      cursor: pointer;
      transition: all 0.3s ease;
      position: relative;
      bottom: -2px;
    }
    
    .tab:hover {
      color: #667eea;
      background: rgba(102, 126, 234, 0.05);
    }
    
    .tab.active {
      color: #667eea;
      border-bottom-color: #667eea;
    }
    
    .tab-content {
      display: none;
    }
    
    .tab-content.active {
      display: block;
      animation: fadeIn 0.3s ease;
    }
    
    @keyframes fadeIn {
      from { opacity: 0; transform: translateY(10px); }
      to { opacity: 1; transform: translateY(0); }
    }
    
    .messages-container {
      background: #f8fafc;
      border-radius: 12px;
      overflow: hidden;
      border: 1px solid #e2e8f0;
    }
    
    .message-item {
      background: white;
      padding: 0.75rem;
      border-bottom: 1px solid #e2e8f0;
      transition: all 0.3s ease;
      cursor: pointer;
    }
    
    .message-item.new-item {
      animation: slideIn 0.4s ease-out;
    }
    
    @keyframes slideIn {
      from { opacity: 0; transform: translateX(-20px); }
      to { opacity: 1; transform: translateX(0); }
    }
    
    .message-item:last-child {
      border-bottom: none;
    }
    
    .message-item:hover {
      background: #f7fafc;
      transform: translateX(4px);
    }
    
    .message-header {
      display: flex;
      justify-content: space-between;
      align-items: flex-start;
      margin-bottom: 0.4rem;
      gap: 0.6rem;
    }
    
    .message-from {
      font-weight: 600;
      color: #2d3748;
      font-size: 0.8rem;
    }
    
    .message-time {
      font-size: 0.7rem;
      color: #718096;
      white-space: nowrap;
    }
    
    .message-subject {
      font-weight: 600;
      color: #4a5568;
      margin-bottom: 0.3rem;
      font-size: 0.85rem;
    }
    
    .message-snippet {
      color: #718096;
      font-size: 0.75rem;
      line-height: 1.4;
      margin-bottom: 0.5rem;
      display: -webkit-box;
      -webkit-line-clamp: 2;
      -webkit-box-orient: vertical;
      overflow: hidden;
    }
    
    .message-actions {
      display: flex;
      gap: 0.5rem;
    }
    
    .btn-small {
      padding: 0.35rem 0.75rem;
      font-size: 0.8rem;
      border-radius: 5px;
      text-decoration: none;
      font-weight: 500;
      transition: all 0.2s ease;
    }
    
    .btn-outline {
      border: 1px solid #667eea;
      color: #667eea;
      background: white;
    }
    
    .btn-outline:hover {
      background: #667eea;
      color: white;
    }
    
    .empty-state {
      text-align: center;
      padding: 2.5rem 1.5rem;
      color: #718096;
    }
    
    .empty-state-icon {
      font-size: 3rem;
      margin-bottom: 0.75rem;
      opacity: 0.5;
    }
    
    .empty-state h3 {
      font-size: 1.1rem;
      margin-bottom: 0.4rem;
      color: #4a5568;
    }
    
    .empty-state p {
      font-size: 0.9rem;
    }
    
    .loading {
      display: inline-block;
      width: 16px;
      height: 16px;
      border: 2px solid #e2e8f0;
      border-top-color: #667eea;
      border-radius: 50%;
      animation: spin 0.8s linear infinite;
    }
    
    @keyframes spin {
      to { transform: rotate(360deg); }
    }
    
    .toast {
      position: fixed;
      bottom: 2rem;
      right: 2rem;
      background: white;
      padding: 1rem 1.5rem;
      border-radius: 12px;
      box-shadow: 0 10px 40px rgba(0,0,0,0.2);
      display: flex;
      align-items: center;
      gap: 0.75rem;
      transform: translateY(100px);
      opacity: 0;
      transition: all 0.3s ease;
      z-index: 1000;
      border-left: 4px solid #48bb78;
    }
    
    .toast.show {
      transform: translateY(0);
      opacity: 1;
    }
    
    @media (max-width: 768px) {
      .header h1 {
        font-size: 2rem;
      }
      
      .card {
        padding: 1.5rem;
      }
      
      .create-section {
        flex-direction: column;
      }
      
      .btn {
        width: 100%;
      }
      
      .address-display {
        flex-direction: column;
        align-items: stretch;
      }
      
      .message-header {
        flex-direction: column;
      }
    }
  </style>
  <script>
    let currentLocal = '';
    let pollInterval = null;
    let messageTTL = 30;
    let lastMessageIds = [];
    
    async function createAddr() {
      const desired = document.getElementById('local').value.trim();
      const btn = event.target;
      btn.disabled = true;
      btn.innerHTML = '<span class="loading"></span>';
      
      try {
        const r = await fetch('/api/address?local=' + encodeURIComponent(desired), {method: 'POST'});
        const j = await r.json();
        currentLocal = j.local;
        messageTTL = j.ttl || 30;
        
        const addrEl = document.getElementById('addr');
        addrEl.textContent = j.address;
        addrEl.classList.remove('empty');
        
        document.getElementById('copy-section').style.display = 'flex';
        document.getElementById('ttl-info').style.display = 'block';
        document.getElementById('ttl-minutes').textContent = messageTTL;
        
        // 设置发件人地址
        document.getElementById('send-from').value = j.address;
        
        loadMsgs();
        startPolling();
        showToast('邮箱地址已创建！邮件保留 ' + messageTTL + ' 分钟');
      } catch (e) {
        showToast('创建失败，请重试', 'error');
      } finally {
        btn.disabled = false;
        btn.textContent = '创建邮箱';
      }
    }
    
    async function loadMsgs() {
      if (!currentLocal) return;
      
      try {
        const r = await fetch('/api/messages/' + currentLocal);
        const msgs = await r.json() || [];
        
        const container = document.getElementById('messages-container');
        const badge = document.getElementById('inbox-badge');
        badge.textContent = msgs.length;
        
        if (msgs.length === 0) {
          container.innerHTML = '<div class="empty-state"><div class="empty-state-icon">📭</div><h3>收件箱为空</h3><p>等待新邮件到达...</p></div>';
          lastMessageIds = [];
          return;
        }
        
        // 检查邮件列表是否有变化（ID 和数量）
        const currentIds = msgs.map(m => m.id);
        const hasChanged = currentIds.length !== lastMessageIds.length || 
                          currentIds.some((id, i) => id !== lastMessageIds[i]);
        
        if (hasChanged) {
          // 只有在邮件列表真正变化时才完全重新渲染
          renderMessages(msgs);
          lastMessageIds = currentIds;
        } else {
          // 如果邮件列表没变，只更新倒计时
          updateTimers(msgs);
        }
      } catch (e) {
        console.error('加载消息失败:', e);
      }
    }
    
    function renderMessages(msgs) {
      const container = document.getElementById('messages-container');
      const existingIds = Array.from(container.querySelectorAll('.message-item')).map(el => el.getAttribute('data-msg-id'));
      container.innerHTML = '';
      
      for (const m of msgs) {
        const div = document.createElement('div');
        const isNew = !existingIds.includes(m.id);
        div.className = isNew ? 'message-item new-item' : 'message-item';
        div.setAttribute('data-msg-id', m.id);
        
        const time = new Date(m.createdAt);
        const expiresAt = new Date(m.expiresAt);
        const now = new Date();
        const minutesLeft = Math.max(0, Math.floor((expiresAt - now) / 60000));
        
        const timeStr = time.toLocaleString('zh-CN', {
          month: 'short',
          day: 'numeric',
          hour: '2-digit',
          minute: '2-digit'
        });
        
        let expiryBadge = '';
        if (minutesLeft > 0) {
          const expiryColor = minutesLeft <= 5 ? '#e53e3e' : (minutesLeft <= 15 ? '#dd6b20' : '#48bb78');
          expiryBadge = '<span class="expiry-timer" style="font-size:0.8rem;color:' + expiryColor + ';margin-left:0.5rem;" data-expires="' + m.expiresAt + '">⏱ ' + minutesLeft + '分钟后过期</span>';
        }
        
        div.innerHTML = 
          '<div class="message-header">' +
            '<div class="message-from">' + escapeHtml(m.from || '未知发件人') + '</div>' +
            '<div class="message-time">' + timeStr + expiryBadge + '</div>' +
          '</div>' +
          '<div class="message-subject">' + escapeHtml(m.subject || '(无主题)') + '</div>' +
          '<div class="message-snippet">' + escapeHtml(m.snippet || '') + '</div>' +
          '<div class="message-actions">' +
            '<a href="/view/' + currentLocal + '/' + m.id + '" class="btn-small btn-outline">查看详情</a>' +
            '<a href="/api/messages/' + currentLocal + '/' + m.id + '?format=raw" download="message.eml" class="btn-small btn-outline" style="margin-left:0.5rem;">下载 EML</a>' +
          '</div>';
        
        container.appendChild(div);
      }
    }
    
    function updateTimers(msgs) {
      // 只更新倒计时，不重新渲染整个列表
      const now = new Date();
      
      msgs.forEach(m => {
        const msgEl = document.querySelector('[data-msg-id="' + m.id + '"]');
        if (!msgEl) return;
        
        const timerEl = msgEl.querySelector('.expiry-timer');
        if (!timerEl) return;
        
        const expiresAt = new Date(m.expiresAt);
        const minutesLeft = Math.max(0, Math.floor((expiresAt - now) / 60000));
        
        if (minutesLeft > 0) {
          const expiryColor = minutesLeft <= 5 ? '#e53e3e' : (minutesLeft <= 15 ? '#dd6b20' : '#48bb78');
          timerEl.style.color = expiryColor;
          timerEl.textContent = '⏱ ' + minutesLeft + '分钟后过期';
        }
      });
    }
    
    function copyAddress() {
      const addr = document.getElementById('addr').textContent;
      if (addr === '(尚未创建)') return;
      
      // 尝试使用现代 Clipboard API
      if (navigator.clipboard && navigator.clipboard.writeText) {
        navigator.clipboard.writeText(addr).then(() => {
          showToast('地址已复制到剪贴板！');
        }).catch((err) => {
          console.error('Clipboard API failed:', err);
          fallbackCopy(addr);
        });
      } else {
        // 降级到传统方法
        fallbackCopy(addr);
      }
    }
    
    function fallbackCopy(text) {
      const textarea = document.createElement('textarea');
      textarea.value = text;
      textarea.style.position = 'fixed';
      textarea.style.opacity = '0';
      document.body.appendChild(textarea);
      textarea.select();
      
      try {
        const successful = document.execCommand('copy');
        if (successful) {
          showToast('地址已复制到剪贴板！');
        } else {
          showToast('复制失败，请手动复制', 'error');
        }
      } catch (err) {
        console.error('Fallback copy failed:', err);
        showToast('复制失败，请手动复制', 'error');
      } finally {
        document.body.removeChild(textarea);
      }
    }
    
    function switchTab(tabName) {
      // 更新选项卡按钮状态
      const tabs = document.querySelectorAll('.tab');
      tabs.forEach(tab => tab.classList.remove('active'));
      event.target.classList.add('active');
      
      // 更新选项卡内容显示
      document.getElementById('inbox-tab').classList.remove('active');
      document.getElementById('compose-tab').classList.remove('active');
      
      if (tabName === 'inbox') {
        document.getElementById('inbox-tab').classList.add('active');
      } else if (tabName === 'compose') {
        document.getElementById('compose-tab').classList.add('active');
        
        // 切换到发送邮件时，检查是否已创建邮箱
        if (!currentLocal) {
          showToast('请先创建邮箱后再发送邮件', 'error');
        }
      }
    }
    
    function clearCompose() {
      // 不清空发件人（自动填充的）
      document.getElementById('send-to').value = '';
      document.getElementById('send-subject').value = '';
      document.getElementById('send-body').value = '';
      document.getElementById('send-status').style.display = 'none';
    }
    
    async function sendEmail() {
      // 检查是否已创建邮箱
      if (!currentLocal) {
        showToast('请先创建邮箱后再发送邮件', 'error');
        return;
      }
      
      const to = document.getElementById('send-to').value.trim();
      const subject = document.getElementById('send-subject').value.trim();
      const body = document.getElementById('send-body').value.trim();
      
      // 验证必填字段
      if (!to) {
        showToast('请输入收件人', 'error');
        return;
      }
      
      if (!subject) {
        showToast('请输入邮件主题', 'error');
        return;
      }
      
      if (!body) {
        showToast('请输入邮件内容', 'error');
        return;
      }
      
      // 解析收件人（支持逗号分隔的多个邮箱）
      const toList = to.split(',').map(e => e.trim()).filter(e => e);
      
      const btn = document.getElementById('send-btn-text');
      const originalText = btn.textContent;
      btn.textContent = '📤 发送中...';
      
      const statusEl = document.getElementById('send-status');
      statusEl.style.display = 'none';
      
      try {
        const response = await fetch('/api/send', {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
          },
          body: JSON.stringify({
            from: currentLocal,  // 使用当前邮箱的本地部分
            to: toList,
            subject: subject,
            body: body,
          }),
        });
        
        const result = await response.json();
        
        if (response.ok && result.success) {
          statusEl.style.display = 'block';
          statusEl.style.background = '#c6f6d5';
          statusEl.style.borderLeft = '4px solid #48bb78';
          statusEl.style.color = '#22543d';
          statusEl.innerHTML = '<strong>✅ 发送成功！</strong><br>' + result.message;
          
          showToast('邮件已发送！');
          
          // 清空表单（不关闭）
          setTimeout(() => {
            clearCompose();
          }, 2000);
        } else {
          throw new Error(result.error || '发送失败');
        }
      } catch (error) {
        statusEl.style.display = 'block';
        statusEl.style.background = '#fed7d7';
        statusEl.style.borderLeft = '4px solid #e53e3e';
        statusEl.style.color = '#742a2a';
        statusEl.innerHTML = '<strong>❌ 发送失败</strong><br>' + error.message;
        
        showToast('发送失败: ' + error.message, 'error');
      } finally {
        btn.textContent = originalText;
      }
    }
    
    function startPolling() {
      if (pollInterval) clearInterval(pollInterval);
      pollInterval = setInterval(loadMsgs, 4000);
    }
    
    function showToast(message, type = 'success') {
      const toast = document.getElementById('toast');
      toast.textContent = message;
      toast.className = 'toast show';
      setTimeout(() => {
        toast.classList.remove('show');
      }, 3000);
    }
    
    function escapeHtml(text) {
      const div = document.createElement('div');
      div.textContent = text;
      return div.innerHTML;
    }
    
    document.addEventListener('DOMContentLoaded', () => {
      document.getElementById('local').addEventListener('keypress', (e) => {
        if (e.key === 'Enter') createAddr();
      });
      
      // 从 URL 参数恢复邮箱地址
      const params = new URLSearchParams(window.location.search);
      const mailbox = params.get('mailbox');
      if (mailbox) {
        currentLocal = mailbox;
        document.getElementById('local').value = mailbox;
        
        // 模拟创建邮箱（不调用API，直接设置界面）
        fetch('/api/address?local=' + encodeURIComponent(mailbox), {method: 'POST'})
          .then(r => r.json())
          .then(j => {
            messageTTL = j.ttl || 30;
            const addrEl = document.getElementById('addr');
            addrEl.textContent = j.address;
            addrEl.classList.remove('empty');
            document.getElementById('copy-section').style.display = 'flex';
            document.getElementById('ttl-info').style.display = 'block';
            document.getElementById('ttl-minutes').textContent = messageTTL;
            document.getElementById('send-from').value = j.address;
            loadMsgs();
            startPolling();
          })
          .catch(e => console.error('恢复邮箱失败:', e));
      }
    });
  </script>
</head>
<body>
  <div class="container">
    <div class="header">
      <h1>📬 临时邮箱</h1>
      <p>快速创建临时邮箱，接收验证邮件</p>
    </div>
    
    <div class="card">
      <div class="create-section">
        <div class="input-wrapper">
          <input id="local" type="text" placeholder="自定义邮箱名称（可选）" autocomplete="off" />
        </div>
        <button class="btn btn-primary" onclick="createAddr()">创建邮箱</button>
      </div>
      
      <div class="address-display">
        <div>
          <div class="address-label">当前邮箱地址</div>
          <div id="addr" class="address-value empty">(尚未创建)</div>
        </div>
        <button id="copy-section" class="btn-copy" onclick="copyAddress()" style="display:none;">复制地址</button>
      </div>
      
      <div id="ttl-info" style="display:none;margin-top:1rem;padding:1rem;background:#fff3cd;border-left:4px solid #ffc107;border-radius:8px;">
        <div style="display:flex;align-items:center;gap:0.5rem;color:#856404;">
          <span style="font-size:1.2rem;">⏰</span>
          <span style="font-weight:600;">重要提示：</span>
          <span>邮件将在收到后 <strong id="ttl-minutes">30</strong> 分钟后自动删除</span>
        </div>
      </div>
    </div>
    
    <div class="card">
      <!-- 选项卡导航 -->
      <div class="tabs">
        <button class="tab active" onclick="switchTab('inbox')">
          📨 收件箱 <span id="inbox-badge" class="badge" style="margin-left:0.5rem;">0</span>
        </button>
        <button class="tab" onclick="switchTab('compose')">
          📤 发送邮件
        </button>
      </div>
      
      <!-- 收件箱选项卡 -->
      <div id="inbox-tab" class="tab-content active">
        <div id="messages-container" class="messages-container">
          <div class="empty-state">
            <div class="empty-state-icon">✨</div>
            <h3>开始使用</h3>
            <p>创建邮箱地址后，邮件将自动出现在这里</p>
          </div>
        </div>
      </div>
      
      <!-- 发送邮件选项卡 -->
      <div id="compose-tab" class="tab-content">
        <div style="display:flex;flex-direction:column;gap:1rem;">
          <div>
            <label style="display:block;margin-bottom:0.5rem;font-weight:600;color:#4a5568;">发件人</label>
            <input id="send-from" type="text" readonly
              style="width:100%;padding:0.75rem;border:2px solid #e2e8f0;border-radius:8px;font-size:1rem;background:#f7fafc;color:#4a5568;cursor:not-allowed;" 
              placeholder="请先创建邮箱" />
            <small style="color:#718096;margin-top:0.25rem;display:block;">使用当前创建的邮箱地址</small>
          </div>
          
          <div>
            <label style="display:block;margin-bottom:0.5rem;font-weight:600;color:#4a5568;">收件人 *</label>
            <input id="send-to" type="email" placeholder="recipient@example.com" required
              style="width:100%;padding:0.75rem;border:2px solid #e2e8f0;border-radius:8px;font-size:1rem;transition:all 0.2s;" 
              onfocus="this.style.borderColor='#667eea'" 
              onblur="this.style.borderColor='#e2e8f0'" />
            <small style="color:#718096;margin-top:0.25rem;display:block;">多个收件人用逗号分隔</small>
          </div>
          
          <div>
            <label style="display:block;margin-bottom:0.5rem;font-weight:600;color:#4a5568;">主题 *</label>
            <input id="send-subject" type="text" placeholder="邮件主题" required
              style="width:100%;padding:0.75rem;border:2px solid #e2e8f0;border-radius:8px;font-size:1rem;transition:all 0.2s;" 
              onfocus="this.style.borderColor='#667eea'" 
              onblur="this.style.borderColor='#e2e8f0'" />
          </div>
          
          <div>
            <label style="display:block;margin-bottom:0.5rem;font-weight:600;color:#4a5568;">内容 *</label>
            <textarea id="send-body" placeholder="邮件正文..." required
              style="width:100%;padding:0.75rem;border:2px solid #e2e8f0;border-radius:8px;font-size:1rem;min-height:150px;resize:vertical;font-family:inherit;transition:all 0.2s;" 
              onfocus="this.style.borderColor='#667eea'" 
              onblur="this.style.borderColor='#e2e8f0'"></textarea>
          </div>
          
          <div style="display:flex;gap:1rem;">
            <button class="btn btn-primary" onclick="sendEmail()" style="flex:1;">
              <span id="send-btn-text">📤 发送</span>
            </button>
            <button class="btn" onclick="clearCompose()" style="flex:1;background:#f7fafc;color:#4a5568;">
              🗑️ 清空
            </button>
          </div>
        </div>
        
        <div id="send-status" style="display:none;margin-top:1rem;padding:1rem;border-radius:8px;"></div>
      </div>
    </div>
  </div>
  
  <div id="toast" class="toast"></div>
</body>
</html>`

const messageDetailTemplate = `<!doctype html>
<html lang="zh-CN">
<head>
  <meta charset="utf-8" />
  <meta name="viewport" content="width=device-width, initial-scale=1" />
  <title>%s - 邮件详情</title>
  <link rel="preconnect" href="https://fonts.googleapis.com">
  <link rel="preconnect" href="https://fonts.gstatic.com" crossorigin>
  <link href="https://fonts.googleapis.com/css2?family=Inter:wght@400;500;600;700&display=swap" rel="stylesheet">
  <style>
    * {
      margin: 0;
      padding: 0;
      box-sizing: border-box;
    }
    
    body {
      font-family: 'Inter', -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
      background: linear-gradient(135deg, #667eea 0%%, #764ba2 100%%);
      min-height: 100vh;
      padding: 1rem 1rem;
      color: #2d3748;
    }
    
    .container {
      max-width: 650px;
      margin: 0 auto;
      animation: fadeIn 0.6s ease-out;
    }
    
    @keyframes fadeIn {
      from { opacity: 0; transform: translateY(20px); }
      to { opacity: 1; transform: translateY(0); }
    }
    
    .back-btn {
      display: inline-flex;
      align-items: center;
      gap: 0.35rem;
      color: white;
      text-decoration: none;
      font-weight: 600;
      margin-bottom: 0.75rem;
      padding: 0.4rem 0.85rem;
      background: rgba(255,255,255,0.2);
      border-radius: 8px;
      transition: all 0.3s ease;
      font-size: 0.8rem;
    }
    
    .back-btn:hover {
      background: rgba(255,255,255,0.3);
      transform: translateX(-4px);
    }
    
    .card {
      background: white;
      border-radius: 12px;
      box-shadow: 0 10px 30px rgba(0,0,0,0.3);
      overflow: hidden;
    }
    
    .email-header {
      background: linear-gradient(135deg, #f6f8fb 0%%, #e9ecef 100%%);
      padding: 1rem;
      border-bottom: 2px solid #e2e8f0;
    }
    
    .subject {
      font-size: 1.1rem;
      font-weight: 700;
      color: #2d3748;
      margin-bottom: 0.75rem;
      line-height: 1.3;
    }
    
    .meta-row {
      display: grid;
      grid-template-columns: 70px 1fr;
      gap: 0.6rem;
      margin-bottom: 0.4rem;
      align-items: start;
    }
    
    .meta-label {
      font-weight: 600;
      color: #718096;
      font-size: 0.7rem;
    }
    
    .meta-value {
      color: #2d3748;
      word-break: break-all;
      font-size: 0.8rem;
    }
    
    .email-body {
      padding: 1rem;
      min-height: 300px;
      background: white;
    }
    
    #html-frame {
      width: 100%%;
      min-height: 350px;
      border: 1px solid #e2e8f0;
      border-radius: 6px;
      background: white;
    }
    
    .text-content {
      font-family: 'Courier New', monospace;
      white-space: pre-wrap;
      word-wrap: break-word;
      color: #2d3748;
      line-height: 1.5;
      background: #f8fafc;
      padding: 0.75rem;
      border-radius: 6px;
      border: 1px solid #e2e8f0;
      font-size: 0.75rem;
    }
    
    .no-content {
      text-align: center;
      padding: 2rem;
      color: #a0aec0;
      font-size: 0.9rem;
    }
    
    .action-bar {
      display: flex;
      gap: 0.6rem;
      padding: 0.75rem 1rem;
      background: #f8fafc;
      border-top: 2px solid #e2e8f0;
      flex-wrap: wrap;
    }
    
    .btn {
      padding: 0.5rem 1rem;
      font-size: 0.75rem;
      font-weight: 600;
      border: none;
      border-radius: 6px;
      cursor: pointer;
      transition: all 0.3s ease;
      font-family: inherit;
      text-decoration: none;
      display: inline-block;
    }
    
    .btn-primary {
      background: linear-gradient(135deg, #667eea 0%%, #764ba2 100%%);
      color: white;
    }
    
    .btn-primary:hover {
      transform: translateY(-2px);
      box-shadow: 0 4px 12px rgba(102, 126, 234, 0.4);
    }
    
    .btn-secondary {
      background: white;
      color: #667eea;
      border: 2px solid #667eea;
    }
    
    .btn-secondary:hover {
      background: #667eea;
      color: white;
    }
    
    @media (max-width: 768px) {
      .meta-row {
        grid-template-columns: 1fr;
        gap: 0.25rem;
      }
      
      .tabs {
        overflow-x: auto;
      }
      
      .tab {
        padding: 0.75rem 1rem;
        font-size: 0.9rem;
        white-space: nowrap;
      }
      
      .tab-content {
        padding: 1rem;
      }
      
      .action-bar {
        flex-direction: column;
      }
      
      .btn {
        width: 100%%;
      }
    }
  </style>
</head>
<body>
  <div class="container">
    <a href="/?mailbox=%s" class="back-btn">← 返回收件箱</a>
    
    <div class="card">
      <div class="email-header">
        <div class="subject">%s</div>
        <div class="meta-row">
          <div class="meta-label">发件人:</div>
          <div class="meta-value">%s</div>
        </div>
        <div class="meta-row">
          <div class="meta-label">时间:</div>
          <div class="meta-value">%s</div>
        </div>
      </div>
      
      <div class="email-body">
        %s
      </div>
      
      <div class="action-bar">
        <a href="/?mailbox=%s" class="btn btn-primary">返回收件箱</a>
        <a href="/api/messages/%s/%s?format=raw" download="message.eml" class="btn btn-secondary">下载 EML 文件</a>
      </div>
    </div>
  </div>
</body>
</html>`
