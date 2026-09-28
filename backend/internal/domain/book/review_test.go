package book_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/mrstsgk/book-management-system/backend/internal/domain/book"
	"github.com/mrstsgk/book-management-system/backend/internal/domain/common"
)

func TestNewComment(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		in      string
		want    string
		wantErr bool
	}{
		{name: "通常の感想は有効", in: "分散システムの設計を体系的に学べた", want: "分散システムの設計を体系的に学べた"},
		{name: "改行とタブを含む本文は有効", in: "第1部\n\t信頼性\r\n第2部", want: "第1部\n\t信頼性\r\n第2部"},
		{name: "前後の空白と改行は取り除く", in: "\n  良書  \n", want: "良書"},
		{name: "5000文字ちょうどは有効", in: strings.Repeat("あ", 5000), want: strings.Repeat("あ", 5000)},
		{name: "5001文字はエラー", in: strings.Repeat("あ", 5001), wantErr: true},
		{name: "空文字はエラー", in: "", wantErr: true},
		{name: "空白と改行だけはエラー", in: " \n\t ", wantErr: true},
		{name: "改行・タブ以外の制御文字はエラー", in: "良書\x00", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := book.NewComment(tt.in)
			if tt.wantErr {
				if !errors.Is(err, common.ErrInvalid) {
					t.Fatalf("err = %v, want ErrInvalid", err)
				}
				return
			}
			if err != nil || got.String() != tt.want {
				t.Fatalf("got (%q, %v), want (%q, nil)", got.String(), err, tt.want)
			}
		})
	}
}

func TestNewRating(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		in      int
		wantErr bool
	}{
		{name: "下限の1は有効", in: 1},
		{name: "上限の5は有効", in: 5},
		{name: "0はエラー", in: 0, wantErr: true},
		{name: "6はエラー", in: 6, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := book.NewRating(tt.in)
			if tt.wantErr {
				if !errors.Is(err, common.ErrInvalid) {
					t.Fatalf("err = %v, want ErrInvalid", err)
				}
				return
			}
			if err != nil || got.Int() != tt.in {
				t.Fatalf("got (%d, %v), want (%d, nil)", got.Int(), err, tt.in)
			}
		})
	}
}
