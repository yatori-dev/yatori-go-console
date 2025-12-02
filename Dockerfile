# ---------- Build stage ----------
FROM golang:1.23-bookworm AS builder

WORKDIR /app

# 安装 GoCV 编译依赖
RUN apt-get update && apt-get install -y --no-install-recommends \
    gcc g++ make cmake pkg-config git \
    libjpeg-dev libpng-dev libtiff-dev \
    libgtk2.0-dev libgtk-3-dev \
    libavcodec-dev libavformat-dev libswscale-dev libavutil-dev libswresample-dev \
    libeigen3-dev libtbb-dev \
    && rm -rf /var/lib/apt/lists/*

# 安装 OpenCV (GoCV 构建需要)
RUN apt-get update && apt-get install -y --no-install-recommends \
    libopencv-dev \
    && rm -rf /var/lib/apt/lists/*

# 复制依赖
COPY go.mod go.sum ./
RUN go mod download

# 复制源码
COPY . .
RUN go mod tidy

ARG TARGETOS
ARG TARGETARCH

# 🔥 ---- 关键：禁用 Aruco (否则必定编译失败) ----
RUN CGO_CPPFLAGS="-DGOCV_DISABLE_ARUCO" \
    CGO_ENABLED=1 GOOS=${TARGETOS} GOARCH=${TARGETARCH} \
    go build -ldflags="-s -w" -o /xvexitong ./main.go


# ---------- Runtime stage ----------
FROM debian:bookworm-slim

WORKDIR /app

RUN apt-get update && apt-get install -y --no-install-recommends \
    libopencv-core406 \
    libopencv-imgproc406 \
    libopencv-imgcodecs406 \
    libasound2 tzdata \
    && rm -rf /var/lib/apt/lists/*

ENV TZ=Asia/Shanghai
RUN ln -snf /usr/share/zoneinfo/$TZ /etc/localtime && echo $TZ > /etc/timezone

COPY --from=builder /xvexitong /usr/local/bin/xvexitong

ENTRYPOINT ["/usr/local/bin/xvexitong"]
