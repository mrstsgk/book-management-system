// parseTagSelection は非公開関数（register.go・update.go の両方から呼ぶ内部ヘルパー）のため、
// 公開APIからは直接検証できず in-package テストにしている。
package command

import (
	"context"
	"errors"
	"reflect"
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

func (f *fakeTagQuery) CountBooks(context.Context) (*tag.TagBookCounts, error) { return nil, nil }

func (f *fakeTagQuery) ExistsAll(_ context.Context, ids []tag.ID) (bool, error) {
	f.called++
	f.gotIDs = ids
	return f.exists, f.existsErr
}

func (f *fakeTagQuery) ExistsByName(context.Context, string, tag.ID) (bool, error) { return false, nil }

func TestParseTagSelection(t *testing.T) {
	t.Parallel()

	t.Run("0個は空のTagSelectionを返す（実在確認は空配列で呼ばれる）", func(t *testing.T) {
		t.Parallel()
		tags := &fakeTagQuery{exists: true}
		var vs common.Violations
		got, err := parseTagSelection(context.Background(), &vs, tags, nil)
		if err != nil || len(vs) != 0 {
			t.Fatalf("err = %v, violations = %+v", err, vs)
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
		var vs common.Violations
		got, err := parseTagSelection(context.Background(), &vs, tags, []int64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10})
		if err != nil || len(vs) != 0 {
			t.Fatalf("err = %v, violations = %+v", err, vs)
		}
		if ids := got.IDs(); len(ids) != 10 {
			t.Fatalf("IDs = %v, want 10 ids", ids)
		}
		if tags.called != 1 || len(tags.gotIDs) != 10 {
			t.Fatalf("ExistsAll called with %v, want 10 ids", tags.gotIDs)
		}
	})

	// 上限超過・重複は前提が成立しないので実在確認をせず、違反だけを集める
	for _, tt := range []struct {
		name string
		ids  []int64
		want common.Violation
	}{
		{name: "11個は max の違反", ids: []int64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11}, want: common.Violation{Field: "tagIds", Rule: "max"}},
		{name: "重複したIDは tagIdsDuplicated の違反", ids: []int64{1, 1}, want: common.Violation{Field: "tagIds", Rule: "tagIdsDuplicated"}},
	} {
		t.Run(tt.name+"でExistsAllを呼ばない", func(t *testing.T) {
			t.Parallel()
			tags := &fakeTagQuery{exists: true}
			var vs common.Violations
			if _, err := parseTagSelection(context.Background(), &vs, tags, tt.ids); err != nil {
				t.Fatalf("err = %v, want nil (violations are collected, not returned)", err)
			}
			if !reflect.DeepEqual(vs, common.Violations{tt.want}) {
				t.Fatalf("violations = %+v, want [%+v]", vs, tt.want)
			}
			if tags.called != 0 {
				t.Fatal("ExistsAll must not be called when the input is already invalid")
			}
		})
	}

	t.Run("存在しないIDが混じると tagIdsNotFound の違反", func(t *testing.T) {
		t.Parallel()
		var vs common.Violations
		if _, err := parseTagSelection(context.Background(), &vs, &fakeTagQuery{exists: false}, []int64{999}); err != nil {
			t.Fatalf("err = %v", err)
		}
		if want := (common.Violations{{Field: "tagIds", Rule: "tagIdsNotFound"}}); !reflect.DeepEqual(vs, want) {
			t.Fatalf("violations = %+v, want %+v", vs, want)
		}
	})

	t.Run("ExistsAllの障害はそのまま返す", func(t *testing.T) {
		t.Parallel()
		wantErr := errors.New("db: timeout")
		var vs common.Violations
		if _, err := parseTagSelection(context.Background(), &vs, &fakeTagQuery{existsErr: wantErr}, []int64{1}); !errors.Is(err, wantErr) {
			t.Fatalf("err = %v, want %v", err, wantErr)
		}
	})
}
