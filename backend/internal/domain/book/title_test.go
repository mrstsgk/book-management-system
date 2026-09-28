package book_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/mrstsgk/book-management-system/backend/internal/domain/book"
	"github.com/mrstsgk/book-management-system/backend/internal/domain/common"
)

func TestNewTitle(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		in      string
		want    string
		wantErr bool
	}{
		{name: "通常のタイトルは有効", in: "人間失格", want: "人間失格"},
		{name: "途中の全角空白は残す", in: "改訂新版　良いコード／悪いコードで学ぶ設計入門", want: "改訂新版　良いコード／悪いコードで学ぶ設計入門"},
		{name: "途中の半角空白は残す", in: "徹底攻略 AWS認定 ソリューションアーキテクト", want: "徹底攻略 AWS認定 ソリューションアーキテクト"},
		{name: "前後の半角・全角空白は取り除く", in: "　 人間失格 　", want: "人間失格"},
		{name: "255文字ちょうどは有効", in: strings.Repeat("あ", 255), want: strings.Repeat("あ", 255)},
		{name: "前後の空白を除いて255文字なら有効", in: "　" + strings.Repeat("あ", 255) + "　", want: strings.Repeat("あ", 255)},
		{name: "256文字はエラー", in: strings.Repeat("あ", 256), wantErr: true},
		{name: "空文字はエラー", in: "", wantErr: true},
		{name: "空白だけはエラー", in: "　 　", wantErr: true},
		{name: "途中のタブはエラー", in: "人間\t失格", wantErr: true},
		{name: "途中の改行はエラー", in: "人間\r\n失格", wantErr: true},
		{name: "先頭のタブはエラー", in: "\t人間失格", wantErr: true},
		{name: "末尾の改行はエラー", in: "人間失格\n", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := book.NewTitle(tt.in)
			if tt.wantErr {
				if !errors.Is(err, common.ErrInvalid) {
					t.Fatalf("err = %v, want ErrInvalid", err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got.String() != tt.want {
				t.Fatalf("String() = %q, want %q", got.String(), tt.want)
			}
		})
	}
}
