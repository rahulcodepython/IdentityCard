GOBIN := $(shell go env GOPATH)/bin
export PATH := $(GOBIN):$(PATH)

ifneq (,$(wildcard .env))
include .env
export
endif

.PHONY: dev build seed migrate-up migrate-down migrate-new sqlc-generate lint

dev:
	go run ./cmd/api

build:
	go build -o bin/api ./cmd/api

seed:
	go run ./cmd/seed

migrate-up:
	migrate -path internal/db/migrations -database "$$DATABASE_URL" up

migrate-down:
	migrate -path internal/db/migrations -database "$$DATABASE_URL" down 1

migrate-new:
	migrate create -ext sql -dir internal/db/migrations -seq $(name)

sqlc-generate:
	sqlc generate

lint:
	go vet ./...
	gofmt -l .
