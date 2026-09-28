package book

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/mrstsgk/book-management-system/backend/internal/domain/common"
)

const (
	commentMaxLength = 5000
	ratingMin        = 1
	ratingMax        = 5
)

// Comment は自分が書く感想の VO。前後の空白を除いて1〜5000文字。本文なので改行とタブは使えるが、
// それ以外の制御文字は受け付けない。
type Comment struct {
	value string
}

func NewComment(raw string) (Comment, error) {
	if strings.IndexFunc(raw, isDisallowedInBody) >= 0 {
		return Comment{}, fmt.Errorf("%w: 感想に使用できない文字が含まれています", common.ErrInvalid)
	}
	v := strings.TrimSpace(raw)
	if n := utf8.RuneCountInString(v); n < 1 || n > commentMaxLength {
		return Comment{}, fmt.Errorf("%w: 感想は1〜%d文字で入力してください", common.ErrInvalid, commentMaxLength)
	}
	return Comment{value: v}, nil
}

func (c Comment) String() string {
	return c.value
}

func isDisallowedInBody(r rune) bool {
	return unicode.IsControl(r) && r != '\n' && r != '\r' && r != '\t'
}

// Rating は自分が付ける評価の VO（1〜5 の5段階）。
type Rating struct {
	value int
}

func NewRating(v int) (Rating, error) {
	if v < ratingMin || v > ratingMax {
		return Rating{}, fmt.Errorf("%w: 評価は%d〜%dで指定してください", common.ErrInvalid, ratingMin, ratingMax)
	}
	return Rating{value: v}, nil
}

func (r Rating) Int() int {
	return r.value
}
