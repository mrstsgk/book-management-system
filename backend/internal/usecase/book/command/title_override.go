package command

import (
	"strings"
	"unicode"

	"github.com/mrstsgk/book-management-system/backend/internal/domain/book"
)

// parseTitleOverride は書名の上書きの入力を解釈する。前後の空白を除いて空なら「上書きしない」、それ以外は書名の規則で検証する。
// タブや改行は Unicode 上は空白の一種だが、制御文字として拒否する（「空白だけ」の特例で見逃さない）。
func parseTitleOverride(raw string) (*book.Title, error) {
	if strings.IndexFunc(raw, unicode.IsControl) < 0 && strings.TrimSpace(raw) == "" {
		return nil, nil
	}
	t, err := book.NewTitle(raw)
	if err != nil {
		return nil, err
	}
	return &t, nil
}
