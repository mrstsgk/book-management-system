#!/usr/bin/env bash
#
# Stop hook: ADR check
# -----------------------------------------------
# Rule: docs/rules/adr.md
#
# Two checks:
#   1. Existence  — dependency/infra trigger files changed with no docs/adr/*.md touched.
#   2. Content    — a touched docs/adr/*.md has an empty/placeholder ## 背景・経緯 or ## 決定 section.
#
# This only verifies structural completeness (required headers have non-empty text),
# not whether the reasoning is actually sound — that ceiling is inherent to a grep-based
# check; a human/AI reviewer still has to judge the content itself.
#
# Bypass: SKIP_ADR_CHECK=1

if ! command -v jq >/dev/null 2>&1; then
    exit 0
fi

case "${SKIP_ADR_CHECK:-}" in
    1|true|TRUE|yes) exit 0 ;;
esac

cat >/dev/null 2>&1 # drain stdin

TOPLEVEL=$(git rev-parse --show-toplevel 2>/dev/null) || exit 0
BRANCH=$(git rev-parse --abbrev-ref HEAD 2>/dev/null) || exit 0
case "$BRANCH" in main|develop) exit 0 ;; esac

cd "$TOPLEVEL" || exit 0

MERGE_BASE=$(git merge-base origin/develop HEAD 2>/dev/null || git merge-base develop HEAD 2>/dev/null)
[ -n "$MERGE_BASE" ] || exit 0

# All changed paths: committed-since-develop + uncommitted (tracked + untracked)
CHANGED=$(
    {
        git diff --name-only "$MERGE_BASE" 2>/dev/null
        git ls-files --others --exclude-standard 2>/dev/null
    } | sort -u
)
[ -n "$CHANGED" ] || exit 0

# --- check 1: dependency/infra trigger without an ADR entry ---
TRIGGER=$(printf '%s\n' "$CHANGED" \
    | grep -E '(^|/)(go\.mod|go\.sum|package\.json|pnpm-lock\.yaml|Dockerfile|Dockerfile\.api)$|^\.github/workflows/.*\.ya?ml$|^infrastructure/.*\.tf$' \
    | grep -vE '/node_modules/' || true)

ADR_FILES=$(printf '%s\n' "$CHANGED" | grep -E '^docs/adr/.*\.md$' | grep -v '/README\.md$' | while IFS= read -r f; do [ -f "$f" ] && printf '%s\n' "$f"; done || true)

if [ -n "$TRIGGER" ] && [ -z "$ADR_FILES" ]; then
    files="$(printf '%s\n' "$TRIGGER" | sed 's/^/  - /' | head -n 20)"
    reason="依存関係/インフラの変更を検知しましたが docs/adr/ に判断記録がありません。
ライブラリ追加・置き換え・見送りなら docs/adr/YYYY-MM-DD-<slug>.md を作成してください（ルール: docs/rules/adr.md）。
判断を伴わない変更（パッチのみのlockfile更新など）なら SKIP_ADR_CHECK=1 で無視できます。

変更ファイル:
${files}"
    jq -n --arg r "$reason" '{decision:"block",reason:$r}'
    exit 0
fi

[ -n "$ADR_FILES" ] || exit 0

# --- check 2: content completeness of touched ADR files ---
extract_section() {
    # $1 = file, $2 = heading text (without "## ")
    awk -v h="$2" '
        /^## /{
            if (insec) exit
            if (index($0, "## " h) == 1) { insec=1; next }
        }
        insec { print }
    ' "$1"
}

is_placeholder() {
    # trim whitespace and strip known placeholder tokens; empty result = placeholder
    stripped=$(printf '%s' "$1" | tr -d '[:space:]')
    case "$stripped" in
        ""|"..."|"…"|TBD|TODO|todo|tbd) return 0 ;;
        *) return 1 ;;
    esac
}

missing=""
for f in $ADR_FILES; do
    [ -f "$f" ] || continue
    background=$(extract_section "$f" "背景・経緯")
    decision=$(extract_section "$f" "決定")
    if is_placeholder "$background"; then
        missing="${missing}\n  - ${f}: ## 背景・経緯 が空またはプレースホルダのままです"
    fi
    if is_placeholder "$decision"; then
        missing="${missing}\n  - ${f}: ## 決定 が空またはプレースホルダのままです"
    fi
done

if [ -n "$missing" ]; then
    reason="$(printf 'docs/adr/ のファイルに必須セクションの中身がありません（ルール: docs/rules/adr.md）。\n背景・経緯 と 決定 は実質的な内容を書いてください（見出しだけのスタブは不可）。%b' "$missing")"
    jq -n --arg r "$reason" '{decision:"block",reason:$r}'
    exit 0
fi

exit 0
