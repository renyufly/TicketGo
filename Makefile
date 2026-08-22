# 统一管理 Go 后端和 Web 前端的开发命令，
# 并同时兼容 Windows 和 Linux/macOS
# Makefile 本身基本不负责真正的业务逻辑，
# 它最大的意义就是把原本很长的命令包装成：make xxx
# 让开发者、本地环境和 CI 都使用同一套标准命令

# 定义两个变量, 选择使用哪个 Go
LOCAL_GO_WIN := .tools/go/bin/go.exe
LOCAL_GO_UNIX := .tools/go/bin/go

# 如果项目根目录存在 .env 文件，就加载它里面的环境变量
ifneq (,$(wildcard .env))
include .env
export
endif

# 自动选择 Go, 如果都不存在，就使用系统安装的 Go
ifneq ($(wildcard $(LOCAL_GO_WIN)),)
GO ?= $(LOCAL_GO_WIN)
else ifneq ($(wildcard $(LOCAL_GO_UNIX)),)
GO ?= $(LOCAL_GO_UNIX)
else
GO ?= go
endif
# 主要是防止路径中出现空格导致 shell 解析错误
GO_CMD := "$(GO)"

# Windows 和 Linux/macOS 分别设置命令
# 设置两个 Go 缓存目录
# Windows 编译出来的服务器：bin/ticketgo.exe
ifeq ($(OS),Windows_NT)
GO_ENV := set "GOMODCACHE=$(CURDIR)/.cache/go-mod" && set "GOCACHE=$(CURDIR)/.cache/go-build" &&
MKDIR_BIN := if not exist bin mkdir bin
CHECK_NAME := if "$(NAME)"=="" (echo NAME is required & exit /b 1)
SERVER_BIN := bin/ticketgo.exe
else
GO_ENV := GOMODCACHE=$(CURDIR)/.cache/go-mod GOCACHE=$(CURDIR)/.cache/go-build
MKDIR_BIN := mkdir -p bin
CHECK_NAME := test -n "$(NAME)" || (echo "NAME is required" && exit 1)
SERVER_BIN := bin/ticketgo
endif

# 定义数据库 migration 命令：golang-migrate 用于执行数据库迁移 SQL
MIGRATE := $(GO_ENV) $(GO_CMD) run -tags postgres github.com/golang-migrate/migrate/v4/cmd/migrate@v4.18.3
# 默认 PostgreSQL. 如果 .env 已经定义：DATABASE_URL=...就不会使用这里的默认值
DATABASE_URL ?= postgres://ticketgo:ticketgo_local_password@localhost:5432/ticketgo?sslmode=disable

# 告诉 Make：这些名字是“命令”，不是文件
.PHONY: help bootstrap-go deps format lint test build run compose-up compose-down migrate-up migrate-down migrate-create web-install web-dev web-check web-e2e check

# make help 打印可用命令
help:
	@echo "bootstrap-go | deps | format | lint | test | build | run | compose-up | compose-down | migrate-up | migrate-down | migrate-create NAME=name | web-install | web-dev | web-check | web-e2e | check"

# make bootstrap-go 给项目自动安装/初始化本地 Go 环境
bootstrap-go:
	powershell -NoProfile -ExecutionPolicy Bypass -File scripts/bootstrap-go.ps1

# 下载 go.mod 中的 Go 依赖
deps:
	$(GO_ENV) $(GO_CMD) mod download

# 相当于 go fmt ./... ：格式化整个 Go 项目
format:
	$(GO_ENV) $(GO_CMD) fmt ./...

# 静态检查代码
# 注意：这里叫 lint，但实际执行的是：go vet 并不是 golangci-lint
lint:
	$(GO_ENV) $(GO_CMD) vet ./...

# 运行整个项目测试：go test ./...
test:
	$(GO_ENV) $(GO_CMD) test ./...

# 把：cmd/server 编译成服务器程序
build:
	@$(MKDIR_BIN)
	$(GO_ENV) $(GO_CMD) build -o $(SERVER_BIN) ./cmd/server

# make run：启动后端
# 注意：make run → 编译 + 立即运行
#       make build → 只生成可执行文件
run:
	$(GO_ENV) $(GO_CMD) run ./cmd/server

# 后台启动 Docker 服务 （例如 PostgreSQL、Redis 等）
compose-up:
	docker compose up -d

# 停止并删除 Docker Compose 创建的容器
compose-down:
	docker compose down

# 执行所有还没有运行的 migration
migrate-up:
	$(MIGRATE) -path migrations -database "$(DATABASE_URL)" up

# 回滚最近的一次 migration
migrate-down:
	$(MIGRATE) -path migrations -database "$(DATABASE_URL)" down 1

# 创建 migration
migrate-create:
	@$(CHECK_NAME)
	$(MIGRATE) create -ext sql -dir migrations -seq "$(NAME)"

# 前端命令：安装 npm 依赖。
# npm ci 一般用于根据：package-lock.json 进行严格、可重复的依赖安装
web-install:
	cd web && npm ci

# 启动前端开发服务器
web-dev:
	cd web && npm run dev

# 执行前端项目定义的："check": "..."
# 通常可能包含 TypeScript、ESLint 等检查
web-check:
	cd web && npm run check

# 运行前端 E2E 测试。比如 Playwright
web-e2e:
	cd web && npm run test:e2e

# 没有自己写命令，而是声明依赖：format lint test build web-check
# make check 即依次执行：make format make lint ... make web-check
# 一个项目整体质量检查命令
check: format lint test build web-check
