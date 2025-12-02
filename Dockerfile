# ---------- Build stage ----------
FROM golang:1.23-bookworm AS builder

WORKDIR /app

# 安装 GoCV 编译依赖（开发环境）---------------------------------------------
RUN apt-get update && apt-get install -y --no-install-recommends \
    gcc g++ make cmake pkg-config git \
    libjpeg-dev libpng-dev libtiff-dev \
    libgtk2.0-dev libgtk-3-dev \
    libavcodec-dev libavformat-dev libswscale-dev libavutil-dev \
    libeigen3-dev libtbb-dev \
    && rm -rf /var/lib/apt/lists/*

# 安装 OpenCV（GoCV 必须）-----------------------------------------------------
RUN apt-get update && apt-get install -y --no-install-recommends \
    libopencv-dev \
    && rm -rf /var/lib/apt/lists/*

# 复制 go.mod 和 go.sum，并下载依赖
COPY go.mod go.sum ./
RUN go mod download

# 复制所有源码
COPY . .

# 清理依赖
RUN go mod tidy

# 架构参数（buildx 自动注入）
ARG TARGETOS
ARG TARGETARCH

# GoCV 需要 CGO，需要链接系统 OpenCV
RUN CGO_ENABLED=1 GOOS=${TARGETOS} GOARCH=${TARGETARCH} \
    go build -ldflags="-s -w" -o /xvexitong ./main.go


# ---------- Runtime stage ----------
FROM debian:bookworm-slim

WORKDIR /app

# 安装 GoCV 运行必需依赖 ---------------------------------------
RUN apt-get update && apt-get install -y --no-install-recommends \
    libopencv-core406 libopencv-imgproc406 libopencv-imgcodecs406 \
    libasound2 tzdata \
    && rm -rf /var/lib/apt/lists/*

# 设置时区为北京时间
ENV TZ=Asia/Shanghai
RUN ln -snf /usr/share/zoneinfo/$TZ /etc/localtime && echo $TZ > /etc/timezone

# 复制可执行文件
COPY --from=builder /xvexitong /usr/local/bin/xvexitong

# 容器启动命令
ENTRYPOINT ["/usr/local/bin/xvexitong"]
