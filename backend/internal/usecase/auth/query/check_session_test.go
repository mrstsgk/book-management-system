package query_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/mrstsgk/book-management-system/backend/internal/domain/auth"
	"github.com/mrstsgk/book-management-system/backend/internal/domain/common"
	"github.com/mrstsgk/book-management-system/backend/internal/usecase/auth/query"
)

type fakeSessions struct {
	found     *auth.Session
	findErr   error
	updatedID auth.SessionID
	updatedAt time.Time
}

func (f *fakeSessions) Save(context.Context, *auth.Session) error { return nil }
func (f *fakeSessions) Find(context.Context, auth.SessionID) (*auth.Session, error) {
	return f.found, f.findErr
}
func (f *fakeSessions) Delete(context.Context, auth.SessionID) error { return nil }
func (f *fakeSessions) UpdateExpiry(_ context.Context, id auth.SessionID, at time.Time) error {
	f.updatedID, f.updatedAt = id, at
	return nil
}

func TestCheckSessionUsecase_Execute(t *testing.T) {
	t.Parallel()
	issued := time.Date(2026, 9, 30, 12, 0, 0, 0, time.UTC)

	t.Run("有効なら延長する", func(t *testing.T) {
		t.Parallel()
		sessions := &fakeSessions{found: auth.NewSession("sid", issued)}
		now := issued.Add(30 * time.Minute)
		uc := &query.CheckSessionUsecaseImpl{Sessions: sessions, Now: func() time.Time { return now }}

		if err := uc.Execute(context.Background(), "sid"); err != nil {
			t.Fatal(err)
		}
		if string(sessions.updatedID) != "sid" || !sessions.updatedAt.Equal(now.Add(time.Hour)) {
			t.Fatalf("updated (%q, %v)", sessions.updatedID, sessions.updatedAt)
		}
	})

	t.Run("アイドル期限切れは401で延長しない", func(t *testing.T) {
		t.Parallel()
		sessions := &fakeSessions{found: auth.NewSession("sid", issued)}
		uc := &query.CheckSessionUsecaseImpl{Sessions: sessions, Now: func() time.Time { return issued.Add(time.Hour) }}

		if err := uc.Execute(context.Background(), "sid"); !errors.Is(err, common.ErrUnauthorized) {
			t.Fatalf("err = %v, want ErrUnauthorized", err)
		}
		if sessions.updatedID != "" {
			t.Fatal("must not extend an expired session")
		}
	})

	t.Run("見つからなければ401", func(t *testing.T) {
		t.Parallel()
		sessions := &fakeSessions{findErr: common.ErrNotFound}
		uc := &query.CheckSessionUsecaseImpl{Sessions: sessions, Now: time.Now}
		if err := uc.Execute(context.Background(), "nope"); !errors.Is(err, common.ErrUnauthorized) {
			t.Fatalf("err = %v, want ErrUnauthorized", err)
		}
	})

	t.Run("Repositoryの他のエラーはそのまま伝わる", func(t *testing.T) {
		t.Parallel()
		boom := errors.New("db down")
		uc := &query.CheckSessionUsecaseImpl{Sessions: &fakeSessions{findErr: boom}, Now: time.Now}
		if err := uc.Execute(context.Background(), "sid"); !errors.Is(err, boom) {
			t.Fatalf("err = %v, want %v", err, boom)
		}
	})
}
