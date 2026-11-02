package domainerr_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/mrstsgk/book-management-system/backend/internal/domainerr"
)

func TestCodedError_分類で判定でき包んでもコードを取り出せる(t *testing.T) {
	t.Parallel()
	err := fmt.Errorf("wrap: %w", domainerr.New(domainerr.ErrNotFound, "bookNotFound", "読んだ本が見つかりません"))

	if !errors.Is(err, domainerr.ErrNotFound) {
		t.Fatal("errors.Is(err, ErrNotFound) = false, want true")
	}
	if errors.Is(err, domainerr.ErrConflict) {
		t.Fatal("errors.Is(err, ErrConflict) = true, want false")
	}
	ce, ok := errors.AsType[*domainerr.CodedError](err)
	if !ok || ce.Code != "bookNotFound" || ce.Params != nil {
		t.Fatalf("CodedError = %+v, want code bookNotFound without params", ce)
	}
}

func TestCodedError_WithParamsは元のエラーを変えない(t *testing.T) {
	t.Parallel()
	base := domainerr.New(domainerr.ErrTooManyAttempts, "tooManyLoginAttempts", "しばらく待ってください")
	got := base.WithParams(map[string]any{"retryAfterSeconds": 30})

	if got.Params["retryAfterSeconds"] != 30 {
		t.Fatalf("params = %v, want retryAfterSeconds 30", got.Params)
	}
	if base.Params != nil {
		t.Fatalf("base params = %v, want nil", base.Params)
	}
}
