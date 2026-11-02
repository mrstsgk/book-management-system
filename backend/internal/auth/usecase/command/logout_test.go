package command_test

import (
	"context"
	"errors"
	"testing"

	authcmd "github.com/mrstsgk/book-management-system/backend/internal/auth/usecase/command"
)

func TestLogoutUsecase_Execute(t *testing.T) {
	t.Parallel()

	t.Run("セッションを削除する", func(t *testing.T) {
		t.Parallel()
		sessions := &fakeSessions{}
		uc := &authcmd.LogoutUsecaseImpl{Sessions: sessions}
		if err := uc.Execute(context.Background(), "sid"); err != nil {
			t.Fatal(err)
		}
		if string(sessions.deletedID) != "sid" {
			t.Fatalf("deleted %q", sessions.deletedID)
		}
	})

	t.Run("Repositoryのエラーはそのまま伝わる", func(t *testing.T) {
		t.Parallel()
		boom := errors.New("db down")
		uc := &authcmd.LogoutUsecaseImpl{Sessions: &fakeSessions{deleteErr: boom}}
		if err := uc.Execute(context.Background(), "sid"); !errors.Is(err, boom) {
			t.Fatalf("err = %v, want %v", err, boom)
		}
	})
}
