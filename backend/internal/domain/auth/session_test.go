package auth_test

import (
	"encoding/base64"
	"testing"
	"time"

	"github.com/mrstsgk/book-management-system/backend/internal/domain/auth"
)

func TestNewSessionID(t *testing.T) {
	t.Parallel()
	a, err := auth.NewSessionID()
	if err != nil {
		t.Fatal(err)
	}
	b, _ := auth.NewSessionID()
	if a == b {
		t.Fatal("two IDs must differ")
	}
	if len(a) != 43 {
		t.Fatalf("len = %d, want 43 (32 bytes base64url without padding)", len(a))
	}
	if _, err := base64.RawURLEncoding.DecodeString(string(a)); err != nil {
		t.Fatalf("not base64url: %v", err)
	}
}

func TestSession(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 9, 30, 12, 0, 0, 0, time.UTC)

	t.Run("発行直後はアイドル1時間・絶対24時間", func(t *testing.T) {
		t.Parallel()
		s := auth.NewSession("id", now)
		if !s.ExpiresAt.Equal(now.Add(time.Hour)) || !s.AbsoluteExpiresAt.Equal(now.Add(24*time.Hour)) {
			t.Fatalf("got %+v", s)
		}
	})

	tests := []struct {
		name  string
		at    time.Time
		valid bool
	}{
		{name: "アイドル期限の直前は有効", at: now.Add(time.Hour - time.Second), valid: true},
		{name: "アイドル期限ちょうどは無効", at: now.Add(time.Hour), valid: false},
		{name: "アイドル期限+1秒は無効", at: now.Add(time.Hour + time.Second), valid: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			s := auth.NewSession("id", now)
			if got := s.IsValid(tt.at); got != tt.valid {
				t.Fatalf("IsValid = %v, want %v", got, tt.valid)
			}
		})
	}

	t.Run("延長するとアイドル期限が今+1時間になる", func(t *testing.T) {
		t.Parallel()
		s := auth.NewSession("id", now)
		s.Extend(now.Add(30 * time.Minute))
		if !s.ExpiresAt.Equal(now.Add(90 * time.Minute)) {
			t.Fatalf("ExpiresAt = %v", s.ExpiresAt)
		}
	})

	t.Run("延長しても絶対期限を超えない（絶対期限に達すれば無効になる）", func(t *testing.T) {
		t.Parallel()
		s := auth.NewSession("id", now)
		s.Extend(now.Add(23*time.Hour + 30*time.Minute))
		if !s.ExpiresAt.Equal(now.Add(24 * time.Hour)) {
			t.Fatalf("ExpiresAt = %v, want the absolute limit", s.ExpiresAt)
		}
		if s.IsValid(now.Add(24 * time.Hour)) {
			t.Fatal("must be invalid at the absolute limit")
		}
	})
}
