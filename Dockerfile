# ---------- Build stage ----------
FROM golang:1.23-bookworm AS builder

WORKDIR /app

# 安装 GoCV 编译依赖（开发环境）
RUN apt-get update && apt-get install -y --no-install-recommends \
    gcc g++ make cmake pkg-config git \
    libjpeg-dev libpng-dev libtiff-dev \
    libgtk2.0-dev libgtk-3-dev \
    libavcodec-dev libavformat-dev libswscale-dev libavutil-dev libswresample-dev \
    libeigen3-dev libtbb-dev \
    && rm -rf /var/lib/apt/lists/*

# 安装 OpenCV (GoCV 编译需要头文件 + pkgconfig)
RUN apt-get update && apt-get install -y --no-install-recommends \
    libopencv-dev \
    && rm -rf /var/lib/apt/lists/*

# 复制 go.mod 和 go.sum
COPY go.mod go.sum ./
RUN go mod download

# 复制所有源码
COPY . .

# 清理依赖
RUN go mod tidy

# 构建架构（由 buildx 注入）
ARG TARGETOS
ARG TARGETARCH

# -------------------------------------------------------------------------------------------------
# 🔥 最关键的地方：禁用 GoCV Aruco（修复 aruco.cpp: cv::aruco 未声明编译失败问题）
# -------------------------------------------------------------------------------------------------
RUN CGO_CPPFLAGS="-DGOCV_DISABLE_ARUCO" \
    CGO_ENABLED=1 GOOS=${TARGETOS} GOARCH=${TARGETARCH} \
    go build -ldflags="-s -w" -o /xvexitong ./main.go


# ---------- Runtime stage ----------
FROM debian:bookworm-slim

WORKDIR /app

# 安装 GoCV 必需的运行库（不用安装 OpenCV 开发库）
RUN apt-get update && apt-get install -y --no-install-recommends \
    libopencv-core406 \
    libopencv-imgproc406 \
    libopencv-imgcodecs406 \
    libasound2 tzdata \
    && rm -rf /var/lib/apt/lists/*

# 设置时区为北京时间
ENV TZ=Asia/Shanghai
RUN ln -snf /usr/share/zoneinfo/$TZ /etc/localtime && echo $TZ > /etc/timezone

# 拷贝构建产物
COPY --from=builder /xvexitong /usr/local/bin/xvexitong

# 容器启动命令
ENTRYPOINT ["/usr/local/bin/xvexitong"]
