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

	// Static files
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

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
		// 为HTML内容注入暗色主题样式
		styledHTML := injectDarkThemeStyles(htmlContent)
		bodyHTML = fmt.Sprintf(`<iframe id="html-frame" srcdoc="%s" sandbox="allow-same-origin"></iframe>`, escapeHTMLAttr(styledHTML))
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
		msg.ID,
		local,
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

// injectDarkThemeStyles 为HTML邮件内容注入暗色主题样式
func injectDarkThemeStyles(htmlContent string) string {
	darkThemeCSS := `<style>
		body {
			background-color: rgba(5, 20, 5, 0.95) !important;
			color: #c0d0c0 !important;
			font-family: 'JetBrains Mono', 'Courier New', monospace !important;
			padding: 1.5rem !important;
			line-height: 1.6 !important;
		}
		* {
			color: #c0d0c0 !important;
			border-color: rgba(57, 255, 20, 0.3) !important;
		}
		a {
			color: #39ff14 !important;
			text-decoration: underline !important;
		}
		a:hover {
			color: #5fff5f !important;
		}
		table {
			border-collapse: collapse !important;
			background: rgba(0, 30, 0, 0.3) !important;
		}
		th, td {
			border: 1px solid rgba(57, 255, 20, 0.2) !important;
			padding: 0.5rem !important;
		}
		th {
			background: rgba(57, 255, 20, 0.1) !important;
			color: #39ff14 !important;
			font-weight: bold !important;
		}
		pre, code {
			background: rgba(0, 40, 0, 0.5) !important;
			border: 1px solid rgba(57, 255, 20, 0.2) !important;
			padding: 0.5rem !important;
			border-radius: 3px !important;
			color: #39ff14 !important;
		}
		img {
			filter: brightness(0.9) contrast(1.1) !important;
			border: 1px solid rgba(57, 255, 20, 0.2) !important;
		}
	</style>`

	// 尝试在 </head> 或 <body> 标签前插入样式
	if strings.Contains(strings.ToLower(htmlContent), "</head>") {
		htmlContent = strings.Replace(htmlContent, "</head>", darkThemeCSS+"</head>", 1)
	} else if strings.Contains(strings.ToLower(htmlContent), "<body") {
		htmlContent = strings.Replace(htmlContent, "<body", darkThemeCSS+"<body", 1)
	} else {
		// 如果没有标准HTML结构，直接在开头插入
		htmlContent = darkThemeCSS + htmlContent
	}

	return htmlContent
}

const indexHTML = `<!doctype html>
<html lang="zh-CN">
<head>
  <meta charset="utf-8" />
  <meta name="viewport" content="width=device-width, initial-scale=1, maximum-scale=1, user-scalable=no" />
  <title>AlienMail | Interstellar Uplink</title>
  <link rel="icon" type="image/svg+xml" href="/static/favicon.svg">
  <link rel="preconnect" href="https://fonts.googleapis.com">
  <link rel="preconnect" href="https://fonts.gstatic.com" crossorigin>
  <link href="https://fonts.googleapis.com/css2?family=JetBrains+Mono:wght@400;600&family=Share+Tech+Mono&family=Orbitron:wght@500;700;900&display=swap" rel="stylesheet">
  <style>
    :root {
      /* 异形飞船配色：生物荧光绿 + 深空黑 */
      --alien-green: #39ff14;
      --alien-dim: rgba(57, 255, 20, 0.2);
      --alien-dark: #051005;
      --hud-border: rgba(57, 255, 20, 0.4);
      --text-main: #e0f0e0;
      --text-muted: #608060;
      --warning: #ff3333;
      --card-bg: rgba(5, 15, 5, 0.85);
    }

    /* 滚动条样式 */
    ::-webkit-scrollbar { width: 4px; height: 4px; }
    ::-webkit-scrollbar-track { background: transparent; }
    ::-webkit-scrollbar-thumb { background: var(--alien-dim); border-radius: 0; }
    ::-webkit-scrollbar-thumb:hover { background: var(--alien-green); }

    * { margin: 0; padding: 0; box-sizing: border-box; }
    
    body {
      font-family: 'Share Tech Mono', monospace; /* 科技感字体 */
      background-color: #000;
      
      /* 飞船舷窗效果：深空背景 */
      background-image: 
        radial-gradient(circle at center, #0a1a0a 0%, #000 100%);
      
      height: 100vh; 
      overflow: hidden; /* 初始锁定，营造仪表盘感 */
      
      padding: 1rem;
      color: var(--text-main);
      display: flex;
      flex-direction: column;
      transition: height 0.3s ease;
    }
    
    /* 混合滚动模式 */
    body.scroll-mode {
      height: auto;
      min-height: 100vh;
      overflow-y: auto;
    }

    /* 星空粒子动画 */
    body::before {
      content: "";
      position: fixed;
      top: 0; left: 0; width: 100%; height: 100%;
      background-image: 
        radial-gradient(2px 2px at 20px 30px, #fff, rgba(0,0,0,0)),
        radial-gradient(2px 2px at 40px 70px, #fff, rgba(0,0,0,0)),
        radial-gradient(2px 2px at 50px 160px, #fff, rgba(0,0,0,0)),
        radial-gradient(2px 2px at 90px 40px, #fff, rgba(0,0,0,0)),
        radial-gradient(2px 2px at 130px 80px, #fff, rgba(0,0,0,0));
      background-size: 200px 200px;
      animation: warpSpeed 20s linear infinite;
      opacity: 0.4;
      pointer-events: none;
      z-index: -1;
    }
    
    /* 扫描线效果 */
    body::after {
      content: " ";
      display: block;
      position: fixed;
      top: 0; left: 0; bottom: 0; right: 0;
      background: linear-gradient(rgba(18, 16, 16, 0) 50%, rgba(0, 255, 0, 0.05) 50%);
      background-size: 100% 4px;
      z-index: 2;
      pointer-events: none;
    }

    @keyframes warpSpeed {
      from { transform: translateY(0); }
      to { transform: translateY(200px); }
    }

    .container {
      max-width: 800px;
      width: 100%;
      margin: 0 auto;
      height: 100%;
      display: flex;
      flex-direction: column;
      gap: 1.5rem;
      position: relative;
      z-index: 3;
    }
    
    .container.scroll-mode { height: auto; }

    .header {
      text-align: center;
      flex-shrink: 0;
      text-transform: uppercase;
    }
    .header h1 {
      font-family: 'Orbitron', sans-serif;
      font-size: 2rem;
      font-weight: 900;
      color: var(--alien-green);
      text-shadow: 0 0 15px var(--alien-green);
      letter-spacing: 2px;
      margin-bottom: 0.5rem;
    }
    .header p {
      color: var(--text-muted);
      font-size: 0.8rem;
      letter-spacing: 4px;
    }
    
    /* 异形飞船风格卡片 */
    .card {
      background: var(--card-bg);
      /* 全息玻璃质感 */
      backdrop-filter: blur(10px);
      border: 1px solid var(--hud-border);
      box-shadow: 0 0 20px rgba(57, 255, 20, 0.05), inset 0 0 30px rgba(57, 255, 20, 0.02);
      clip-path: polygon(
        0 0, 
        100% 0, 
        100% calc(100% - 20px), 
        calc(100% - 20px) 100%, 
        0 100%
      ); /* 切角设计 */
      position: relative;
    }
    
    /* 装饰性角落 */
    .card::before {
      content: '';
      position: absolute;
      bottom: 0; right: 0;
      width: 20px; height: 20px;
      border-bottom: 2px solid var(--alien-green);
      border-right: 2px solid var(--alien-green);
      clip-path: polygon(0 100%, 100% 0, 100% 100%);
    }

    .control-card {
      flex-shrink: 0;
      padding: 1.5rem;
    }

    .content-card {
      flex: 1; 
      min-height: 0;
      max-height: 500px;
      display: flex;
      flex-direction: column;
      padding: 0;
      transition: all 0.3s ease;
    }
    
    .content-card.auto-height {
      flex: none;
      height: auto;
      min-height: 500px;
      max-height: none;
      overflow: visible;
    }

    .card-header-strip {
      padding: 1rem 1.5rem;
      border-bottom: 1px solid var(--hud-border);
      background: rgba(57, 255, 20, 0.05);
      flex-shrink: 0;
    }
    
    /* 输入框：终端风格 */
    .create-section { display: flex; gap: 1rem; margin-bottom: 1.2rem; }
    .input-wrapper { flex: 1; position: relative; }
    
    input, textarea {
      width: 100%; padding: 0.8rem 1rem; font-size: 1rem;
      background: rgba(0, 20, 0, 0.6); 
      border: 1px solid var(--text-muted);
      color: var(--alien-green); 
      font-family: 'JetBrains Mono', monospace;
      transition: all 0.3s ease;
      text-transform: lowercase;
    }
    input:focus, textarea:focus {
      outline: none; border-color: var(--alien-green);
      box-shadow: 0 0 15px var(--alien-dim);
      background: rgba(0, 30, 0, 0.8);
    }
    input::placeholder, textarea::placeholder { color: rgba(57, 255, 20, 0.3); text-transform: none; }
    
    /* 按钮：全息按钮 */
    .btn {
      padding: 0.8rem 1.5rem; font-size: 0.9rem; font-weight: 700;
      border: 1px solid var(--alien-green);
      background: rgba(57, 255, 20, 0.1);
      color: var(--alien-green);
      cursor: pointer; transition: all 0.2s;
      font-family: 'Orbitron', sans-serif; 
      text-transform: uppercase; letter-spacing: 1px;
      clip-path: polygon(10px 0, 100% 0, 100% calc(100% - 10px), calc(100% - 10px) 100%, 0 100%, 0 10px);
    }
    .btn:hover {
      background: var(--alien-green);
      color: black;
      box-shadow: 0 0 20px var(--alien-green);
    }
    
    .address-display {
      background: rgba(0, 20, 0, 0.4); padding: 0.8rem 1rem;
      border: 1px dashed var(--text-muted);
      display: flex; align-items: center; justify-content: space-between; gap: 1rem;
    }
    .address-info { display: flex; flex-direction: column; overflow: hidden; }
    .address-label { font-size: 0.6rem; color: var(--text-muted); letter-spacing: 1px; margin-bottom: 0.2rem; }
    .address-value { font-size: 1.1rem; font-weight: 600; color: var(--alien-green); font-family: 'JetBrains Mono', monospace; text-shadow: 0 0 5px var(--alien-green); }
    
    .btn-copy {
      font-size: 0.7rem; padding: 0.5rem 1rem;
    }

    .tabs { display: flex; gap: 2px; }
    .tab {
      padding: 0.6rem 1.5rem; background: rgba(0,0,0,0.3); border: 1px solid transparent; border-bottom: none;
      color: var(--text-muted); font-weight: 600; font-size: 0.9rem;
      cursor: pointer; transition: all 0.3s; font-family: 'Orbitron', sans-serif;
    }
    .tab:hover { color: var(--alien-green); }
    .tab.active { 
      color: var(--alien-green); 
      background: rgba(57, 255, 20, 0.05);
      border-color: var(--hud-border);
      border-bottom: 1px solid var(--card-bg); /* Merge with content */
      margin-bottom: -1px; z-index: 10;
      box-shadow: 0 -5px 10px rgba(57, 255, 20, 0.05);
    }
    
    .tab-content { 
      display: none; 
      flex: 1; flex-direction: column; 
      overflow: hidden;
    }
    .tab-content.active { display: flex; }
    
    /* 收件箱：内部滚动 */
    .scroll-area-internal {
      flex: 1;
      overflow-y: auto; 
      padding: 1.5rem;
    }
    
    /* 发送页：自然高度 */
    .scroll-area-natural {
      height: auto;
      padding: 1.5rem;
    }

    .messages-container { display: flex; flex-direction: column; gap: 0.5rem; }
    
    .message-item {
      background: rgba(0, 20, 0, 0.3); padding: 1rem;
      border-left: 2px solid var(--text-muted);
      transition: all 0.2s ease; cursor: pointer;
    }
    .message-item:hover { 
      background: rgba(57, 255, 20, 0.05); 
      border-left-color: var(--alien-green);
      box-shadow: inset 5px 0 10px rgba(57, 255, 20, 0.05);
    }
    
    .message-header { display: flex; justify-content: space-between; align-items: baseline; margin-bottom: 0.4rem; }
    .message-from { font-weight: 700; color: #fff; font-size: 0.95rem; letter-spacing: 0.5px; }
    .message-time { font-size: 0.7rem; color: var(--text-muted); }
    .message-subject { color: var(--alien-green); font-size: 0.9rem; }
    
    .compose-form { display: flex; flex-direction: column; gap: 1.2rem; }
    .compose-editor { min-height: 350px; resize: vertical; border-left: 3px solid var(--alien-dim); }
    
    .empty-state { text-align: center; margin-top: 4rem; color: var(--text-muted); }
    .empty-state-icon { font-size: 3rem; margin-bottom: 1rem; opacity: 0.3; animation: float 3s ease-in-out infinite; }
    @keyframes float { 0% { transform: translateY(0px); } 50% { transform: translateY(-10px); } 100% { transform: translateY(0px); } }

    .toast {
      position: fixed; bottom: 2rem; right: 2rem; 
      background: #000; border: 1px solid var(--alien-green);
      color: var(--alien-green); padding: 1rem 2rem;
      box-shadow: 0 0 30px rgba(57, 255, 20, 0.2);
      transform: translateY(100px); opacity: 0; transition: all 0.3s;
      z-index: 1000; font-family: 'JetBrains Mono', monospace;
      text-transform: uppercase;
    }
    .toast.error { border-color: var(--warning); color: var(--warning); box-shadow: 0 0 30px rgba(255, 51, 51, 0.2); }
    .toast.show { transform: translateY(0); opacity: 1; }

    .badge { background: var(--alien-green); color: black; padding: 2px 6px; font-weight: bold; border-radius: 2px; font-size: 0.7rem; }
    
    /* 呼吸灯动画 */
    .loading { 
      display: inline-block; width: 14px; height: 14px; 
      border: 2px solid var(--alien-green); border-radius: 50%; 
      animation: pulse 1s infinite;
    }
    @keyframes pulse { 0% { transform: scale(0.8); opacity: 0.5; } 100% { transform: scale(1.2); opacity: 0; } }

    #ttl-info { margin-top: 1rem; color: var(--warning); font-size: 0.8rem; display: none; text-align: center; border: 1px solid var(--warning); padding: 0.5rem; background: rgba(255, 51, 51, 0.05); }

    /* 移动端/小屏适配 */
    @media (max-height: 700px), (max-width: 600px) {
      body { height: auto; overflow-y: auto; }
      .container { height: auto; padding: 0.5rem; gap: 1rem; }
      .content-card { max-height: 400px; min-height: 300px; }
      .address-display { flex-direction: column; align-items: stretch; gap: 0.5rem; }
      .address-value { font-size: 0.85rem; word-break: break-all; }
      .btn-copy { width: 100%; }
      .create-section { flex-direction: column; }
      .input-wrapper { min-width: 100%; }
    }
  </style>
  <script>
    let currentLocal = '';
    let currentDomain = '';
    let pollInterval = null;
    let messageTTL = 30;
    let lastMessageIds = [];
    
    async function createAddr() {
      const input = document.getElementById('local');
      const desired = input.value.trim();
      
      // UI Feedback
      const btn = document.querySelector('.create-section .btn');
      const originalBtnText = btn.textContent;
      btn.disabled = true;
      btn.innerHTML = '<span class="loading"></span>';
      
      try {
        const r = await fetch('/api/address?local=' + encodeURIComponent(desired), {method: 'POST'});
        const j = await r.json();
        currentLocal = j.local;
        currentDomain = j.address.split('@')[1];
        messageTTL = j.ttl || 30;
        document.getElementById('addr').textContent = j.address;
        document.getElementById('copy-section').style.display = 'block';
        document.getElementById('ttl-info').style.display = 'block';
        document.getElementById('ttl-minutes').textContent = messageTTL;
        const sendFromInput = document.getElementById('send-from');
        if (sendFromInput) {
          sendFromInput.value = j.address;
          sendFromInput.style.opacity = '1';
          sendFromInput.placeholder = '';
          console.log('Set send-from to:', j.address, 'Element:', sendFromInput);
        } else {
          console.error('send-from element not found!');
        }
        loadMsgs();
        startPolling();
        showToast('>>> LINK ESTABLISHED');
      } catch (e) { 
        console.error('Create address error:', e);
        showToast('CONNECTION ERROR', 'error'); 
      } 
      finally { btn.disabled = false; btn.textContent = originalBtnText; }
    }
    
    async function loadMsgs() {
      if (!currentLocal) return;
      try {
        const r = await fetch('/api/messages/' + currentLocal);
        const msgs = await r.json() || [];
        document.getElementById('inbox-badge').textContent = msgs.length;
        if (msgs.length === 0) {
          document.getElementById('messages-container').innerHTML = '<div class="empty-state"><div class="empty-state-icon">👾</div><h3>SILENCE</h3><p>Scanning void for data...</p></div>';
          lastMessageIds = [];
          return;
        }
        const currentIds = msgs.map(m => m.id);
        if (currentIds.length !== lastMessageIds.length || currentIds.some((id, i) => id !== lastMessageIds[i])) {
          renderMessages(msgs);
          lastMessageIds = currentIds;
        } else { updateTimers(msgs); }
      } catch (e) { console.error(e); }
    }
    
    function renderMessages(msgs) {
      const container = document.getElementById('messages-container');
      container.innerHTML = '';
      for (const m of msgs) {
        const div = document.createElement('div');
        div.className = 'message-item';
        div.setAttribute('data-msg-id', m.id);
        div.onclick = function() { window.location.href = '/view/' + currentLocal + '/' + m.id; };
        const now = new Date();
        const minutesLeft = Math.max(0, Math.floor((new Date(m.expiresAt) - now) / 60000));
        div.innerHTML = 
          '<div class="message-header">' +
            '<div class="message-from">' + escapeHtml(m.from || 'UNKNOWN ENTITY') + '</div>' +
            '<div class="message-time">' + new Date(m.createdAt).toLocaleTimeString() + '</div>' +
          '</div>' +
          '<div class="message-subject">' + escapeHtml(m.subject || 'ENCRYPTED') + '</div>' +
          '<div class="message-snippet">' + escapeHtml(m.snippet || '') + '</div>' + 
          (minutesLeft > 0 ? '<div style="margin-top:0.5rem;font-size:0.65rem;color:#ff3333;text-align:right;" class="expiry-timer" data-expires="'+m.expiresAt+'">SELF-DESTRUCT IN ' + minutesLeft + ' MIN</div>' : '');
        container.appendChild(div);
      }
    }

    function updateTimers(msgs) {
        const now = new Date();
        msgs.forEach(m => {
            const el = document.querySelector('[data-msg-id="' + m.id + '"] .expiry-timer');
            if (el) {
                const left = Math.max(0, Math.floor((new Date(m.expiresAt) - now) / 60000));
                el.textContent = 'SELF-DESTRUCT IN ' + left + ' MIN';
            }
        });
    }
    
    function copyAddress() {
      const addr = document.getElementById('addr').textContent;
      navigator.clipboard.writeText(addr).then(() => showToast('>>> COPIED TO CLIPBOARD')).catch(() => showToast('COPY FAILED', 'error'));
    }
    
    function switchTab(tabName) {
      document.querySelectorAll('.tab').forEach(tab => tab.classList.remove('active'));
      event.target.classList.add('active');
      document.querySelectorAll('.tab-content').forEach(c => c.classList.remove('active'));
      
      const body = document.body;
      const container = document.querySelector('.container');
      const contentCard = document.querySelector('.content-card');

      if (tabName === 'inbox') {
        document.getElementById('inbox-tab').classList.add('active');
        body.classList.remove('scroll-mode');
        container.classList.remove('scroll-mode');
        contentCard.classList.remove('auto-height');
      } else {
        document.getElementById('compose-tab').classList.add('active');
        body.classList.add('scroll-mode');
        container.classList.add('scroll-mode');
        contentCard.classList.add('auto-height');
        if (!currentLocal) showToast('! ERROR: NO UPLINK DETECTED !', 'error');
      }
    }
    
    async function sendEmail() {
      if (!currentLocal) { showToast('Initialize Uplink First', 'error'); return; }
      const to = document.getElementById('send-to').value;
      const subject = document.getElementById('send-subject').value;
      const body = document.getElementById('send-body').value;
      if (!to || !subject || !body) { showToast('Input Error: Missing Data', 'error'); return; }
      const btn = document.getElementById('send-btn');
      const originalText = btn.textContent;
      btn.textContent = 'TRANSMITTING...';
      try {
        const res = await fetch('/api/send', {
          method: 'POST',
          headers: {'Content-Type': 'application/json'},
          body: JSON.stringify({ from: currentLocal, to: to.split(','), subject, body }),
        });
        const json = await res.json();
        if (res.ok && json.success) {
          showToast('>>> SENT SUCCESSFULLY');
          document.getElementById('send-body').value = '';
          document.getElementById('send-subject').value = '';
        } else { throw new Error(json.error); }
      } catch (e) { showToast('TRANSMISSION FAILED', 'error'); } 
      finally { btn.textContent = originalText; }
    }
    
    function startPolling() {
      if (pollInterval) clearInterval(pollInterval);
      pollInterval = setInterval(loadMsgs, 4000);
    }
    
    function showToast(msg, type='success') {
      const t = document.getElementById('toast');
      t.textContent = msg;
      t.className = 'toast show ' + type;
      setTimeout(() => t.classList.remove('show'), 3000);
    }
    
    function escapeHtml(text) {
      const div = document.createElement('div'); div.innerText = text; return div.innerHTML;
    }

    document.addEventListener('DOMContentLoaded', () => {
        const params = new URLSearchParams(window.location.search);
        if (params.get('mailbox')) {
            const mailbox = params.get('mailbox');
            currentLocal = mailbox;
            document.getElementById('local').value = mailbox;
            // Try to determine domain - call API to get full address
            fetch('/api/address?local=' + encodeURIComponent(mailbox), {method: 'POST'})
                .then(r => r.json())
                .then(j => {
                    currentDomain = j.address.split('@')[1];
                    document.getElementById('addr').textContent = j.address;
                    document.getElementById('copy-section').style.display = 'block';
                    document.getElementById('ttl-info').style.display = 'block';
                    document.getElementById('ttl-minutes').textContent = j.ttl || 30;
                    document.getElementById('send-from').value = j.address;
                    loadMsgs();
                    startPolling();
                })
                .catch(() => {
                    // Fallback if API fails
                    showToast('CONNECTION ERROR', 'error');
                });
        }
    });
  </script>
</head>
<body>
  <div class="container">
    <div class="header">
      <h1>ALIEN MAIL</h1>
      <p>SECURE / ANONYMOUS / EPHEMERAL</p>
    </div>
    
    <div class="card control-card">
      <div class="create-section">
        <div class="input-wrapper">
          <input id="local" type="text" placeholder="ENTER ALIAS..." autocomplete="off" onkeyup="if(event.key === 'Enter') createAddr()" />
        </div>
        <button class="btn" onclick="createAddr()">INITIALIZE</button>
      </div>
      
      <div class="address-display">
        <div class="address-info">
            <div class="address-label">ACTIVE FREQUENCY</div>
            <div id="addr" class="address-value empty">WAITING FOR INPUT...</div>
        </div>
        <button id="copy-section" class="btn btn-copy" onclick="copyAddress()" style="display:none;">COPY</button>
      </div>
      
      <div id="ttl-info">
         ⚠️ WARNING: DATA PURGE IN <strong id="ttl-minutes">30</strong> MIN
      </div>
    </div>

    <div class="card content-card">
      <div class="card-header-strip">
        <div class="tabs">
          <button class="tab active" onclick="switchTab('inbox')">
            📥 INBOX <span id="inbox-badge" class="badge" style="margin-left:0.4rem;">0</span>
          </button>
          <button class="tab" onclick="switchTab('compose')">
            📤 TRANSMIT
          </button>
        </div>
      </div>
      
      <div id="inbox-tab" class="tab-content active">
        <div class="scroll-area-internal">
          <div id="messages-container" class="messages-container">
            <div class="empty-state">
              <div class="empty-state-icon">👾</div>
              <h3>SYSTEM READY</h3>
              <p>Scanning for incoming signals...</p>
            </div>
          </div>
        </div>
      </div>
      
      <div id="compose-tab" class="tab-content">
        <div class="scroll-area-natural">
          <div class="compose-form">
            <div>
              <label style="font-size:0.7rem;color:var(--alien-green);font-weight:700;margin-bottom:0.3rem;display:block;">ORIGIN</label>
              <input id="send-from" type="text" readonly style="cursor:not-allowed;border-style:dashed;" placeholder="NO UPLINK" />
            </div>
            <div>
              <label style="font-size:0.7rem;color:var(--alien-green);font-weight:700;margin-bottom:0.3rem;display:block;">TARGET</label>
              <input id="send-to" type="email" placeholder="RECIPIENT@GALAXY.COM" />
            </div>
            <div>
              <label style="font-size:0.7rem;color:var(--alien-green);font-weight:700;margin-bottom:0.3rem;display:block;">SUBJECT</label>
              <input id="send-subject" type="text" placeholder="HEADER INFO..." />
            </div>
            <div style="flex:1; display:flex; flex-direction:column;">
              <label style="font-size:0.7rem;color:var(--alien-green);font-weight:700;margin-bottom:0.3rem;display:block;">PAYLOAD</label>
              <textarea id="send-body" class="compose-editor" placeholder="ENTER DATA STREAM..."></textarea>
            </div>
            <div style="display:flex; gap:1rem;">
              <button id="send-btn" class="btn" onclick="sendEmail()" style="flex:2">TRANSMIT DATA</button>
              <button class="btn" onclick="document.getElementById('send-body').value=''" style="flex:1;border-color:var(--text-muted);color:var(--text-muted);">PURGE</button>
            </div>
          </div>
        </div>
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
  <title>%s - DECODED MESSAGE</title>
  <link rel="icon" type="image/svg+xml" href="/static/favicon.svg">
  <link rel="preconnect" href="https://fonts.googleapis.com">
  <link rel="preconnect" href="https://fonts.gstatic.com" crossorigin>
  <link href="https://fonts.googleapis.com/css2?family=JetBrains+Mono:wght@400;600&family=Share+Tech+Mono&display=swap" rel="stylesheet">
  <style>
    :root {
      --alien-green: #39ff14;
      --bg-dark: #020502;
      --card-bg: rgba(5, 20, 5, 0.9);
      --text-main: #e0f0e0;
      --border-color: rgba(57, 255, 20, 0.3);
    }
    
    ::-webkit-scrollbar { width: 4px; }
    ::-webkit-scrollbar-track { background: transparent; }
    ::-webkit-scrollbar-thumb { background: var(--border-color); }

    * { margin: 0; padding: 0; box-sizing: border-box; }
    
    body {
      font-family: 'Share Tech Mono', monospace;
      background-color: #000;
      background-image: radial-gradient(circle at 50%% 30%%, #0f2010 0%%, #000 100%%);
      height: 100vh;
      overflow: hidden; 
      padding: 1.5rem;
      color: var(--text-main);
      display: flex;
      flex-direction: column;
    }
    
    .scanline {
      position: fixed; top: 0; left: 0; width: 100%%; height: 100%%;
      background: linear-gradient(to bottom, transparent 50%%, rgba(0, 255, 0, 0.02) 50%%);
      background-size: 100%% 4px; pointer-events: none; z-index: 999;
    }

    .container {
      max-width: 900px;
      width: 100%%;
      margin: 0 auto;
      height: 100%%;
      display: flex;
      flex-direction: column;
      animation: fadeIn 0.3s ease-out;
    }
    @keyframes fadeIn { from { opacity: 0; transform: scale(0.98); } to { opacity: 1; transform: scale(1); } }
    
    .top-bar {
      display: flex; justify-content: space-between; align-items: center;
      margin-bottom: 1rem;
      flex-shrink: 0;
    }

    .back-btn {
      color: var(--alien-green); text-decoration: none; font-weight: bold;
      font-size: 1rem; text-transform: uppercase; letter-spacing: 2px;
      display: flex; align-items: center; gap: 0.5rem;
      text-shadow: 0 0 5px var(--alien-green);
    }
    .back-btn:hover { text-decoration: underline; }
    
    .card {
      background: var(--card-bg);
      border: 1px solid var(--border-color);
      flex: 1;
      display: flex;
      flex-direction: column;
      overflow: hidden;
      box-shadow: 0 0 30px rgba(0, 50, 0, 0.5);
      clip-path: polygon(0 0, 100%% 0, 100%% calc(100%% - 20px), calc(100%% - 20px) 100%%, 0 100%%);
    }
    
    .email-header {
      padding: 1.5rem;
      border-bottom: 1px solid var(--border-color);
      background: rgba(0, 50, 0, 0.2);
      flex-shrink: 0;
    }
    
    .subject {
      font-size: 1.4rem; font-weight: 700; color: white; margin-bottom: 0.8rem; letter-spacing: 1px;
    }
    
    .meta-row {
      font-family: 'JetBrains Mono', monospace; font-size: 0.85rem;
      color: #80a080; margin-bottom: 0.3rem; display: flex; gap: 1rem;
    }
    .meta-label { color: var(--alien-green); min-width: 60px; }
    
    .email-content-wrapper {
      flex: 1;
      position: relative;
      background: rgba(5, 20, 5, 0.9);
      overflow: hidden;
      border-top: 1px solid rgba(57, 255, 20, 0.1);
    }
    
    #html-frame {
      width: 100%%; height: 100%%; border: none;
      background: rgba(5, 20, 5, 0.95);
    }
    
    .text-content {
      padding: 2rem; font-family: 'JetBrains Mono', monospace;
      color: #c0d0c0; overflow-y: auto; height: 100%%; white-space: pre-wrap;
      line-height: 1.6;
    }
    
    .no-content {
      padding: 3rem;
      text-align: center;
      color: var(--text-muted);
      font-style: italic;
    }
    
    .action-bar {
      padding: 1rem; background: rgba(0, 20, 0, 0.6);
      border-top: 1px solid var(--border-color);
      display: flex; gap: 1rem; justify-content: flex-end;
      flex-shrink: 0;
    }
    
    .btn {
      padding: 0.6rem 1.2rem; border: 1px solid var(--alien-green); 
      text-decoration: none; font-size: 0.8rem; font-weight: bold; 
      color: var(--alien-green); text-transform: uppercase;
      transition: 0.3s; background: rgba(0, 50, 0, 0.2);
    }
    .btn:hover { background: var(--alien-green); color: black; box-shadow: 0 0 15px var(--alien-green); }
    .btn-outline { border-style: dashed; }

  </style>
</head>
<body>
  <div class="scanline"></div>
  <div class="container">
    <div class="top-bar">
      <a href="/?mailbox=%s" class="back-btn">← RETURN TO BRIDGE</a>
      <div style="font-size:0.7rem; color:#446644;">SECURE CHANNEL ESTABLISHED</div>
    </div>
    
    <div class="card">
      <div class="email-header">
        <div class="subject">%s</div>
        <div class="meta-row"><span class="meta-label">FROM:</span> <span>%s</span></div>
        <div class="meta-row"><span class="meta-label">TIME:</span> <span>%s</span></div>
      </div>
      
      <div class="email-content-wrapper">
        %s
      </div>
      
      <div class="action-bar">
        <a href="/api/messages/%s/%s?format=raw" download="message.eml" class="btn btn-outline">DOWNLOAD RAW</a>
        <a href="/?mailbox=%s" class="btn">CLOSE VIEWER</a>
      </div>
    </div>
  </div>
</body>
</html>`
