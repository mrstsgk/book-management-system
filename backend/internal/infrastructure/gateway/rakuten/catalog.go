package rakuten

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"

	domainbook "github.com/mrstsgk/book-management-system/backend/internal/domain/book"
	"github.com/mrstsgk/book-management-system/backend/internal/domain/common"
)

// DefaultBaseURL is Rakuten Web Service's API host. Screens showing its data must
// carry Rakuten's credit (told apart by CoverSourceRakuten).
const DefaultBaseURL = "https://openapi.rakuten.co.jp"

const searchPath = "/services/api/BooksBook/Search/20170404"

type catalog struct {
	baseURL       string
	applicationID string
	accessKey     string
	client        *http.Client
}

func NewCatalog(baseURL, applicationID, accessKey string, client *http.Client) domainbook.BookCatalog {
	return &catalog{baseURL: baseURL, applicationID: applicationID, accessKey: accessKey, client: client}
}

type response struct {
	Items []struct {
		Title         string `json:"title"`
		Author        string `json:"author"`
		PublisherName string `json:"publisherName"`
		SalesDate     string `json:"salesDate"`
		ItemPrice     int64  `json:"itemPrice"`
		LargeImageURL string `json:"largeImageUrl"`
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
	entry := &domainbook.CatalogEntry{
		ISBN:        isbn,
		Title:       it.Title,
		Authors:     it.Author,
		Publisher:   it.PublisherName,
		PublishedOn: it.SalesDate,
	}
	if it.ItemPrice > 0 {
		p := it.ItemPrice
		entry.Price = &p
	}
	if it.LargeImageURL != "" {
		cover, err := domainbook.NewCover(it.LargeImageURL, domainbook.CoverSourceRakuten)
		if err != nil {
			return nil, fmt.Errorf("rakuten: cover: %w", err)
		}
		entry.Cover = &cover
	}
	return entry, nil
}

// withoutURL drops *url.Error's URL, which carries applicationId/accessKey and would leak them into logs.
func withoutURL(err error) error {
	var ue *url.Error
	if errors.As(err, &ue) {
		return fmt.Errorf("%s: %w", ue.Op, ue.Err)
	}
	return err
}
