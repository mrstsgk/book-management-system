package command

import (
	"time"

	"github.com/mrstsgk/book-management-system/backend/internal/domain/author"
)

var jst = time.FixedZone("Asia/Tokyo", 9*60*60)

// newBirthDate は任意項目の生年月日を VO に変換する。now が nil なら日本時間の現在時刻で判定する。
func newBirthDate(date *time.Time, now func() time.Time) (*author.BirthDate, error) {
	if date == nil {
		return nil, nil
	}
	current := time.Now().In(jst)
	if now != nil {
		current = now()
	}
	b, err := author.NewBirthDate(*date, current)
	if err != nil {
		return nil, err
	}
	return &b, nil
}
