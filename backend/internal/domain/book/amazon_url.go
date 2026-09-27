package book

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/mrstsgk/book-management-system/backend/internal/domain/common"
)

const amazonURLMaxLength = 2048

// Short-link hosts Amazon hands out from its share button.
var amazonShortHosts = map[string]bool{"amzn.asia": true, "amzn.to": true}

// Store domains; subdomains such as www. and smile. are accepted too.
var amazonStoreDomains = []string{"amazon.co.jp", "amazon.com"}

// AmazonURL is a value object for a book's Amazon product link (https on an Amazon host only).
type AmazonURL struct {
	value string
}

func NewAmazonURL(raw string) (AmazonURL, error) {
	if len(raw) > amazonURLMaxLength {
		return AmazonURL{}, fmt.Errorf("%w: AmazonのURLは%d文字以内で入力してください", common.ErrInvalid, amazonURLMaxLength)
	}
	u, err := url.Parse(raw)
	// Userinfo is rejected so "https://amazon.co.jp@evil.example/" can't pass as Amazon.
	if err != nil || u.Scheme != "https" || u.User != nil || u.Port() != "" || !isAmazonHost(u.Hostname()) {
		return AmazonURL{}, fmt.Errorf("%w: AmazonのURL（https://www.amazon.co.jp/... など）を入力してください", common.ErrInvalid)
	}
	return AmazonURL{value: raw}, nil
}

func (u AmazonURL) String() string {
	return u.value
}

func isAmazonHost(host string) bool {
	host = strings.ToLower(host)
	if amazonShortHosts[host] {
		return true
	}
	for _, d := range amazonStoreDomains {
		if host == d || strings.HasSuffix(host, "."+d) {
			return true
		}
	}
	return false
}
