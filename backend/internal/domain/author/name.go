package author

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/mrstsgk/book-management-system/backend/internal/domain/common"
)

const nameMaxLength = 100

// Name is a value object for author names: 1..100 runes after trimming surrounding
// whitespace. Inner spaces are kept so names like "Martin Kleppmann" fit; control
// characters (tabs, newlines) are rejected.
type Name struct {
	value string
}

func NewName(raw string) (Name, error) {
	// Checked before trimming so a stray tab or newline is rejected rather than silently dropped.
	if strings.IndexFunc(raw, unicode.IsControl) >= 0 {
		return Name{}, fmt.Errorf("%w: 著者名にタブや改行は使用できません", common.ErrInvalid)
	}
	v := strings.TrimSpace(raw)
	n := utf8.RuneCountInString(v)
	if n < 1 || n > nameMaxLength {
		return Name{}, fmt.Errorf("%w: 著者名は1〜%d文字で入力してください", common.ErrInvalid, nameMaxLength)
	}
	return Name{value: v}, nil
}

func (n Name) String() string {
	return n.value
}
