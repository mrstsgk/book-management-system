package auth_test

import (
	"context"
	"errors"
	"testing"
	"time"

	domainauth "github.com/mrstsgk/book-management-system/backend/internal/domain/auth"
	domaincommon "github.com/mrstsgk/book-management-system/backend/internal/domain/common"
	pgauth "github.com/mrstsgk/book-management-system/backend/internal/infrastructure/postgres/auth"
)

func newSession(t *testing.T) *domainauth.Session {
	t.Helper()
	id, err := domainauth.NewSessionID()
	if err != nil {
		t.Fatal(err)
	}
	return domainauth.NewSession(id, time.Now().UTC().Truncate(time.Microsecond))
}

func TestRepository_SaveThenFind(t *testing.T) {
	db := connectTestDB(t)
	repo := pgauth.NewRepository(db)
	s := newSession(t)
	t.Cleanup(func() { db.Exec("DELETE FROM admin_session WHERE id = ?", string(s.ID)) })

	if err := repo.Save(context.Background(), s); err != nil {
		t.Fatalf("Save: %v", err)
	}
	got, err := repo.Find(context.Background(), s.ID)
	if err != nil {
		t.Fatalf("Find: %v", err)
	}
	if got.ID != s.ID || !got.ExpiresAt.Equal(s.ExpiresAt) || !got.AbsoluteExpiresAt.Equal(s.AbsoluteExpiresAt) {
		t.Fatalf("got %+v, want %+v", got, s)
	}
}

func TestRepository_Find_NotFound(t *testing.T) {
	repo := pgauth.NewRepository(connectTestDB(t))
	if _, err := repo.Find(context.Background(), "no-such-session"); !errors.Is(err, domaincommon.ErrNotFound) {
		t.Fatalf("err = %v, want ErrNotFound", err)
	}
}

func TestRepository_UpdateExpiry(t *testing.T) {
	db := connectTestDB(t)
	repo := pgauth.NewRepository(db)
	s := newSession(t)
	t.Cleanup(func() { db.Exec("DELETE FROM admin_session WHERE id = ?", string(s.ID)) })
	if err := repo.Save(context.Background(), s); err != nil {
		t.Fatal(err)
	}

	next := s.ExpiresAt.Add(30 * time.Minute)
	if err := repo.UpdateExpiry(context.Background(), s.ID, next); err != nil {
		t.Fatalf("UpdateExpiry: %v", err)
	}
	got, _ := repo.Find(context.Background(), s.ID)
	if !got.ExpiresAt.Equal(next) {
		t.Fatalf("ExpiresAt = %v, want %v", got.ExpiresAt, next)
	}
	if !got.AbsoluteExpiresAt.Equal(s.AbsoluteExpiresAt) {
		t.Fatal("absolute limit must not change")
	}
}

func TestRepository_Delete(t *testing.T) {
	db := connectTestDB(t)
	repo := pgauth.NewRepository(db)
	s := newSession(t)
	if err := repo.Save(context.Background(), s); err != nil {
		t.Fatal(err)
	}

	if err := repo.Delete(context.Background(), s.ID); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if _, err := repo.Find(context.Background(), s.ID); !errors.Is(err, domaincommon.ErrNotFound) {
		t.Fatalf("Find after delete: err = %v, want ErrNotFound", err)
	}
	if err := repo.Delete(context.Background(), s.ID); err != nil {
		t.Fatalf("Delete again must not fail: %v", err)
	}
}

// テーブルにパスワードもハッシュも入らないことを列名で確かめる（設計の「秘密はサーバーの環境変数にだけ置く」の担保）。
func TestRepository_TableHasNoSecretColumns(t *testing.T) {
	db := connectTestDB(t)
	var cols []string
	if err := db.Raw("SELECT column_name FROM information_schema.columns WHERE table_name = 'admin_session'").Scan(&cols).Error; err != nil {
		t.Fatal(err)
	}
	for _, c := range cols {
		if c == "password" || c == "password_hash" {
			t.Fatalf("admin_session must not have column %q", c)
		}
	}
}
