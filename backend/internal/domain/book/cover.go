package book

import (
	"fmt"
	"net/url"

	"github.com/mrstsgk/book-management-system/backend/internal/domain/common"
)

// CoverSource は書影の提供元。画面に出すクレジットを決める。
type CoverSource string

const (
	CoverSourceOpenBD CoverSource = "openbd"
)

const coverURLMaxLength = 2048

// Cover は提供元がホストする書影の VO。提供元が認めているのは本の紹介のためにそのまま表示することだけなので、
// 画像そのものは保存・加工しない。
type Cover struct {
	url    string
	source CoverSource
}

// NewCover は提供元の書影を作る。
func NewCover(rawURL string, source CoverSource) (Cover, error) {
	if source != CoverSourceOpenBD {
		return Cover{}, fmt.Errorf("%w: 書影の提供元が不正です", common.ErrInvalid)
	}
	if !validCoverURL(rawURL) {
		return Cover{}, fmt.Errorf("%w: 書影のURLが不正です", common.ErrInvalid)
	}
	return Cover{url: rawURL, source: source}, nil
}

func validCoverURL(raw string) bool {
	u, err := url.Parse(raw)
	return err == nil && len(raw) <= coverURLMaxLength && u.Scheme == "https" && u.Hostname() != "" && u.User == nil
}

func (c Cover) URL() string {
	return c.url
}

func (c Cover) Source() CoverSource {
	return c.source
}
