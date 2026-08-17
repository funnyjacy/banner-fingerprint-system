# 多阶段构建：编译阶段
FROM golang:1.21-alpine AS builder

WORKDIR /build

# 配置 Go 代理
ENV GOPROXY=https://goproxy.cn,direct
ENV GOSUMDB=off

# 复制 go.mod 并下载依赖
COPY go.mod ./
RUN go mod download

# 复制源代码
COPY . .

# 编译二进制文件，静态链接
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o server main.go

# 生产阶段：使用最小化镜像
FROM alpine:3.19

# 创建非 root 用户
RUN addgroup -g 1000 appuser && \
    adduser -D -u 1000 -G appuser appuser

WORKDIR /app

# 从构建阶段复制二进制文件
COPY --from=builder --chown=appuser:appuser /build/server /app/server

# 切换到非 root 用户
USER appuser

# 暴露端口
EXPOSE 8080

# 健康检查
HEALTHCHECK --interval=10s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# 启动服务
CMD ["/app/server"]
