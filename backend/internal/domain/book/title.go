package book

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/mrstsgk/book-management-system/backend/internal/domain/common"
)

const titleMaxLength = 255

// Title is a value object for book titles (1..255 runes, no whitespace).
type Title struct {
	value string
}

func NewTitle(raw string) (Title, error) {
	n := utf8.RuneCountInString(raw)
	if n < 1 || n > titleMaxLength {
		return Title{}, fmt.Errorf("%w: 書籍タイトルは1〜%d文字で入力してください", common.ErrInvalid, titleMaxLength)
	}
	if strings.IndexFunc(raw, unicode.IsSpace) >= 0 {
		return Title{}, fmt.Errorf("%w: 書籍タイトルに空白は使用できません", common.ErrInvalid)
	}
	return Title{value: raw}, nil
}

func (t Title) String() string {
	return t.value
}
