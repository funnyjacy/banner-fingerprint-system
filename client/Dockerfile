# 多阶段构建：编译阶段
FROM golang:1.21-alpine AS builder

WORKDIR /build

# 复制 go.mod 并下载依赖
COPY go.mod ./
RUN go mod download

# 复制源代码
COPY . .

# 编译二进制文件，静态链接
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o client main.go

# 生产阶段：使用最小化镜像
FROM alpine:3.19

# 创建非 root 用户
RUN addgroup -g 1000 appuser && \
    adduser -D -u 1000 -G appuser appuser

WORKDIR /app

# 从构建阶段复制二进制文件
COPY --from=builder --chown=appuser:appuser /build/client /app/client

# 创建数据目录
RUN mkdir -p /data && chown appuser:appuser /data

# 切换到非 root 用户
USER appuser

# 启动客户端
CMD ["/app/client"]
