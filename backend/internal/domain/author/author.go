package author

import "context"

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
