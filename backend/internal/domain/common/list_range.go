package common

import "fmt"

const (
	DefaultListLimit = 20
	// MaxListLimit はクライアントの指定にかかわらず、一覧の取得件数に上限をかける。
	MaxListLimit = 100
)

// ListRange は一覧の取得範囲の VO。取得開始位置 offset（0 以上）から、最大 limit 件（1〜MaxListLimit）を取る。
type ListRange struct {
	limit  int
	offset int
}

func NewListRange(limit, offset int) (ListRange, error) {
	if limit < 1 || limit > MaxListLimit {
		return ListRange{}, fmt.Errorf("%w: 取得件数は1〜%d件で指定してください", ErrInvalid, MaxListLimit)
	}
	if offset < 0 {
		return ListRange{}, fmt.Errorf("%w: 取得開始位置は0以上で指定してください", ErrInvalid)
	}
	return ListRange{limit: limit, offset: offset}, nil
}

func (r ListRange) Limit() int {
	return r.limit
}

func (r ListRange) Offset() int {
	return r.offset
}
