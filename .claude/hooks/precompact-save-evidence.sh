#!/usr/bin/env bash
#
# PreCompact(manual) hook: 圧縮前の差分・状況をファイルへ退避
# ----------------------------------------------------------
# 手動 /compact の直前に、その時点の作業ツリーの状態（git status / diff）を
# .claude/_evidence/precompact_manual.txt へ保存する。
#
# 目的: 会話が長引いて /compact した後でも「その時点の差分なんだっけ？」を
#       .claude/_evidence/ から復元できるようにする。
#
# 保存先はカレントの git リポジトリ（worktree を含む）のトップ直下。
# .claude/* は .gitignore 済みのため _evidence はコミット対象外。

set -euo pipefail

command -v jq >/dev/null 2>&1 || exit 0

input="$(cat 2>/dev/null || true)"
session_id="$(printf '%s' "$input" | jq -r '.session_id // empty' 2>/dev/null || true)"
[ -n "$session_id" ] || exit 0

top="$(git rev-parse --show-toplevel 2>/dev/null || true)"
[ -n "$top" ] || exit 0

evidence_dir="$top/.claude/_evidence"
mkdir -p "$evidence_dir"

# session_id ごとにファイルを分ける（同じ worktree で複数セッションが並行して
# compact すると、分けないと後から実行した方が先のセッションのevidenceを上書きし、
# 別セッションの内容が誤って注入される）
out="$evidence_dir/precompact_manual.${session_id}.txt"

{
    date
    echo
    echo "# branch"
    git -C "$top" rev-parse --abbrev-ref HEAD 2>/dev/null || true
    echo
    echo "# git status --porcelain"
    git -C "$top" status --porcelain 2>/dev/null || true
    echo
    echo "# git diff --stat"
    git -C "$top" diff --stat 2>/dev/null || true
    echo
    echo "# git diff (unstaged, full)"
    git -C "$top" diff --binary 2>/dev/null || true
    echo
    echo "# git diff --cached (staged, full)"
    git -C "$top" diff --cached --binary 2>/dev/null || true
    echo
    echo "# untracked files (content)"
    git -C "$top" status --porcelain 2>/dev/null | awk '/^\?\? /{print substr($0,4)}' | while IFS= read -r f; do
        echo "--- untracked: $f ---"
        git -C "$top" diff --binary --no-index -- /dev/null "$top/$f" 2>/dev/null || true
    done
} >"$out"

exit 0
