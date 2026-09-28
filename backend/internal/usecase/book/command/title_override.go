package command

import (
	"strings"

	"github.com/mrstsgk/book-management-system/backend/internal/domain/book"
)

// parseTitleOverride は書名の上書きの入力を解釈する。前後の空白を除いて空なら「上書きしない」、それ以外は書名の規則で検証する。
func parseTitleOverride(raw string) (*book.Title, error) {
	if strings.TrimSpace(raw) == "" {
		return nil, nil
	}
	t, err := book.NewTitle(raw)
	if err != nil {
		return nil, err
	}
	return &t, nil
}
