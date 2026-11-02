.DEFAULT_GOAL := help
.PHONY: help lint bundle bundle-check breaking

# 版はここで固定し、CI（.github/workflows/openapi-ci.yml）と揃える。
# Redocly は Node（mise）、oasdiff は Go（mise）で実行し、グローバルには入れない。
REDOCLY_VERSION := 2.63.0
OASDIFF_VERSION := v1.33.0

REDOCLY := npx -y @redocly/cli@$(REDOCLY_VERSION)

# 破壊的変更の比較元（既定は develop のバンドル）
BASE_REF ?= origin/develop

help: ## このヘルプを表示
	@grep -E '^[a-zA-Z0-9_-]+:.*?## .*$$' $(MAKEFILE_LIST) \
		| awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-14s\033[0m %s\n", $$1, $$2}'

lint: ## 構文・スキーマ・operationId の一意性を検査
	$(REDOCLY) lint --config redocly.yaml

bundle: ## 実行サービスごとに dist/ へまとめる（dist/ は手編集禁止）
	$(REDOCLY) bundle --config redocly.yaml

bundle-check: bundle ## バンドル結果にドリフトがないか検査（CI と同じ）
	git diff --exit-code -- dist/
	@test -z "$$(git status --porcelain -- dist/)" || (git status --porcelain -- dist/ && exit 1)

# 破壊的変更があっても --fail-on を付けないので成功扱い（警告のみ）。oasdiff 自体の失敗は失敗として返す。
# 比較元と変更後の dist/ の和集合を走査し、削除されたバンドルも報告する。
breaking: ## BASE_REF のバンドルと比べて破壊的変更を表示（本番リリースまでは警告のみ）
	@git rev-parse --verify --quiet "$(BASE_REF)^{commit}" >/dev/null || { \
		echo "BASE_REF=$(BASE_REF) を解決できないため比較できない（git fetch origin を実行するか BASE_REF を指定する）" >&2; \
		exit 1; \
	}
	@status=0; \
	files=$$( { git ls-tree --name-only "$(BASE_REF)" -- dist/ 2>/dev/null; ls dist/*.yaml 2>/dev/null; } | grep '\.yaml$$' | sort -u ); \
	for f in $$files; do \
		if [ ! -f "$$f" ]; then \
			echo "== $$f: 破壊的変更 — バンドルが削除された（$(BASE_REF) の全エンドポイントが無くなる）"; \
			continue; \
		fi; \
		base=$$(mktemp); \
		if git show "$(BASE_REF):openapi/$$f" > "$$base" 2>/dev/null; then \
			echo "== $$f"; \
			go run github.com/oasdiff/oasdiff@$(OASDIFF_VERSION) breaking "$$base" "$$f" || status=$$?; \
		else \
			echo "== $$f: $(BASE_REF) に無い新しいバンドルのため比較しない"; \
		fi; \
		rm -f "$$base"; \
	done; \
	exit $$status
