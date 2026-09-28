package common

import "fmt"

const (
	DefaultListLimit = 20
	// MaxListLimit keeps list queries bounded no matter what the client asks for.
	MaxListLimit = 100
)

// ListRange is a value object for the slice of a list to fetch: up to limit items
// (1..MaxListLimit) starting at offset (>= 0).
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
