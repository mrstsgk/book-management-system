#!/usr/bin/env bash
#
# SessionStart(compact) hook: 退避した圧縮前の差分をコンテキストへ再注入
# ------------------------------------------------------------------
# /compact 完了直後、precompact-save-evidence.sh が保存した
# .claude/_evidence/precompact_manual.txt が存在すれば、その内容を
# additionalContext として圧縮後のコンテキストへ戻す。
#
# PostCompact イベントは additionalContext をサポートしないため、
# additionalContext をサポートする SessionStart(matcher: "compact") を使う。
#
# これにより圧縮後も「圧縮直前の作業ツリーの差分」を引き継げる。
# ファイルが無い（手動 compact でない等）場合は何もしない。

set -euo pipefail

command -v jq >/dev/null 2>&1 || exit 0

input="$(cat 2>/dev/null || true)"
session_id="$(printf '%s' "$input" | jq -r '.session_id // empty' 2>/dev/null || true)"
[ -n "$session_id" ] || exit 0

top="$(git rev-parse --show-toplevel 2>/dev/null || true)"
[ -n "$top" ] || exit 0

evidence="$top/.claude/_evidence/precompact_manual.${session_id}.txt"
[ -s "$evidence" ] || exit 0

ctx="$(printf '圧縮直前の作業ツリーの状態（%s より）:\n\n%s' "$evidence" "$(cat "$evidence")")"

jq -n --arg ctx "$ctx" '{hookSpecificOutput:{hookEventName:"SessionStart",additionalContext:$ctx}}'

exit 0
