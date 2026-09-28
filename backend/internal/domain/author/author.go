package author

import (
	"context"

	"github.com/mrstsgk/book-management-system/backend/internal/domain/common"
)

type ID int64

type Author struct {
	ID        ID
	Name      Name
	BirthDate *BirthDate
	Version   int
}

func New(name Name, birthDate *BirthDate) *Author {
	return &Author{Name: name, BirthDate: birthDate}
}

// Repository は Author の書き込み系（Command）を永続化するポート。
type Repository interface {
	// FindByID は id の Author を取得する。存在しなければ ErrNotFound を返す。
	FindByID(ctx context.Context, id ID) (*Author, error)
	// CountByIDs は ids のうち存在する Author の件数を返す。
	CountByIDs(ctx context.Context, ids []ID) (int, error)
	// Create は a を新規作成し、採番された ID と初期バージョンを a に設定する。
	Create(ctx context.Context, a *Author) error
	// Update は a.Version が一致する行を更新し、a.Version を進める。不一致なら ErrConflict を返す。
	Update(ctx context.Context, a *Author) error
}

// Query is the read-side persistence port (Query).
type Query interface {
	Exists(ctx context.Context, id ID) (bool, error)
}

// AuthorListItem は著者一覧の1行分の Read Model。
type AuthorListItem struct {
	ID        ID
	Name      string
	BirthDate *string // YYYY-MM-DD
}

// AuthorList は著者一覧のうち取得範囲の分と、条件に一致した総件数。
type AuthorList struct {
	Items []*AuthorListItem
	Total int
}

// ListQuery は著者一覧（書籍登録時に著者を選ぶ画面など）の参照系ポート。
type ListQuery interface {
	// FindList は名前に nameContains を含む著者（空なら全件）を ID 順に、取得範囲 r の分だけ返す。
	FindList(ctx context.Context, nameContains string, r common.ListRange) (*AuthorList, error)
}

// NameQuery は名前（外部カタログの著者の候補など）を既存の著者に照合する参照系ポート。
type NameQuery interface {
	// FindIDsByNames は名前が完全一致する著者の ID を名前ごとに返す。一致しない名前はマップに含めない。
	// 同名の著者が複数いる場合は最小の ID を使う。
	FindIDsByNames(ctx context.Context, names []string) (map[string]ID, error)
}
