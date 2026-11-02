package domain_test

import (
	"reflect"
	"testing"

	bookdomain "github.com/mrstsgk/book-management-system/backend/internal/book/domain"
	tagdomain "github.com/mrstsgk/book-management-system/backend/internal/tag/domain"
)

func mustBook(t *testing.T) *bookdomain.Book {
	t.Helper()
	isbn, err := bookdomain.NewISBN("9784873118703")
	if err != nil {
		t.Fatal(err)
	}
	bib, err := bookdomain.NewBibliography("データ指向アプリケーションデザイン", "Kleppmann,Martin", "オーム社", "201907")
	if err != nil {
		t.Fatal(err)
	}
	cover, err := bookdomain.NewCover("https://cover.openbd.jp/9784873118703.jpg", bookdomain.CoverSourceOpenBD)
	if err != nil {
		t.Fatal(err)
	}
	summary, err := bookdomain.NewSummary("分散データの設計を体系的に学べる")
	if err != nil {
		t.Fatal(err)
	}
	comment, err := bookdomain.NewComment("良書")
	if err != nil {
		t.Fatal(err)
	}
	rating, err := bookdomain.NewRating(5)
	if err != nil {
		t.Fatal(err)
	}
	b := bookdomain.New(isbn, bib, &cover, summary, comment, rating, bookdomain.TagSelection{})
	cover = bookdomain.Cover{} // 呼び出し側の変数を変えても本には影響しないこと
	return b
}

func TestNew_SetsFieldsCopiesCoverAndLeavesIdentityUnassigned(t *testing.T) {
	t.Parallel()
	b := mustBook(t)

	if b.ID != 0 || b.Version != 0 {
		t.Fatalf("ID/Version = %d/%d, want zero values before saving", b.ID, b.Version)
	}
	if b.ISBN.String() != "9784873118703" || b.Bibliography.Title() != "データ指向アプリケーションデザイン" ||
		b.Summary.String() != "分散データの設計を体系的に学べる" || b.Comment.String() != "良書" || b.Rating.Int() != 5 {
		t.Fatalf("got %+v", b)
	}
	if b.Cover == nil || b.Cover.URL() != "https://cover.openbd.jp/9784873118703.jpg" {
		t.Fatalf("Cover = %+v, want a copy of the given cover", b.Cover)
	}
	if b.TitleOverride != nil {
		t.Fatalf("TitleOverride = %v, want nil for a newly created book", b.TitleOverride)
	}
	if len(b.Tags.IDs()) != 0 {
		t.Fatalf("Tags = %v, want empty for a newly created book", b.Tags.IDs())
	}
}

func TestBook_ChangeReview(t *testing.T) {
	t.Parallel()
	b := mustBook(t)
	summary, _ := bookdomain.NewSummary("読み返して見方が変わった")
	comment, _ := bookdomain.NewComment("読み返して評価が変わった")
	rating, _ := bookdomain.NewRating(4)
	tags, _ := bookdomain.NewTagSelection([]tagdomain.ID{1, 2})

	b.ChangeReview(summary, comment, rating, tags, 3)

	if b.Summary != summary || b.Comment != comment || b.Rating != rating || b.Version != 3 {
		t.Fatalf("got summary=%q comment=%q rating=%d version=%d", b.Summary.String(), b.Comment.String(), b.Rating.Int(), b.Version)
	}
	if got := b.Tags.IDs(); len(got) != 2 || got[0] != 1 || got[1] != 2 {
		t.Fatalf("Tags = %v, want [1 2]", got)
	}
	if b.Bibliography.Title() != "データ指向アプリケーションデザイン" || b.Cover == nil {
		t.Fatal("changing the review must not touch the catalog data")
	}
}

func TestBook_OverrideTitle(t *testing.T) {
	t.Parallel()

	t.Run("上書きを付けるとコピーして持つ", func(t *testing.T) {
		t.Parallel()
		b := mustBook(t)
		title, _ := bookdomain.NewTitle("正しい書名")
		b.OverrideTitle(&title)
		title = bookdomain.Title{}
		if b.TitleOverride == nil || b.TitleOverride.String() != "正しい書名" {
			t.Fatalf("TitleOverride = %v, want a copy of the given title", b.TitleOverride)
		}
	})

	t.Run("nilで上書きを外す", func(t *testing.T) {
		t.Parallel()
		b := mustBook(t)
		title, _ := bookdomain.NewTitle("正しい書名")
		b.OverrideTitle(&title)
		b.OverrideTitle(nil)
		if b.TitleOverride != nil {
			t.Fatalf("TitleOverride = %v, want nil", b.TitleOverride)
		}
	})

	t.Run("外部カタログを取り直しても上書きは残る", func(t *testing.T) {
		t.Parallel()
		b := mustBook(t)
		title, _ := bookdomain.NewTitle("正しい書名")
		b.OverrideTitle(&title)
		bib, _ := bookdomain.NewBibliography("カタログの別の書名", "", "", "")
		b.RefreshCatalog(bib, nil)
		if b.TitleOverride == nil || b.TitleOverride.String() != "正しい書名" {
			t.Fatalf("TitleOverride = %v, want it kept after refreshing the catalog", b.TitleOverride)
		}
	})
}

func TestBook_RefreshCatalog(t *testing.T) {
	t.Parallel()
	b := mustBook(t)
	bib, _ := bookdomain.NewBibliography("データ指向アプリケーションデザイン 第2版", "Kleppmann,Martin", "オライリー・ジャパン", "202601")

	b.RefreshCatalog(bib, nil)

	if b.Bibliography != bib || b.Cover != nil {
		t.Fatalf("got bibliography=%+v cover=%v, want the refreshed bibliography and no cover", b.Bibliography, b.Cover)
	}
	if b.Comment.String() != "良書" || b.Rating.Int() != 5 {
		t.Fatal("refreshing the catalog must not touch the review")
	}
}

func TestBook_RefreshCatalog_ReplacesCover(t *testing.T) {
	t.Parallel()
	b := mustBook(t)
	bib, _ := bookdomain.NewBibliography("データ指向アプリケーションデザイン", "Kleppmann,Martin", "オーム社", "201907")
	oc, _ := bookdomain.NewCover("https://cover.openbd.jp/new.jpg", bookdomain.CoverSourceOpenBD)

	b.RefreshCatalog(bib, &oc)
	oc = bookdomain.Cover{}

	if b.Cover == nil || b.Cover.URL() != "https://cover.openbd.jp/new.jpg" {
		t.Fatalf("cover=%v, want the refreshed openBD cover kept independent of the caller's variable", b.Cover)
	}
}

// Read Model は書き込み側の VO・Entity に依存しない（docs/rules/testing.md）。
func TestReadModels_UsePlainFieldTypes(t *testing.T) {
	t.Parallel()
	writeSide := map[reflect.Type]bool{
		reflect.TypeFor[bookdomain.ISBN]():         true,
		reflect.TypeFor[bookdomain.Bibliography](): true,
		reflect.TypeFor[bookdomain.Cover]():        true,
		reflect.TypeFor[bookdomain.CoverSource]():  true,
		reflect.TypeFor[bookdomain.Title]():        true,
		reflect.TypeFor[bookdomain.Summary]():      true,
		reflect.TypeFor[bookdomain.Comment]():      true,
		reflect.TypeFor[bookdomain.Rating]():       true,
		reflect.TypeFor[bookdomain.TagSelection](): true,
		reflect.TypeFor[bookdomain.Book]():         true,
	}
	for _, typ := range []reflect.Type{reflect.TypeFor[bookdomain.BookDetail](), reflect.TypeFor[bookdomain.BookListItem]()} {
		for i := range typ.NumField() {
			ft := typ.Field(i).Type
			for ft.Kind() == reflect.Pointer || ft.Kind() == reflect.Slice {
				ft = ft.Elem()
			}
			if writeSide[ft] {
				t.Errorf("%s.%s uses write-side type %s", typ.Name(), typ.Field(i).Name, ft)
			}
		}
	}
}
