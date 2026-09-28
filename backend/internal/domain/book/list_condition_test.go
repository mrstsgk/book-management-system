package book_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/mrstsgk/book-management-system/backend/internal/domain/book"
	"github.com/mrstsgk/book-management-system/backend/internal/domain/common"
	"github.com/mrstsgk/book-management-system/backend/internal/domain/tag"
)

func TestNewListCondition(t *testing.T) {
	t.Parallel()
	id := func(v int64) *tag.ID { x := tag.ID(v); return &x }
	tests := []struct {
		name        string
		keyword     string
		tagID       *tag.ID
		wantKeyword string
		wantTag     tag.ID
		wantHasTag  bool
		wantErr     bool
	}{
		{name: "ゼロ値は絞り込まない", keyword: ""},
		{name: "空白だけは検索しない", keyword: " 　"},
		{name: "前後の空白を除く", keyword: " 設計 ", wantKeyword: "設計"},
		{name: "100文字ちょうどは有効", keyword: strings.Repeat("あ", 100), wantKeyword: strings.Repeat("あ", 100)},
		{name: "101文字はエラー", keyword: strings.Repeat("あ", 101), wantErr: true},
		{name: "改行はエラー", keyword: "設計\n入門", wantErr: true},
		{name: "前後のタブもエラー", keyword: "\t設計", wantErr: true},
		{name: "タグを指定できる", tagID: id(3), wantTag: 3, wantHasTag: true},
		{name: "タグID 0はエラー", tagID: id(0), wantErr: true},
		{name: "キーワードとタグを同時に指定できる", keyword: "設計", tagID: id(1), wantKeyword: "設計", wantTag: 1, wantHasTag: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := book.NewListCondition(tt.keyword, tt.tagID)
			if tt.wantErr {
				if !errors.Is(err, common.ErrInvalid) {
					t.Fatalf("err = %v, want ErrInvalid", err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got.Keyword() != tt.wantKeyword {
				t.Fatalf("Keyword() = %q, want %q", got.Keyword(), tt.wantKeyword)
			}
			if id, ok := got.TagID(); ok != tt.wantHasTag || id != tt.wantTag {
				t.Fatalf("TagID() = (%d, %v), want (%d, %v)", id, ok, tt.wantTag, tt.wantHasTag)
			}
		})
	}
}

func TestNewListCondition_CopiesTagID(t *testing.T) {
	t.Parallel()
	v := tag.ID(3)
	c, err := book.NewListCondition("", &v)
	if err != nil {
		t.Fatal(err)
	}
	v = 99
	if id, _ := c.TagID(); id != 3 {
		t.Fatalf("TagID() = %d, want 3 (changing the caller's variable must not change the condition)", id)
	}
}
