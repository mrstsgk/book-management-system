package domain

import (
	"fmt"
	"strings"

	"github.com/mrstsgk/book-management-system/backend/internal/domainerr"
)

// Credentials はログイン画面から受け取った ID とパスワード。空を拒否するだけで、正しいかどうかは判定しない。
type Credentials struct {
	// ログイン ID。
	id string
	// 平文のパスワード。
	password string
}

// NewCredentials は空欄を弾く。パスワードはトリムしない（先頭や末尾の空白もパスワードの一部として扱う）。
func NewCredentials(id, password string) (Credentials, error) {
	if strings.TrimSpace(id) == "" {
		return Credentials{}, fmt.Errorf("%w: IDを入力してください", domainerr.ErrInvalid)
	}
	if password == "" {
		return Credentials{}, fmt.Errorf("%w: パスワードを入力してください", domainerr.ErrInvalid)
	}
	return Credentials{id: id, password: password}, nil
}

func (c Credentials) ID() string       { return c.id }
func (c Credentials) Password() string { return c.password }
