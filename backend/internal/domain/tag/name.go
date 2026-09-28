package tag

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/mrstsgk/book-management-system/backend/internal/domain/common"
)

const nameMaxLength = 30

// Name はタグの名前の VO。書名など他のドメインの命名規則には依存しない
// （タグの命名規則は本の書名と揃える理由が無いため、独立して検証する）。
type Name struct {
	value string
}

// NewName は前後の空白を除いて1〜30文字のタグ名を作る。タブや改行などの制御文字は受け付けない。
func NewName(raw string) (Name, error) {
	// 前後にあるタブや改行も、黙って取り除かずに拒否するためトリム前に検査する
	if strings.IndexFunc(raw, unicode.IsControl) >= 0 {
		return Name{}, fmt.Errorf("%w: タグ名にタブや改行は使用できません", common.ErrInvalid)
	}
	v := strings.TrimSpace(raw)
	if n := utf8.RuneCountInString(v); n < 1 || n > nameMaxLength {
		return Name{}, fmt.Errorf("%w: タグ名は1〜%d文字にしてください", common.ErrInvalid, nameMaxLength)
	}
	return Name{value: v}, nil
}

func (n Name) String() string {
	return n.value
}
