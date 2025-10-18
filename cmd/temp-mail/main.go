package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"temp_mail/internal/httpapi"
	"temp_mail/internal/smtpclient"
	"temp_mail/internal/smtpserver"
	"temp_mail/internal/storage"
)

func main() {
	// Config via env with defaults
	httpAddr := getenv("HTTP_ADDR", ":8080")
	smtpAddr := getenv("SMTP_ADDR", ":2525")
	domain := getenv("DOMAIN", "tmp.local")
	ttlStr := getenv("MESSAGE_TTL", "30m")
	ttl, err := time.ParseDuration(ttlStr)
	if err != nil {
		log.Fatalf("invalid MESSAGE_TTL: %v", err)
	}

	store := storage.NewMemoryStore(ttl)

	// SMTP客户端配置（用于发送邮件）
	// 使用本地域名创建发送客户端
	smtpClient := smtpclient.NewClient(domain)
	log.Printf("SMTP发送客户端已启用，使用域名: %s", domain)

	// HTTP server
	mux := httpapi.NewMux(store, domain, smtpClient)
	httpSrv := &http.Server{Addr: httpAddr, Handler: mux}

	// SMTP server
	smtpSrv := smtpserver.NewServer(store, domain)

	// Run servers
	go func() {
		log.Printf("HTTP listening on %s", httpAddr)
		if err := httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("http server: %v", err)
		}
	}()

	go func() {
		log.Printf("SMTP listening on %s for domain %s", smtpAddr, domain)
		if err := smtpSrv.ListenAndServe(smtpAddr); err != nil {
			log.Fatalf("smtp server: %v", err)
		}
	}()

	// Graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop
	log.Printf("Shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = httpSrv.Shutdown(ctx)
	_ = smtpSrv.Shutdown()
	store.Close()
}

func getenv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
