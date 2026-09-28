package book

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/mrstsgk/book-management-system/backend/internal/domain/common"
)

const (
	titleMaxLength       = 255
	authorsMaxLength     = 500
	publisherMaxLength   = 255
	publishedOnMaxLength = 32
)

// Bibliography は外部カタログから取得した書誌の VO。著者は独立したエンティティにせず、
// 提供元の文字列のまま持つ（著者と訳者の区別が無く、書式も提供元ごとに違うため）。
type Bibliography struct {
	title       string
	authors     string
	publisher   string
	publishedOn string
}

// NewBibliography は書誌を作る。書名は必須、それ以外は空でもよい（著者情報が無い本もある）。
// いずれも前後の空白を除き、タブや改行などの制御文字は受け付けない。
func NewBibliography(title, authors, publisher, publishedOn string) (Bibliography, error) {
	t, err := text("書名", title, 1, titleMaxLength)
	if err != nil {
		return Bibliography{}, err
	}
	a, err := text("著者", authors, 0, authorsMaxLength)
	if err != nil {
		return Bibliography{}, err
	}
	p, err := text("出版社", publisher, 0, publisherMaxLength)
	if err != nil {
		return Bibliography{}, err
	}
	d, err := text("発売日", publishedOn, 0, publishedOnMaxLength)
	if err != nil {
		return Bibliography{}, err
	}
	return Bibliography{title: t, authors: a, publisher: p, publishedOn: d}, nil
}

func (b Bibliography) Title() string       { return b.title }
func (b Bibliography) Authors() string     { return b.authors }
func (b Bibliography) Publisher() string   { return b.publisher }
func (b Bibliography) PublishedOn() string { return b.publishedOn }

// text は前後の空白を除いた1行の文字列を、文字数の範囲と制御文字の有無で検証する。
func text(label, raw string, minLen, maxLen int) (string, error) {
	// 前後にあるタブや改行も、黙って取り除かずに拒否するためトリム前に検査する
	if strings.IndexFunc(raw, unicode.IsControl) >= 0 {
		return "", fmt.Errorf("%w: %sにタブや改行は使用できません", common.ErrInvalid, label)
	}
	v := strings.TrimSpace(raw)
	if n := utf8.RuneCountInString(v); n < minLen || n > maxLen {
		return "", fmt.Errorf("%w: %sは%d〜%d文字にしてください", common.ErrInvalid, label, minLen, maxLen)
	}
	return v, nil
}
