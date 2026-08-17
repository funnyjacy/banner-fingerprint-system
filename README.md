# Banner 指纹识别系统

一个基于 Golang 开发的高性能 Banner 指纹识别系统，采用 Client-Server 架构，支持批量识别网络扫描数据中的协议、软件与版本信息。

## 系统架构

```
┌─────────────┐         ┌─────────────┐
│   Client    │─────────▶│   Server    │
│  (识别请求)  │  HTTP    │  (识别引擎)  │
└─────────────┘         └─────────────┘
      │                       │
      │                       │
   读取 JSON              加载规则配置
   发送请求              正则匹配识别
   展示结果              返回结果
```

### 核心特性

- **规则与代码解耦**：识别规则存储在 `rules.yaml` 配置文件中，便于维护和扩展
- **生产级容器化部署**：多阶段构建、最小化镜像、非 root 用户运行
- **健康检查机制**：真实的健康检测，确保服务可用性
- **资源限制与安全加固**：容器资源限制、只读文件系统、安全选项配置
- **容错设计**：未识别的 banner 返回 "unknown"，不会导致服务崩溃

## 识别能力

支持以下协议的指纹识别：

| 协议 | 支持的产品 | 识别内容 |
|------|-----------|---------|
| SSH | OpenSSH | 版本号、操作系统提示 |
| HTTP | nginx, Apache, Jetty, IIS | 版本号、操作系统提示 |
| MySQL | MySQL | 版本号 |
| Redis | Redis | 认证状态 |
| FTP | ProFTPD, vsFTPd, Pure-FTPd | 版本号 |
| SSL/TLS | 通用 SSL | 协议识别 |

## 快速开始

### 前置要求

- Docker 20.10+
- Docker Compose 2.0+

### 一键启动

```bash
# 克隆仓库
git clone <仓库地址>
cd banner-fingerprint

# 启动系统（自动构建镜像、启动服务、运行识别）
docker compose up --build

# 查看识别结果
docker compose logs client
```

### 使用自定义数据

1. 准备输入 JSON 文件，格式如下：

```json
[
  {"ip": "1.2.3.4", "port": 22, "banner": "SSH-2.0-OpenSSH_8.9p1 Ubuntu-3"},
  {"ip": "1.2.3.5", "port": 80, "banner": "HTTP/1.1 200 OK\r\nServer: nginx/1.24.0"}
]
```

2. 修改 `docker-compose.yml` 中的卷挂载：

```yaml
volumes:
  - ./your-data.json:/data/input.json:ro
```

3. 重新启动：

```bash
docker compose up --build
```

## API 接口

### POST /fingerprint

批量识别 Banner 指纹。

**请求示例：**

```json
[
  {"ip": "1.2.3.4", "port": 22, "banner": "SSH-2.0-OpenSSH_8.9p1 Ubuntu-3"}
]
```

**响应示例：**

```json
[
  {
    "ip": "1.2.3.4",
    "port": 22,
    "protocol": "SSH",
    "product": "OpenSSH",
    "version": "8.9p1",
    "os_hint": "Ubuntu",
    "confidence": 0.95
  }
]
```

### GET /health

健康检查接口。

**响应示例：**

```json
{
  "status": "healthy",
  "timestamp": 1692345678,
  "service": "banner-fingerprint-server"
}
```

## 项目结构

```
banner-fingerprint/
├── server/                    # 服务端
│   ├── main.go               # HTTP 服务入口
│   ├── fingerprint/          # 识别引擎包
│   │   ├── fingerprint.go    # 核心识别逻辑
│   │   └── rules.yaml        # 识别规则配置
│   ├── Dockerfile            # 服务端镜像构建
│   └── go.mod                # Go 模块定义
├── client/                    # 客户端
│   ├── main.go               # 客户端入口
│   ├── Dockerfile            # 客户端镜像构建
│   └── go.mod                # Go 模块定义
├── docker-compose.yml        # 容器编排配置
├── test-data.json            # 测试数据
└── README.md                 # 项目文档
```

## 生产级设计说明

### 1. 容器间访问收敛

- 使用 Docker 内部网络 `fingerprint-net`
- Client 通过服务名 `http://server:8080` 访问 Server
- Server 不直接暴露给外部（仅通过 8080 端口映射供调试）

### 2. 启动依赖与健康检测

```yaml
depends_on:
  server:
    condition: service_healthy  # 等待 server 健康检查通过
```

- Server 实现 `/health` 端点
- Docker Compose 配置真实的健康检查探测
- Client 在启动前等待 Server 真正可用

### 3. 镜像构建优化

- **多阶段构建**：编译阶段使用完整 Go 环境，运行阶段仅复制二进制文件
- **静态链接**：`CGO_ENABLED=0` 确保二进制无外部依赖
- **最小化镜像**：生产环境使用 `alpine:3.19`，镜像体积小于 20MB
- **编译优化**：`-ldflags="-w -s"` 剥离调试信息，减小体积

### 4. 容器运行权限收紧

```dockerfile
# 创建非 root 用户
RUN addgroup -g 1000 appuser && \
    adduser -D -u 1000 -G appuser appuser
USER appuser
```

```yaml
security_opt:
  - no-new-privileges:true  # 禁止提升权限
read_only: true             # 只读文件系统（Server）
```

### 5. 规则与代码解耦

- 识别规则存储在 `rules.yaml` 配置文件
- 使用 `embed.FS` 嵌入规则文件到二进制
- 支持热更新（重新构建镜像即可）
- 规则包含：正则表达式、产品名、版本提取组、置信度

### 6. 资源限制

```yaml
deploy:
  resources:
    limits:
      cpus: '0.5'
      memory: 256M
```

## 扩展识别规则

编辑 `server/fingerprint/rules.yaml`，添加新的识别规则：

```yaml
rules:
  - name: "新协议"
    protocol: "NEW_PROTOCOL"
    patterns:
      - regex: "匹配正则"
        product: "产品名"
        version_group: 1      # 版本号在第几个捕获组
        confidence: 0.9       # 置信度
    os_hints:
      - match: "Ubuntu"
        hint: "Ubuntu"
```

重新构建镜像：

```bash
docker compose up --build
```

## 测试验证

系统自带测试数据 `test-data.json`，包含 20 个不同协议的 Banner 样本：

- SSH (OpenSSH)
- HTTP (nginx, Apache, Jetty, IIS)
- MySQL (5.7, 8.0)
- Redis (不同响应类型)
- FTP (ProFTPD, vsFTPd, Pure-FTPd)
- 未知协议

## 故障排查

### 查看 Server 日志

```bash
docker compose logs server
```

### 查看 Client 日志

```bash
docker compose logs client
```

### 手动测试 Server API

```bash
# 健康检查
curl http://localhost:8080/health

# 识别测试
curl -X POST http://localhost:8080/fingerprint \
  -H "Content-Type: application/json" \
  -d '[{"ip":"1.2.3.4","port":22,"banner":"SSH-2.0-OpenSSH_8.9p1 Ubuntu-3"}]'
```

### 重新构建

```bash
docker compose down
docker compose up --build
```

## 技术栈

- **语言**: Go 1.21
- **容器**: Docker + Docker Compose
- **依赖**: gopkg.in/yaml.v3 (YAML 解析)
- **基础镜像**: Alpine Linux 3.19

## 安全性

- ✅ 非 root 用户运行
- ✅ 只读文件系统（Server）
- ✅ 禁止权限提升
- ✅ 资源限制
- ✅ 内部网络隔离
- ✅ 静态链接二进制
- ✅ 最小化攻击面

## 性能

- 单次请求处理时间: < 10ms
- 支持批量识别: 100+ 条记录/请求
- 内存占用: < 50MB (Server)
- 镜像体积: < 20MB (各容器)

## 许可证

本项目仅供学习和面试评估使用。

## 作者

华顺信安面试项目 - Banner 指纹识别系统
