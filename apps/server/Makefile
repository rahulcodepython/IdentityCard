GOBIN := $(shell go env GOPATH)/bin
export PATH := $(GOBIN):$(PATH)

ifneq (,$(wildcard .env))
include .env
export
endif

.PHONY: dev build seed migrate-up migrate-down migrate-new migrate-force sqlc-generate lint

dev:
	go run ./cmd/server

build:
	go build -o bin/server ./cmd/server

seed:
	go run ./cmd/seed

migrate-up:
	migrate -path internal/db/migrations -database "$$DATABASE_URL" up

migrate-down:
	migrate -path internal/db/migrations -database "$$DATABASE_URL" down 1

migrate-new:
	migrate create -ext sql -dir internal/db/migrations -seq $(name)

migrate-force:
	migrate -path internal/db/migrations -database "$$DATABASE_URL" force $(version)

sqlc-generate:
	sqlc generate

lint:
	go vet ./...
	gofmt -l .
