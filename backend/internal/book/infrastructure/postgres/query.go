package postgres

import (
	"context"
	"errors"
	"strings"

	"gorm.io/gorm"

	bookdomain "github.com/mrstsgk/book-management-system/backend/internal/book/domain"
	"github.com/mrstsgk/book-management-system/backend/internal/domainerr"
)

// query は、読んだ本の一覧・詳細を PostgreSQL から読み取る（bookdomain.Query の実装）。
type query struct {
	db *gorm.DB
}

// NewQuery は、db を使う Query を作る。
func NewQuery(db *gorm.DB) bookdomain.Query {
	return &query{db: db}
}

// FindDetailByID は、id の本の詳細（タグ名を含む）を 1 件返す。存在しなければ domainerr.ErrNotFound を返す。
func (q *query) FindDetailByID(ctx context.Context, id bookdomain.ID) (*bookdomain.BookDetail, error) {
	var row model
	if err := q.db.WithContext(ctx).First(&row, int64(id)).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domainerr.New(domainerr.ErrNotFound, "bookNotFound", "読んだ本が見つかりません")
		}
		return nil, err
	}
	tagNames, err := tagNamesByBookID(ctx, q.db, []int64{row.ID})
	if err != nil {
		return nil, err
	}
	return &bookdomain.BookDetail{
		ID: bookdomain.ID(row.ID), ISBN: row.ISBN, Title: displayTitle(row.TitleOverride, row.Title), Authors: row.Authors,
		Publisher: row.Publisher, PublishedOn: row.PublishedOn, AmazonURL: amazonURLOf(row.ISBN),
		CoverURL: row.CoverURL, CoverSource: row.CoverSource, TitleOverride: row.TitleOverride, Summary: row.Summary,
		Tags: orEmpty(tagNames[row.ID]), Comment: row.Comment, Rating: row.Rating, Version: row.Version,
	}, nil
}

// ExistsByISBN は同じ ISBN の行が 1 件以上あるかを返す。
func (q *query) ExistsByISBN(ctx context.Context, isbn bookdomain.ISBN) (bool, error) {
	var count int64
	if err := q.db.WithContext(ctx).Model(&model{}).Where("isbn = ?", isbn.String()).Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

// displayTitle は画面に出す書名を決める。自分で上書きした書名があればそれ、無ければ外部カタログの書名。
func displayTitle(override *string, catalogTitle string) string {
	if override != nil {
		return *override
	}
	return catalogTitle
}

// bookTagRow は book_tags と tags を結合した1行（どの本にどのタグ名が付くか）。
type bookTagRow struct {
	BookID  int64
	TagName string
}

// tagNamesByBookID は bookIDs に対応するタグ名を book_id ごとにまとめて返す（1回の別クエリ、N+1にしない）。
func tagNamesByBookID(ctx context.Context, db *gorm.DB, bookIDs []int64) (map[int64][]string, error) {
	result := make(map[int64][]string, len(bookIDs))
	if len(bookIDs) == 0 {
		return result, nil
	}
	var rows []bookTagRow
	err := db.WithContext(ctx).Table("book_tags").
		Select("book_tags.book_id AS book_id, tags.name AS tag_name").
		Joins("JOIN tags ON tags.id = book_tags.tag_id").
		Where("book_tags.book_id IN ?", bookIDs).
		Order("tags.name").
		Find(&rows).Error
	if err != nil {
		return nil, err
	}
	for _, row := range rows {
		result[row.BookID] = append(result[row.BookID], row.TagName)
	}
	return result, nil
}

// orEmpty は nil のスライスを空スライスに揃える（Read Model は「タグ無し」と「未取得」を区別しないため）。
func orEmpty(names []string) []string {
	if names == nil {
		return []string{}
	}
	return names
}

// Search は条件に合う本を新しく登録した順（同時刻は ID の大きい順）に、取得範囲の分だけ返す。総件数も条件に合う件数。
func (q *query) Search(ctx context.Context, c bookdomain.ListCondition, r bookdomain.ListRange) (*bookdomain.BookList, error) {
	db := q.db.WithContext(ctx)
	// 総件数と取得範囲の取得は別の SQL なので、同時に登録されると1件ずれうる。
	// 一覧画面では許容でき、スナップショットのトランザクションより軽い。
	var total int64
	if err := applyCondition(db.Model(&model{}), c).Count(&total).Error; err != nil {
		return nil, err
	}
	var rows []model
	err := applyCondition(db.Model(&model{}), c).
		Select("id, isbn, title, title_override, authors, cover_url, cover_source, summary, rating").
		Order("created_at DESC, id DESC").Limit(r.Limit()).Offset(r.Offset()).
		Find(&rows).Error
	if err != nil {
		return nil, err
	}
	bookIDs := make([]int64, len(rows))
	for i, row := range rows {
		bookIDs[i] = row.ID
	}
	tagNames, err := tagNamesByBookID(ctx, q.db, bookIDs)
	if err != nil {
		return nil, err
	}
	list := &bookdomain.BookList{Items: make([]*bookdomain.BookListItem, 0, len(rows)), Total: int(total)}
	for _, row := range rows {
		list.Items = append(list.Items, &bookdomain.BookListItem{
			ID: bookdomain.ID(row.ID), ISBN: row.ISBN, Title: displayTitle(row.TitleOverride, row.Title), Summary: row.Summary,
			Authors: row.Authors, AmazonURL: amazonURLOf(row.ISBN), CoverURL: row.CoverURL, CoverSource: row.CoverSource,
			Rating: row.Rating, Tags: orEmpty(tagNames[row.ID]),
		})
	}
	return list, nil
}

// applyCondition は一覧の検索・絞り込みの条件を WHERE に足す。
// 書名は画面に出す書名（上書き優先）で探す。上書き前の書名で当たると、画面に見えない書名で当たって利用者が混乱するため。
// タグは JOIN ではなく EXISTS にする（JOIN だと本の行が重複しうるため）。
func applyCondition(tx *gorm.DB, c bookdomain.ListCondition) *gorm.DB {
	if kw := c.Keyword(); kw != "" {
		p := "%" + likeEscaper.Replace(kw) + "%"
		tx = tx.Where(`(COALESCE(title_override, title) ILIKE ? ESCAPE '\' OR authors ILIKE ? ESCAPE '\')`, p, p)
	}
	if id, ok := c.TagID(); ok {
		tx = tx.Where("EXISTS (SELECT 1 FROM book_tags WHERE book_tags.book_id = books.id AND book_tags.tag_id = ?)", int64(id))
	}
	return tx
}

// likeEscaper はキーワード中の LIKE のワイルドカードを文字として扱うためにエスケープする。
var likeEscaper = strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`)

// amazonURLOf は保存済みの ISBN から商品ページのリンクを導出する（bookdomain.ISBN.AmazonURL）。
// 保存済みの ISBN は書き込み時に検証済みなので、解釈できなければリンクなしとする。
func amazonURLOf(isbn string) *string {
	v, err := bookdomain.NewISBN(isbn)
	if err != nil {
		return nil
	}
	u, ok := v.AmazonURL()
	if !ok {
		return nil
	}
	return &u
}
