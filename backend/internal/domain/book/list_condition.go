package book

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/mrstsgk/book-management-system/backend/internal/domain/common"
	"github.com/mrstsgk/book-management-system/backend/internal/domain/tag"
)

const keywordMaxLength = 100

// ListCondition は一覧の検索・絞り込みの条件（参照専用）。ゼロ値は「絞り込まない」。
type ListCondition struct {
	keyword string
	tagID   *tag.ID
}

// NewListCondition はキーワード（前後の空白を除いて100文字まで。空なら検索しない）と、任意の分野タグで条件を作る。
func NewListCondition(keyword string, tagID *tag.ID) (ListCondition, error) {
	// 前後のタブや改行も、黙って取り除かずに拒否するためトリム前に検査する
	if strings.IndexFunc(keyword, unicode.IsControl) >= 0 {
		return ListCondition{}, fmt.Errorf("%w: 検索キーワードにタブや改行は使用できません", common.ErrInvalid)
	}
	v := strings.TrimSpace(keyword)
	if utf8.RuneCountInString(v) > keywordMaxLength {
		return ListCondition{}, fmt.Errorf("%w: 検索キーワードは%d文字以内にしてください", common.ErrInvalid, keywordMaxLength)
	}
	c := ListCondition{keyword: v}
	if tagID != nil {
		if *tagID < 1 {
			return ListCondition{}, fmt.Errorf("%w: 分野タグのIDが不正です", common.ErrInvalid)
		}
		id := *tagID
		c.tagID = &id
	}
	return c, nil
}

// Keyword は書名・著者の部分一致に使う文字列。空なら検索しない。
func (c ListCondition) Keyword() string {
	return c.keyword
}

// TagID は絞り込む分野タグ。絞り込まないなら false。
func (c ListCondition) TagID() (tag.ID, bool) {
	if c.tagID == nil {
		return 0, false
	}
	return *c.tagID, true
}
