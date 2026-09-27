package book

import (
	"fmt"

	"github.com/mrstsgk/book-management-system/backend/internal/domain/common"
)

type PublishStatus int

const (
	Unpublished PublishStatus = 1
	Published   PublishStatus = 2
)

func NewPublishStatus(v int) (PublishStatus, error) {
	s := PublishStatus(v)
	if s != Unpublished && s != Published {
		return 0, fmt.Errorf("%w: 出版状況が不正です", common.ErrInvalid)
	}
	return s, nil
}

// CanChangeTo reports whether the status may move to next; published books can never go back to unpublished.
func (s PublishStatus) CanChangeTo(next PublishStatus) bool {
	return s != Published || next == Published
}
