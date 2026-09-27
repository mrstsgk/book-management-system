.DEFAULT_GOAL := help
.PHONY: help tidy run db-up db-down migrate-up migrate-down migrate-create swagger swagger-check fmt-check lint vuln test build tools

# CLI ツールは go.mod に入れず go install する。版はここで固定し、CI と揃える。
SWAG_VERSION        := v1.16.6
MIGRATE_VERSION     := v4.20.1
GOLANGCI_VERSION    := v2.12.2
GOVULNCHECK_VERSION := v1.7.0

DB_URL ?= postgres://postgres:postgres@localhost:5432/book_management?sslmode=disable

help: ## このヘルプを表示
	@grep -E '^[a-zA-Z0-9_-]+:.*?## .*$$' $(MAKEFILE_LIST) \
		| awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-22s\033[0m %s\n", $$1, $$2}'

tidy: ## go mod tidy
	go mod tidy

run: ## API を起動（:8080）
	go run ./cmd/api

db-up: ## PostgreSQL を起動（docker compose）
	docker compose up -d

db-down: ## PostgreSQL を停止
	docker compose down

migrate-up: ## マイグレーションを適用
	migrate -path ./migrations -database "$(DB_URL)" up

migrate-down: ## マイグレーションを 1 つ戻す
	migrate -path ./migrations -database "$(DB_URL)" down 1

migrate-create: ## マイグレーションを新規作成（make migrate-create NAME=create_books_table）
	@test -n "$(NAME)" || (echo "NAME を指定してください（例: make migrate-create NAME=create_books_table）" && exit 1)
	migrate create -ext sql -dir ./migrations -seq $(NAME)

# Output is committed; regenerate and check git diff in CI.
swagger: ## OpenAPI を排出（api/docs は手編集禁止）
	swag init -g cmd/api/main.go -o api/docs --parseDependency --parseInternal --outputTypes yaml,json

swagger-check: swagger ## 排出物にドリフトがないか検査（CI と同じ）
	git diff --exit-code -- api/docs/swagger.yaml api/docs/swagger.json

fmt-check: ## 未フォーマットのファイルがないか検査（CI と同じ）
	@out=$$(gofmt -l ./cmd ./config ./internal) || exit 1; \
	if [ -n "$$out" ]; then echo "gofmt が必要なファイル:"; echo "$$out"; exit 1; fi

lint: ## golangci-lint + go vet
	golangci-lint run
	go vet ./...

vuln: ## 依存の脆弱性チェック
	govulncheck ./...

test: ## 全テストを実行
	go test ./...

build: ## API をビルド（bin/）
	go build -o bin/api ./cmd/api

tools: ## CLI ツールを導入（swag / migrate / golangci-lint / govulncheck）
	go install github.com/swaggo/swag/cmd/swag@$(SWAG_VERSION)
	go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@$(MIGRATE_VERSION)
	go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@$(GOLANGCI_VERSION)
	go install golang.org/x/vuln/cmd/govulncheck@$(GOVULNCHECK_VERSION)
