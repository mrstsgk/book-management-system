package author_test

import (
	"testing"

	"github.com/mrstsgk/book-management-system/backend/internal/domain/author"
)

func TestNew_SetsFieldsAndLeavesIdentityUnassigned(t *testing.T) {
	t.Parallel()
	name, err := author.NewName("太宰治")
	if err != nil {
		t.Fatalf("NewName: %v", err)
	}

	a := author.New(name, nil)

	if a.ID != 0 || a.Version != 0 {
		t.Fatalf("ID/Version = %d/%d, want zero values for an unsaved Author", a.ID, a.Version)
	}
	if a.Name != name || a.BirthDate != nil {
		t.Fatalf("got %+v, want name=%v and no birth date", a, name)
	}
}
