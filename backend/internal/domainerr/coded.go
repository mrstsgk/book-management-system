package domainerr

// CodedError は、業務ルールの違反に FE が文言を引くためのコードと、文言に差し込む値を付けたエラー。
// 分類（ErrNotFound など）を Unwrap で返すので、errors.Is による HTTP ステータスへの変換は従来どおり働く。
type CodedError struct {
	// Kind は分類（ErrNotFound / ErrConflict / ErrUnauthorized / ErrTooManyAttempts）。
	Kind error
	// Code は業務ルールの違反のコード（lowerCamel。OpenAPI の x-error-codes の code）。
	Code string
	// Message は開発者・ログ向けの説明。
	Message string
	// Params は文言に差し込む値（BE しか知らない値だけ）。無ければ nil。
	Params map[string]any
}

// New は、分類とコードと説明からエラーを作る。
func New(kind error, code, message string) *CodedError {
	return &CodedError{Kind: kind, Code: code, Message: message}
}

// WithParams は、文言に差し込む値を付けたエラーを返す。
func (e *CodedError) WithParams(params map[string]any) *CodedError {
	cp := *e
	cp.Params = params
	return &cp
}

func (e *CodedError) Error() string {
	return e.Kind.Error() + ": " + e.Message
}

func (e *CodedError) Unwrap() error {
	return e.Kind
}
