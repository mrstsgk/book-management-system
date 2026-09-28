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
		{name: "前後の空白を除いて保持する", in: " 徹底攻略 AWS認定 ソリューションアーキテクト アソシエイト教科書 第3版 ", want: "徹底攻略 AWS認定 ソリューションアーキテクト アソシエイト教科書 第3版"},
		{name: "途中の全角空白は残す", in: "改訂新版　良いコード／悪いコードで学ぶ設計入門", want: "改訂新版　良いコード／悪いコードで学ぶ設計入門"},
		{name: "1文字は有効", in: "本", want: "本"},
		{name: "255文字ちょうどは有効", in: strings.Repeat("あ", 255), want: strings.Repeat("あ", 255)},
		{name: "256文字はエラー", in: strings.Repeat("あ", 256), wantErr: true},
		{name: "空文字はエラー", in: "", wantErr: true},
		{name: "空白だけはエラー", in: "　 ", wantErr: true},
		{name: "改行はエラー", in: "人間\n失格", wantErr: true},
		{name: "前後のタブもエラー", in: "\t人間失格", wantErr: true},
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
				t.Fatalf("got %q, want %q", got.String(), tt.want)
			}
		})
	}
}
