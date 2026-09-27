package author_test

import (
	"errors"
	"testing"
	"time"

	"github.com/mrstsgk/book-management-system/backend/internal/domain/author"
	"github.com/mrstsgk/book-management-system/backend/internal/domain/common"
)

func TestNewBirthDate(t *testing.T) {
	t.Parallel()
	jst := time.FixedZone("Asia/Tokyo", 9*60*60)
	now := time.Date(2026, 9, 28, 0, 30, 0, 0, jst) // 2026-09-27 15:30 in UTC
	date := func(y int, m time.Month, d int) time.Time { return time.Date(y, m, d, 0, 0, 0, 0, time.UTC) }

	tests := []struct {
		name    string
		in      time.Time
		wantErr bool
	}{
		{name: "昨日は有効", in: date(2026, 9, 27)},
		{name: "今日はエラー（当日は含まない）", in: date(2026, 9, 28), wantErr: true},
		{name: "未来日はエラー", in: date(2026, 9, 29), wantErr: true},
		{name: "十分に過去の日付は有効", in: date(1909, 6, 19)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := author.NewBirthDate(tt.in, now)
			if tt.wantErr {
				if !errors.Is(err, common.ErrInvalid) {
					t.Fatalf("err = %v, want ErrInvalid", err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !got.Time().Equal(tt.in) {
				t.Fatalf("Time() = %v, want %v", got.Time(), tt.in)
			}
		})
	}
}

func TestRestoreBirthDate_KeepsCalendarDateWithoutCheckingToday(t *testing.T) {
	t.Parallel()
	in := time.Date(2999, 1, 2, 15, 4, 5, 0, time.UTC)

	got := author.RestoreBirthDate(in)

	if want := time.Date(2999, 1, 2, 0, 0, 0, 0, time.UTC); !got.Time().Equal(want) {
		t.Fatalf("Time() = %v, want %v", got.Time(), want)
	}
}
