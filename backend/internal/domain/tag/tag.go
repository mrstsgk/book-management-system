package tag

import "context"

type ID int64

// Tag は自分が定義した分野タグ。本には ID の集合（book.TagSelection）で付ける。
type Tag struct {
	ID      ID
	Name    Name
	Version int
}

// New は登録前のタグを作る（ID とバージョンは保存時に採番する）。
func New(name Name) *Tag {
	return &Tag{Name: name}
}

// Rename は名前を差し替える。version は更新元が読んだバージョン（楽観的ロックに使う）。
func (t *Tag) Rename(name Name, version int) {
	t.Name = name
	t.Version = version
}

// Repository はタグの書き込み系を永続化するポート。
type Repository interface {
	// FindByID は id のタグを取得する。存在しなければ common.ErrNotFound を返す。
	FindByID(ctx context.Context, id ID) (*Tag, error)
	// Create は t を新規作成し、採番した ID と初期バージョンを t に設定する。同じ名前があれば common.ErrConflict を返す。
	Create(ctx context.Context, t *Tag) error
	// Update は t.Version が一致する行を更新し、t.Version を進める。不一致・同名なら common.ErrConflict を返す。
	Update(ctx context.Context, t *Tag) error
	// Delete は id のタグを削除する。存在しなければ common.ErrNotFound を返す。付いていた本からは自動で外れる。
	Delete(ctx context.Context, id ID) error
}

// TagListItem はタグ一覧の1行分の Read Model。Version は一覧から改名（version が要る）へ進むのに必要
// （個別取得の API が無いため、一覧の応答だけが version の取得元になる）。
type TagListItem struct {
	ID      ID
	Name    string
	Version int
}

// TagList はタグの一覧全体（想定件数が少ないため取得範囲は絞らない）。
type TagList struct {
	Items []*TagListItem
}

// TagBookCount は分野別の集計の1行分の Read Model（公開画面向け。管理画面の TagListItem とは目的が違うため分ける）。
type TagBookCount struct {
	ID        ID
	Name      string
	BookCount int
}

// TagBookCounts は分野別の集計全体。本が付いていないタグは含めない。
type TagBookCounts struct {
	Items []*TagBookCount
}

// Query はタグの参照系ポート。
type Query interface {
	// FindList は全件を返す。
	FindList(ctx context.Context) (*TagList, error)
	// ExistsAll は ids がすべて実在するかを返す（本の登録・更新時の入力検証に使う）。
	ExistsAll(ctx context.Context, ids []ID) (bool, error)
	// CountBooks はタグごとの冊数を、冊数の多い順（同数ならタグ名順）に返す。
	CountBooks(ctx context.Context) (*TagBookCounts, error)
}
