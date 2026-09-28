package book_test

import (
	"errors"
	"testing"

	"github.com/mrstsgk/book-management-system/backend/internal/domain/book"
	"github.com/mrstsgk/book-management-system/backend/internal/domain/common"
)

func TestNewISBN(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		in      string
		want    string
		wantErr bool
	}{
		{name: "ISBN-13はそのまま保持する", in: "9784873118703", want: "9784873118703"},
		{name: "ハイフンと空白を取り除く", in: " 978-4-87311-870-3 ", want: "9784873118703"},
		{name: "ISBN-10はISBN-13に変換する", in: "4873118700", want: "9784873118703"},
		{name: "末尾がXのISBN-10を受け付ける", in: "486354104X", want: "9784863541047"},
		{name: "末尾が小文字xのISBN-10も受け付ける", in: "486354104x", want: "9784863541047"},
		{name: "979で始まるISBN-13を受け付ける", in: "9791032305690", want: "9791032305690"},
		{name: "ISBN-13のチェックディジット誤りはエラー", in: "9784873118704", wantErr: true},
		{name: "ISBN-10のチェックディジット誤りはエラー", in: "4873118701", wantErr: true},
		{name: "978/979以外で始まる13桁はエラー", in: "1234567890128", wantErr: true},
		{name: "12桁はエラー", in: "978487311870", wantErr: true},
		{name: "数字以外を含むとエラー", in: "97848731187A3", wantErr: true},
		{name: "空文字はエラー", in: "", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := book.NewISBN(tt.in)
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

func TestISBN_ISBN10AndAmazonURL(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		isbn       string
		wantISBN10 string
		wantURL    string
		wantOK     bool
	}{
		{name: "978のISBNはISBN-10と商品ページのURLを返す", isbn: "9784873118703", wantISBN10: "4873118700", wantURL: "https://www.amazon.co.jp/dp/4873118700", wantOK: true},
		{name: "チェックディジットがXになるISBN-10を返す", isbn: "9784863541047", wantISBN10: "486354104X", wantURL: "https://www.amazon.co.jp/dp/486354104X", wantOK: true},
		{name: "979のISBNにはISBN-10もURLも無い", isbn: "9791032305690", wantOK: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			isbn, err := book.NewISBN(tt.isbn)
			if err != nil {
				t.Fatal(err)
			}
			isbn10, ok := isbn.ISBN10()
			if ok != tt.wantOK || isbn10 != tt.wantISBN10 {
				t.Fatalf("ISBN10() = (%q, %v), want (%q, %v)", isbn10, ok, tt.wantISBN10, tt.wantOK)
			}
			u, ok := isbn.AmazonURL()
			if ok != tt.wantOK || u != tt.wantURL {
				t.Fatalf("AmazonURL() = (%q, %v), want (%q, %v)", u, ok, tt.wantURL, tt.wantOK)
			}
		})
	}
}
