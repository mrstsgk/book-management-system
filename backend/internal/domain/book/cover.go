package book

import (
	"fmt"
	"net/url"

	"github.com/mrstsgk/book-management-system/backend/internal/domain/common"
)

// CoverSource is where a cover image comes from. It decides the credit a screen must show
// (e.g. Rakuten Web Service requires one) and which service to ask when refreshing.
type CoverSource string

const (
	CoverSourceOpenBD  CoverSource = "openbd"
	CoverSourceRakuten CoverSource = "rakuten"
)

const coverURLMaxLength = 2048

// Cover is a value object for a cover image hosted by its provider. The image itself is
// never copied: providers only allow showing it as-is for introducing the book.
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
