package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"time"
)

const (
	// IdleTimeout は操作が無いまま失効するまでの時間。書き込み API を通るたびにここまで延びる。
	IdleTimeout = time.Hour
	// AbsoluteTimeout はログインからの上限。延長しても超えない（盗まれたセッションが延命され続けないため）。
	AbsoluteTimeout = 24 * time.Hour
)

// SessionID はブラウザに渡す唯一の値。推測できないよう乱数から作る。
type SessionID string

// NewSessionID は 32 byte の乱数を base64url（パディング無し・43 文字）にする。
func NewSessionID() (SessionID, error) {
	var b [32]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", err
	}
	return SessionID(base64.RawURLEncoding.EncodeToString(b[:])), nil
}

// Session はログイン済みの状態。ID 以外はいつまで有効かだけを持つ（誰のセッションかは、利用者が自分 1 人なので持たない）。
type Session struct {
	ID                SessionID
	ExpiresAt         time.Time
	AbsoluteExpiresAt time.Time
}

func NewSession(id SessionID, now time.Time) *Session {
	return &Session{ID: id, ExpiresAt: now.Add(IdleTimeout), AbsoluteExpiresAt: now.Add(AbsoluteTimeout)}
}

// IsValid はアイドル期限に達していなければ true（期限ちょうどは無効）。Extend が絶対期限で頭打ちにするので、
// アイドル期限は常に絶対期限以下であり、絶対期限を別に比べる必要は無い。
func (s *Session) IsValid(now time.Time) bool {
	return now.Before(s.ExpiresAt)
}

// Extend はアイドル期限を今 + IdleTimeout に延ばす。絶対期限は超えない。
func (s *Session) Extend(now time.Time) {
	next := now.Add(IdleTimeout)
	if next.After(s.AbsoluteExpiresAt) {
		next = s.AbsoluteExpiresAt
	}
	s.ExpiresAt = next
}

// SessionRepository はセッションを永続化するポート。
type SessionRepository interface {
	// Save は s を新規保存する。
	Save(ctx context.Context, s *Session) error
	// Find は id のセッションを返す。無ければ common.ErrNotFound を返す（期限切れでも行があれば返す）。
	Find(ctx context.Context, id SessionID) (*Session, error)
	// Delete は id のセッションを消す。無くてもエラーにしない。
	Delete(ctx context.Context, id SessionID) error
	// UpdateExpiry は id のアイドル期限を expiresAt にする。
	UpdateExpiry(ctx context.Context, id SessionID, expiresAt time.Time) error
}

// PasswordVerifier はパスワードとハッシュの照合。ハッシュ方式（bcrypt）を Domain から隠すための ExternalGateway 相当。
type PasswordVerifier interface {
	// Matches は password が hash に対応していれば true。
	Matches(hash, password string) bool
}
