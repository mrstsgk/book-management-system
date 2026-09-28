package tag_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/mrstsgk/book-management-system/backend/internal/domain/common"
	"github.com/mrstsgk/book-management-system/backend/internal/domain/tag"
)

func TestNewName(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		in      string
		want    string
		wantErr bool
	}{
		{name: "前後の空白を除いて保持する", in: " データベース ", want: "データベース"},
		{name: "途中の空白は残す", in: "分散 システム", want: "分散 システム"},
		{name: "1文字は有効", in: "本", want: "本"},
		{name: "30文字ちょうどは有効", in: strings.Repeat("あ", 30), want: strings.Repeat("あ", 30)},
		{name: "31文字はエラー", in: strings.Repeat("あ", 31), wantErr: true},
		{name: "空文字はエラー", in: "", wantErr: true},
		{name: "空白だけはエラー", in: "　 ", wantErr: true},
		{name: "改行はエラー", in: "分散\nシステム", wantErr: true},
		{name: "前後のタブもエラー", in: "\tデータベース", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := tag.NewName(tt.in)
			if tt.wantErr {
				if !errors.Is(err, common.ErrInvalid) {
					t.Fatalf("err = %v, want ErrInvalid", err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got.String() != tt.want {
				t.Fatalf("got %q, want %q", got.String(), tt.want)
			}
		})
	}
}
