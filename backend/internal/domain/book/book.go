package book

import (
	"context"

	"github.com/mrstsgk/book-management-system/backend/internal/domain/common"
)

type ID int64

// Book は自分が読んだ本1冊。外部カタログから取得した書誌・書影と、自分が書いた感想・評価を持つ。
type Book struct {
	ID           ID
	ISBN         ISBN
	Bibliography Bibliography
	// Cover は書影。提供元に書影が無い本は nil。
	Cover *Cover
	// TitleOverride は自分で上書きした書名。上書きしていなければ nil。外部カタログを取り直しても消さない。
	TitleOverride *Title
	Summary       Summary
	Comment       Comment
	Rating        Rating
	Tags          TagSelection
	Version       int
}

// New は登録前の読んだ本を作る（ID とバージョンは保存時に採番する）。
func New(isbn ISBN, bibliography Bibliography, cover *Cover, summary Summary, comment Comment, rating Rating, tags TagSelection) *Book {
	return &Book{ISBN: isbn, Bibliography: bibliography, Cover: copyOf(cover), Summary: summary, Comment: comment, Rating: rating, Tags: tags}
}

// ChangeReview は一言まとめ・感想・評価・分野タグを差し替える。version は更新元が読んだバージョン（楽観的ロックに使う）。
func (b *Book) ChangeReview(summary Summary, comment Comment, rating Rating, tags TagSelection, version int) {
	b.Summary = summary
	b.Comment = comment
	b.Rating = rating
	b.Tags = tags
	b.Version = version
}

// OverrideTitle は書名を自分で上書きする。nil なら上書きを外し、外部カタログの書名に戻す。
func (b *Book) OverrideTitle(title *Title) {
	b.TitleOverride = copyOf(title)
}

// RefreshCatalog は外部カタログから取り直した書誌と書影に差し替える（提供元の変更を反映するため）。
func (b *Book) RefreshCatalog(bibliography Bibliography, cover *Cover) {
	b.Bibliography = bibliography
	b.Cover = copyOf(cover)
}

// copyOf は呼び出し側の変数を後から変えても本が変わらないよう、値をコピーして持つ。
func copyOf[T any](v *T) *T {
	if v == nil {
		return nil
	}
	c := *v
	return &c
}

// Repository は読んだ本の書き込み系（Command）を永続化するポート。
type Repository interface {
	// FindByID は id の本を取得する。存在しなければ ErrNotFound を返す。
	FindByID(ctx context.Context, id ID) (*Book, error)
	// Create は b を新規作成し、採番した ID と初期バージョンを b に設定する。同じ ISBN の本があれば ErrConflict を返す。
	Create(ctx context.Context, b *Book) error
	// CreateAll は books を1つのトランザクションで全冊作成する。1冊でも同じ ISBN があれば1冊も残さず ErrConflict を返す。
	CreateAll(ctx context.Context, books []*Book) error
	// Update は b.Version が一致する行を更新し、b.Version を進める。不一致なら ErrConflict を返す。
	Update(ctx context.Context, b *Book) error
	// Delete は id の本を削除する。存在しなければ ErrNotFound を返す。
	Delete(ctx context.Context, id ID) error
}

// BookDetail は読んだ本1冊の Read Model。
type BookDetail struct {
	ID   ID
	ISBN string
	// Title は表示する書名（上書きがあればそれ、無ければ外部カタログの書名）。
	Title       string
	Authors     string
	Publisher   string
	PublishedOn string
	// AmazonURL は ISBN から導出する（保存しない）。ISBN-10 の形式が無ければ nil。
	AmazonURL *string
	// CoverURL・CoverSource は書影が無ければ nil。
	CoverURL    *string
	CoverSource *string
	// TitleOverride は自分で上書きした書名。上書きしていなければ nil（編集画面が今の上書きを送り直すのに使う）。
	TitleOverride *string
	Summary       string
	Tags          []string
	Comment       string
	Rating        int
	Version       int
}

// BookListItem は読んだ本の一覧の1行分の Read Model。感想の本文は詳細でだけ返す。
type BookListItem struct {
	ID   ID
	ISBN string
	// Title は表示する書名（上書きがあればそれ、無ければ外部カタログの書名）。
	Title       string
	Summary     string
	Tags        []string
	Authors     string
	AmazonURL   *string
	CoverURL    *string
	CoverSource *string
	Rating      int
}

// BookList は一覧のうち取得範囲の分と、全体の総件数。
type BookList struct {
	Items []*BookListItem
	Total int
}

// Query は読んだ本の参照系ポート。
type Query interface {
	// FindDetailByID は存在しなければ ErrNotFound を返す。
	FindDetailByID(ctx context.Context, id ID) (*BookDetail, error)
	// FindList は条件 c に合う本を新しく登録した順に、取得範囲 r の分だけ返す。総件数も c に合う件数。
	FindList(ctx context.Context, c ListCondition, r common.ListRange) (*BookList, error)
}

// CatalogEntry は外部カタログが ISBN について持っている書誌と書影。
type CatalogEntry struct {
	ISBN         ISBN
	Bibliography Bibliography
	// Cover はどの提供元にも書影が無ければ nil。
	Cover *Cover
}

// BookCatalog は外部の書籍カタログ（openBD など）へのポート（ExternalGateway）。
type BookCatalog interface {
	// Lookup はどのカタログにも ISBN が無ければ ErrNotFound を返す。
	Lookup(ctx context.Context, isbn ISBN) (*CatalogEntry, error)
}
