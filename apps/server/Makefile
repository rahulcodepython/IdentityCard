GOBIN ?= $(shell go env GOPATH 2>/dev/null)/bin
export PATH := $(GOBIN):$(PATH)

ifneq (,$(wildcard .env))
include .env
export
endif

.PHONY: dev build seed migrate-up migrate-down migrate-new migrate-force lint

dev:
	go run ./cmd/server

build:
	go build -o bin/server ./cmd/server

seed:
	go run ./cmd/seed

migrate-up:
	migrate -path internal/migrations -database "$$DATABASE_URL" up

migrate-down:
	migrate -path internal/migrations -database "$$DATABASE_URL" down 1

migrate-new:
	migrate create -ext sql -dir internal/migrations -seq $(name)

migrate-force:
	migrate -path internal/migrations -database "$$DATABASE_URL" force $(version)

lint:
	go vet ./...
	gofmt -l .
