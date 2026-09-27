package author_test

import (
	"context"
	"testing"

	domainauthor "github.com/mrstsgk/book-management-system/backend/internal/domain/author"
	pgauthor "github.com/mrstsgk/book-management-system/backend/internal/infrastructure/postgres/author"
)

func TestQuery_Exists(t *testing.T) {
	db := connectTestDB(t)
	a := newAuthor(t, "query-test-exists", nil)
	createAuthor(t, db, pgauthor.NewRepository(db), a)
	qry := pgauthor.NewQuery(db)

	tests := []struct {
		name string
		id   domainauthor.ID
		want bool
	}{
		{name: "存在する著者はtrue", id: a.ID, want: true},
		{name: "存在しない著者はfalse", id: -1, want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := qry.Exists(context.Background(), tt.id)
			if err != nil {
				t.Fatalf("Exists: %v", err)
			}
			if got != tt.want {
				t.Fatalf("Exists = %v, want %v", got, tt.want)
			}
		})
	}
}
