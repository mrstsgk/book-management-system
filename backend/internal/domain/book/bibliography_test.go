package book_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/mrstsgk/book-management-system/backend/internal/domain/book"
	"github.com/mrstsgk/book-management-system/backend/internal/domain/common"
)

func TestNewBibliography(t *testing.T) {
	t.Parallel()
	type in struct{ title, authors, publisher, publishedOn string }
	tests := []struct {
		name    string
		in      in
		want    in
		wantErr bool
	}{
		{name: "すべて指定すると前後の空白を除いて保持する",
			in:   in{" データ指向アプリケーションデザイン ", "Kleppmann,Martin 斉藤,太郎", "オーム社", "201907"},
			want: in{"データ指向アプリケーションデザイン", "Kleppmann,Martin 斉藤,太郎", "オーム社", "201907"}},
		{name: "書名の途中の全角空白は残す", in: in{"改訂新版　良いコード／悪いコードで学ぶ設計入門", "", "", ""}, want: in{"改訂新版　良いコード／悪いコードで学ぶ設計入門", "", "", ""}},
		{name: "著者・出版社・発売日は空でもよい", in: in{"徹底攻略 AWS認定", "", "", ""}, want: in{"徹底攻略 AWS認定", "", "", ""}},
		{name: "書名255文字ちょうどは有効", in: in{strings.Repeat("あ", 255), "", "", ""}, want: in{strings.Repeat("あ", 255), "", "", ""}},
		{name: "書名256文字はエラー", in: in{strings.Repeat("あ", 256), "", "", ""}, wantErr: true},
		{name: "書名が空白だけはエラー", in: in{"　 ", "", "", ""}, wantErr: true},
		{name: "著者500文字ちょうどは有効", in: in{"本", strings.Repeat("あ", 500), "", ""}, want: in{"本", strings.Repeat("あ", 500), "", ""}},
		{name: "著者501文字はエラー", in: in{"本", strings.Repeat("あ", 501), "", ""}, wantErr: true},
		{name: "出版社256文字はエラー", in: in{"本", "", strings.Repeat("あ", 256), ""}, wantErr: true},
		{name: "発売日33文字はエラー", in: in{"本", "", "", strings.Repeat("1", 33)}, wantErr: true},
		{name: "書名の改行はエラー", in: in{"人間\n失格", "", "", ""}, wantErr: true},
		{name: "著者の前後のタブもエラー", in: in{"本", "\t太宰治", "", ""}, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := book.NewBibliography(tt.in.title, tt.in.authors, tt.in.publisher, tt.in.publishedOn)
			if tt.wantErr {
				if !errors.Is(err, common.ErrInvalid) {
					t.Fatalf("err = %v, want ErrInvalid", err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if g := (in{got.Title(), got.Authors(), got.Publisher(), got.PublishedOn()}); g != tt.want {
				t.Fatalf("got %+v, want %+v", g, tt.want)
			}
		})
	}
}
