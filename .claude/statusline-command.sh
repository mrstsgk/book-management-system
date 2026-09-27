#!/bin/bash
if ! command -v jq >/dev/null 2>&1; then
    echo "📁 (jq not found)"
    exit 0
fi

input=$(cat)

MODEL=$(echo "$input" | jq -r '.model.display_name')
DIR=$(echo "$input" | jq -r '.workspace.current_dir')
EFFORT=$(echo "$input" | jq -r '.effort.level // "—"')
PCT=$(echo "$input" | jq -r '.context_window.used_percentage // 0' | cut -d. -f1)

GREEN=$'\033[32m'
YELLOW=$'\033[33m'
RED=$'\033[31m'
RESET=$'\033[0m'

if [ "$PCT" -ge 70 ]; then BAR_COLOR="$RED"
elif [ "$PCT" -ge 50 ]; then BAR_COLOR="$YELLOW"
else BAR_COLOR="$GREEN"; fi

FILLED=$((PCT / 10)); EMPTY=$((10 - FILLED))
printf -v FILL "%${FILLED}s"; printf -v PAD "%${EMPTY}s"
BAR="${FILL// /█}${PAD// /░}"

BRANCH=""
if [[ "$DIR" == *"/.claude/worktrees/"* ]]; then
    WORKTREE="${DIR##*/.claude/worktrees/}"
    DIR_LABEL="🌲 ${WORKTREE}"
else
    DIR_LABEL="📁 ${DIR##*/}"
    if git -C "$DIR" rev-parse --git-dir > /dev/null 2>&1; then
        branch=$(git -C "$DIR" branch --show-current 2>/dev/null)
        [ -n "$branch" ] && BRANCH=" | 🌿 $branch"
    fi
fi

printf '%s\n' "${DIR_LABEL}${BRANCH}"
printf '🤖 %s%s | 💪 %s%s | 📚 %s%s%s %s%%\n' "$MODEL" "$RESET" "$EFFORT" "$RESET" "$BAR_COLOR" "$BAR" "$RESET" "$PCT"
