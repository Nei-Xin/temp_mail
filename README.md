# Temp Mail - 临时邮箱服务# Temp Mail - 临时邮箱服务# Temp Mail - 临时邮箱服务# Temp Mail - 临时邮箱服务# Temp Mail - 临时邮箱服务



一个轻量级的临时邮箱服务，使用 Go 语言实现。支持收发邮件，提供简洁的 Web 界面。



## ✨ 特性一个轻量级的临时邮箱服务，使用 Go 语言实现。支持收发邮件，提供简洁的 Web 界面。



- 📧 SMTP 邮件接收（端口 25）

- 📤 SMTP 邮件发送（自动查找 MX 记录 + STARTTLS 加密）

- 🎯 标签页界面（收件箱/发送邮件切换方便）## ✨ 特性一个轻量级的临时邮箱服务，使用 Go 语言实现。支持收发邮件，提供简洁的 Web 界面。

- 💾 内存存储（邮件自动过期，默认 1 小时）

- 🐳 Docker 支持（一键部署，镜像仅 24.5MB）

- 🔧 环境变量配置（无需修改代码）

- 📧 **SMTP 邮件接收**：监听 25 端口，接收外部邮件

---

- 📤 **SMTP 邮件发送**：使用临时邮箱地址发送邮件，自动查找 MX 记录

## 🚀 快速开始

- 🔒 **STARTTLS 支持**：支持加密连接，提高安全性## ✨ 特性一个轻量级的临时邮箱服务，使用 Go 语言实现。支持收发邮件，提供简洁的 Web 界面。一个轻量级的临时邮箱服务，使用 Go 语言实现。

### Docker Compose 部署（推荐）

- 🎯 **标签页界面**：收件箱和发送邮件切换方便

```bash

# 1. 克隆项目- 💾 **内存存储**：邮件自动过期（默认 1 小时）

git clone https://github.com/Nei-Xin/temp_mail.git

cd temp_mail- 🐳 **Docker 支持**：一键部署，镜像仅 24.5MB



# 2. 修改配置（可选）- 🔧 **灵活配置**：支持 IP 或自定义域名，无需修改代码- 📧 **SMTP 邮件接收**：监听 25 端口，接收外部邮件

nano .env  # 修改 DOMAIN 为你的域名或 IP



# 3. 启动服务

docker-compose up -d---- 📤 **SMTP 邮件发送**：使用临时邮箱地址发送邮件，自动查找 MX 记录



# 4. 访问 Web 界面

http://your-server:8080

```## 🚀 快速开始- 🎯 **标签页界面**：收件箱和发送邮件切换方便## ✨ 特性## ✨ 特性



### Docker Run 部署



```bash### Docker Compose 部署（推荐）- 💾 **内存存储**：邮件自动过期（默认 1 小时）

docker pull neixin/temp-mail:v2.1



docker run -d \

  --name temp-mail \```bash- 🐳 **Docker 支持**：一键部署，镜像仅 24.5MB

  -p 8080:8080 \

  -p 25:25 \# 1. 克隆项目

  -e DOMAIN=your-domain.com \

  -e MESSAGE_TTL=1h \git clone https://github.com/Nei-Xin/temp_mail.git- 🔧 **灵活配置**：支持 IP 或自定义域名，无需修改代码

  neixin/temp-mail:v2.1

```cd temp_mail



### 本地编译运行- 📧 **SMTP 邮件接收**：监听 25 端口，接收外部邮件- 📧 SMTP 邮件接收（端口 25）



```bash# 2. 修改配置（可选）

# 1. 安装依赖

go mod tidynano .env  # 修改 DOMAIN 为你的域名或 IP---



# 2. 编译

go build -o temp-mail ./cmd/temp-mail

# 3. 启动服务- 📤 **SMTP 邮件发送**：使用临时邮箱地址发送邮件，自动查找 MX 记录- 📤 SMTP 邮件发送（支持 Gmail、163、QQ 等）

# 3. 运行

./temp-maildocker-compose up -d



# 4. 访问## 🚀 快速开始

http://localhost:8080

```# 4. 访问 Web 界面



---http://your-server:8080- 🎯 **标签页界面**：收件箱和发送邮件切换方便- 🌐 Web 界面（端口 8080）



## ⚙️ 配置```



通过 `.env` 文件或环境变量配置：### Docker Compose 部署（推荐）



| 变量 | 默认值 | 说明 |### Docker Run 部署

|------|--------|------|

| `HTTP_ADDR` | `:8080` | HTTP 服务监听地址 |- 💾 **内存存储**：邮件自动过期（默认 1 小时）- 💾 内存存储，自动过期（默认 30 分钟）

| `SMTP_ADDR` | `:25` | SMTP 接收服务监听地址 |

| `DOMAIN` | `localhost` | 邮件域名（支持 IP 地址） |```bash

| `MESSAGE_TTL` | `1h` | 邮件保留时间（支持 `s`/`m`/`h`） |

| `TZ` | `Asia/Shanghai` | 时区 |docker pull neixin/temp-mail:v2.1```bash



**配置文件示例（.env）：**



```bashdocker run -d \# 1. 克隆项目- 🐳 **Docker 支持**：一键部署，镜像仅 24.5MB- 🐳 Docker 支持，一键部署

HTTP_ADDR=:8080

SMTP_ADDR=:25  --name temp-mail \

DOMAIN=mail.example.com

MESSAGE_TTL=1h  -p 8080:8080 \git clone <your-repo>

TZ=Asia/Shanghai

```  -p 25:25 \



**MESSAGE_TTL 格式示例：**  -e DOMAIN=your-domain.com \cd temp_mail- 🔧 **灵活配置**：支持 IP 或自定义域名，无需修改代码- 🔧 环境变量配置，无需修改代码

- `30m` - 30 分钟

- `1h` - 1 小时  -e MESSAGE_TTL=1h \

- `2h` - 2 小时

- `180s` - 180 秒  neixin/temp-mail:v2.1



---```



## 📖 使用说明# 2. 修改配置（可选）



### 接收邮件### 本地编译运行



1. 访问 Web 界面（http://your-server:8080）nano .env  # 修改 DOMAIN 为你的域名或 IP

2. 输入自定义邮箱名称，点击"创建邮箱"

3. 复制生成的邮箱地址（如：test@your-domain.com）```bash

4. 使用该地址接收邮件

5. 邮件会自动显示在"📨 收件箱"标签页中# 1. 安装依赖------



### 发送邮件go mod tidy



1. 切换到"📤 发送邮件"标签页# 3. 启动服务

2. 填写邮件信息：

   - **发件人**：输入邮箱本地部分（如：test）# 2. 编译

   - **收件人**：目标邮箱地址（多个用逗号分隔）

   - **主题**：邮件主题go build -o temp-mail ./cmd/temp-maildocker-compose up -d

   - **内容**：邮件正文

3. 点击"📤 发送"按钮

4. 如果发件人邮箱不存在，系统会自动创建

# 3. 运行

**工作原理**：

- 使用临时邮箱地址作为发件人./temp-mail

- 自动查找收件人的 MX 记录

- 直接连接对方邮件服务器发送（支持 STARTTLS 加密）# 4. 访问 Web 界面## 🚀 快速开始## 🚀 快速开始

- 无需配置第三方 SMTP 服务

# 4. 访问

**⚠️ 发送限制说明**：

- ✅ **可以发送**：QQ、163、126、Sina 等国内邮箱（成功率 95%+）http://localhost:8080http://your-server:8080

- ❌ **无法发送**：Gmail、Outlook 等国际邮箱（被拒绝）

  - **原因**：这些服务商要求高信誉 IP，会拒绝云服务器 IP```

  - **错误提示**：`Gmail 拒绝，建议用 QQ/163`

- 💡 **建议**：主要用于测试和接收邮件，发送功能建议使用国内邮箱```



------



## 🌐 生产环境部署



### 端口要求## ⚙️ 配置



| 端口 | 用途 | 说明 |### Docker Run 部署### Docker Compose 部署（推荐）### 方式 1: Docker Compose（推荐）

|------|------|------|

| 25 | SMTP 邮件接收 | 接收外部邮件 |通过 `.env` 文件或环境变量配置：

| 8080 | HTTP Web 界面 | 查看和发送邮件 |



### DNS 配置（推荐）

| 变量 | 默认值 | 说明 |

如果使用自定义域名，建议配置以下 DNS 记录：

|------|--------|------|```bash

```

类型    名称                值                  优先级| `HTTP_ADDR` | `:8080` | HTTP 服务监听地址 |

A       mail.example.com    服务器IP            -

MX      example.com         mail.example.com    10| `SMTP_ADDR` | `:25` | SMTP 接收服务监听地址 |docker pull neixin/temp-mail:v2.0

TXT     example.com         v=spf1 ip4:服务器IP -all

```| `DOMAIN` | `localhost` | 邮件域名（支持 IP 地址） |



**说明**：| `MESSAGE_TTL` | `1h` | 邮件保留时间（支持 `s`/`m`/`h`，如 `30m`、`2h`） |```bash```bash

- **A 记录**：域名指向服务器 IP（必需）

- **MX 记录**：邮件路由记录（必需）| `TZ` | `Asia/Shanghai` | 时区 |

- **SPF 记录**：提高发送邮件可信度（推荐）

docker run -d \

### 防火墙配置

**配置文件示例（.env）：**

```bash

# 开放端口  --name temp-mail \# 1. 克隆项目# 1. 修改配置

sudo ufw allow 25/tcp

sudo ufw allow 8080/tcp```bash

sudo ufw enable

```HTTP_ADDR=:8080  -p 8080:8080 \



---SMTP_ADDR=:25



## 📡 API 接口DOMAIN=mail.example.com  -p 25:25 \git clone <your-repo>nano .env  # 修改 DOMAIN 为你的域名



### 创建邮箱地址MESSAGE_TTL=1h



```httpTZ=Asia/Shanghai  -e DOMAIN=101.44.160.108 \

POST /api/address?local=test

``````



响应：  -e MESSAGE_TTL=3600 \cd temp_mail

```json

{---

  "address": "test@example.com",

  "local": "test",  neixin/temp-mail:v2.0

  "ttl": 3600

}## 📖 使用说明

```

```# 2. 启动服务

### 获取邮件列表

### 接收邮件

```http

GET /api/messages/{local}

```

1. 访问 Web 界面（http://your-server:8080）

响应：

```json2. 输入自定义邮箱名称，点击"创建邮箱"### 本地编译运行# 2. 修改配置（可选）docker-compose up -d

[

  {3. 复制生成的邮箱地址（如：`test@your-domain.com`）

    "id": "abc123",

    "from": "sender@gmail.com",4. 使用该地址接收邮件

    "to": ["test@example.com"],

    "subject": "测试邮件",5. 邮件会自动显示在"📨 收件箱"标签页中

    "received": "2025-10-18T10:30:00Z"

  }```bashnano .env  # 修改 DOMAIN 为你的域名或 IP

]

```### 发送邮件



### 获取邮件详情# 1. 安装依赖



```http1. 切换到"📤 发送邮件"标签页

GET /api/messages/{local}/{id}

GET /api/messages/{local}/{id}?format=raw2. 填写以下信息：go mod tidy# 3. 访问 Web 界面

```

   - **发件人**：输入邮箱本地部分（如：`test`）

### 发送邮件

   - **收件人**：目标邮箱地址（多个用逗号分隔）

```http

POST /api/send   - **主题**：邮件主题

Content-Type: application/json

   - **内容**：邮件正文# 2. 编译# 3. 启动服务http://localhost:8080

{

  "from": "test",3. 点击"📤 发送"按钮

  "to": ["recipient@example.com"],

  "subject": "测试邮件",4. 如果发件人邮箱不存在，系统会自动创建go build -o temp-mail ./cmd/temp-mail

  "body": "邮件正文内容",

  "html": "<p>HTML内容</p>"

}

```**工作原理**：docker-compose up -d```



响应（成功）：- 使用您创建的临时邮箱地址作为发件人

```json

{- 系统自动查找收件人的 MX 记录# 3. 运行

  "success": true,

  "message": "邮件已发送",- 直接连接对方邮件服务器发送邮件（支持 STARTTLS 加密）

  "from": "test@example.com"

}- 无需配置第三方 SMTP 服务./temp-mail

```



响应（失败）：

```json**⚠️ 发送限制说明**：

{

  "error": "Gmail 拒绝，建议用 QQ/163"- ✅ **可以发送**：QQ.com、163.com、126.com、Sina.com.cn 等国内邮箱（成功率 95%+）

}

```- ❌ **无法发送**：Gmail、Outlook、Hotmail 等国际邮箱# 4. 访问# 4. 访问 Web 界面### 方式 2: Docker Run



---  - **原因**：这些服务商要求发件 IP 具有高信誉度，会拒绝云服务器 IP 的直接连接



## 🐳 Docker 镜像  - **错误提示**：`Gmail 拒绝，建议用 QQ/163` 或 `Outlook 拒绝，建议用 QQ/163`http://localhost:8080



### 镜像信息- 💡 **建议**：本服务主要用于测试和接收邮件，发送功能建议使用国内邮箱



- **镜像名称**：neixin/temp-mail```http://your-server:8080

- **最新版本**：v2.1

- **镜像大小**：~24.5MB---

- **基础镜像**：Alpine Linux



### 拉取镜像

## 🌐 生产环境部署

```bash

docker pull neixin/temp-mail:v2.1---``````bash

docker pull neixin/temp-mail:latest

```### 端口要求



### 构建镜像



```bash| 端口 | 用途 | 说明 |

# 构建

docker build -t neixin/temp-mail:v2.1 .|------|------|------|## ⚙️ 配置docker run -d \



# 运行| **25** | SMTP 邮件接收 | 接收外部邮件（Gmail、Outlook 等） |

docker run -d -p 8080:8080 -p 25:25 \

  -e DOMAIN=your-domain.com \| **8080** | HTTP Web 界面 | 查看和发送邮件 |

  neixin/temp-mail:v2.1

```



---### DNS 配置（推荐）通过 `.env` 文件或环境变量配置：### Docker Run 部署  --name temp-mail \



## 🔧 开发



### 项目结构如果使用自定义域名，建议配置以下 DNS 记录以提高邮件送达率：



```

temp_mail/

├── cmd/temp-mail/      # 主程序入口```| 变量 | 默认值 | 说明 |  -p 8080:8080 \

│   └── main.go

├── internal/类型    名称                    值                      优先级

│   ├── httpapi/        # HTTP 服务和 Web 界面

│   ├── smtpserver/     # SMTP 接收服务器A       mail.example.com        服务器IP                -|------|--------|------|

│   ├── smtpclient/     # SMTP 发送客户端（MX 查询 + STARTTLS）

│   └── storage/        # 内存存储MX      example.com             mail.example.com        10

├── Dockerfile          # Docker 镜像构建

├── docker-compose.yml  # Docker Compose 配置TXT     example.com             v=spf1 ip4:服务器IP -all   -| `HTTP_ADDR` | `:8080` | HTTP 服务监听地址 |```bash  -p 25:25 \

├── .env                # 环境变量配置

└── README.md           # 项目文档```

```

| `SMTP_ADDR` | `:25` | SMTP 接收服务监听地址 |

### 本地开发

**说明**：

```bash

# 安装依赖- **A 记录**：域名指向服务器 IP（必需）| `DOMAIN` | `localhost` | 邮件域名（支持 IP 地址） |docker pull neixin/temp-mail:v2.0  -e DOMAIN=mail.example.com \

go mod tidy

- **MX 记录**：邮件路由记录（必需）

# 运行

go run ./cmd/temp-mail- **SPF 记录**：提高发送邮件的可信度，减少被拒绝的概率（推荐）| `MESSAGE_TTL` | `3600` | 邮件保留时间（秒） |



# 测试- **PTR 记录**：反向解析，需要联系服务器提供商配置（推荐）

go test ./...

| `TZ` | `Asia/Shanghai` | 时区 |  neixin/temp-mail:latest

# 编译

go build -o temp-mail ./cmd/temp-mail### 防火墙配置

```



---

```bash

## 💡 注意事项

# 开放端口**配置文件示例（.env）：**docker run -d \```

1. **数据存储**：所有邮件存储在内存中，重启服务后数据会丢失

2. **端口 25**：监听 25 端口需要 root 权限或使用 Dockersudo ufw allow 25/tcp

3. **防火墙**：确保开放 25 和 8080 端口

4. **云服务商**：部分云服务商（阿里云、腾讯云）封禁 25 端口sudo ufw allow 8080/tcp

5. **发送限制**：Gmail/Outlook 等国际邮箱会拒绝云服务器 IP

sudo ufw enable

---

``````bash  --name temp-mail \

## 🔒 安全建议



- 使用 Nginx 反向代理，配置 SSL/TLS

- 限制访问来源（防火墙规则）---HTTP_ADDR=:8080

- 定期更新 Docker 镜像

- 监控日志，防止滥用

- 不要用于生产环境敏感数据

## 📡 API 接口SMTP_ADDR=:25  -p 8080:8080 \### 方式 3: 本地编译运行

---



## 📚 常见问题

### 创建邮箱地址DOMAIN=mail.example.com

### Q1: 无法接收外部邮件？



**检查**：

- DNS MX 记录是否正确```httpMESSAGE_TTL=3600  -p 25:25 \

- 防火墙是否开放 25 端口

- 云服务商是否封禁 25 端口（阿里云、腾讯云默认封禁）POST /api/address?local=test

- 查看日志：`docker logs temp-mail`

```TZ=Asia/Shanghai

### Q2: 为什么无法发送到 Gmail/Outlook？



这是正常现象，不是代码问题。

**响应**：```  -e DOMAIN=101.44.160.108 \```bash

**原因**：

- Gmail、Outlook 要求发件 IP 具有高信誉度```json

- 云服务器 IP 通常被归类为"低信誉 IP"

- 即使配置了完整 DNS 记录，仍可能被拒绝{



**错误示例**：  "address": "test@example.com",

```

550 5.7.1 The IP you're using to send mail is not authorized  "local": "test",---  -e MESSAGE_TTL=3600 \# 1. 编译

550 5.7.1 Client host blocked using Spamhaus

```  "ttl": 3600



**解决方案**：}

- ✅ 使用国内邮箱（QQ/163/126）进行测试

- ⏳ 配置 DKIM + 建立 IP 信誉（需要 2-4 周）```

- 💰 使用 SendGrid、Amazon SES 等专业服务

## 📖 使用说明  neixin/temp-mail:v2.0go mod tidy

### Q3: 如何使用自定义域名？

### 获取邮件列表

1. 修改 `.env` 文件中的 `DOMAIN`

2. 配置 DNS A 记录和 MX 记录

3. 重启服务：`docker-compose restart`

```http

### Q4: 邮件保留多久？

GET /api/messages/{local}### 接收邮件```go build ./cmd/temp-mail

默认 1 小时，可通过 `MESSAGE_TTL` 配置（如 `30m`、`2h`）。

```

### Q5: 发送邮件的成功率如何？



| 邮箱服务商 | 成功率 | 说明 |

|-----------|-------|------|**响应**：

| QQ、163、126 | ✅ 95%+ | 国内邮箱，稳定 |

| Gmail、Outlook | ❌ <5% | 严格的 IP 信誉要求 |```json1. 访问 Web 界面（http://your-server:8080）

| 企业邮箱 | 🔄 50-80% | 取决于具体配置 |

[

---

  {2. 输入自定义邮箱名称，点击"创建邮箱"

## 📄 许可证

    "id": "abc123",

MIT License

    "from": "sender@gmail.com",3. 复制生成的邮箱地址（如：test@your-domain.com）### 本地编译运行# 2. 运行

---

    "to": ["test@example.com"],

## 🙏 致谢

    "subject": "测试邮件",4. 使用该地址接收邮件

基于 [emersion/go-smtp](https://github.com/emersion/go-smtp) 构建

    "received": "2025-10-18T10:30:00Z"

---

  }5. 邮件会自动显示在"📨 收件箱"标签页中./temp-mail

## 🔗 链接

]

- **GitHub**: https://github.com/Nei-Xin/temp_mail

- **Docker Hub**: https://hub.docker.com/r/neixin/temp-mail```



---



**开始使用临时邮箱服务！** 🎉### 获取邮件详情### 发送邮件```bash




```http

GET /api/messages/{local}/{id}

GET /api/messages/{local}/{id}?format=raw  # 获取原始邮件1. 切换到"📤 发送邮件"标签页# 1. 安装依赖# 3. 访问

```

2. 填写以下信息：

### 发送邮件

   - **发件人**：输入邮箱本地部分（如：test）go mod tidyhttp://localhost:8080

```http

POST /api/send   - **收件人**：目标邮箱地址（多个用逗号分隔）

Content-Type: application/json

   - **主题**：邮件主题```

{

  "from": "test",   - **内容**：邮件正文

  "to": ["recipient@example.com"],

  "subject": "测试邮件",3. 点击"📤 发送"按钮# 2. 编译

  "body": "邮件正文内容",

  "html": "<p>HTML内容</p>"4. 如果发件人邮箱不存在，系统会自动创建

}

```go build -o temp-mail ./cmd/temp-mail---



**响应（成功）**：**工作原理**：

```json

{- 使用您创建的临时邮箱地址作为发件人

  "success": true,

  "message": "邮件已发送",- 系统自动查找收件人的 MX 记录

  "from": "test@example.com"

}- 直接连接对方邮件服务器发送邮件# 3. 运行## ⚙️ 配置

```

- 无需配置第三方 SMTP 服务

**响应（失败）**：

```json./temp-mail

{

  "error": "Gmail 拒绝，建议用 QQ/163"**⚠️ 发送限制说明**：

}

```- ✅ **可以发送**：QQ.com、163.com、126.com 等国内邮箱通过环境变量配置，无需修改代码：



---- ⚠️ **可能失败**：Gmail、Hotmail、Outlook 等国际邮箱（它们有严格的反垃圾邮件策略，会拒绝未认证 IP 的连接）



## 🐳 Docker 镜像- 💡 **提高成功率**：配置 DNS（SPF、PTR 记录）、使用有信誉的服务器 IP# 4. 访问



### 镜像信息



- **镜像名称**：`neixin/temp-mail`---http://localhost:8080### 基础配置

- **最新版本**：`v2.1`

- **镜像大小**：~24.5MB

- **基础镜像**：Alpine Linux

## 🌐 生产环境部署```

### 拉取镜像



```bash

docker pull neixin/temp-mail:v2.1### 端口要求| 变量 | 默认值 | 说明 |

docker pull neixin/temp-mail:latest

```



### 构建镜像| 端口 | 用途 | 说明 |---|------|--------|------|



```bash|------|------|------|

# 构建

docker build -t neixin/temp-mail:v2.1 .| **25** | SMTP 邮件接收 | 接收外部邮件（Gmail、Outlook 等） || `HTTP_ADDR` | `:8080` | HTTP 服务监听地址 |



# 运行| **8080** | HTTP Web 界面 | 查看和发送邮件 |

docker run -d -p 8080:8080 -p 25:25 \

  -e DOMAIN=your-domain.com \## ⚙️ 配置| `SMTP_ADDR` | `:25` | SMTP 接收服务监听地址 |

  neixin/temp-mail:v2.1

```### DNS 配置（可选）



---| `DOMAIN` | `tmp.local` | 邮件域名 |



## 🔧 开发如果使用自定义域名，建议配置以下 DNS 记录以提高邮件送达率：



### 项目结构通过 `.env` 文件或环境变量配置：| `MESSAGE_TTL` | `30m` | 邮件保留时间 |



``````

temp_mail/

├── cmd/temp-mail/      # 主程序入口类型    名称                    值                      优先级| `TZ` | `Asia/Shanghai` | 时区 |

│   └── main.go

├── internal/A       mail.example.com        服务器IP                -

│   ├── httpapi/        # HTTP 服务和 Web 界面

│   ├── smtpserver/     # SMTP 接收服务器MX      example.com             mail.example.com        10| 变量 | 默认值 | 说明 |

│   ├── smtpclient/     # SMTP 发送客户端（MX 查询 + STARTTLS）

│   └── storage/        # 内存存储TXT     example.com             v=spf1 ip4:服务器IP -all   -

├── Dockerfile          # Docker 镜像构建

├── docker-compose.yml  # Docker Compose 配置```|------|--------|------|**配置文件示例（.env）：**

├── .env                # 环境变量配置

└── README.md           # 项目文档

```

**说明**：| `HTTP_ADDR` | `:8080` | HTTP 服务监听地址 |

### 本地开发

- **A 记录**：域名指向服务器 IP

```bash

# 安装依赖- **MX 记录**：邮件路由记录| `SMTP_ADDR` | `:25` | SMTP 接收服务监听地址 |```bash

go mod tidy

- **SPF 记录**：提高发送邮件的可信度，减少被拒绝的概率

# 运行

go run ./cmd/temp-mail- **PTR 记录**：反向解析，需要联系服务器提供商配置| `DOMAIN` | `localhost` | 邮件域名（支持 IP 地址） |# 基础配置



# 测试

go test ./...

### 防火墙配置| `MESSAGE_TTL` | `3600` | 邮件保留时间（秒） |HTTP_ADDR=:8080

# 编译

go build -o temp-mail ./cmd/temp-mail

```

```bash| `TZ` | `Asia/Shanghai` | 时区 |SMTP_ADDR=:25

---

# 开放端口

## 💡 注意事项

sudo ufw allow 25/tcpDOMAIN=mail.example.com

1. **数据存储**：所有邮件存储在内存中，重启服务后数据会丢失

2. **端口 25**：监听 25 端口需要 root 权限或使用 Dockersudo ufw allow 8080/tcp

3. **防火墙**：确保开放 25 和 8080 端口

4. **云服务商**：部分云服务商（阿里云、腾讯云）封禁 25 端口，无法接收外部邮件sudo ufw enable**配置文件示例（.env）：**MESSAGE_TTL=30m

5. **DNS 配置**：使用 IP 地址也可以正常工作，DNS 仅用于提高送达率

6. **发送限制**：Gmail/Outlook 等国际邮箱会拒绝云服务器 IP 的直接连接```



---TZ=Asia/Shanghai



## 🔒 安全建议---



- 使用 Nginx 反向代理，配置 SSL/TLS```bash```

- 限制访问来源（防火墙规则）

- 定期更新 Docker 镜像## 📡 API 接口

- 监控日志，防止滥用

- 不要用于生产环境敏感数据HTTP_ADDR=:8080



---### 创建邮箱地址



## 📚 常见问题SMTP_ADDR=:25---



### Q1: 无法接收外部邮件？```http



**检查**：POST /api/address?local=testDOMAIN=mail.example.com

- DNS MX 记录是否正确

- 防火墙是否开放 25 端口```

- 云服务商是否封禁 25 端口（阿里云、腾讯云默认封禁）

- SMTP 服务是否正常运行：`docker logs temp-mail`MESSAGE_TTL=3600## 🌐 生产环境部署



### Q2: 为什么无法发送到 Gmail/Outlook？**响应**：



这是正常现象，不是代码问题。```jsonTZ=Asia/Shanghai



**原因**：{

- Gmail、Outlook 等大型邮件服务商要求发件 IP 具有高信誉度

- 云服务器 IP（如华为云、阿里云）通常被归类为"低信誉 IP"  "address": "test@example.com",```### 端口要求

- 即使配置了完整的 DNS 记录（SPF、PTR、DKIM），仍可能被拒绝

  "local": "test",

**错误示例**：

```  "ttl": 3600

550 5.7.1 The IP you're using to send mail is not authorized

550 5.7.1 Client host blocked using Spamhaus}

```

```---| 端口 | 用途 | 说明 |

**解决方案**：

- ✅ **接受现实**：使用国内邮箱（QQ/163/126）进行测试

- ⏳ **长期方案**：配置 DKIM + 建立 IP 信誉（需要 2-4 周）

- 💰 **生产环境**：使用 SendGrid、Amazon SES 等专业邮件服务### 获取邮件列表|------|------|------|



### Q3: 如何使用自定义域名？



1. 修改 `.env` 文件中的 `DOMAIN````http## 📖 使用说明| **25** | SMTP 邮件接收 | 接收外部邮件（Gmail、Outlook 等） |

2. 配置 DNS A 记录和 MX 记录

3. 重启服务：`docker-compose restart`GET /api/messages/{local}



### Q4: 邮件保留多久？```| **8080** | HTTP Web 界面 | 查看邮件 |



默认 1 小时，可通过 `MESSAGE_TTL` 配置（支持 `30m`、`2h`、`24h` 等）。



### Q5: 如何配置 HTTPS？**响应**：### 接收邮件



使用 Nginx 反向代理：```json



```nginx[### DNS 配置

server {

    listen 443 ssl http2;  {

    server_name mail.example.com;

        "id": "abc123",1. 访问 Web 界面（http://your-server:8080）

    ssl_certificate /path/to/cert.pem;

    ssl_certificate_key /path/to/key.pem;    "from": "sender@gmail.com",

    

    location / {    "to": ["test@example.com"],2. 输入自定义邮箱名称，点击"创建邮箱"添加以下 DNS 记录：

        proxy_pass http://localhost:8080;

        proxy_set_header Host $host;    "subject": "测试邮件",

        proxy_set_header X-Real-IP $remote_addr;

    }    "received": "2025-10-18T10:30:00Z"3. 复制生成的邮箱地址（如：test@your-domain.com）

}

```  }



### Q6: 发送邮件的成功率如何？]4. 使用该地址接收邮件```



| 邮件服务商 | 成功率 | 说明 |```

|-----------|-------|------|

| QQ.com, 163.com, 126.com | ✅ 95%+ | 配置良好的国内邮箱 |5. 邮件会自动显示在"📨 收件箱"标签页中类型    名称                    值                  优先级

| Sina.com.cn, Sohu.com | ✅ 90%+ | 国内邮箱，较稳定 |

| Gmail, Outlook, Hotmail | ❌ <5% | 严格的 IP 信誉要求 |### 获取邮件详情

| 企业邮箱 | 🔄 50-80% | 取决于具体配置 |

A       mail.example.com        服务器IP            -

---

```http

## 📄 许可证

GET /api/messages/{local}/{id}### 发送邮件MX      example.com             mail.example.com    10

MIT License

GET /api/messages/{local}/{id}?format=raw  # 获取原始邮件

---

``````

## 🙏 致谢



基于 [emersion/go-smtp](https://github.com/emersion/go-smtp) 构建

### 发送邮件1. 切换到"📤 发送邮件"标签页

---



## 🔗 链接

```http2. 填写以下信息：### 部署步骤

- **GitHub**: https://github.com/Nei-Xin/temp_mail

- **Docker Hub**: https://hub.docker.com/r/neixin/temp-mailPOST /api/send

- **问题反馈**: https://github.com/Nei-Xin/temp_mail/issues

Content-Type: application/json   - **发件人**：输入邮箱本地部分（如：test）

---



**开始使用临时邮箱服务！** 🎉

{   - **收件人**：目标邮箱地址（多个用逗号分隔）```bash

  "from": "test",

  "to": ["recipient@example.com"],   - **主题**：邮件主题# 1. 克隆项目

  "subject": "测试邮件",

  "body": "邮件正文内容",   - **内容**：邮件正文git clone <your-repo>

  "html": "<p>HTML内容</p>"

}3. 点击"📤 发送"按钮cd temp_mail

```

4. 如果发件人邮箱不存在，系统会自动创建

**响应（成功）**：

```json# 2. 修改配置

{

  "success": true,**工作原理**：cp .env .env.local

  "message": "邮件已发送",

  "from": "test@example.com"- 使用您创建的临时邮箱地址作为发件人nano .env.local  # 修改 DOMAIN

}

```- 系统自动查找收件人的 MX 记录



---- 直接连接对方邮件服务器发送邮件# 3. 启动服务



## 🐳 Docker 镜像- 无需配置第三方 SMTP 服务docker-compose up -d



### 镜像信息



- **镜像名称**：`neixin/temp-mail`---# 4. 查看日志

- **最新版本**：`v2.0`

- **镜像大小**：24.5MBdocker-compose logs -f

- **基础镜像**：Alpine Linux

## 🌐 生产环境部署

### 拉取镜像

# 5. 测试

```bash

docker pull neixin/temp-mail:v2.0### 端口要求# 发送邮件到 test@example.com

docker pull neixin/temp-mail:latest

```# 访问 http://mail.example.com:8080



### 构建镜像| 端口 | 用途 | 说明 |```



```bash|------|------|------|

# 构建

docker build -t neixin/temp-mail:v2.0 .| **25** | SMTP 邮件接收 | 接收外部邮件（Gmail、Outlook 等） |---



# 运行| **8080** | HTTP Web 界面 | 查看和发送邮件 |

docker run -d -p 8080:8080 -p 25:25 \

  -e DOMAIN=your-domain.com \## � 使用说明

  neixin/temp-mail:v2.0

```### DNS 配置（可选）



---### 接收邮件



## 🔧 开发如果使用自定义域名，建议配置以下 DNS 记录以提高邮件送达率：



### 项目结构1. 访问 Web 界面（http://your-server:8080）



``````2. 输入自定义邮箱名称（可选），点击"创建邮箱"

temp_mail/

├── cmd/temp-mail/      # 主程序入口类型    名称                    值                  优先级3. 复制生成的邮箱地址

│   └── main.go

├── internal/A       mail.example.com        服务器IP            -4. 使用该地址接收邮件

│   ├── httpapi/        # HTTP 服务和 Web 界面

│   ├── smtpserver/     # SMTP 接收服务器MX      example.com             mail.example.com    105. 邮件会自动显示在收件箱中

│   ├── smtpclient/     # SMTP 发送客户端（MX 查询）

│   └── storage/        # 内存存储TXT     example.com             v=spf1 ip4:服务器IP -all

├── Dockerfile          # Docker 镜像构建

├── docker-compose.yml  # Docker Compose 配置```### 发送邮件

├── .env                # 环境变量配置

└── README.md           # 项目文档

```

**说明**：1. **首先创建一个临时邮箱**（如果还没有）

### 本地开发

- **A 记录**：域名指向服务器 IP2. 在 Web 界面中，找到"📤 发送邮件"卡片

```bash

# 安装依赖- **MX 记录**：邮件路由记录3. 点击"✍️ 写邮件"按钮展开表单

go mod tidy

- **SPF 记录**：提高发送邮件的可信度（可选）4. 填写以下信息：

# 运行

go run ./cmd/temp-mail   - **发件人**：自动使用您创建的邮箱地址（不可修改）



# 测试### 防火墙配置   - **收件人**：目标邮箱地址（必填，多个收件人用逗号分隔）

go test ./...

   - **主题**：邮件主题（必填）

# 编译

go build -o temp-mail ./cmd/temp-mail```bash   - **内容**：邮件正文（必填）

```

# 开放端口5. 点击"📤 发送"按钮发送邮件

---

sudo ufw allow 25/tcp

## 💡 注意事项

sudo ufw allow 8080/tcp**工作原理**：

1. **数据存储**：所有邮件存储在内存中，重启服务后数据会丢失

2. **端口 25**：监听 25 端口需要 root 权限或使用 Dockersudo ufw enable- 使用您创建的临时邮箱地址作为发件人

3. **防火墙**：确保开放 25 和 8080 端口

4. **云服务商**：部分云服务商（阿里云、腾讯云）封禁 25 端口，无法接收外部邮件```- 系统通过查找收件人的 MX 记录，直接发送邮件到对方邮件服务器

5. **DNS 配置**：使用 IP 地址也可以正常工作，DNS 仅用于提高送达率

6. **发送限制**：Gmail/Hotmail 等大型邮箱可能拒绝连接，建议先测试国内邮箱- 无需配置第三方 SMTP 服务



------



## 🔒 安全建议---



- 使用 Nginx 反向代理，配置 SSL/TLS## 📡 API 接口

- 限制访问来源（防火墙规则）

- 定期更新 Docker 镜像## �📡 API

- 监控日志，防止滥用

- 不要用于生产环境敏感数据### 创建邮箱地址



---### 创建邮箱地址



## 📚 常见问题```http



### Q1: 无法接收外部邮件？POST /api/address?local=test```bash



**检查**：```POST /api/address?local=test

- DNS MX 记录是否正确

- 防火墙是否开放 25 端口

- 云服务商是否封禁 25 端口（阿里云、腾讯云默认封禁）

- SMTP 服务是否正常运行：`docker logs temp-mail`**响应**：返回:



### Q2: 发送邮件失败？```json{



**常见原因及解决方案**：{  "address": "test@example.com",



#### 1. Gmail/Hotmail/Outlook 拒绝连接（最常见）  "address": "test@example.com",  "local": "test",



**现象**：  "local": "test",  "ttl": 30

```

发送失败: 无法连接到 gmail.com 的任何MX服务器  "ttl": 3600}

```

}```

**原因**：

- Gmail、Hotmail、Outlook 等大型邮件服务商有严格的反垃圾邮件策略```

- 会直接拒绝来自未认证 IP 的 SMTP 连接

- 需要 IP 具有良好的发信信誉### 获取邮件列表



**解决方案**：### 获取邮件列表

- ✅ **先测试国内邮箱**：QQ.com、163.com、126.com（成功率高）

- ✅ **配置 PTR 记录**：联系服务器提供商配置反向 DNS 解析```bash

- ✅ **配置 SPF 记录**：在域名 DNS 中添加 `v=spf1 ip4:你的IP -all`

- ✅ **配置 DKIM/DMARC**：进一步提高邮件信誉```httpGET /api/messages/{local}

- ✅ **使用有信誉的服务器**：避免使用廉价 VPS 的共享 IP

- ⏳ **建立 IP 信誉**：Gmail 通常需要较长时间（数周到数月）才能接受新 IPGET /api/messages/{local}



#### 2. 其他常见问题```返回:



- **收件人邮箱地址错误**：检查邮箱格式[

- **MX 记录查询失败**：网络或 DNS 问题

- **IP 被列入黑名单**：使用 [MXToolbox](https://mxtoolbox.com/blacklists.aspx) 检查**响应**：  {

- **对方服务器临时故障**：稍后重试

```json    "id": "abc123",

**测试建议**：

```bash[    "from": "sender@gmail.com",

# 1. 先测试国内邮箱（成功率高）

curl -X POST http://localhost:8080/api/send -H "Content-Type: application/json" -d '{  {    "to": ["test@example.com"],

  "from": "test",

  "to": ["your@qq.com"],    "id": "abc123",    "subject": "测试邮件",

  "subject": "测试邮件",

  "body": "这是一封测试邮件"    "from": "sender@gmail.com",    "received": "2025-10-17T10:30:00Z"

}'

    "to": ["test@example.com"],  }

# 2. 查看详细日志

docker logs -f temp-mail    "subject": "测试邮件",]

```

    "received": "2025-10-18T10:30:00Z"```

### Q3: 如何使用自定义域名？

  }

1. 修改 `.env` 文件中的 `DOMAIN`

2. 配置 DNS A 记录和 MX 记录]### 获取邮件详情

3. 重启服务：`docker-compose restart`

```

### Q4: 邮件保留多久？

```bash

默认 1 小时（3600 秒），可通过 `MESSAGE_TTL` 配置。

### 获取邮件详情GET /api/messages/{local}/{id}

### Q5: 如何配置 HTTPS？



使用 Nginx 反向代理：

```http# 获取原始邮件

```nginx

server {GET /api/messages/{local}/{id}GET /api/messages/{local}/{id}?format=raw

    listen 443 ssl http2;

    server_name mail.example.com;GET /api/messages/{local}/{id}?format=raw  # 获取原始邮件```

    

    ssl_certificate /path/to/cert.pem;```

    ssl_certificate_key /path/to/key.pem;

    ### 发送邮件

    location / {

        proxy_pass http://localhost:8080;### 发送邮件

        proxy_set_header Host $host;

        proxy_set_header X-Real-IP $remote_addr;```bash

    }

}```httpPOST /api/send

```

POST /api/send

### Q6: 为什么可以发送到 QQ 但不能发送到 Gmail？

Content-Type: application/json请求体:

这是正常现象：

{

| 邮件服务商 | 发送成功率 | 说明 |

|-----------|----------|------|{  "from": "test",                    // 必填，您创建的邮箱本地部分

| QQ.com, 163.com, 126.com | ✅ 高 | 反垃圾策略相对宽松 |

| Gmail, Hotmail, Outlook | ⚠️ 低 | 需要 IP 信誉、PTR、SPF、DKIM 等配置 |  "from": "test",  "to": ["recipient@example.com"],   // 必填，收件人列表

| 企业邮箱 | 🔄 中等 | 取决于具体配置 |

  "to": ["recipient@example.com"],  "subject": "测试邮件",             // 必填

**建议**：

- 用于测试和开发：使用国内邮箱即可  "subject": "测试邮件",  "body": "邮件正文内容",            // 必填

- 用于生产环境：需要完整配置 DNS 记录并建立 IP 信誉

- 发送重要邮件：建议使用专业邮件服务（SendGrid、Amazon SES 等）  "body": "邮件正文内容",  "html": "<p>HTML内容</p>"          // 可选，HTML格式邮件



---  "html": "<p>HTML内容</p>"}



## 📄 许可证}



MIT License```返回（成功）:



---{



## 🙏 致谢**响应（成功）**：  "success": true,



基于 [emersion/go-smtp](https://github.com/emersion/go-smtp) 构建```json  "message": "邮件已发送",



---{  "from": "test@mail.example.com"



**开始使用临时邮箱服务！** 🎉  "success": true,}


  "message": "邮件已发送",

  "from": "test@example.com"返回（失败）:

}{

```  "error": "发件人邮箱 test@mail.example.com 不存在，请先创建邮箱"

}

---```



## 🐳 Docker 镜像---



### 镜像信息## 🐳 Docker 镜像



- **镜像名称**：`neixin/temp-mail`### 拉取镜像

- **最新版本**：`v2.0`

- **镜像大小**：24.5MB```bash

- **基础镜像**：Alpine Linuxdocker pull neixin/temp-mail:latest

```

### 拉取镜像

### 构建镜像

```bash

docker pull neixin/temp-mail:v2.0```bash

docker pull neixin/temp-mail:latest# 构建

```docker build -t neixin/temp-mail:latest .



### 构建镜像# 推送到 Docker Hub

docker push neixin/temp-mail:latest

```bash```

# 构建

docker build -t neixin/temp-mail:v2.0 .---



# 运行## 🔧 开发

docker run -d -p 8080:8080 -p 25:25 \

  -e DOMAIN=your-domain.com \### 项目结构

  neixin/temp-mail:v2.0

``````

temp_mail/

---├── cmd/temp-mail/      # 主程序入口

├── internal/

## 🔧 开发│   ├── httpapi/        # HTTP 服务和 Web 界面

│   ├── smtpserver/     # SMTP 接收服务器

### 项目结构│   ├── smtpclient/     # SMTP 发送客户端

│   └── storage/        # 内存存储

```├── Dockerfile          # Docker 镜像构建

temp_mail/├── docker-compose.yml  # Docker Compose 配置

├── cmd/temp-mail/      # 主程序入口└── .env                # 环境变量配置

│   └── main.go```

├── internal/

│   ├── httpapi/        # HTTP 服务和 Web 界面### 本地开发

│   ├── smtpserver/     # SMTP 接收服务器

│   ├── smtpclient/     # SMTP 发送客户端（MX 查询）```bash

│   └── storage/        # 内存存储# 安装依赖

├── Dockerfile          # Docker 镜像构建go mod tidy

├── docker-compose.yml  # Docker Compose 配置

├── .env                # 环境变量配置# 运行

└── README.md           # 项目文档go run ./cmd/temp-mail

```

# 测试

### 本地开发go test ./...

```

```bash

# 安装依赖---

go mod tidy

## � 注意事项

# 运行

go run ./cmd/temp-mail1. **数据存储**：所有邮件存储在内存中，重启服务后数据会丢失

2. **端口 25**：监听 25 端口需要 root 权限或 Docker

# 测试3. **防火墙**：确保开放 25 和 8080 端口

go test ./...4. **DNS 配置**：接收外部邮件需要正确配置 MX 记录

5. **安全性**：建议配置 Nginx 反向代理和 SSL 证书

# 编译

go build -o temp-mail ./cmd/temp-mail---

```

## 🔒 安全建议

---

- 使用 Nginx 反向代理，配置 SSL/TLS

## 💡 注意事项- 限制 8080 端口的访问（仅内网或特定 IP）

- 定期更新 Docker 镜像

1. **数据存储**：所有邮件存储在内存中，重启服务后数据会丢失- 监控日志，防止滥用

2. **端口 25**：监听 25 端口需要 root 权限或使用 Docker

3. **防火墙**：确保开放 25 和 8080 端口---

4. **云服务商**：部分云服务商（阿里云、腾讯云）封禁 25 端口，无法接收外部邮件

5. **DNS 配置**：使用 IP 地址也可以正常工作，DNS 仅用于提高送达率## 📚 常见问题



---### Q1: 无法接收外部邮件？



## 🔒 安全建议检查：

- DNS MX 记录是否正确

- 使用 Nginx 反向代理，配置 SSL/TLS- 防火墙是否开放 25 端口

- 限制访问来源（防火墙规则）- 云服务商是否封禁 25 端口（Aliyun、Tencent 默认封禁）

- 定期更新 Docker 镜像

- 监控日志，防止滥用### Q2: 如何修改域名？

- 不要用于生产环境敏感数据

修改 `.env` 文件中的 `DOMAIN`，重启服务即可，无需重新编译。

---

### Q3: 如何配置 SSL？

## 📚 常见问题

使用 Nginx 反向代理：

### Q1: 无法接收外部邮件？

```nginx

**检查**：server {

- DNS MX 记录是否正确    listen 443 ssl http2;

- 防火墙是否开放 25 端口    server_name mail.example.com;

- 云服务商是否封禁 25 端口（阿里云、腾讯云默认封禁）    

- SMTP 服务是否正常运行：`docker logs temp-mail`    ssl_certificate /path/to/cert.pem;

    ssl_certificate_key /path/to/key.pem;

### Q2: 发送邮件失败？    

    location / {

**可能原因**：        proxy_pass http://localhost:8080;

- 收件人邮箱地址错误    }

- 对方邮件服务器 MX 记录查询失败}

- 对方服务器拒绝接收（IP 信誉问题）```

- 检查日志：`docker logs temp-mail`

---

### Q3: 如何使用自定义域名？

## 📄 许可证

1. 修改 `.env` 文件中的 `DOMAIN`

2. 配置 DNS A 记录和 MX 记录MIT License

3. 重启服务：`docker-compose restart`

---

### Q4: 邮件保留多久？

## 🙏 致谢

默认 1 小时（3600 秒），可通过 `MESSAGE_TTL` 配置。

基于 [emersion/go-smtp](https://github.com/emersion/go-smtp) 构建

### Q5: 如何配置 HTTPS？

---

使用 Nginx 反向代理：

**开始使用临时邮箱服务！** 🎉

```nginx| **2525** | SMTP 邮件接收 | ✅ 必需 | 普通用户 |

server {| **80** | HTTP（通过 Nginx） | ⚪ 可选 | 需要 root |

    listen 443 ssl http2;| **443** | HTTPS（通过 Nginx） | ⚪ 可选 | 需要 root |

    server_name mail.example.com;| **25** | SMTP 标准端口 | ⚪ 可选 | 需要 root |

    

    ssl_certificate /path/to/cert.pem;**✅ 25端口封禁？没问题！**

    ssl_certificate_key /path/to/key.pem;- 本项目默认使用 **2525 端口**，不依赖 25 端口

    - 可以在**阿里云、腾讯云、AWS** 等封禁 25 端口的服务器上正常运行

    location / {- 适用于内部测试、开发环境、团队协作

        proxy_pass http://localhost:8080;- 详细说明请查看 [25端口封禁环境部署指南](PORT25_BLOCKED.md)

        proxy_set_header Host $host;

        proxy_set_header X-Real-IP $remote_addr;**⚠️ 需要接收 Gmail/Outlook 等外部邮件？**

    }- 外部邮件服务器**只使用 25 端口**发送邮件（无法更改）

}- 必须部署在支持 25 端口的服务器上

```- **无需修改代码**：通过端口转发即可（25 → 2525）

- 推荐服务商：Vultr ($5/月)、DigitalOcean ($6/月)

---- 详细说明请查看 [接收外部邮件部署指南](EXTERNAL_EMAIL.md)

- 无需修改代码的配置方法：[NO_CODE_CHANGE.md](NO_CODE_CHANGE.md)

## 📄 许可证

**说明**：

MIT License- 默认使用 **8080** 和 **2525** 端口，无需 root 权限

- 生产环境建议使用 Nginx 反向代理实现 80/443 端口访问

---- 如需接收外部邮件服务器的邮件，需要 25 端口（或使用支持25端口的云服务商）



## 🙏 致谢详细端口配置方案请参考 [PORT_REQUIREMENTS.md](PORT_REQUIREMENTS.md)



基于 [emersion/go-smtp](https://github.com/emersion/go-smtp) 构建### 快速部署（推荐）



---1. **本地编译 Linux 版本**



**开始使用临时邮箱服务！** 🎉```powershell

# Windows PowerShell
$env:GOOS="linux"
$env:GOARCH="amd64"
go build -o temp-mail-linux ./cmd/temp-mail
```

2. **上传到服务器**

```bash
scp temp-mail-linux deploy.sh user@your-server:/opt/temp-mail/
```

3. **服务器上部署**

```bash
cd /opt/temp-mail
sudo bash deploy.sh
```

### Docker 部署

```bash
# 使用 Docker Compose
docker-compose up -d

# 或使用 Docker
docker build -t temp-mail .
docker run -d -p 8080:8080 -p 2525:2525 temp-mail
```

### 配置说明

- **环境变量**：编辑 `/opt/temp-mail/.env`
- **域名配置**：设置 `DOMAIN` 为你的域名（无需修改代码）
- **端口配置**：修改 `HTTP_ADDR` 和 `SMTP_ADDR`
- **DNS 设置**：添加 MX 记录指向你的服务器

#### 自定义域名示例

```bash
# 编辑配置文件
sudo nano /opt/temp-mail/.env

# 修改域名
DOMAIN=mail.yourdomain.com
HTTP_ADDR=:8080
SMTP_ADDR=:2525
MESSAGE_TTL=30

# 重启服务
sudo systemctl restart temp-mail
```

详细域名配置教程：[DOMAIN_CONFIG.md](DOMAIN_CONFIG.md)

更多详情请参考 [DEPLOYMENT.md](DEPLOYMENT.md)
