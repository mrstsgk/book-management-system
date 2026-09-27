#!/usr/bin/env bash
# Stop hook: 判断が確定した後（作業終了時）に docs/adr/ 記録を促す。
# ルール: docs/rules/adr.md
#
# require-adr-on-decision.sh (Stop) は依存関係ファイル変更という機械的トリガーしか
# 検知できない。デバッグの根本原因判断・レビュー指摘の見送りは diff だけでは
# 「判断があったか」自体を検知できないため、該当スキル使用のリマインドで補う。
#
# PostToolUse(Skill) はスキル呼び出し直後に発火し、根本原因の特定やレビュー指摘の
# 採否が確定する前に届いてしまうため、ターン終了時点の Stop で確認する
# （直近のユーザー発言以降に呼ばれた Skill だけを対象にする）。
set -euo pipefail

input="$(cat)"
transcript="$(printf '%s' "$input" | jq -r '.transcript_path // empty')"
[ -n "$transcript" ] && [ -f "$transcript" ] || exit 0

skill=$(jq -n '
  [inputs] as $entries
  | ($entries | to_entries
      | map(select(.value.type=="user" and (.value.message.content|type)=="string"))
      | last.key // -1) as $boundary
  | [$entries[($boundary+1):][]
      | select(.type=="assistant") | .message.content[]?
      | select(.type=="tool_use" and .name=="Skill") | .input.skill]
  | last // empty
' "$transcript" 2>/dev/null || true)
skill="${skill#\"}"
skill="${skill%\"}"

case "$skill" in
  *systematic-debugging*)
    msg="【ADR】根本原因の特定・対応方針の決定が終わったら、複数の対応案を検討したなら docs/adr/YYYY-MM-DD-<slug>.md に記録すること（ルール: docs/rules/adr.md）。原因がひとつしかなく検討の余地がなかったなら不要。" ;;
  *code-review*|*coderabbit-review*)
    msg="【ADR】レビュー指摘を「不採用」で見送った場合、理由が「一時的に無視」を超えるものなら docs/adr/YYYY-MM-DD-<slug>.md に記録すること（ルール: docs/rules/adr.md）。全指摘を反映したなら不要。" ;;
  *)
    exit 0 ;;
esac

jq -n --arg m "$msg" '{hookSpecificOutput:{hookEventName:"Stop",additionalContext:$m}}'
