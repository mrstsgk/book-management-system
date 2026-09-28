package book

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/mrstsgk/book-management-system/backend/internal/domain/common"
)

const titleMaxLength = 255

// Title is a value object for book titles: 1..255 runes after trimming surrounding
// whitespace. Inner spaces (including full-width ones, common in Japanese titles such
// as "改訂新版　…") are kept; control characters (tabs, newlines) are rejected.
type Title struct {
	value string
}

func NewTitle(raw string) (Title, error) {
	// Checked before trimming so a stray tab or newline is rejected rather than silently dropped.
	if strings.IndexFunc(raw, unicode.IsControl) >= 0 {
		return Title{}, fmt.Errorf("%w: 書籍タイトルにタブや改行は使用できません", common.ErrInvalid)
	}
	v := strings.TrimSpace(raw)
	n := utf8.RuneCountInString(v)
	if n < 1 || n > titleMaxLength {
		return Title{}, fmt.Errorf("%w: 書籍タイトルは1〜%d文字で入力してください", common.ErrInvalid, titleMaxLength)
	}
	return Title{value: v}, nil
}

func (t Title) String() string {
	return t.value
}
