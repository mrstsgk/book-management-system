// Package googlebooks は Google Books API から書影だけを取る。書誌は openBD に一本化しているので使わない。
package googlebooks

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

// DefaultBaseURL は Google Books API。キーの無い呼び出しは全員で1つの日次枠を共有しており使い切られているため、キー前提にする。
const DefaultBaseURL = "https://www.googleapis.com"

// Cover は Google Books が持つ書影。PR-C で domain の Cover に詰め替える。
type Cover struct {
	ImageURL string // https に直した imageLinks.thumbnail
	PageURL  string // volumeInfo.infoLink（その本の Google Books のページ）
}

type Client struct {
	baseURL string
	apiKey  string
	client  *http.Client
}

// NewClient は baseURL（既定 https://www.googleapis.com）と API キーで作る。
func NewClient(baseURL, apiKey string, httpClient *http.Client) *Client {
	if baseURL == "" {
		baseURL = DefaultBaseURL
	}
	return &Client{baseURL: baseURL, apiKey: apiKey, client: httpClient}
}

// FindCover は ISBN の書影を返す。該当なし・imageLinks なしは (nil, nil)。通信・HTTP の失敗は error。
func (c *Client) FindCover(ctx context.Context, isbn string) (*Cover, error) {
	q := url.Values{"q": {"isbn:" + isbn}, "key": {c.apiKey}}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/books/v1/volumes?"+q.Encode(), nil)
	if err != nil {
		// 解析エラーはキー入りの URL をそのまま含むので、元のエラーは包まない
		return nil, errors.New("googlebooks: invalid request URL")
	}
	res, err := c.client.Do(req)
	if err != nil {
		// url.Error はキーを含む URL を持つので、エラーメッセージにキーを出さないよう中身だけ包む
		var uerr *url.Error
		if errors.As(err, &uerr) {
			err = uerr.Err
		}
		return nil, fmt.Errorf("googlebooks: %w", err)
	}
	defer func() { _ = res.Body.Close() }()
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("googlebooks: unexpected status %d", res.StatusCode)
	}
	var body volumes
	if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
		return nil, fmt.Errorf("googlebooks: decode: %w", err)
	}
	if len(body.Items) == 0 {
		return nil, nil
	}
	v := body.Items[0].VolumeInfo
	// Google の規約では書影と一緒に本ごとの Google Books へのリンクが要るので、リンクが無ければ書影として使わない
	if v.ImageLinks.Thumbnail == "" || v.InfoLink == "" {
		return nil, nil
	}
	return &Cover{ImageURL: toHTTPS(v.ImageLinks.Thumbnail), PageURL: toHTTPS(v.InfoLink)}, nil
}

type volumes struct {
	Items []struct {
		VolumeInfo struct {
			ImageLinks struct {
				Thumbnail string `json:"thumbnail"`
			} `json:"imageLinks"`
			InfoLink string `json:"infoLink"`
		} `json:"volumeInfo"`
	} `json:"items"`
}

// toHTTPS は API が http で返す URL を https にする（同じ URL が https でも配信されており、画面は https だけを出すため）。
func toHTTPS(raw string) string {
	if rest, ok := strings.CutPrefix(raw, "http://"); ok {
		return "https://" + rest
	}
	return raw
}
