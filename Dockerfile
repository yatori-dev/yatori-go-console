# ============================
# 第1阶段：构建 Go 可执行文件
# ============================
FROM golang:1.23-alpine AS builder

WORKDIR /app

# 拷贝依赖文件并下载模块
COPY go.mod go.sum ./
RUN go mod download

# 拷贝项目全部源代码
COPY . .

# 编译（静态链接，减小依赖）
RUN go build -ldflags="-s -w" -o /xvexitong .

# ============================
# 第2阶段：精简运行镜像
# ============================
FROM alpine:latest

WORKDIR /app

# 设置时区（可选）
RUN apk add --no-cache tzdata && \
    ln -sf /usr/share/zoneinfo/Asia/Shanghai /etc/localtime

# 拷贝可执行文件
COPY --from=builder /xvexitong /app/xvexitong

# 入口命令（无端口暴露）
ENTRYPOINT ["/app/xvexitong"]
