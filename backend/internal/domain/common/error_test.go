package common_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/mrstsgk/book-management-system/backend/internal/domain/common"
)

func TestSentinelErrors(t *testing.T) {
	t.Parallel()
	sentinels := []error{common.ErrNotFound, common.ErrInvalid, common.ErrConflict}

	for i, target := range sentinels {
		t.Run(target.Error(), func(t *testing.T) {
			t.Parallel()
			wrapped := fmt.Errorf("%w: detail", target)
			for j, other := range sentinels {
				if got, want := errors.Is(wrapped, other), i == j; got != want {
					t.Errorf("errors.Is(wrap(%v), %v) = %v, want %v", target, other, got, want)
				}
			}
		})
	}
}
