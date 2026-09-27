package command_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/mrstsgk/book-management-system/backend/internal/domain/common"
	"github.com/mrstsgk/book-management-system/backend/internal/usecase/author/command"
)

func TestCreateUsecase_BirthDateIsJudgedByTheInjectedClock(t *testing.T) {
	t.Parallel()
	jst := time.FixedZone("Asia/Tokyo", 9*60*60)
	// 2026-09-28 00:30 JST is still 2026-09-27 in UTC.
	now := func() time.Time { return time.Date(2026, 9, 28, 0, 30, 0, 0, jst) }
	uc := &command.CreateUsecaseImpl{Authors: &fakeRepository{}, Now: now}

	if _, err := uc.Execute(context.Background(), command.CreateCommand{Name: "太宰治", BirthDate: date(2026, 9, 27)}); err != nil {
		t.Fatalf("the day before today (JST) must be accepted: %v", err)
	}
	if _, err := uc.Execute(context.Background(), command.CreateCommand{Name: "太宰治", BirthDate: date(2026, 9, 28)}); !errors.Is(err, common.ErrInvalid) {
		t.Fatalf("today (JST) must be rejected: err = %v", err)
	}
}

func TestCreateUsecase_WithoutClockUsesCurrentTime(t *testing.T) {
	t.Parallel()
	uc := &command.CreateUsecaseImpl{Authors: &fakeRepository{}}
	tomorrow := time.Now().AddDate(0, 0, 2)

	if _, err := uc.Execute(context.Background(), command.CreateCommand{Name: "太宰治", BirthDate: &tomorrow}); !errors.Is(err, common.ErrInvalid) {
		t.Fatalf("a future birth date must be rejected: err = %v", err)
	}
}
