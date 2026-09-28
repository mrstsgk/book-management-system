package book

import (
	"fmt"
	"slices"

	"github.com/mrstsgk/book-management-system/backend/internal/domain/common"
	"github.com/mrstsgk/book-management-system/backend/internal/domain/tag"
)

// tagMaxCount は1冊に付けられる分野タグの上限。
const tagMaxCount = 10

// TagSelection は本に付ける分野タグの ID の集合の VO。タグの実体は持たず、ID の妥当性（上限・重複）だけを保証する。
// タグが実在するかどうかは、外部データを見る必要があるため VO では検証できず、ユースケースが確認する。
type TagSelection struct {
	ids []tag.ID
}

// NewTagSelection は0〜10個、重複の無いタグIDの集合を作る。
func NewTagSelection(ids []tag.ID) (TagSelection, error) {
	if len(ids) > tagMaxCount {
		return TagSelection{}, fmt.Errorf("%w: 分野タグは%d個までにしてください", common.ErrInvalid, tagMaxCount)
	}
	seen := make(map[tag.ID]bool, len(ids))
	for _, id := range ids {
		if seen[id] {
			return TagSelection{}, fmt.Errorf("%w: 同じ分野タグが重複しています", common.ErrInvalid)
		}
		seen[id] = true
	}
	cp := make([]tag.ID, len(ids))
	copy(cp, ids)
	return TagSelection{ids: cp}, nil
}

// IDs は呼び出し側が変更してもVOの不変条件を壊さないよう、複製を返す。
func (s TagSelection) IDs() []tag.ID {
	return slices.Clone(s.ids)
}
