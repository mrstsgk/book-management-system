package author_test

import (
	"reflect"
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

// 一覧の Read Model は書き込み側の VO・Entity に依存しない（docs/rules/testing.md）。
func TestAuthorListItem_UsesPlainFieldTypes(t *testing.T) {
	t.Parallel()
	writeSide := map[reflect.Type]bool{
		reflect.TypeOf(author.Name{}):      true,
		reflect.TypeOf(author.BirthDate{}): true,
		reflect.TypeOf(author.Author{}):    true,
	}
	typ := reflect.TypeOf(author.AuthorListItem{})
	for i := range typ.NumField() {
		ft := typ.Field(i).Type
		for ft.Kind() == reflect.Pointer || ft.Kind() == reflect.Slice {
			ft = ft.Elem()
		}
		if writeSide[ft] {
			t.Errorf("AuthorListItem.%s uses write-side type %s", typ.Field(i).Name, ft)
		}
	}
}
