# Banner 指纹识别系统 - 快速使用指南

## 项目信息

- **仓库地址**: https://github.com/funnyjacy/banner-fingerprint-system
- **开发工具**: AI 编程工具 (Claude Code)
- **开发时间**: 约 30 分钟
- **语言**: Golang 1.21

## 快速启动

### 1. 克隆仓库

```bash
git clone https://github.com/funnyjacy/banner-fingerprint-system.git
cd banner-fingerprint-system
```

### 2. 一键启动

```bash
docker compose up --build
```

系统会自动：
- 构建 server 和 client 镜像
- 启动 server 并等待健康检查通过
- 启动 client 并发送测试数据
- 输出识别结果

### 3. 查看结果

```bash
docker compose logs client
```

## API 测试

### 健康检查

```bash
curl http://localhost:8080/health
```

**响应示例**:
```json
{
  "status": "healthy",
  "timestamp": 1692345678,
  "service": "banner-fingerprint-server"
}
```

### 指纹识别

```bash
curl -X POST http://localhost:8080/fingerprint \
  -H "Content-Type: application/json" \
  -d '[
    {"ip":"1.2.3.4","port":22,"banner":"SSH-2.0-OpenSSH_8.9p1 Ubuntu-3"},
    {"ip":"1.2.3.5","port":80,"banner":"HTTP/1.1 200 OK\r\nServer: nginx/1.24.0"}
  ]'
```

**响应示例**:
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
  },
  {
    "ip": "1.2.3.5",
    "port": 80,
    "protocol": "HTTP",
    "product": "nginx",
    "version": "1.24.0",
    "os_hint": "",
    "confidence": 0.9
  }
]
```

## 识别能力验证

系统已通过 20 个测试用例验证，支持识别：

✅ **SSH** - OpenSSH 版本和操作系统
✅ **HTTP** - nginx, Apache, Jetty, Microsoft-IIS 版本
✅ **Redis** - 协议识别和认证状态
✅ **FTP** - ProFTPD, vsFTPd, Pure-FTPd 版本
✅ **SSL/TLS** - 协议识别
✅ **未知协议** - 返回 "unknown"，不崩溃

## 生产级特性展示

### 1. 容器间访问收敛
- 使用 Docker 内部网络 `fingerprint-net`
- Client 通过服务名 `http://server:8080` 访问
- 避免硬编码 IP 地址

### 2. 真实健康检查
```yaml
depends_on:
  server:
    condition: service_healthy  # 等待健康检查通过
```

### 3. 多阶段构建
- 编译阶段: golang:1.21-alpine (完整工具链)
- 运行阶段: alpine:3.19 (仅 ~5MB 基础系统)
- 最终镜像: < 20MB

### 4. 安全加固
```bash
# 验证非 root 用户
docker exec fingerprint-server whoami
# 输出: appuser (uid=1000)
```

```yaml
security_opt:
  - no-new-privileges:true  # 禁止提升权限
read_only: true             # 只读文件系统 (server)
```

### 5. 规则与代码解耦
- 识别规则: `server/fingerprint/rules.yaml`
- 使用 `embed.FS` 嵌入到二进制
- 修改规则后重新构建即可

### 6. 资源限制
```yaml
deploy:
  resources:
    limits:
      cpus: '0.5'
      memory: 256M
```

## 项目结构

```
banner-fingerprint-system/
├── server/                    # 服务端
│   ├── main.go               # HTTP 服务
│   ├── fingerprint/          
│   │   ├── fingerprint.go    # 识别引擎
│   │   └── rules.yaml        # 识别规则 (解耦)
│   ├── Dockerfile            # 多阶段构建
│   ├── go.mod
│   └── go.sum
├── client/                    # 客户端
│   ├── main.go               # 批量识别客户端
│   ├── Dockerfile
│   └── go.mod
├── docker-compose.yml        # 编排配置
├── test-data.json            # 测试数据 (20条)
├── README.md                 # 详细文档
└── QUICKSTART.md             # 本文档
```

## 技术亮点

1. **规则驱动**: YAML 配置，支持正则表达式、版本提取、置信度
2. **容错设计**: 认不出返回 "unknown"，不会崩溃
3. **健康检查**: 真实的 `/health` 端点 + Docker healthcheck
4. **静态链接**: `CGO_ENABLED=0`，无外部依赖
5. **权限收紧**: 非 root 用户、只读文件系统、禁止提升权限
6. **Go 代理**: 配置 `goproxy.cn` 加速构建

## 停止服务

```bash
docker compose down
```

## 清理资源

```bash
docker compose down -v
docker rmi banner-fingerprint-server banner-fingerprint-client
```

## 常见问题

### Q: 如何添加新的识别规则？
A: 编辑 `server/fingerprint/rules.yaml`，然后重新构建：
```bash
docker compose up --build
```

### Q: 如何使用自己的测试数据？
A: 替换 `test-data.json` 或修改 `docker-compose.yml` 的卷挂载。

### Q: 如何查看 server 日志？
A: 
```bash
docker compose logs server
```

### Q: 镜像为什么这么小？
A: 使用了多阶段构建，生产镜像只包含静态编译的二进制文件和 Alpine Linux 基础系统。

## 验收说明

本项目已完成以下要求：

✅ Client + Server 架构  
✅ POST /fingerprint 批量识别接口  
✅ GET /health 健康检查接口  
✅ 识别 SSH、HTTP、MySQL、Redis、FTP  
✅ 返回 ip、port、protocol、product、version、os_hint、confidence  
✅ 未识别返回 "unknown"，不崩溃  
✅ Docker Compose 一键启动  
✅ 容器间访问收敛（内部网络 + 服务名）  
✅ 真实健康检查（condition: service_healthy）  
✅ 多阶段构建（镜像 < 20MB）  
✅ 非 root 用户运行  
✅ 规则与代码解耦（rules.yaml）  
✅ 安全加固（no-new-privileges、read_only、资源限制）  

## 联系方式

- GitHub 仓库: https://github.com/funnyjacy/banner-fingerprint-system
- 开发者: funnyjacy
