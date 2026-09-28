package book

const summaryMaxLength = 100

// Summary は一覧で見せる一言まとめの VO。1行の短い文なので、感想と違い改行も受け付けない。
type Summary struct {
	value string
}

// NewSummary は前後の空白を除いて1〜100文字の一言まとめを作る。
func NewSummary(raw string) (Summary, error) {
	v, err := text("一言まとめ", raw, 1, summaryMaxLength)
	if err != nil {
		return Summary{}, err
	}
	return Summary{value: v}, nil
}

func (s Summary) String() string {
	return s.value
}
