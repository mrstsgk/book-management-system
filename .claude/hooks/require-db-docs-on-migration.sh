#!/usr/bin/env bash
#
# Stop hook: マイグレーション変更時にDBスキーマドキュメント更新を要求する
# -----------------------------------------------------------------
# ルール: docs/rules/db-documentation.md
#
# backend のマイグレーション（backend/migrations/*.sql）が変更されたのに、対応する
# docs/db/backend-schema.json / .md の両方が同じ変更に含まれていなければブロックする。
# 構造的な有無のみのチェックであり、内容の正確性は保証しない。
#
# バイパス: SKIP_DB_DOCS_CHECK=1

command -v jq >/dev/null 2>&1 || exit 0

case "${SKIP_DB_DOCS_CHECK:-}" in
    1|true|TRUE|yes) exit 0 ;;
esac

TOPLEVEL=$(git rev-parse --show-toplevel 2>/dev/null) || exit 0
BRANCH=$(git rev-parse --abbrev-ref HEAD 2>/dev/null) || exit 0
case "$BRANCH" in main|develop) exit 0 ;; esac

cd "$TOPLEVEL" || exit 0

MERGE_BASE=$(git merge-base origin/develop HEAD 2>/dev/null || git merge-base develop HEAD 2>/dev/null)
[ -n "$MERGE_BASE" ] || exit 0

CHANGED=$(
    {
        git diff --name-only "$MERGE_BASE" 2>/dev/null
        git ls-files --others --exclude-standard 2>/dev/null
    } | sort -u
)
[ -n "$CHANGED" ] || exit 0

_check_stack() {
    local stack="$1" migration_regex="$2"
    local migrated json_doc md_doc

    migrated=$(printf '%s\n' "$CHANGED" | grep -E "$migration_regex") || return 0
    [ -n "$migrated" ] || return 0

    json_doc=""
    md_doc=""
    printf '%s\n' "$CHANGED" | grep -Fxq "docs/db/${stack}-schema.json" && [ -f "docs/db/${stack}-schema.json" ] && json_doc="docs/db/${stack}-schema.json"
    printf '%s\n' "$CHANGED" | grep -Fxq "docs/db/${stack}-schema.md" && [ -f "docs/db/${stack}-schema.md" ] && md_doc="docs/db/${stack}-schema.md"

    if [ -n "$json_doc" ] && [ -n "$md_doc" ]; then
        return 0
    fi

    local missing=""
    [ -z "$json_doc" ] && missing="${missing}\n  - docs/db/${stack}-schema.json"
    [ -z "$md_doc" ] && missing="${missing}\n  - docs/db/${stack}-schema.md"

    local files
    files="$(printf '%s\n' "$migrated" | sed 's/^/  - /' | head -n 20)"
    reason="$(printf '%s の DBマイグレーションを変更しましたが、対応するスキーマドキュメントの更新がありません（ルール: docs/rules/db-documentation.md）。\n\n変更したマイグレーション:\n%s\n\n更新が必要なドキュメント:%b\n\nドキュメントを伴わない理由がある場合は SKIP_DB_DOCS_CHECK=1 で無視できます。' "$stack" "$files" "$missing")"
    jq -n --arg r "$reason" '{decision:"block", reason:$r}'
    exit 0
}

_check_stack "backend" '^backend/migrations/.*\.sql$'

exit 0
