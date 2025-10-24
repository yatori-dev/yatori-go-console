# ---------- Build stage ----------
FROM golang:1.23-bookworm AS builder

WORKDIR /app

# 安装编译依赖
RUN apt-get update && apt-get install -y --no-install-recommends \
    gcc libc6-dev libasound2-dev \
    && rm -rf /var/lib/apt/lists/*

# 复制 go.mod 和 go.sum，并下载依赖
COPY go.mod go.sum ./
RUN go mod download

# 复制所有源码
COPY . .

# ✅ 同步和清理依赖，避免 go.mod 不一致问题
RUN go mod tidy

# ✅ 启用 CGO 编译，并生成 Linux 可执行文件
RUN CGO_ENABLED=1 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-s -w" -o /xvexitong ./main.go


# ---------- Runtime stage ----------
FROM debian:bookworm-slim

WORKDIR /app

# ✅ 安装运行依赖 + 时区支持
RUN apt-get update && apt-get install -y --no-install-recommends \
    libasound2 tzdata \
    && rm -rf /var/lib/apt/lists/*

# ✅ 设置时区为北京时间
ENV TZ=Asia/Shanghai
RUN ln -snf /usr/share/zoneinfo/$TZ /etc/localtime && echo $TZ > /etc/timezone

# ✅ 拷贝可执行文件到系统路径
COPY --from=builder /xvexitong /usr/local/bin/xvexitong

# ✅ 容器启动命令
ENTRYPOINT ["/usr/local/bin/xvexitong"]
