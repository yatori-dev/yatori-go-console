# ---------- Build stage ----------
FROM golang:1.23-buster AS builder

WORKDIR /app

# 安装系统依赖（支持 cgo、OTO、ONNX Runtime 等库）
RUN apt-get update && apt-get install -y --no-install-recommends \
    gcc \
    libc6-dev \
    libasound2-dev \
    && rm -rf /var/lib/apt/lists/*

# 拷贝模块定义并下载依赖
COPY go.mod go.sum ./
RUN go mod download

# 拷贝源代码
COPY . .

# 编译可执行文件（开启 CGO 支持）
RUN CGO_ENABLED=1 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-s -w" -o /xvexitong ./main.go

# ---------- Runtime stage ----------
FROM debian:buster-slim

WORKDIR /app

# 安装运行时系统库（如果需要音频或其他依赖）
RUN apt-get update && apt-get install -y --no-install-recommends \
    libasound2 \
    && rm -rf /var/lib/apt/lists/*

# 拷贝可执行文件
COPY --from=builder /xvexitong /app/xvexitong

# 入口命令
ENTRYPOINT ["/app/xvexitong"]
