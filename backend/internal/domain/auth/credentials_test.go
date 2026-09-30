package auth_test

import (
	"errors"
	"testing"

	"github.com/mrstsgk/book-management-system/backend/internal/domain/auth"
	"github.com/mrstsgk/book-management-system/backend/internal/domain/common"
)

func TestNewCredentials(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		id, pw  string
		wantID  string
		wantPW  string
		wantErr bool
	}{
		{name: "IDとパスワードをそのまま持つ", id: "admin", pw: "pa ss ", wantID: "admin", wantPW: "pa ss "},
		{name: "パスワードの前後の空白は落とさない", id: "admin", pw: " secret ", wantID: "admin", wantPW: " secret "},
		{name: "IDが空はエラー", id: "", pw: "x", wantErr: true},
		{name: "パスワードが空はエラー", id: "admin", pw: "", wantErr: true},
		{name: "IDが空白だけはエラー", id: " ", pw: "x", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := auth.NewCredentials(tt.id, tt.pw)
			if tt.wantErr {
				if !errors.Is(err, common.ErrInvalid) {
					t.Fatalf("err = %v, want ErrInvalid", err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got.ID() != tt.wantID || got.Password() != tt.wantPW {
				t.Fatalf("got (%q, %q), want (%q, %q)", got.ID(), got.Password(), tt.wantID, tt.wantPW)
			}
		})
	}
}
