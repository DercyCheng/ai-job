FROM golang:1.21-alpine AS builder

# 接收服务参数
ARG SERVICE=gateway

WORKDIR /app

# 安装依赖
RUN apk add --no-cache git

# 复制依赖文件
COPY go.mod go.sum ./

# 下载依赖
RUN go mod download

# 复制源代码
COPY . .

# 构建指定的服务
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/main ./cmd/${SERVICE}/main.go

# 使用轻量级的基础镜像
FROM alpine:latest

# 接收服务参数
ARG SERVICE=gateway

WORKDIR /app

# 从builder阶段复制构建好的二进制文件
COPY --from=builder /app/main /app/main

# 复制配置文件
COPY configs /app/configs

# 创建日志目录
RUN mkdir -p /app/logs

# 暴露相应的端口
EXPOSE 8080
EXPOSE 8081
EXPOSE 8082

# 运行应用
CMD ["/app/main"]
