package book

import (
	"context"
	"fmt"

	"github.com/mrstsgk/book-management-system/backend/internal/domain/author"
	"github.com/mrstsgk/book-management-system/backend/internal/domain/common"
)

type ID int64

type Book struct {
	ID        ID
	Title     Title
	Price     Price
	AuthorIDs []author.ID
	Status    PublishStatus
	ISBN      *ISBN
	Cover     *Cover
	Version   int
}

func New(title Title, price Price, authorIDs []author.ID, status PublishStatus) (*Book, error) {
	if err := validateAuthorIDs(authorIDs); err != nil {
		return nil, err
	}
	return &Book{Title: title, Price: price, AuthorIDs: authorIDs, Status: status}, nil
}

// Change は書籍の内容を差し替える。出版済みから未出版への変更は拒否し、失敗時は b を変更しない。
func (b *Book) Change(title Title, price Price, authorIDs []author.ID, status PublishStatus, version int) error {
	if err := validateAuthorIDs(authorIDs); err != nil {
		return err
	}
	if !b.Status.CanChangeTo(status) {
		return fmt.Errorf("%w: 出版状況は「出版済み」から「未出版」に変更できません", common.ErrInvalid)
	}
	b.Title = title
	b.Price = price
	b.AuthorIDs = authorIDs
	b.Status = status
	b.Version = version
	return nil
}

// ChangeCatalogInfo sets the ISBN and the cover found for it; nil clears each.
// It keeps copies so later changes to the caller's variables don't alter the book.
func (b *Book) ChangeCatalogInfo(isbn *ISBN, cover *Cover) {
	b.ISBN = copyOf(isbn)
	b.Cover = copyOf(cover)
}

func copyOf[T any](v *T) *T {
	if v == nil {
		return nil
	}
	c := *v
	return &c
}

func validateAuthorIDs(ids []author.ID) error {
	if len(ids) == 0 {
		return fmt.Errorf("%w: 著者は1人以上指定してください", common.ErrInvalid)
	}
	seen := make(map[author.ID]struct{}, len(ids))
	for _, id := range ids {
		if _, ok := seen[id]; ok {
			return fmt.Errorf("%w: 著者IDが重複しています", common.ErrInvalid)
		}
		seen[id] = struct{}{}
	}
	return nil
}

// Repository は Book の書き込み系（Command）を永続化するポート。
// Book と著者との関連（author_book）は1つの集約として、実装側で同一トランザクションに保存する。
type Repository interface {
	// FindByID は id の Book を著者ID込みで取得する。存在しなければ ErrNotFound を返す。
	FindByID(ctx context.Context, id ID) (*Book, error)
	// Create は b と著者との関連を新規作成し、採番された ID と初期バージョンを b に設定する。
	Create(ctx context.Context, b *Book) error
	// Update は b.Version が一致する行と著者との関連を更新し、b.Version を進める。不一致なら ErrConflict を返す。
	Update(ctx context.Context, b *Book) error
}

// AuthorSummary is the read model of an author embedded in BookDetail.
type AuthorSummary struct {
	ID        author.ID
	Name      string
	BirthDate *string // YYYY-MM-DD
	Version   int
}

// BookDetail is the read model of a single book.
type BookDetail struct {
	ID      ID
	Title   string
	Price   int64
	Authors []AuthorSummary
	Status  int
	ISBN    *string
	// AmazonURL is derived from ISBN (not stored); nil when there is no ISBN-10 form.
	AmazonURL   *string
	CoverURL    *string
	CoverSource *string
	Version     int
}

// BookSummary is the read model for a book list; it omits authors and version.
type BookSummary struct {
	ID     ID
	Title  string
	Price  int64
	Status int
}

// Query is the read-side persistence port (Query).
type Query interface {
	// FindDetailByID returns ErrNotFound when the book does not exist.
	FindDetailByID(ctx context.Context, id ID) (*BookDetail, error)
	FindSummariesByAuthorID(ctx context.Context, authorID author.ID) ([]*BookSummary, error)
}

// CatalogEntry is what an external book catalog knows about an ISBN. It is input
// assistance for registering a book, so it keeps the provider's raw text (e.g. an
// authors string that mixes authors and translators) instead of domain VOs.
type CatalogEntry struct {
	ISBN      ISBN
	Title     string
	Authors   string
	Publisher string
	// PublishedOn is as the provider gives it (e.g. "201907" or "20170807").
	PublishedOn string
	// Price is nil when the provider has no usable yen amount.
	Price *int64
	// Cover is nil when no provider has an image for this ISBN.
	Cover *Cover
}

// BookCatalog is the port to external book catalogs such as openBD (ExternalGateway).
type BookCatalog interface {
	// Lookup returns ErrNotFound when no catalog knows the ISBN.
	Lookup(ctx context.Context, isbn ISBN) (*CatalogEntry, error)
}
