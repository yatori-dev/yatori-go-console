# 第1阶段：构建阶段
FROM golang:1.23-alpine AS builder

# 设置工作目录
WORKDIR /app

# 拷贝 go.mod 和 go.sum 先行下载依赖
COPY go.mod go.sum ./
RUN go mod download

# 拷贝全部代码
COPY . .

# 编译
RUN go build -o /xvexitong .

# 第2阶段：运行阶段
FROM alpine:latest

# 设置时区（可选）
RUN apk add --no-cache tzdata && \
    ln -sf /usr/share/zoneinfo/Asia/Shanghai /etc/localtime

WORKDIR /app

# 从构建阶段拷贝编译好的二进制文件
COPY --from=builder /xvexitong /app/xvexitong

# 暴露端口（如程序监听 8080，可自行修改）
EXPOSE 8080

# 入口命令
ENTRYPOINT ["/app/xvexitong"]
