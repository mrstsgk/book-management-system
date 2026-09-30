package command

import (
	"context"
	"crypto/subtle"
	"fmt"
	"sync"
	"time"

	"github.com/mrstsgk/book-management-system/backend/internal/domain/auth"
	"github.com/mrstsgk/book-management-system/backend/internal/domain/common"
)

const (
	maxFailures  = 5
	lockDuration = time.Minute
)

// LoginCommand は LoginUsecase の入力。データだけを持ち、ロジックは持たない。
type LoginCommand struct {
	ID       string
	Password string
}

// AdminAccount は環境変数から読んだ管理者の ID とパスワードのハッシュ。
type AdminAccount struct {
	ID           string
	PasswordHash string
}

type LoginUsecase interface {
	// Execute は照合に成功したらセッションを発行し、その ID を返す。
	Execute(ctx context.Context, cmd LoginCommand) (string, error)
}

// LoginUsecaseImpl は 1 インスタンスを全リクエストで共有する（失敗回数を持つため）。
type LoginUsecaseImpl struct {
	Admin    AdminAccount
	Verifier auth.PasswordVerifier
	Sessions auth.SessionRepository
	Now      func() time.Time

	// 総当たり対策。利用者は 1 人なので ID ごとに分けない。
	// ponytail: プロセス内メモリ。再起動で消え、複数プロセスでは共有されない。公開して多プロセスにするなら DB に移す
	mu          sync.Mutex
	failures    int
	lockedUntil time.Time
}

// Execute は空欄 → ロック → ID とパスワードの照合 → セッション保存の順に進む。
// ID が違ってもパスワードの照合を必ず行い、応答時間から ID の当たり外れが分からないようにする。
func (u *LoginUsecaseImpl) Execute(ctx context.Context, cmd LoginCommand) (string, error) {
	creds, err := auth.NewCredentials(cmd.ID, cmd.Password)
	if err != nil {
		return "", err
	}
	now := u.Now()
	if u.locked(now) {
		return "", fmt.Errorf("%w: しばらく待ってからやり直してください", common.ErrTooManyAttempts)
	}
	if !u.matches(creds) {
		u.fail(now)
		return "", fmt.Errorf("%w: IDかパスワードが違います", common.ErrUnauthorized)
	}
	u.reset()

	id, err := auth.NewSessionID()
	if err != nil {
		return "", err
	}
	s := auth.NewSession(id, now)
	if err := u.Sessions.Save(ctx, s); err != nil {
		return "", err
	}
	return string(id), nil
}

// matches は ID を定数時間で比べ、その結果に関わらずパスワードも照合する。設定が空なら常に false。
func (u *LoginUsecaseImpl) matches(creds auth.Credentials) bool {
	idOK := subtle.ConstantTimeCompare([]byte(creds.ID()), []byte(u.Admin.ID)) == 1
	pwOK := u.Verifier.Matches(u.Admin.PasswordHash, creds.Password())
	return u.Admin.ID != "" && u.Admin.PasswordHash != "" && idOK && pwOK
}

func (u *LoginUsecaseImpl) locked(now time.Time) bool {
	u.mu.Lock()
	defer u.mu.Unlock()
	return now.Before(u.lockedUntil)
}

// fail は失敗を数え、maxFailures に達したら lockDuration の間ロックして回数を戻す。
func (u *LoginUsecaseImpl) fail(now time.Time) {
	u.mu.Lock()
	defer u.mu.Unlock()
	u.failures++
	if u.failures >= maxFailures {
		u.lockedUntil = now.Add(lockDuration)
		u.failures = 0
	}
}

func (u *LoginUsecaseImpl) reset() {
	u.mu.Lock()
	defer u.mu.Unlock()
	u.failures = 0
}
