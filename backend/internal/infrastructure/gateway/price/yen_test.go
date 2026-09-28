package price_test

import (
	"testing"

	"github.com/mrstsgk/book-management-system/backend/internal/infrastructure/gateway/price"
)

func TestParseYen(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		in   string
		want *int64
	}{
		{name: "数字だけ", in: "4600", want: ptr(4600)},
		{name: "末尾の注記は無視する", in: "310 (税込)", want: ptr(310)},
		{name: "カンマ区切り", in: "1,650", want: ptr(1650)},
		{name: "前後の空白", in: "  2600 ", want: ptr(2600)},
		{name: "0円", in: "0", want: ptr(0)},
		{name: "数字で始まらなければnil", in: "価格未定", want: nil},
		{name: "空文字はnil", in: "", want: nil},
		{name: "int64を超える桁数はnil", in: "99999999999999999999", want: nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := price.ParseYen(tt.in)
			if (got == nil) != (tt.want == nil) || (got != nil && *got != *tt.want) {
				t.Fatalf("ParseYen(%q) = %v, want %v", tt.in, deref(got), deref(tt.want))
			}
		})
	}
}

func ptr(n int64) *int64 { return &n }

func deref(p *int64) any {
	if p == nil {
		return nil
	}
	return *p
}
