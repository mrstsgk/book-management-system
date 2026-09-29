package tag_test

import (
	"testing"

	"github.com/mrstsgk/book-management-system/backend/internal/domain/tag"
)

func mustName(t *testing.T, raw string) tag.Name {
	t.Helper()
	n, err := tag.NewName(raw)
	if err != nil {
		t.Fatal(err)
	}
	return n
}

func TestNew_SetsNameAndLeavesIdentityUnassigned(t *testing.T) {
	t.Parallel()
	tg := tag.New(mustName(t, "データベース"))

	if tg.ID != 0 || tg.Version != 0 {
		t.Fatalf("ID/Version = %d/%d, want zero values before saving", tg.ID, tg.Version)
	}
	if tg.Name.String() != "データベース" {
		t.Fatalf("Name = %q, want データベース", tg.Name.String())
	}
}

func TestTag_Rename(t *testing.T) {
	t.Parallel()
	tg := tag.New(mustName(t, "データベース"))

	tg.Rename(mustName(t, "分散システム"), 3)

	if tg.Name.String() != "分散システム" || tg.Version != 3 {
		t.Fatalf("got name=%q version=%d", tg.Name.String(), tg.Version)
	}
}
