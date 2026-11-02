package command_test

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/mrstsgk/book-management-system/backend/internal/domain/book"
	"github.com/mrstsgk/book-management-system/backend/internal/domain/common"
	"github.com/mrstsgk/book-management-system/backend/internal/domain/tag"
	"github.com/mrstsgk/book-management-system/backend/internal/usecase/book/command"
)

// fakeBooks は book.Repository の手書き Fake（docs/rules/testing.md）。
type fakeBooks struct {
	created   *book.Book
	createErr error

	findByID    *book.Book
	findByIDErr error

	updated   *book.Book
	updateErr error

	deletedID book.ID
	deleteErr error
}

func (f *fakeBooks) FindByID(context.Context, book.ID) (*book.Book, error) {
	return f.findByID, f.findByIDErr
}

func (f *fakeBooks) Create(_ context.Context, b *book.Book) error {
	if f.createErr != nil {
		return f.createErr
	}
	b.ID, b.Version = 1, 1
	f.created = b
	return nil
}

func (f *fakeBooks) CreateAll(context.Context, []*book.Book) error { return nil }

func (f *fakeBooks) Update(_ context.Context, b *book.Book) error {
	if f.updateErr != nil {
		return f.updateErr
	}
	b.Version++
	f.updated = b
	return nil
}

func (f *fakeBooks) Delete(_ context.Context, id book.ID) error {
	f.deletedID = id
	return f.deleteErr
}

// fakeCatalog は book.BookCatalog の手書き Fake。
type fakeCatalog struct {
	entry  *book.CatalogEntry
	err    error
	called int
}

func (f *fakeCatalog) Lookup(context.Context, book.ISBN) (*book.CatalogEntry, error) {
	f.called++
	return f.entry, f.err
}

// fakeDetails は book.Query の手書き Fake。
type fakeDetails struct {
	detail *book.BookDetail
	gotID  book.ID
	// isbnExists は ExistsByISBN の応答（登録済みの ISBN を模す）。
	isbnExists bool
	existsErr  error
}

func (f *fakeDetails) ExistsByISBN(context.Context, book.ISBN) (bool, error) {
	return f.isbnExists, f.existsErr
}

func (f *fakeDetails) FindDetailByID(_ context.Context, id book.ID) (*book.BookDetail, error) {
	f.gotID = id
	return f.detail, nil
}

func (f *fakeDetails) Search(context.Context, book.ListCondition, common.ListRange) (*book.BookList, error) {
	return nil, nil
}

func catalogEntry(t *testing.T, title string, cover *book.Cover) *book.CatalogEntry {
	t.Helper()
	isbn, _ := book.NewISBN("9784873118703")
	bib, err := book.NewBibliography(title, "Kleppmann,Martin", "オーム社", "201907")
	if err != nil {
		t.Fatal(err)
	}
	return &book.CatalogEntry{ISBN: isbn, Bibliography: bib, Cover: cover}
}

// fakeTagQuery は tag.Query の手書き Fake。
type fakeTagQuery struct {
	exists    bool
	existsErr error
	gotIDs    []tag.ID
}

func (f *fakeTagQuery) FindAll(context.Context) (*tag.TagList, error) { return nil, nil }

func (f *fakeTagQuery) CountBooks(context.Context) (*tag.TagBookCounts, error) { return nil, nil }

func (f *fakeTagQuery) ExistsAll(_ context.Context, ids []tag.ID) (bool, error) {
	f.gotIDs = ids
	return f.exists, f.existsErr
}

func (f *fakeTagQuery) ExistsByName(context.Context, string, tag.ID) (bool, error) { return false, nil }

// violationsOf はエラーから違反の一覧を取り出す（Violations でなければテストを落とす）。
func violationsOf(t *testing.T, err error) common.Violations {
	t.Helper()
	var vs common.Violations
	if !errors.As(err, &vs) {
		t.Fatalf("err = %v, want Violations", err)
	}
	return vs
}

func mustCover(t *testing.T, url string) *book.Cover {
	t.Helper()
	c, err := book.NewCover(url, book.CoverSourceOpenBD)
	if err != nil {
		t.Fatal(err)
	}
	return &c
}

func TestRegisterBookUsecase_Execute(t *testing.T) {
	t.Parallel()
	valid := command.RegisterBookCommand{ISBN: "4873118700", Summary: "分散データの設計を学べる", TagIDs: []int64{1, 2}, Comment: "良書", Rating: 5}

	t.Run("カタログの書誌・書影と感想・評価を組み合わせて登録し詳細を返す", func(t *testing.T) {
		t.Parallel()
		cover := mustCover(t, "https://cover.openbd.jp/9784873118703.jpg")
		books := &fakeBooks{}
		want := &book.BookDetail{ID: 1}
		details := &fakeDetails{detail: want}
		uc := &command.RegisterBookUsecaseImpl{
			Books: books, Catalog: &fakeCatalog{entry: catalogEntry(t, "データ指向アプリケーションデザイン", cover)},
			Details: details, Tags: &fakeTagQuery{exists: true},
		}

		got, err := uc.Execute(context.Background(), valid)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		c := books.created
		if c == nil || c.ISBN.String() != "9784873118703" || c.Bibliography.Title() != "データ指向アプリケーションデザイン" ||
			c.Cover == nil || *c.Cover != *cover || c.Summary.String() != "分散データの設計を学べる" ||
			c.Comment.String() != "良書" || c.Rating.Int() != 5 || c.TitleOverride != nil {
			t.Fatalf("created %+v", c)
		}
		if ids := c.Tags.IDs(); len(ids) != 2 || ids[0] != 1 || ids[1] != 2 {
			t.Fatalf("Tags = %v, want [1 2]", ids)
		}
		if got != want || details.gotID != 1 {
			t.Fatalf("got %+v (detail looked up for %d), want the created book's detail", got, details.gotID)
		}
	})

	t.Run("書名の上書きを指定すると付けて登録する", func(t *testing.T) {
		t.Parallel()
		cmd := valid
		cmd.TitleOverride = "正しい書名"
		books := &fakeBooks{}
		uc := &command.RegisterBookUsecaseImpl{Books: books, Catalog: &fakeCatalog{entry: catalogEntry(t, "カタログの書名", nil)}, Details: &fakeDetails{detail: &book.BookDetail{}}, Tags: &fakeTagQuery{exists: true}}

		if _, err := uc.Execute(context.Background(), cmd); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if books.created == nil || books.created.TitleOverride == nil || books.created.TitleOverride.String() != "正しい書名" {
			t.Fatalf("created %+v, want TitleOverride = 正しい書名", books.created)
		}
	})

	t.Run("空白だけの書名の上書きはエラーにせず上書きなしで登録する", func(t *testing.T) {
		t.Parallel()
		cmd := valid
		cmd.TitleOverride = "　"
		books := &fakeBooks{}
		uc := &command.RegisterBookUsecaseImpl{Books: books, Catalog: &fakeCatalog{entry: catalogEntry(t, "カタログの書名", nil)}, Details: &fakeDetails{detail: &book.BookDetail{}}, Tags: &fakeTagQuery{exists: true}}

		if _, err := uc.Execute(context.Background(), cmd); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if books.created == nil || books.created.TitleOverride != nil {
			t.Fatalf("created %+v, want TitleOverride = nil", books.created)
		}
	})

	invalid := []struct {
		name string
		cmd  command.RegisterBookCommand
		tags *fakeTagQuery
		want common.Violations
	}{
		{name: "不正なISBN", cmd: command.RegisterBookCommand{ISBN: "123", Summary: "要約", Comment: "良書", Rating: 5}, want: common.Violations{{Field: "isbn", Rule: "format"}}},
		{name: "空の感想", cmd: command.RegisterBookCommand{ISBN: "4873118700", Summary: "要約", Comment: " ", Rating: 5}, want: common.Violations{{Field: "comment", Rule: "invalid"}}},
		{name: "範囲外の評価", cmd: command.RegisterBookCommand{ISBN: "4873118700", Summary: "要約", Comment: "良書", Rating: 6}, want: common.Violations{{Field: "rating", Rule: "invalid"}}},
		{name: "空の一言まとめ", cmd: command.RegisterBookCommand{ISBN: "4873118700", Summary: " ", Comment: "良書", Rating: 5}, want: common.Violations{{Field: "summary", Rule: "invalid"}}},
		{name: "改行を含む一言まとめ", cmd: command.RegisterBookCommand{ISBN: "4873118700", Summary: "a\nb", Comment: "良書", Rating: 5}, want: common.Violations{{Field: "summary", Rule: "invalid"}}},
		{name: "256文字の書名の上書き", cmd: command.RegisterBookCommand{ISBN: "4873118700", Summary: "要約", TitleOverride: strings.Repeat("あ", 256), Comment: "良書", Rating: 5}, want: common.Violations{{Field: "titleOverride", Rule: "invalid"}}},
		{name: "存在しないタグID", cmd: command.RegisterBookCommand{ISBN: "4873118700", Summary: "要約", TagIDs: []int64{999}, Comment: "良書", Rating: 5}, tags: &fakeTagQuery{exists: false}, want: common.Violations{{Field: "tagIds", Rule: "tagIdsNotFound"}}},
		{name: "タグ11個", cmd: command.RegisterBookCommand{ISBN: "4873118700", Summary: "要約", TagIDs: []int64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11}, Comment: "良書", Rating: 5}, want: common.Violations{{Field: "tagIds", Rule: "max"}}},
		{name: "重複したタグID", cmd: command.RegisterBookCommand{ISBN: "4873118700", Summary: "要約", TagIDs: []int64{1, 1}, Comment: "良書", Rating: 5}, want: common.Violations{{Field: "tagIds", Rule: "tagIdsDuplicated"}}},
	}
	for _, tt := range invalid {
		t.Run(tt.name+"はカタログも保存も呼ばずに違反を返す", func(t *testing.T) {
			t.Parallel()
			books := &fakeBooks{}
			catalog := &fakeCatalog{}
			tags := tt.tags
			if tags == nil {
				tags = &fakeTagQuery{exists: true}
			}
			uc := &command.RegisterBookUsecaseImpl{Books: books, Catalog: catalog, Details: &fakeDetails{}, Tags: tags}

			_, err := uc.Execute(context.Background(), tt.cmd)
			if !errors.Is(err, common.ErrInvalid) {
				t.Fatalf("err = %v, want ErrInvalid", err)
			}
			if got := violationsOf(t, err); !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("violations = %+v, want %+v", got, tt.want)
			}
			if catalog.called != 0 || books.created != nil {
				t.Fatal("neither the catalog nor the repository may be called for invalid input")
			}
		})
	}

	t.Run("複数の違反は一度にすべて返す（評価・タグ・ISBN の重複）", func(t *testing.T) {
		t.Parallel()
		books := &fakeBooks{}
		catalog := &fakeCatalog{}
		uc := &command.RegisterBookUsecaseImpl{Books: books, Catalog: catalog, Details: &fakeDetails{isbnExists: true}, Tags: &fakeTagQuery{exists: false}}
		cmd := valid
		cmd.Rating = 0
		cmd.TagIDs = []int64{999}

		_, err := uc.Execute(context.Background(), cmd)
		want := common.Violations{{Field: "rating", Rule: "invalid"}, {Field: "tagIds", Rule: "tagIdsNotFound"}, {Field: "isbn", Rule: "isbnDuplicated"}}
		if got := violationsOf(t, err); !reflect.DeepEqual(got, want) {
			t.Fatalf("violations = %+v, want %+v", got, want)
		}
		if catalog.called != 0 || books.created != nil {
			t.Fatal("neither the catalog nor the repository may be called when there are violations")
		}
	})

	t.Run("登録済みのISBNはカタログを呼ばず isbnDuplicated の違反", func(t *testing.T) {
		t.Parallel()
		books := &fakeBooks{}
		catalog := &fakeCatalog{entry: catalogEntry(t, "x", nil)}
		uc := &command.RegisterBookUsecaseImpl{Books: books, Catalog: catalog, Details: &fakeDetails{isbnExists: true}, Tags: &fakeTagQuery{exists: true}}

		_, err := uc.Execute(context.Background(), valid)
		if want := (common.Violations{{Field: "isbn", Rule: "isbnDuplicated"}}); !reflect.DeepEqual(violationsOf(t, err), want) {
			t.Fatalf("err = %v, want %+v", err, want)
		}
		if catalog.called != 0 || books.created != nil {
			t.Fatal("neither the catalog nor the repository may be called for a duplicated ISBN")
		}
	})

	t.Run("不正なISBNでは重複の検査をしない", func(t *testing.T) {
		t.Parallel()
		details := &fakeDetails{existsErr: errors.New("must not be called")}
		uc := &command.RegisterBookUsecaseImpl{Books: &fakeBooks{}, Catalog: &fakeCatalog{}, Details: details, Tags: &fakeTagQuery{exists: true}}
		cmd := valid
		cmd.ISBN = "123"

		_, err := uc.Execute(context.Background(), cmd)
		if want := (common.Violations{{Field: "isbn", Rule: "format"}}); !reflect.DeepEqual(violationsOf(t, err), want) {
			t.Fatalf("err = %v, want %+v", err, want)
		}
	})

	t.Run("重複検査の障害はそのまま返す", func(t *testing.T) {
		t.Parallel()
		wantErr := errors.New("db: timeout")
		uc := &command.RegisterBookUsecaseImpl{Books: &fakeBooks{}, Catalog: &fakeCatalog{}, Details: &fakeDetails{existsErr: wantErr}, Tags: &fakeTagQuery{exists: true}}
		if _, err := uc.Execute(context.Background(), valid); !errors.Is(err, wantErr) {
			t.Fatalf("err = %v, want %v", err, wantErr)
		}
	})

	t.Run("カタログに無いISBNは isbnNotInCatalog の違反（NotFound ではない）", func(t *testing.T) {
		t.Parallel()
		books := &fakeBooks{}
		uc := &command.RegisterBookUsecaseImpl{Books: books, Catalog: &fakeCatalog{err: common.ErrNotFound}, Details: &fakeDetails{}, Tags: &fakeTagQuery{exists: true}}

		_, err := uc.Execute(context.Background(), valid)
		if errors.Is(err, common.ErrNotFound) {
			t.Fatalf("err = %v, must not be ErrNotFound (the resource being created isn't missing)", err)
		}
		if want := (common.Violations{{Field: "isbn", Rule: "isbnNotInCatalog"}}); !reflect.DeepEqual(violationsOf(t, err), want) {
			t.Fatalf("err = %v, want %+v", err, want)
		}
		if books.created != nil {
			t.Fatal("Repository.Create must not be called")
		}
	})

	t.Run("カタログの障害はそのまま返り保存しない", func(t *testing.T) {
		t.Parallel()
		wantErr := errors.New("openbd: timeout")
		books := &fakeBooks{}
		uc := &command.RegisterBookUsecaseImpl{Books: books, Catalog: &fakeCatalog{err: wantErr}, Details: &fakeDetails{}, Tags: &fakeTagQuery{exists: true}}

		if _, err := uc.Execute(context.Background(), valid); !errors.Is(err, wantErr) {
			t.Fatalf("err = %v, want %v", err, wantErr)
		}
		if books.created != nil {
			t.Fatal("Repository.Create must not be called")
		}
	})

	t.Run("同時登録で一意制約に当たったときはConflictを返し詳細は取得しない", func(t *testing.T) {
		t.Parallel()
		details := &fakeDetails{}
		uc := &command.RegisterBookUsecaseImpl{Books: &fakeBooks{createErr: common.ErrConflict}, Catalog: &fakeCatalog{entry: catalogEntry(t, "x", nil)}, Details: details, Tags: &fakeTagQuery{exists: true}}

		if _, err := uc.Execute(context.Background(), valid); !errors.Is(err, common.ErrConflict) {
			t.Fatalf("err = %v, want ErrConflict", err)
		}
		if details.gotID != 0 {
			t.Fatal("the detail must not be queried when saving fails")
		}
	})
}
