package book

import (
	"fmt"
	"net/url"

	"github.com/mrstsgk/book-management-system/backend/internal/domain/common"
)

// CoverSource は書影の提供元。画面に出すクレジット（楽天ウェブサービスは表記が必須）と、
// 取り直すときにどのサービスに問い合わせるかを決める。
type CoverSource string

const (
	CoverSourceOpenBD  CoverSource = "openbd"
	CoverSourceRakuten CoverSource = "rakuten"
)

const coverURLMaxLength = 2048

// Cover は提供元がホストする書影の VO。提供元が認めているのは本の紹介のためにそのまま表示することだけなので、
// 画像そのものは保存・加工しない。
type Cover struct {
	url    string
	source CoverSource
}

func NewCover(rawURL string, source CoverSource) (Cover, error) {
	if source != CoverSourceOpenBD && source != CoverSourceRakuten {
		return Cover{}, fmt.Errorf("%w: 書影の提供元が不正です", common.ErrInvalid)
	}
	u, err := url.Parse(rawURL)
	if err != nil || len(rawURL) > coverURLMaxLength || u.Scheme != "https" || u.Hostname() == "" || u.User != nil {
		return Cover{}, fmt.Errorf("%w: 書影のURLが不正です", common.ErrInvalid)
	}
	return Cover{url: rawURL, source: source}, nil
}

func (c Cover) URL() string {
	return c.url
}

func (c Cover) Source() CoverSource {
	return c.source
}
