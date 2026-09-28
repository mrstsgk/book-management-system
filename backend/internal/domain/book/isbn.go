package book

import (
	"fmt"
	"strings"

	"github.com/mrstsgk/book-management-system/backend/internal/domain/common"
)

// ISBN is a value object for a book's ISBN, always held as the 13-digit form.
// ISBN-10 input is converted to ISBN-13 (978 prefix).
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

// ISBN10 returns the 10-digit form. Only 978-prefixed ISBNs have one; 979 ones do not.
func (i ISBN) ISBN10() (string, bool) {
	if !strings.HasPrefix(i.value, "978") {
		return "", false
	}
	body := i.value[3:12]
	return body + isbn10CheckDigit(body), true
}

// AmazonURL returns the amazon.co.jp product page. For printed books Amazon's
// ASIN is the ISBN-10, so the link is derived rather than entered by hand.
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
