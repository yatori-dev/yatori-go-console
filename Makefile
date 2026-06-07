.PHONY: all build-all build-linux build-linux-arm64 build-windows frontend-build gotool clean help

BINARY_NAME="yvexitong"
RELEASE_DIR="release"

all: frontend-build build-all

build-all: gotool
	mkdir -p $(RELEASE_DIR)
	make build-linux
	make build-linux-arm64
	make build-windows

build-linux:
	CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o $(RELEASE_DIR)/$(BINARY_NAME)-linux-amd64 ./main.go

build-linux-arm64:
	CGO_ENABLED=1 GOOS=linux GOARCH=arm64 go build -ldflags="-s -w" -o $(RELEASE_DIR)/$(BINARY_NAME)-linux-arm64 ./main.go

build-windows:
	CGO_ENABLED=1 GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o $(RELEASE_DIR)/$(BINARY_NAME)-windows-amd64.exe ./main.go

frontend-build:
	@echo "Building frontend..."
	cd frontend && pnpm install && pnpm run build
	mkdir -p assets/web
	rm -rf assets/web/*
	cp -r frontend/out/* assets/web/

run:
	@go run ./

gotool:
	go fmt ./
	go vet ./

clean:
	@if [ -d $(RELEASE_DIR) ] ; then rm -rf $(RELEASE_DIR)/* ; fi
	@if [ -d assets/web ] ; then rm -rf assets/web/* ; fi

help:
	@echo "make - 构建前端并编译所有平台的二进制文件"
	@echo "make build-all - 编译 Go 代码，生成所有平台的二进制文件"
	@echo "make build-linux - 编译 Linux AMD64 二进制文件"
	@echo "make build-linux-arm64 - 编译 Linux ARM64 二进制文件 (当前环境支持)"
	@echo "make build-windows - 编译 Windows AMD64 二进制文件"
	@echo "make frontend-build - 编译前端并输出到 assets/web"
	@echo "make run - 直接运行 Go 代码"
	@echo "make clean - 移除二进制文件和构建产物"
	@echo "make gotool - 运行 Go 工具 'fmt' and 'vet'"
