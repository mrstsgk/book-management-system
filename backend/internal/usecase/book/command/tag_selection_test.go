// parseTagSelection は非公開関数（register.go・update.go の両方から呼ぶ内部ヘルパー）のため、
// 公開APIからは直接検証できず in-package テストにしている。
package command

import (
	"context"
	"errors"
	"testing"

	"github.com/mrstsgk/book-management-system/backend/internal/domain/common"
	"github.com/mrstsgk/book-management-system/backend/internal/domain/tag"
)

type fakeTagQuery struct {
	exists    bool
	existsErr error
	gotIDs    []tag.ID
	called    int
}

func (f *fakeTagQuery) FindList(context.Context) (*tag.TagList, error) { return nil, nil }

func (f *fakeTagQuery) ExistsAll(_ context.Context, ids []tag.ID) (bool, error) {
	f.called++
	f.gotIDs = ids
	return f.exists, f.existsErr
}

func TestParseTagSelection(t *testing.T) {
	t.Parallel()

	t.Run("0個は空のTagSelectionを返す（実在確認は空配列で呼ばれる）", func(t *testing.T) {
		t.Parallel()
		tags := &fakeTagQuery{exists: true}
		got, err := parseTagSelection(context.Background(), tags, nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if tags.called != 1 || len(tags.gotIDs) != 0 {
			t.Fatalf("ExistsAll called=%d with %v, want called once with an empty slice", tags.called, tags.gotIDs)
		}
		if len(got.IDs()) != 0 {
			t.Fatalf("IDs = %v, want empty", got.IDs())
		}
	})

	t.Run("10個は実在確認して返す", func(t *testing.T) {
		t.Parallel()
		tags := &fakeTagQuery{exists: true}
		raw := []int64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
		got, err := parseTagSelection(context.Background(), tags, raw)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if ids := got.IDs(); len(ids) != 10 {
			t.Fatalf("IDs = %v, want 10 ids", ids)
		}
		if tags.called != 1 || len(tags.gotIDs) != 10 {
			t.Fatalf("ExistsAll called with %v, want 10 ids", tags.gotIDs)
		}
	})

	t.Run("11個は上限違反でExistsAllを呼ばずにエラー", func(t *testing.T) {
		t.Parallel()
		tags := &fakeTagQuery{exists: true}
		raw := []int64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11}
		if _, err := parseTagSelection(context.Background(), tags, raw); !errors.Is(err, common.ErrInvalid) {
			t.Fatalf("err = %v, want ErrInvalid", err)
		}
		if tags.called != 0 {
			t.Fatal("ExistsAll must not be called when the count is already invalid")
		}
	})

	t.Run("重複したIDはExistsAllを呼ばずにエラー", func(t *testing.T) {
		t.Parallel()
		tags := &fakeTagQuery{exists: true}
		if _, err := parseTagSelection(context.Background(), tags, []int64{1, 1}); !errors.Is(err, common.ErrInvalid) {
			t.Fatalf("err = %v, want ErrInvalid", err)
		}
		if tags.called != 0 {
			t.Fatal("ExistsAll must not be called when the input already has duplicates")
		}
	})

	t.Run("存在しないIDが混じるとErrInvalid", func(t *testing.T) {
		t.Parallel()
		tags := &fakeTagQuery{exists: false}
		if _, err := parseTagSelection(context.Background(), tags, []int64{999}); !errors.Is(err, common.ErrInvalid) {
			t.Fatalf("err = %v, want ErrInvalid", err)
		}
	})

	t.Run("ExistsAllの障害はそのまま返す", func(t *testing.T) {
		t.Parallel()
		wantErr := errors.New("db: timeout")
		tags := &fakeTagQuery{existsErr: wantErr}
		if _, err := parseTagSelection(context.Background(), tags, []int64{1}); !errors.Is(err, wantErr) {
			t.Fatalf("err = %v, want %v", err, wantErr)
		}
	})
}
