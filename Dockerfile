# ===== Build stage =====
FROM golang:1.23-alpine AS builder

WORKDIR /app

# 丢弃 go sum/cach 可以加快
COPY go.mod go.sum ./
RUN go mod download

COPY . .

# 如果需要静态编译可加 -ldflags, 并指定输出文件
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-s -w" -o /xvexitong ./main.go

# ===== Final stage =====
FROM alpine:latest

WORKDIR /app

# 如果不需要暴露端口，不加 EXPOSE
# RUN apk add --no-cache tzdata && ln -sf /usr/share/zoneinfo/Asia/Shanghai /etc/localtime

COPY --from=builder /xvexitong /app/xvexitong

ENTRYPOINT ["/app/xvexitong"]
