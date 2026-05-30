# ---------- Frontend build stage ----------
FROM node:22-bookworm-slim AS frontend-builder
WORKDIR /frontend

# 安装 pnpm
RUN npm install -g pnpm

# 复制依赖定义
COPY frontend/package.json frontend/pnpm-lock.yaml* ./

# 安装依赖
RUN pnpm install --frozen-lockfile

# 复制源码并构建
COPY frontend/ ./
RUN pnpm run build

# ---------- Backend build stage ----------
FROM golang:1.24-bookworm AS backend-builder
WORKDIR /app

# 安装编译依赖 (音频等需要 CGO)
RUN apt-get update && apt-get install -y --no-install-recommends \
    gcc libc6-dev libasound2-dev pkg-config \
    && rm -rf /var/lib/apt/lists/*

# 复制 go.mod 和 go.sum，并下载依赖
COPY go.mod go.sum ./
RUN go mod download

# 复制所有源码
COPY . .
RUN go mod tidy

# 声明目标平台参数
ARG TARGETOS
ARG TARGETARCH

# 编译后端
RUN CGO_ENABLED=1 GOOS=${TARGETOS} GOARCH=${TARGETARCH} \
    go build -ldflags="-s -w" -o /xvexitong ./main.go


# ---------- Runtime stage ----------
FROM debian:bookworm-slim
WORKDIR /app

# 安装运行依赖 + 时区支持
RUN apt-get update && apt-get install -y --no-install-recommends \
    libasound2 tzdata ca-certificates \
    && rm -rf /var/lib/apt/lists/*

# 设置时区为北京时间
ENV TZ=Asia/Shanghai
RUN ln -snf /usr/share/zoneinfo/$TZ /etc/localtime && echo $TZ > /etc/timezone

# 拷贝后端二进制
COPY --from=backend-builder /xvexitong /usr/local/bin/xvexitong

# 拷贝静态资源目录 (包含 DLL/SO 和 其他 assets)
COPY --from=backend-builder /app/assets /app/assets

# 拷贝前端构建产物到 assets/web (后端 ServerInit.go 期望该路径)
COPY --from=frontend-builder /frontend/out /app/assets/web

# 设置动态库路径，以便程序运行时能找到 assets 目录下的 .so 文件
ENV LD_LIBRARY_PATH=/app/assets:$LD_LIBRARY_PATH

# 容器启动命令
ENTRYPOINT ["/usr/local/bin/xvexitong"]
