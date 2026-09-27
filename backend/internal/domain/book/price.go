package book

import (
	"fmt"

	"github.com/mrstsgk/book-management-system/backend/internal/domain/common"
)

const priceMax = 99_999_999

// Price is a value object for book prices in yen (0..99,999,999).
type Price struct {
	value int64
}

func NewPrice(v int64) (Price, error) {
	if v < 0 || v > priceMax {
		return Price{}, fmt.Errorf("%w: 書籍価格は0〜%dの範囲で入力してください", common.ErrInvalid, priceMax)
	}
	return Price{value: v}, nil
}

func (p Price) Int64() int64 {
	return p.value
}
