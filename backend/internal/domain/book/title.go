package book

// Title は自分で上書きする書名の VO。規則は外部カタログの書名（Bibliography）と同じなので、同じ検証を使う。
type Title struct {
	value string
}

// NewTitle は前後の空白を除いて1〜255文字の書名を作る。タブや改行などの制御文字は受け付けない。
func NewTitle(raw string) (Title, error) {
	v, err := text("書名", raw, 1, titleMaxLength)
	if err != nil {
		return Title{}, err
	}
	return Title{value: v}, nil
}

func (t Title) String() string {
	return t.value
}
