package command_test

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/mrstsgk/book-management-system/backend/internal/domain/auth"
	"github.com/mrstsgk/book-management-system/backend/internal/domain/common"
	"github.com/mrstsgk/book-management-system/backend/internal/usecase/auth/command"
)

// fakeSessions は auth.SessionRepository の手書き Fake（docs/rules/testing.md）。
type fakeSessions struct {
	saved   *auth.Session
	saveErr error

	found   *auth.Session
	findErr error

	deletedID auth.SessionID
	deleteErr error

	updatedID auth.SessionID
	updatedAt time.Time
}

func (f *fakeSessions) Save(_ context.Context, s *auth.Session) error {
	if f.saveErr != nil {
		return f.saveErr
	}
	f.saved = s
	return nil
}
func (f *fakeSessions) Find(context.Context, auth.SessionID) (*auth.Session, error) {
	return f.found, f.findErr
}
func (f *fakeSessions) Delete(_ context.Context, id auth.SessionID) error {
	f.deletedID = id
	return f.deleteErr
}
func (f *fakeSessions) UpdateExpiry(_ context.Context, id auth.SessionID, at time.Time) error {
	f.updatedID, f.updatedAt = id, at
	return nil
}

// fakeVerifier は呼ばれた回数を数える（ID 不一致でも照合が走ることを確かめるため）。
type fakeVerifier struct {
	calls atomic.Int32
	ok    bool
	delay time.Duration // bcrypt の遅さを模す（並列テストで照合中に他の試行を割り込ませる）
}

func (f *fakeVerifier) Matches(string, string) bool {
	f.calls.Add(1)
	time.Sleep(f.delay)
	return f.ok
}

var fixedNow = time.Date(2026, 9, 30, 12, 0, 0, 0, time.UTC)

func newLogin(sessions *fakeSessions, v *fakeVerifier) *command.LoginUsecaseImpl {
	return &command.LoginUsecaseImpl{
		Admin:    command.AdminAccount{ID: "admin", PasswordHash: "$hash"},
		Verifier: v,
		Sessions: sessions,
		Now:      func() time.Time { return fixedNow },
	}
}

// failTimes は失敗を n 回積む（ロックのテストの前提を作る）。
func failTimes(t *testing.T, uc *command.LoginUsecaseImpl, n int) {
	t.Helper()
	for i := 0; i < n; i++ {
		_, _ = uc.Execute(context.Background(), command.LoginCommand{ID: "admin", Password: "bad"})
	}
}

func TestLoginUsecase_Execute(t *testing.T) {
	t.Parallel()

	t.Run("IDとパスワードが合えばセッションを保存してIDを返す", func(t *testing.T) {
		t.Parallel()
		sessions := &fakeSessions{}
		uc := newLogin(sessions, &fakeVerifier{ok: true})

		id, err := uc.Execute(context.Background(), command.LoginCommand{ID: "admin", Password: "pw"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if sessions.saved == nil || string(sessions.saved.ID) != id || len(id) != 43 {
			t.Fatalf("saved %+v, returned %q", sessions.saved, id)
		}
		if !sessions.saved.ExpiresAt.Equal(uc.Now().Add(time.Hour)) {
			t.Fatalf("ExpiresAt = %v", sessions.saved.ExpiresAt)
		}
	})

	t.Run("IDが違えば401、保存しない、それでも照合は1回走る", func(t *testing.T) {
		t.Parallel()
		sessions := &fakeSessions{}
		v := &fakeVerifier{ok: true}
		uc := newLogin(sessions, v)

		_, err := uc.Execute(context.Background(), command.LoginCommand{ID: "other", Password: "pw"})
		if !errors.Is(err, common.ErrUnauthorized) {
			t.Fatalf("err = %v, want ErrUnauthorized", err)
		}
		if sessions.saved != nil {
			t.Fatal("must not save a session")
		}
		if v.calls.Load() != 1 {
			t.Fatalf("verifier calls = %d, want 1 (constant work regardless of ID)", v.calls.Load())
		}
	})

	t.Run("パスワードが違えば401で保存しない", func(t *testing.T) {
		t.Parallel()
		sessions := &fakeSessions{}
		uc := newLogin(sessions, &fakeVerifier{ok: false})

		if _, err := uc.Execute(context.Background(), command.LoginCommand{ID: "admin", Password: "bad"}); !errors.Is(err, common.ErrUnauthorized) {
			t.Fatalf("err = %v, want ErrUnauthorized", err)
		}
		if sessions.saved != nil {
			t.Fatal("must not save a session")
		}
	})

	t.Run("空欄は400で照合しない", func(t *testing.T) {
		t.Parallel()
		v := &fakeVerifier{ok: true}
		uc := newLogin(&fakeSessions{}, v)

		if _, err := uc.Execute(context.Background(), command.LoginCommand{ID: "admin", Password: ""}); !errors.Is(err, common.ErrInvalid) {
			t.Fatalf("err = %v, want ErrInvalid", err)
		}
		if v.calls.Load() != 0 {
			t.Fatal("verifier must not be called for empty input")
		}
	})

	t.Run("4回失敗まではロックしない", func(t *testing.T) {
		t.Parallel()
		v := &fakeVerifier{ok: false}
		uc := newLogin(&fakeSessions{}, v)
		failTimes(t, uc, 4)
		if _, err := uc.Execute(context.Background(), command.LoginCommand{ID: "admin", Password: "bad"}); !errors.Is(err, common.ErrUnauthorized) {
			t.Fatalf("err = %v, want ErrUnauthorized (not locked yet)", err)
		}
	})

	t.Run("5回失敗すると6回目は照合せず429", func(t *testing.T) {
		t.Parallel()
		v := &fakeVerifier{ok: false}
		uc := newLogin(&fakeSessions{}, v)
		failTimes(t, uc, 5)
		if _, err := uc.Execute(context.Background(), command.LoginCommand{ID: "admin", Password: "bad"}); !errors.Is(err, common.ErrTooManyAttempts) {
			t.Fatalf("err = %v, want ErrTooManyAttempts", err)
		}
		if v.calls.Load() != 5 {
			t.Fatalf("verifier calls = %d, want 5 (locked attempt must not verify)", v.calls.Load())
		}
	})

	t.Run("ロック中は正しいパスワードでも429でセッションを作らない", func(t *testing.T) {
		t.Parallel()
		sessions := &fakeSessions{}
		v := &fakeVerifier{ok: false}
		uc := newLogin(sessions, v)
		failTimes(t, uc, 5)
		v.ok = true
		if _, err := uc.Execute(context.Background(), command.LoginCommand{ID: "admin", Password: "right"}); !errors.Is(err, common.ErrTooManyAttempts) {
			t.Fatalf("err = %v, want ErrTooManyAttempts", err)
		}
		if sessions.saved != nil {
			t.Fatal("must not save a session while locked")
		}
	})

	t.Run("1分経てばロックが解ける", func(t *testing.T) {
		t.Parallel()
		v := &fakeVerifier{ok: false}
		uc := newLogin(&fakeSessions{}, v)
		failTimes(t, uc, 5)
		uc.Now = func() time.Time { return fixedNow.Add(time.Minute) }
		v.ok = true
		if _, err := uc.Execute(context.Background(), command.LoginCommand{ID: "admin", Password: "right"}); err != nil {
			t.Fatalf("unexpected error after the lock expired: %v", err)
		}
	})

	t.Run("成功すると失敗回数が戻る", func(t *testing.T) {
		t.Parallel()
		v := &fakeVerifier{ok: false}
		uc := newLogin(&fakeSessions{}, v)
		failTimes(t, uc, 4)
		v.ok = true
		if _, err := uc.Execute(context.Background(), command.LoginCommand{ID: "admin", Password: "right"}); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		v.ok = false
		// reset が無ければ 1 回目で 5 回目の失敗になりロックされ、2 回目が 429 になる。
		for i := 0; i < 2; i++ {
			if _, err := uc.Execute(context.Background(), command.LoginCommand{ID: "admin", Password: "bad"}); !errors.Is(err, common.ErrUnauthorized) {
				t.Fatalf("attempt %d: err = %v, want ErrUnauthorized (count must have been reset)", i+1, err)
			}
		}
	})

	t.Run("並列の総当たりでも照合は5回で止まる", func(t *testing.T) {
		t.Parallel()
		v := &fakeVerifier{ok: false, delay: 10 * time.Millisecond}
		uc := newLogin(&fakeSessions{}, v)

		const n = 20
		var unauthorized, tooMany atomic.Int32
		var wg sync.WaitGroup
		for i := 0; i < n; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				_, err := uc.Execute(context.Background(), command.LoginCommand{ID: "admin", Password: "bad"})
				switch {
				case errors.Is(err, common.ErrUnauthorized):
					unauthorized.Add(1)
				case errors.Is(err, common.ErrTooManyAttempts):
					tooMany.Add(1)
				}
			}()
		}
		wg.Wait()

		if v.calls.Load() != 5 || unauthorized.Load() != 5 || tooMany.Load() != n-5 {
			t.Fatalf("calls=%d unauthorized=%d tooMany=%d, want 5/5/%d", v.calls.Load(), unauthorized.Load(), tooMany.Load(), n-5)
		}
	})

	t.Run("設定が空なら常に401", func(t *testing.T) {
		t.Parallel()
		uc := newLogin(&fakeSessions{}, &fakeVerifier{ok: true})
		uc.Admin = command.AdminAccount{}
		if _, err := uc.Execute(context.Background(), command.LoginCommand{ID: "", Password: "pw"}); err == nil {
			t.Fatal("expected error")
		}
		if _, err := uc.Execute(context.Background(), command.LoginCommand{ID: "admin", Password: "pw"}); !errors.Is(err, common.ErrUnauthorized) {
			t.Fatalf("err = %v, want ErrUnauthorized", err)
		}
	})

	t.Run("Repositoryのエラーはそのまま伝わる", func(t *testing.T) {
		t.Parallel()
		boom := errors.New("db down")
		uc := newLogin(&fakeSessions{saveErr: boom}, &fakeVerifier{ok: true})
		if _, err := uc.Execute(context.Background(), command.LoginCommand{ID: "admin", Password: "pw"}); !errors.Is(err, boom) {
			t.Fatalf("err = %v, want %v", err, boom)
		}
	})
}
