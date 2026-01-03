# Dockerfile
FROM golang:1.24-alpine AS builder

# 设置工作目录
WORKDIR /app

# 复制go mod和sum文件
COPY go.mod go.sum ./

# 下载依赖
RUN go mod download
RUN go mod verify

# 复制源代码
COPY . .

# 构建应用（优化构建命令，移除-a标志以利用缓存）
RUN CGO_ENABLED=0 GOOS=linux go build -installsuffix cgo -ldflags="-s -w" -o main .

# 使用更小的基础镜像
FROM alpine:latest

# 创建工作目录
WORKDIR /root/

# 从builder阶段复制编译好的二进制文件
COPY --from=builder /app/main .

# 复制模板和静态文件
COPY ./templates ./templates
COPY ./static ./static

# 暴露端口
EXPOSE 8080

# 运行应用
CMD ["./main"]
