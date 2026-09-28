package book

import (
	"fmt"
	"strings"

	"github.com/mrstsgk/book-management-system/backend/internal/domain/common"
)

// ISBN は書籍の ISBN の VO。常に13桁で保持し、10桁の入力は 978 付きの13桁に変換する。
type ISBN struct {
	value string
}

func NewISBN(raw string) (ISBN, error) {
	s := strings.NewReplacer("-", "", " ", "").Replace(strings.TrimSpace(raw))
	switch {
	case len(s) == 13 && isDigits(s) && (strings.HasPrefix(s, "978") || strings.HasPrefix(s, "979")) && s[12] == isbn13CheckDigit(s[:12]):
		return ISBN{value: s}, nil
	case len(s) == 10 && isDigits(s[:9]) && strings.ToUpper(s[9:]) == isbn10CheckDigit(s[:9]):
		body := "978" + s[:9]
		return ISBN{value: body + string(isbn13CheckDigit(body))}, nil
	}
	return ISBN{}, fmt.Errorf("%w: ISBNが不正です（13桁または10桁で、チェックディジットが正しいもの）", common.ErrInvalid)
}

func (i ISBN) String() string {
	return i.value
}

// ISBN10 は10桁の形式を返す。10桁の形式があるのは 978 で始まる ISBN だけで、979 には無い。
func (i ISBN) ISBN10() (string, bool) {
	if !strings.HasPrefix(i.value, "978") {
		return "", false
	}
	body := i.value[3:12]
	return body + isbn10CheckDigit(body), true
}

// AmazonURL は amazon.co.jp の商品ページを返す。紙の本の ASIN は ISBN-10 と同じなので、
// 手入力させずに導出する。
func (i ISBN) AmazonURL() (string, bool) {
	isbn10, ok := i.ISBN10()
	if !ok {
		return "", false
	}
	return "https://www.amazon.co.jp/dp/" + isbn10, true
}

func isDigits(s string) bool {
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

// isbn13CheckDigit はISBN-13の先頭12桁から、重み1と3を交互に掛けた和でチェックディジットを求める。
func isbn13CheckDigit(first12 string) byte {
	sum := 0
	for i, r := range first12 {
		d := int(r - '0')
		if i%2 == 1 {
			d *= 3
		}
		sum += d
	}
	return byte('0' + (10-sum%10)%10)
}

// isbn10CheckDigit はISBN-10の先頭9桁から、重み10〜2を掛けた和でチェックディジット（10はX）を求める。
func isbn10CheckDigit(first9 string) string {
	sum := 0
	for i, r := range first9 {
		sum += int(r-'0') * (10 - i)
	}
	switch c := (11 - sum%11) % 11; c {
	case 10:
		return "X"
	default:
		return string(rune('0' + c))
	}
}
