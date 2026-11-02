package domainerr

import "errors"

// UseCase・Infrastructure が返す業務の分類エラー。HTTP ステータスへの変換は Presentation の HTTPErrorHandler だけが行う。
var (
	ErrNotFound        = errors.New("not found")         // → 404
	ErrInvalid         = errors.New("invalid")           // → 400
	ErrConflict        = errors.New("conflict")          // → 409
	ErrUnauthorized    = errors.New("unauthorized")      // → 401
	ErrTooManyAttempts = errors.New("too many attempts") // → 429
)
