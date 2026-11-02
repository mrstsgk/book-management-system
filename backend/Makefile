.DEFAULT_GOAL := help
.PHONY: help tidy run-public run-admin db-up db-down migrate-up migrate-down migrate-create openapi api-gen api-gen-check fmt-check lint vuln test build docker-build-public docker-build-admin tools

# CLI ツールは go.mod に入れず go install / go run する。版はここで固定し、CI と揃える。
OAPI_CODEGEN_VERSION := v2.8.1
MIGRATE_VERSION     := v4.20.1
GOLANGCI_VERSION    := v2.15.0
GOVULNCHECK_VERSION := v1.7.0

DB_URL ?= postgres://postgres:postgres@localhost:5432/book_management?sslmode=disable

help: ## このヘルプを表示
	@grep -E '^[a-zA-Z0-9_-]+:.*?## .*$$' $(MAKEFILE_LIST) \
		| awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-22s\033[0m %s\n", $$1, $$2}'

tidy: ## go mod tidy
	go mod tidy

run-public: ## 閲覧の API（public-api）を起動（:8080）
	go run ./cmd/public-api

run-admin: ## 管理の API（admin-api）を起動（:8081）
	go run ./cmd/admin-api

db-up: ## PostgreSQL を起動（docker compose。healthcheck が通るまで待つ）
	docker compose up -d --wait

db-down: ## PostgreSQL を停止
	docker compose down

migrate-up: ## マイグレーションを適用
	migrate -path ./migrations -database "$(DB_URL)" up

migrate-down: ## マイグレーションを 1 つ戻す
	migrate -path ./migrations -database "$(DB_URL)" down 1

migrate-create: ## マイグレーションを新規作成（make migrate-create NAME=create_books_table）
	@test -n "$(NAME)" || (echo "NAME を指定してください（例: make migrate-create NAME=create_books_table）" && exit 1)
	migrate create -ext sql -dir ./migrations -seq $(NAME)

# 正本は ../openapi/ の YAML。生成物（api/gen/）はコミットし、CI で git diff を検査する。
openapi: ## OpenAPI を lint・バンドルし、BE のコードを再生成（YAML 変更後はこれ）
	$(MAKE) -C ../openapi lint bundle
	$(MAKE) api-gen

api-gen: ## ../openapi/dist から HTTP 境界のコードを生成（api/gen は手編集禁止）
	go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@$(OAPI_CODEGEN_VERSION) \
		-config api/gen/public/config.yaml ../openapi/dist/public.yaml
	go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@$(OAPI_CODEGEN_VERSION) \
		-config api/gen/admin/config.yaml ../openapi/dist/admin.yaml

api-gen-check: api-gen ## 生成物にドリフトがないか検査（CI と同じ）
	git diff --exit-code -- api/gen/
	@test -z "$$(git status --porcelain -- api/gen/)" || (git status --porcelain -- api/gen/ && exit 1)

fmt-check: ## 未フォーマットのファイルがないか検査（CI と同じ）
	@out=$$(gofmt -l ./cmd ./config ./internal) || exit 1; \
	if [ -n "$$out" ]; then echo "gofmt が必要なファイル:"; echo "$$out"; exit 1; fi

lint: ## golangci-lint + go vet
	golangci-lint run
	go vet ./...

vuln: ## 依存の脆弱性チェック
	govulncheck ./...

test: ## 全テストを実行（DB を使うテストは make db-up migrate-up 済みが前提。DB の状態はテストキャッシュの入力に含まれないためキャッシュを使わず、DB を使うパッケージが同じテーブルを触るので 1 つずつ走らせる）
	go test -p 1 -count=1 ./...

build: ## API をビルド（bin/）
	go build -o bin/public-api ./cmd/public-api
	go build -o bin/admin-api ./cmd/admin-api

# 本番の形のイメージ（日常の開発では使わない。デプロイはしない）
docker-build-public: ## public-api のイメージをビルド
	docker build -f Dockerfile.api --build-arg CMD=public-api -t book-management-public-api:local .

docker-build-admin: ## admin-api のイメージをビルド
	docker build -f Dockerfile.api --build-arg CMD=admin-api -t book-management-admin-api:local .

tools: ## CLI ツールを導入（migrate / golangci-lint / govulncheck。oapi-codegen は go run で版固定）
	go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@$(MIGRATE_VERSION)
	go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@$(GOLANGCI_VERSION)
	go install golang.org/x/vuln/cmd/govulncheck@$(GOVULNCHECK_VERSION)
