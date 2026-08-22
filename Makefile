LOCAL_GO_WIN := .tools/go/bin/go.exe
LOCAL_GO_UNIX := .tools/go/bin/go

ifneq (,$(wildcard .env))
include .env
export
endif

ifneq ($(wildcard $(LOCAL_GO_WIN)),)
GO ?= $(LOCAL_GO_WIN)
else ifneq ($(wildcard $(LOCAL_GO_UNIX)),)
GO ?= $(LOCAL_GO_UNIX)
else
GO ?= go
endif
GO_CMD := "$(GO)"

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

MIGRATE := $(GO_ENV) $(GO_CMD) run -tags postgres github.com/golang-migrate/migrate/v4/cmd/migrate@v4.18.3
DATABASE_URL ?= postgres://ticketgo:ticketgo_local_password@localhost:5432/ticketgo?sslmode=disable

.PHONY: help bootstrap-go deps format lint test build run compose-up compose-down migrate-up migrate-down migrate-create check

help:
	@echo "bootstrap-go | deps | format | lint | test | build | run | compose-up | compose-down | migrate-up | migrate-down | migrate-create NAME=name | check"

bootstrap-go:
	powershell -NoProfile -ExecutionPolicy Bypass -File scripts/bootstrap-go.ps1

deps:
	$(GO_ENV) $(GO_CMD) mod download

format:
	$(GO_ENV) $(GO_CMD) fmt ./...

lint:
	$(GO_ENV) $(GO_CMD) vet ./...

test:
	$(GO_ENV) $(GO_CMD) test ./...

build:
	@$(MKDIR_BIN)
	$(GO_ENV) $(GO_CMD) build -o $(SERVER_BIN) ./cmd/server

run:
	$(GO_ENV) $(GO_CMD) run ./cmd/server

compose-up:
	docker compose up -d

compose-down:
	docker compose down

migrate-up:
	$(MIGRATE) -path migrations -database "$(DATABASE_URL)" up

migrate-down:
	$(MIGRATE) -path migrations -database "$(DATABASE_URL)" down 1

migrate-create:
	@$(CHECK_NAME)
	$(MIGRATE) create -ext sql -dir migrations -seq "$(NAME)"

check: format lint test build
