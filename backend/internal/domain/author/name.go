package author

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/mrstsgk/book-management-system/backend/internal/domain/common"
)

const nameMaxLength = 100

// Name is a value object for author names (1..100 runes, no whitespace).
type Name struct {
	value string
}

func NewName(raw string) (Name, error) {
	n := utf8.RuneCountInString(raw)
	if n < 1 || n > nameMaxLength {
		return Name{}, fmt.Errorf("%w: 著者名は1〜%d文字で入力してください", common.ErrInvalid, nameMaxLength)
	}
	// unicode.IsSpace covers the full-width space (U+3000) as well.
	if strings.IndexFunc(raw, unicode.IsSpace) >= 0 {
		return Name{}, fmt.Errorf("%w: 著者名に空白は使用できません", common.ErrInvalid)
	}
	return Name{value: raw}, nil
}

func (n Name) String() string {
	return n.value
}
