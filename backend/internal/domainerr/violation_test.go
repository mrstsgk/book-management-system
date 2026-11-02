package domainerr_test

import (
	"errors"
	"fmt"
	"reflect"
	"testing"

	"github.com/mrstsgk/book-management-system/backend/internal/domainerr"
)

func TestViolations(t *testing.T) {
	t.Parallel()

	t.Run("空ならErrはnil", func(t *testing.T) {
		t.Parallel()
		var v domainerr.Violations
		if err := v.Err(); err != nil {
			t.Fatalf("Err() = %v, want nil", err)
		}
	})

	t.Run("足した違反をすべて保持しErrInvalidに一致する", func(t *testing.T) {
		t.Parallel()
		var v domainerr.Violations
		v.Add("isbn", "isbnDuplicated")
		v.Add("tagIds", "tagIdsNotFound")
		err := v.Err()
		if !errors.Is(err, domainerr.ErrInvalid) {
			t.Fatalf("errors.Is(err, ErrInvalid) = false, want true")
		}
		if errors.Is(err, domainerr.ErrConflict) || errors.Is(err, domainerr.ErrNotFound) {
			t.Fatal("must not match other sentinels")
		}
		var got domainerr.Violations
		if !errors.As(fmt.Errorf("wrap: %w", err), &got) {
			t.Fatal("errors.As must unwrap to Violations")
		}
		want := domainerr.Violations{{Field: "isbn", Rule: "isbnDuplicated"}, {Field: "tagIds", Rule: "tagIdsNotFound"}}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("got %+v, want %+v", got, want)
		}
	})

	t.Run("同じfieldとruleは1つにまとめる", func(t *testing.T) {
		t.Parallel()
		var v domainerr.Violations
		v.Add("name", "tagNameDuplicated")
		v.Add("name", "tagNameDuplicated")
		v.Add("name", "invalid")
		if len(v) != 2 {
			t.Fatalf("len = %d, want 2 (%+v)", len(v), v)
		}
	})

	t.Run("メッセージに違反数が入る", func(t *testing.T) {
		t.Parallel()
		v := domainerr.Violations{{Field: "a", Rule: "b"}}
		if v.Error() != "invalid: 1 violation(s)" {
			t.Fatalf("Error() = %q", v.Error())
		}
	})
}
