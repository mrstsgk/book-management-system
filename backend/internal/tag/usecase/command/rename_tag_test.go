package command_test

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/mrstsgk/book-management-system/backend/internal/domainerr"
	tagdomain "github.com/mrstsgk/book-management-system/backend/internal/tag/domain"
	tagcmd "github.com/mrstsgk/book-management-system/backend/internal/tag/usecase/command"
)

func TestRenameTagUsecase_Execute(t *testing.T) {
	t.Parallel()

	existing := func(t *testing.T) *tagdomain.Tag {
		t.Helper()
		n, err := tagdomain.NewName("旧名")
		if err != nil {
			t.Fatal(err)
		}
		tg := tagdomain.New(n)
		tg.ID, tg.Version = 5, 2
		return tg
	}

	t.Run("名前を差し替えて返す", func(t *testing.T) {
		t.Parallel()
		tags := &fakeTags{findByID: existing(t)}
		q := &fakeTagQuery{}
		uc := &tagcmd.RenameTagUsecaseImpl{Tags: tags, TagQuery: q}

		got, err := uc.Execute(context.Background(), tagcmd.RenameTagCommand{ID: 5, Name: "新名", Version: 2})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if tags.updated == nil || tags.updated.Name.String() != "新名" {
			t.Fatalf("updated %+v", tags.updated)
		}
		if got.Name != "新名" || got.Version != 3 {
			t.Fatalf("got %+v", got)
		}
		// 自分の今の名前と同じ名前に「改名」しても重複にならないよう、自分を除いて検査する
		if q.gotName != "新名" || q.gotExclude != 5 {
			t.Fatalf("ExistsByName called with (%q, %d), want (新名, 5)", q.gotName, q.gotExclude)
		}
	})

	t.Run("空文字は保存を呼ばずに invalid の違反", func(t *testing.T) {
		t.Parallel()
		tags := &fakeTags{findByID: existing(t)}
		uc := &tagcmd.RenameTagUsecaseImpl{Tags: tags, TagQuery: &fakeTagQuery{}}

		_, err := uc.Execute(context.Background(), tagcmd.RenameTagCommand{ID: 5, Name: " ", Version: 2})
		if want := (domainerr.Violations{{Field: "name", Rule: "invalid"}}); !reflect.DeepEqual(violationsOf(t, err), want) {
			t.Fatalf("err = %v, want %+v", err, want)
		}
		if tags.updated != nil {
			t.Fatal("Repository.Update must not be called")
		}
	})

	t.Run("他のタグと同名なら保存を呼ばずに tagNameDuplicated の違反", func(t *testing.T) {
		t.Parallel()
		existingTag := existing(t)
		tags := &fakeTags{findByID: existingTag}
		uc := &tagcmd.RenameTagUsecaseImpl{Tags: tags, TagQuery: &fakeTagQuery{nameExists: true}}

		_, err := uc.Execute(context.Background(), tagcmd.RenameTagCommand{ID: 5, Name: "新名", Version: 2})
		if want := (domainerr.Violations{{Field: "name", Rule: "tagNameDuplicated"}}); !reflect.DeepEqual(violationsOf(t, err), want) {
			t.Fatalf("err = %v, want %+v", err, want)
		}
		if tags.updated != nil || existingTag.Name.String() != "旧名" {
			t.Fatal("nothing may change when the name is duplicated")
		}
	})

	t.Run("存在しないタグはNotFound", func(t *testing.T) {
		t.Parallel()
		tags := &fakeTags{findByIDErr: domainerr.ErrNotFound}
		uc := &tagcmd.RenameTagUsecaseImpl{Tags: tags, TagQuery: &fakeTagQuery{}}

		if _, err := uc.Execute(context.Background(), tagcmd.RenameTagCommand{ID: 5, Name: "新名", Version: 2}); !errors.Is(err, domainerr.ErrNotFound) {
			t.Fatalf("err = %v, want ErrNotFound", err)
		}
	})

	t.Run("バージョン競合はConflict", func(t *testing.T) {
		t.Parallel()
		tags := &fakeTags{findByID: existing(t), updateErr: domainerr.ErrConflict}
		uc := &tagcmd.RenameTagUsecaseImpl{Tags: tags, TagQuery: &fakeTagQuery{}}

		if _, err := uc.Execute(context.Background(), tagcmd.RenameTagCommand{ID: 5, Name: "新名", Version: 2}); !errors.Is(err, domainerr.ErrConflict) {
			t.Fatalf("err = %v, want ErrConflict", err)
		}
	})
}
