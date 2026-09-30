package auth_test

import (
	"strings"
	"testing"

	infraauth "github.com/mrstsgk/book-management-system/backend/internal/infrastructure/auth"
)

func TestBcryptVerifier(t *testing.T) {
	t.Parallel()
	hash, err := infraauth.HashPassword("correct horse")
	if err != nil {
		t.Fatal(err)
	}
	// "$2a$12$" のようにコストが 12 で始まる（総当たりを遅くする設計値）
	if !strings.HasPrefix(hash, "$2a$12$") {
		t.Fatalf("hash = %q, want bcrypt cost 12", hash)
	}

	v := infraauth.NewBcryptVerifier()
	tests := []struct {
		name string
		pw   string
		want bool
	}{
		{name: "正しいパスワードは一致", pw: "correct horse", want: true},
		{name: "違うパスワードは不一致", pw: "wrong", want: false},
		{name: "前後に空白が付くと不一致", pw: " correct horse ", want: false},
		{name: "空は不一致", pw: "", want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := v.Matches(hash, tt.pw); got != tt.want {
				t.Fatalf("Matches = %v, want %v", got, tt.want)
			}
		})
	}

	t.Run("壊れたハッシュは不一致（panic しない）", func(t *testing.T) {
		t.Parallel()
		if v.Matches("not-a-hash", "x") {
			t.Fatal("must be false")
		}
	})
}
