package rakuten

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"

	domainbook "github.com/mrstsgk/book-management-system/backend/internal/domain/book"
	"github.com/mrstsgk/book-management-system/backend/internal/domain/common"
)

// DefaultBaseURL は楽天ウェブサービスの API のホスト。取得したデータを表示する画面には楽天のクレジット表示が必要
// （CoverSourceRakuten で見分ける）。
const DefaultBaseURL = "https://openapi.rakuten.co.jp"

const searchPath = "/services/api/BooksBook/Search/20170404"

type catalog struct {
	baseURL       string
	applicationID string
	accessKey     string
	client        *http.Client
	// now は書影の取得日時（保持期限の起点）に使う。テストで固定できるよう注入する。
	now func() time.Time
}

func NewCatalog(baseURL, applicationID, accessKey string, client *http.Client, now func() time.Time) domainbook.BookCatalog {
	return &catalog{baseURL: baseURL, applicationID: applicationID, accessKey: accessKey, client: client, now: now}
}

type response struct {
	Items []struct {
		Title         string `json:"title"`
		Author        string `json:"author"`
		PublisherName string `json:"publisherName"`
		SalesDate     string `json:"salesDate"`
		LargeImageURL string `json:"largeImageUrl"`
		ItemURL       string `json:"itemUrl"`
	} `json:"Items"`
}

// Lookup は楽天ブックス書籍検索 API に ISBN を問い合わせる。該当 0 件は ErrNotFound にする。
func (c *catalog) Lookup(ctx context.Context, isbn domainbook.ISBN) (*domainbook.CatalogEntry, error) {
	q := url.Values{
		"format":        {"json"},
		"formatVersion": {"2"},
		"isbn":          {isbn.String()},
		"applicationId": {c.applicationID},
		"accessKey":     {c.accessKey},
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+searchPath+"?"+q.Encode(), nil)
	if err != nil {
		return nil, err
	}
	res, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("rakuten: %w", withoutURL(err))
	}
	defer func() { _ = res.Body.Close() }()
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("rakuten: unexpected status %d", res.StatusCode)
	}
	var body response
	if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
		return nil, fmt.Errorf("rakuten: decode: %w", err)
	}
	if len(body.Items) == 0 {
		return nil, fmt.Errorf("%w: 楽天ブックスに該当する書籍がありません", common.ErrNotFound)
	}
	it := body.Items[0]
	bib, err := domainbook.NewBibliography(it.Title, it.Author, it.PublisherName, it.SalesDate)
	if err != nil {
		return nil, fmt.Errorf("rakuten: bibliography: %w", err)
	}
	entry := &domainbook.CatalogEntry{ISBN: isbn, Bibliography: bib}
	// 商品ページへリンクできない楽天の書影は規約上使えないので、商品ページが無ければ書影なしにする
	if it.LargeImageURL != "" && it.ItemURL != "" {
		cover, err := domainbook.NewRakutenCover(it.LargeImageURL, it.ItemURL, c.now())
		if err != nil {
			return nil, fmt.Errorf("rakuten: cover: %w", err)
		}
		entry.Cover = &cover
	}
	return entry, nil
}

// withoutURL は *url.Error の URL を落とす。URL には applicationId と accessKey が載るので、ログに漏らさないため。
func withoutURL(err error) error {
	var ue *url.Error
	if errors.As(err, &ue) {
		return fmt.Errorf("%s: %w", ue.Op, ue.Err)
	}
	return err
}
