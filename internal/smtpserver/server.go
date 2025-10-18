package smtpserver

import (
    "bytes"
    "crypto/tls"
    "encoding/base64"
    "fmt"
    "io"
    "mime"
    "net"
    stdmail "net/mail"
    "strings"
    "sync/atomic"

    "github.com/emersion/go-smtp"
    "temp_mail/internal/storage"
)

type Server struct {
    srv    *smtp.Server
    ln     net.Listener
    domain string
    open   atomic.Bool
}

func NewServer(store storage.Store, domain string) *Server {
    be := &backend{store: store, domain: strings.ToLower(domain)}
    s := smtp.NewServer(be)
    s.Domain = domain
    s.ReadTimeout = 0
    s.WriteTimeout = 0
    s.MaxMessageBytes = 20 * 1024 * 1024
    s.AllowInsecureAuth = true
    return &Server{srv: s, domain: domain}
}

func (s *Server) ListenAndServe(addr string) error {
    ln, err := net.Listen("tcp", addr)
    if err != nil {
        return err
    }
    s.ln = ln
    s.open.Store(true)
    return s.srv.Serve(ln)
}

func (s *Server) Shutdown() error {
    if s.open.Swap(false) {
        if s.ln != nil {
            _ = s.ln.Close()
        }
        s.srv.Close()
    }
    return nil
}

type backend struct {
    store  storage.Store
    domain string
}

func (b *backend) NewSession(c *smtp.Conn) (smtp.Session, error) {
    return &session{store: b.store, domain: b.domain}, nil
}

type session struct {
    store     storage.Store
    domain    string
    from      string
    rcptLocal string
}

func (s *session) AuthPlain(username, password string) error { return nil }
func (s *session) Mail(from string, opts *smtp.MailOptions) error {
    s.from = from
    return nil
}
func (s *session) Rcpt(to string, _ *smtp.RcptOptions) error {
    // Accept recipient if domain matches or if no domain provided (catch-all)
    addr, err := stdmail.ParseAddress(to)
    if err != nil {
        return err
    }
    parts := strings.Split(addr.Address, "@")
    local := parts[0]
    if len(parts) == 2 {
        dom := strings.ToLower(parts[1])
        if s.domain != "" && dom != strings.ToLower(s.domain) {
            return fmt.Errorf("recipient domain not accepted: %s", dom)
        }
    }
    // Normalize plus addressing (local+tag)
    if i := strings.IndexByte(local, '+'); i > 0 {
        local = local[:i]
    }
    s.rcptLocal = s.store.CreateAddress(local)
    return nil
}
func (s *session) Data(r io.Reader) error {
    buf := new(bytes.Buffer)
    if _, err := io.Copy(buf, r); err != nil {
        return err
    }
    raw := buf.Bytes()
    // Parse headers using net/mail to get From and Subject
    var subj string
    var snippet string
    from := s.from
    dec := new(mime.WordDecoder)
    
    if msg, err := stdmail.ReadMessage(bytes.NewReader(raw)); err == nil {
        // Decode MIME encoded subject
        if rawSubj := msg.Header.Get("Subject"); rawSubj != "" {
            if decodedSubj, err := dec.DecodeHeader(rawSubj); err == nil {
                subj = decodedSubj
            } else {
                subj = rawSubj
            }
        }
        
        // Decode MIME encoded from
        if h := msg.Header.Get("From"); h != "" {
            if decodedFrom, err := dec.DecodeHeader(h); err == nil {
                from = decodedFrom
            } else {
                from = h
            }
        }
        
        // Extract snippet from body with encoding support
        contentType := msg.Header.Get("Content-Type")
        contentTransferEncoding := strings.ToLower(msg.Header.Get("Content-Transfer-Encoding"))
        
        if b, err := io.ReadAll(msg.Body); err == nil {
            bodyBytes := b
            bodyText := ""
            
            // Handle multipart messages
            if strings.Contains(contentType, "multipart") {
                bodyText = extractTextFromMultipart(string(b), contentType)
            } else {
                // Decode body if encoded
                switch contentTransferEncoding {
                case "base64":
                    bodyStr := strings.ReplaceAll(string(b), "\r\n", "")
                    bodyStr = strings.ReplaceAll(bodyStr, "\n", "")
                    if decoded, err := base64.StdEncoding.DecodeString(bodyStr); err == nil {
                        bodyBytes = decoded
                    }
                case "quoted-printable":
                    bodyBytes = decodeQuotedPrintableSimple(b)
                }
                bodyText = string(bodyBytes)
            }
            
            // Clean up text for snippet
            t := strings.ReplaceAll(bodyText, "\r", "")
            t = strings.ReplaceAll(t, "\n", " ")
            t = strings.TrimSpace(t)
            
            // Remove HTML tags if present
            if strings.Contains(strings.ToLower(contentType), "html") {
                t = stripHTMLTags(t)
            }
            
            if len(t) > 160 {
                t = t[:160] + "..."
            }
            snippet = t
        }
    }
    _, saveErr := s.store.Save(s.rcptLocal, storage.Message{
        From:    from,
        Subject: subj,
        Snippet: snippet,
        Raw:     raw,
    })
    return saveErr
}
func (s *session) Reset() {}
func (s *session) Logout() error { return nil }

// Optional: STARTTLS config placeholder (not used for local dev)
func (s *Server) SetTLSConfig(cfg *tls.Config) {
    s.srv.TLSConfig = cfg
}

// extractTextFromMultipart extracts plain text from multipart email
func extractTextFromMultipart(body, contentType string) string {
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
        return body
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
        
        // Only process text/plain parts
        if !strings.Contains(headers, "text/plain") {
            continue
        }
        
        // Decode content based on Content-Transfer-Encoding
        if strings.Contains(headers, "base64") {
            content = strings.ReplaceAll(content, "\r\n", "")
            content = strings.ReplaceAll(content, "\n", "")
            if decoded, err := base64.StdEncoding.DecodeString(content); err == nil {
                return string(decoded)
            }
        } else if strings.Contains(headers, "quoted-printable") {
            return string(decodeQuotedPrintableSimple([]byte(content)))
        }
        
        return content
    }
    
    return ""
}

// stripHTMLTags removes HTML tags from text
func stripHTMLTags(s string) string {
    // Simple HTML tag removal
    inTag := false
    var result strings.Builder
    for _, r := range s {
        if r == '<' {
            inTag = true
            continue
        }
        if r == '>' {
            inTag = false
            continue
        }
        if !inTag {
            result.WriteRune(r)
        }
    }
    return result.String()
}

// decodeQuotedPrintableSimple decodes quoted-printable encoded text
func decodeQuotedPrintableSimple(data []byte) []byte {
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
