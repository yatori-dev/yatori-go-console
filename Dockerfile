# ---------- Build stage ----------
FROM golang:1.23-bookworm AS builder

WORKDIR /app

# 安装编译依赖
RUN apt-get update && apt-get install -y --no-install-recommends \
    gcc libc6-dev libasound2-dev \
    && rm -rf /var/lib/apt/lists/*

COPY go.mod go.sum ./
RUN go mod download
COPY . .

RUN CGO_ENABLED=1 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-s -w" -o /xvexitong ./main.go

# ---------- Runtime stage ----------
FROM debian:bookworm-slim

WORKDIR /app

# 安装运行依赖与时区支持 ✅
RUN apt-get update && apt-get install -y --no-install-recommends \
    libasound2 tzdata \
    && rm -rf /var/lib/apt/lists/*

# 设置时区为北京时间（Asia/Shanghai） ✅
ENV TZ=Asia/Shanghai
RUN ln -snf /usr/share/zoneinfo/$TZ /etc/localtime && echo $TZ > /etc/timezone

# 拷贝二进制到不被挂载覆盖的位置
COPY --from=builder /xvexitong /usr/local/bin/xvexitong

# 入口命令：启动程序
ENTRYPOINT ["/usr/local/bin/xvexitong"]
