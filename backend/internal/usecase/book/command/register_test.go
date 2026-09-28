package command_test

import (
	"context"
	"errors"
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
}

func (f *fakeDetails) FindDetailByID(_ context.Context, id book.ID) (*book.BookDetail, error) {
	f.gotID = id
	return f.detail, nil
}

func (f *fakeDetails) FindList(context.Context, common.ListRange) (*book.BookList, error) {
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

func (f *fakeTagQuery) FindList(context.Context) (*tag.TagList, error) { return nil, nil }

func (f *fakeTagQuery) ExistsAll(_ context.Context, ids []tag.ID) (bool, error) {
	f.gotIDs = ids
	return f.exists, f.existsErr
}

func mustCover(t *testing.T, url string) *book.Cover {
	t.Helper()
	c, err := book.NewCover(url, book.CoverSourceOpenBD)
	if err != nil {
		t.Fatal(err)
	}
	return &c
}

func TestRegisterUsecase_Execute(t *testing.T) {
	t.Parallel()
	valid := command.RegisterCommand{ISBN: "4873118700", Summary: "分散データの設計を学べる", TagIDs: []int64{1, 2}, Comment: "良書", Rating: 5}

	t.Run("カタログの書誌・書影と感想・評価を組み合わせて登録し詳細を返す", func(t *testing.T) {
		t.Parallel()
		cover := mustCover(t, "https://cover.openbd.jp/9784873118703.jpg")
		books := &fakeBooks{}
		want := &book.BookDetail{ID: 1}
		details := &fakeDetails{detail: want}
		uc := &command.RegisterUsecaseImpl{
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
		uc := &command.RegisterUsecaseImpl{Books: books, Catalog: &fakeCatalog{entry: catalogEntry(t, "カタログの書名", nil)}, Details: &fakeDetails{detail: &book.BookDetail{}}, Tags: &fakeTagQuery{exists: true}}

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
		uc := &command.RegisterUsecaseImpl{Books: books, Catalog: &fakeCatalog{entry: catalogEntry(t, "カタログの書名", nil)}, Details: &fakeDetails{detail: &book.BookDetail{}}, Tags: &fakeTagQuery{exists: true}}

		if _, err := uc.Execute(context.Background(), cmd); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if books.created == nil || books.created.TitleOverride != nil {
			t.Fatalf("created %+v, want TitleOverride = nil", books.created)
		}
	})

	invalid := []struct {
		name string
		cmd  command.RegisterCommand
	}{
		{name: "不正なISBN", cmd: command.RegisterCommand{ISBN: "123", Summary: "要約", Comment: "良書", Rating: 5}},
		{name: "空の感想", cmd: command.RegisterCommand{ISBN: "4873118700", Summary: "要約", Comment: " ", Rating: 5}},
		{name: "範囲外の評価", cmd: command.RegisterCommand{ISBN: "4873118700", Summary: "要約", Comment: "良書", Rating: 6}},
		{name: "空の一言まとめ", cmd: command.RegisterCommand{ISBN: "4873118700", Summary: " ", Comment: "良書", Rating: 5}},
		{name: "改行を含む一言まとめ", cmd: command.RegisterCommand{ISBN: "4873118700", Summary: "a\nb", Comment: "良書", Rating: 5}},
		{name: "256文字の書名の上書き", cmd: command.RegisterCommand{ISBN: "4873118700", Summary: "要約", TitleOverride: strings.Repeat("あ", 256), Comment: "良書", Rating: 5}},
	}
	for _, tt := range invalid {
		t.Run(tt.name+"はカタログも保存も呼ばずにエラー", func(t *testing.T) {
			t.Parallel()
			books := &fakeBooks{}
			catalog := &fakeCatalog{}
			uc := &command.RegisterUsecaseImpl{Books: books, Catalog: catalog, Details: &fakeDetails{}, Tags: &fakeTagQuery{exists: true}}

			if _, err := uc.Execute(context.Background(), tt.cmd); !errors.Is(err, common.ErrInvalid) {
				t.Fatalf("err = %v, want ErrInvalid", err)
			}
			if catalog.called != 0 || books.created != nil {
				t.Fatal("neither the catalog nor the repository may be called for invalid input")
			}
		})
	}

	t.Run("存在しないタグIDはカタログも保存も呼ばずにエラー", func(t *testing.T) {
		t.Parallel()
		books := &fakeBooks{}
		catalog := &fakeCatalog{}
		tags := &fakeTagQuery{exists: false}
		uc := &command.RegisterUsecaseImpl{Books: books, Catalog: catalog, Details: &fakeDetails{}, Tags: tags}
		cmd := valid
		cmd.TagIDs = []int64{999}

		if _, err := uc.Execute(context.Background(), cmd); !errors.Is(err, common.ErrInvalid) {
			t.Fatalf("err = %v, want ErrInvalid", err)
		}
		if catalog.called != 0 || books.created != nil {
			t.Fatal("neither the catalog nor the repository may be called for invalid input")
		}
	})

	t.Run("タグを11個指定するとカタログも保存も呼ばずにエラー", func(t *testing.T) {
		t.Parallel()
		books := &fakeBooks{}
		catalog := &fakeCatalog{}
		uc := &command.RegisterUsecaseImpl{Books: books, Catalog: catalog, Details: &fakeDetails{}, Tags: &fakeTagQuery{exists: true}}
		cmd := valid
		cmd.TagIDs = []int64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11}

		if _, err := uc.Execute(context.Background(), cmd); !errors.Is(err, common.ErrInvalid) {
			t.Fatalf("err = %v, want ErrInvalid", err)
		}
		if catalog.called != 0 || books.created != nil {
			t.Fatal("neither the catalog nor the repository may be called for invalid input")
		}
	})

	t.Run("カタログに無いISBNは登録できない入力として扱う", func(t *testing.T) {
		t.Parallel()
		books := &fakeBooks{}
		uc := &command.RegisterUsecaseImpl{Books: books, Catalog: &fakeCatalog{err: common.ErrNotFound}, Details: &fakeDetails{}, Tags: &fakeTagQuery{exists: true}}

		_, err := uc.Execute(context.Background(), valid)
		if !errors.Is(err, common.ErrInvalid) || errors.Is(err, common.ErrNotFound) {
			t.Fatalf("err = %v, want ErrInvalid (not ErrNotFound: the resource being created isn't missing)", err)
		}
		if books.created != nil {
			t.Fatal("Repository.Create must not be called")
		}
	})

	t.Run("カタログの障害はそのまま返り保存しない", func(t *testing.T) {
		t.Parallel()
		wantErr := errors.New("openbd: timeout")
		books := &fakeBooks{}
		uc := &command.RegisterUsecaseImpl{Books: books, Catalog: &fakeCatalog{err: wantErr}, Details: &fakeDetails{}, Tags: &fakeTagQuery{exists: true}}

		if _, err := uc.Execute(context.Background(), valid); !errors.Is(err, wantErr) {
			t.Fatalf("err = %v, want %v", err, wantErr)
		}
		if books.created != nil {
			t.Fatal("Repository.Create must not be called")
		}
	})

	t.Run("同じISBNの登録済みはConflictを返し詳細は取得しない", func(t *testing.T) {
		t.Parallel()
		details := &fakeDetails{}
		uc := &command.RegisterUsecaseImpl{Books: &fakeBooks{createErr: common.ErrConflict}, Catalog: &fakeCatalog{entry: catalogEntry(t, "x", nil)}, Details: details, Tags: &fakeTagQuery{exists: true}}

		if _, err := uc.Execute(context.Background(), valid); !errors.Is(err, common.ErrConflict) {
			t.Fatalf("err = %v, want ErrConflict", err)
		}
		if details.gotID != 0 {
			t.Fatal("the detail must not be queried when saving fails")
		}
	})
}
