package author

import (
	"fmt"
	"time"

	"github.com/mrstsgk/book-management-system/backend/internal/domain/common"
)

// BirthDate is a value object for a date strictly before today (today is rejected).
type BirthDate struct {
	value time.Time
}

// NewBirthDate takes now explicitly so "today" (and its time zone) is decided by the
// caller's clock, not hidden here. Only the calendar date of each argument is compared.
func NewBirthDate(date, now time.Time) (BirthDate, error) {
	d := dateOnly(date)
	if !d.Before(dateOnly(now)) {
		return BirthDate{}, fmt.Errorf("%w: 生年月日は現在の日付よりも過去である必要があります", common.ErrInvalid)
	}
	return BirthDate{value: d}, nil
}

// RestoreBirthDate rebuilds a persisted value without re-checking it against today:
// the rule applies when the date is entered, and a stored date only gets further in the past.
func RestoreBirthDate(date time.Time) BirthDate {
	return BirthDate{value: dateOnly(date)}
}

func (b BirthDate) Time() time.Time {
	return b.value
}

func dateOnly(t time.Time) time.Time {
	y, m, d := t.Date()
	return time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
}
