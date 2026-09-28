package openbd

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	domainbook "github.com/mrstsgk/book-management-system/backend/internal/domain/book"
	"github.com/mrstsgk/book-management-system/backend/internal/domain/common"
	"github.com/mrstsgk/book-management-system/backend/internal/infrastructure/gateway/price"
)

// DefaultBaseURL is openBD's public API. It needs no key; its terms allow the data
// (including covers) only for promoting or introducing books.
const DefaultBaseURL = "https://api.openbd.jp"

type catalog struct {
	baseURL string
	client  *http.Client
}

func NewCatalog(baseURL string, client *http.Client) domainbook.BookCatalog {
	return &catalog{baseURL: baseURL, client: client}
}

type record struct {
	Summary struct {
		Title     string `json:"title"`
		Author    string `json:"author"`
		Publisher string `json:"publisher"`
		PubDate   string `json:"pubdate"`
		Cover     string `json:"cover"`
	} `json:"summary"`
	Onix struct {
		ProductSupply struct {
			SupplyDetail struct {
				Price []struct {
					PriceAmount string `json:"PriceAmount"`
				} `json:"Price"`
			} `json:"SupplyDetail"`
		} `json:"ProductSupply"`
	} `json:"onix"`
}

// Lookup は openBD に ISBN を問い合わせる。openBD は該当なしを配列の null で返すので、それを ErrNotFound にする。
func (c *catalog) Lookup(ctx context.Context, isbn domainbook.ISBN) (*domainbook.CatalogEntry, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet,
		c.baseURL+"/v1/get?isbn="+url.QueryEscape(isbn.String()), nil)
	if err != nil {
		return nil, err
	}
	res, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("openbd: %w", err)
	}
	defer func() { _ = res.Body.Close() }()
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("openbd: unexpected status %d", res.StatusCode)
	}
	var records []*record
	if err := json.NewDecoder(res.Body).Decode(&records); err != nil {
		return nil, fmt.Errorf("openbd: decode: %w", err)
	}
	if len(records) == 0 || records[0] == nil {
		return nil, fmt.Errorf("%w: openBDに該当する書籍がありません", common.ErrNotFound)
	}
	r := records[0]
	entry := &domainbook.CatalogEntry{
		ISBN:        isbn,
		Title:       r.Summary.Title,
		Authors:     r.Summary.Author,
		Publisher:   r.Summary.Publisher,
		PublishedOn: r.Summary.PubDate,
	}
	if prices := r.Onix.ProductSupply.SupplyDetail.Price; len(prices) > 0 {
		entry.Price = price.ParseYen(prices[0].PriceAmount)
	}
	if r.Summary.Cover != "" {
		cover, err := domainbook.NewCover(r.Summary.Cover, domainbook.CoverSourceOpenBD)
		if err != nil {
			return nil, fmt.Errorf("openbd: cover: %w", err)
		}
		entry.Cover = &cover
	}
	return entry, nil
}
