// parseTitleOverride は非公開関数で、登録・更新の両ユースケースが共有する解釈規則そのものを検証する対象のため、
// 公開 API 越しではなく in-package で直接テストする。
package command

import (
	"errors"
	"strings"
	"testing"

	"github.com/mrstsgk/book-management-system/backend/internal/domain/common"
)

func TestParseTitleOverride(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		in      string
		want    string
		wantNil bool
		wantErr bool
	}{
		{name: "書名を渡すと上書きになる", in: "正しい書名", want: "正しい書名"},
		{name: "前後の空白は除いて保持する", in: " 正しい書名 ", want: "正しい書名"},
		{name: "空文字は上書きしない", in: "", wantNil: true},
		{name: "空白だけは上書きしない", in: "　 ", wantNil: true},
		{name: "タブだけは空白ではなく制御文字としてエラー", in: "\t", wantErr: true},
		{name: "改行だけは空白ではなく制御文字としてエラー", in: "\n", wantErr: true},
		{name: "256文字は書名の規則違反でエラー", in: strings.Repeat("あ", 256), wantErr: true},
		{name: "改行を含むとエラー", in: "正しい\n書名", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := parseTitleOverride(tt.in)
			if tt.wantErr {
				if !errors.Is(err, common.ErrInvalid) {
					t.Fatalf("err = %v, want ErrInvalid", err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tt.wantNil {
				if got != nil {
					t.Fatalf("got %v, want nil", got)
				}
				return
			}
			if got == nil || got.String() != tt.want {
				t.Fatalf("got %v, want %q", got, tt.want)
			}
		})
	}
}
