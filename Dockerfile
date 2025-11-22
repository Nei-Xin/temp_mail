# 多阶段构建 Dockerfile - Temp Mail
# 生产环境镜像，直接监听 25 端口

# ============================================
# 阶段 1: 构建阶段
# ============================================
FROM golang:1.22-alpine AS builder

# 安装构建依赖
RUN apk add --no-cache git ca-certificates tzdata

# 设置工作目录
WORKDIR /build

# 复制 go mod 文件
COPY go.mod go.sum ./

# 下载依赖
RUN go mod download

# 复制源代码
COPY . .

# 编译程序（静态链接，减小体积）
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s" \
    -o temp-mail \
    ./cmd/temp-mail

# ============================================
# 阶段 2: 运行阶段
# ============================================
FROM alpine:latest

# 安装运行时依赖
RUN apk --no-cache add ca-certificates tzdata

# 设置时区
ENV TZ=Asia/Shanghai

# 创建工作目录
WORKDIR /app

# 从构建阶段复制二进制文件
COPY --from=builder /build/temp-mail /app/temp-mail

# 复制静态文件
COPY --from=builder /build/static /app/static

# 设置文件权限
RUN chmod +x /app/temp-mail

# 以 root 运行（监听 25 端口需要特权）
# USER root (默认就是 root)

# 暴露端口
EXPOSE 8080 25

# 设置默认环境变量
ENV HTTP_ADDR=:8080 \
    SMTP_ADDR=:25 \
    DOMAIN=tmp.local \
    MESSAGE_TTL=30m

# 健康检查
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/ || exit 1

# 启动程序
CMD ["/app/temp-mail"]
