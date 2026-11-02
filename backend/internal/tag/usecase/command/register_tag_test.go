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

// fakeTags は tagdomain.Repository の手書き Fake（docs/rules/testing.md）。
type fakeTags struct {
	created   *tagdomain.Tag
	createErr error

	findByID    *tagdomain.Tag
	findByIDErr error

	updated   *tagdomain.Tag
	updateErr error

	deletedID tagdomain.ID
	deleteErr error
}

func (f *fakeTags) FindByID(context.Context, tagdomain.ID) (*tagdomain.Tag, error) {
	return f.findByID, f.findByIDErr
}

func (f *fakeTags) Create(_ context.Context, t *tagdomain.Tag) error {
	if f.createErr != nil {
		return f.createErr
	}
	t.ID, t.Version = 1, 1
	f.created = t
	return nil
}

func (f *fakeTags) Update(_ context.Context, t *tagdomain.Tag) error {
	if f.updateErr != nil {
		return f.updateErr
	}
	t.Version++
	f.updated = t
	return nil
}

func (f *fakeTags) Delete(_ context.Context, id tagdomain.ID) error {
	f.deletedID = id
	return f.deleteErr
}

// fakeTagQuery は tagdomain.Query の手書き Fake（同名の有無だけを模す）。
type fakeTagQuery struct {
	nameExists bool
	existsErr  error
	gotName    string
	gotExclude tagdomain.ID
}

func (f *fakeTagQuery) FindAll(context.Context) (*tagdomain.TagList, error)          { return nil, nil }
func (f *fakeTagQuery) CountBooks(context.Context) (*tagdomain.TagBookCounts, error) { return nil, nil }
func (f *fakeTagQuery) ExistsAll(context.Context, []tagdomain.ID) (bool, error)      { return true, nil }
func (f *fakeTagQuery) ExistsByName(_ context.Context, name string, exclude tagdomain.ID) (bool, error) {
	f.gotName, f.gotExclude = name, exclude
	return f.nameExists, f.existsErr
}

// violationsOf はエラーから違反の一覧を取り出す（Violations でなければテストを落とす）。
func violationsOf(t *testing.T, err error) domainerr.Violations {
	t.Helper()
	var vs domainerr.Violations
	if !errors.As(err, &vs) {
		t.Fatalf("err = %v, want Violations", err)
	}
	return vs
}

func TestRegisterTagUsecase_Execute(t *testing.T) {
	t.Parallel()

	t.Run("タグ名を登録して返す", func(t *testing.T) {
		t.Parallel()
		tags := &fakeTags{}
		q := &fakeTagQuery{}
		uc := &tagcmd.RegisterTagUsecaseImpl{Tags: tags, TagQuery: q}

		got, err := uc.Execute(context.Background(), tagcmd.RegisterTagCommand{Name: "データベース"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if tags.created == nil || tags.created.Name.String() != "データベース" {
			t.Fatalf("created %+v", tags.created)
		}
		if got.ID != 1 || got.Name != "データベース" || got.Version != 1 {
			t.Fatalf("got %+v", got)
		}
		if q.gotName != "データベース" || q.gotExclude != 0 {
			t.Fatalf("ExistsByName called with (%q, %d), want (データベース, 0)", q.gotName, q.gotExclude)
		}
	})

	t.Run("空文字は重複検査も保存も呼ばずに invalid の違反", func(t *testing.T) {
		t.Parallel()
		tags := &fakeTags{}
		q := &fakeTagQuery{existsErr: errors.New("must not be called")}
		uc := &tagcmd.RegisterTagUsecaseImpl{Tags: tags, TagQuery: q}

		_, err := uc.Execute(context.Background(), tagcmd.RegisterTagCommand{Name: " "})
		if want := (domainerr.Violations{{Field: "name", Rule: "invalid"}}); !reflect.DeepEqual(violationsOf(t, err), want) {
			t.Fatalf("err = %v, want %+v", err, want)
		}
		if tags.created != nil {
			t.Fatal("Repository.Create must not be called")
		}
	})

	t.Run("同名のタグがあれば保存を呼ばずに tagNameDuplicated の違反", func(t *testing.T) {
		t.Parallel()
		tags := &fakeTags{}
		uc := &tagcmd.RegisterTagUsecaseImpl{Tags: tags, TagQuery: &fakeTagQuery{nameExists: true}}

		_, err := uc.Execute(context.Background(), tagcmd.RegisterTagCommand{Name: "データベース"})
		if want := (domainerr.Violations{{Field: "name", Rule: "tagNameDuplicated"}}); !reflect.DeepEqual(violationsOf(t, err), want) {
			t.Fatalf("err = %v, want %+v", err, want)
		}
		if tags.created != nil {
			t.Fatal("Repository.Create must not be called")
		}
	})

	t.Run("重複検査の障害はそのまま返す", func(t *testing.T) {
		t.Parallel()
		wantErr := errors.New("db: timeout")
		uc := &tagcmd.RegisterTagUsecaseImpl{Tags: &fakeTags{}, TagQuery: &fakeTagQuery{existsErr: wantErr}}
		if _, err := uc.Execute(context.Background(), tagcmd.RegisterTagCommand{Name: "データベース"}); !errors.Is(err, wantErr) {
			t.Fatalf("err = %v, want %v", err, wantErr)
		}
	})

	t.Run("同時追加で一意制約に当たったときはConflict", func(t *testing.T) {
		t.Parallel()
		tags := &fakeTags{createErr: domainerr.ErrConflict}
		uc := &tagcmd.RegisterTagUsecaseImpl{Tags: tags, TagQuery: &fakeTagQuery{}}

		if _, err := uc.Execute(context.Background(), tagcmd.RegisterTagCommand{Name: "データベース"}); !errors.Is(err, domainerr.ErrConflict) {
			t.Fatalf("err = %v, want ErrConflict", err)
		}
	})
}
