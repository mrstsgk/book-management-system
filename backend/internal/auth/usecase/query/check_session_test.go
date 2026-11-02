package query_test

import (
	"context"
	"errors"
	"testing"
	"time"

	authdomain "github.com/mrstsgk/book-management-system/backend/internal/auth/domain"
	authqry "github.com/mrstsgk/book-management-system/backend/internal/auth/usecase/query"
	"github.com/mrstsgk/book-management-system/backend/internal/domainerr"
)

type fakeSessions struct {
	found     *authdomain.Session
	findErr   error
	updatedID authdomain.SessionID
	updatedAt time.Time
}

func (f *fakeSessions) Create(context.Context, *authdomain.Session) error { return nil }
func (f *fakeSessions) FindByID(context.Context, authdomain.SessionID) (*authdomain.Session, error) {
	return f.found, f.findErr
}
func (f *fakeSessions) Delete(context.Context, authdomain.SessionID) error { return nil }
func (f *fakeSessions) UpdateExpiry(_ context.Context, id authdomain.SessionID, at time.Time) error {
	f.updatedID, f.updatedAt = id, at
	return nil
}

func TestCheckSessionUsecase_Execute(t *testing.T) {
	t.Parallel()
	issued := time.Date(2026, 9, 30, 12, 0, 0, 0, time.UTC)

	t.Run("有効なら延長する", func(t *testing.T) {
		t.Parallel()
		sessions := &fakeSessions{found: authdomain.NewSession("sid", issued)}
		now := issued.Add(30 * time.Minute)
		uc := &authqry.CheckSessionUsecaseImpl{Sessions: sessions, Now: func() time.Time { return now }}

		if err := uc.Execute(context.Background(), "sid"); err != nil {
			t.Fatal(err)
		}
		if string(sessions.updatedID) != "sid" || !sessions.updatedAt.Equal(now.Add(time.Hour)) {
			t.Fatalf("updated (%q, %v)", sessions.updatedID, sessions.updatedAt)
		}
	})

	t.Run("アイドル期限切れは401で延長しない", func(t *testing.T) {
		t.Parallel()
		sessions := &fakeSessions{found: authdomain.NewSession("sid", issued)}
		uc := &authqry.CheckSessionUsecaseImpl{Sessions: sessions, Now: func() time.Time { return issued.Add(time.Hour) }}

		if err := uc.Execute(context.Background(), "sid"); !errors.Is(err, domainerr.ErrUnauthorized) {
			t.Fatalf("err = %v, want ErrUnauthorized", err)
		}
		if sessions.updatedID != "" {
			t.Fatal("must not extend an expired session")
		}
	})

	t.Run("見つからなければ401", func(t *testing.T) {
		t.Parallel()
		sessions := &fakeSessions{findErr: domainerr.ErrNotFound}
		uc := &authqry.CheckSessionUsecaseImpl{Sessions: sessions, Now: time.Now}
		if err := uc.Execute(context.Background(), "nope"); !errors.Is(err, domainerr.ErrUnauthorized) {
			t.Fatalf("err = %v, want ErrUnauthorized", err)
		}
	})

	t.Run("Repositoryの他のエラーはそのまま伝わる", func(t *testing.T) {
		t.Parallel()
		boom := errors.New("db down")
		uc := &authqry.CheckSessionUsecaseImpl{Sessions: &fakeSessions{findErr: boom}, Now: time.Now}
		if err := uc.Execute(context.Background(), "sid"); !errors.Is(err, boom) {
			t.Fatalf("err = %v, want %v", err, boom)
		}
	})
}
