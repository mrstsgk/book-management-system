package book_test

import (
	"reflect"
	"testing"

	"github.com/mrstsgk/book-management-system/backend/internal/domain/book"
)

func mustBook(t *testing.T) *book.Book {
	t.Helper()
	isbn, err := book.NewISBN("9784873118703")
	if err != nil {
		t.Fatal(err)
	}
	bib, err := book.NewBibliography("データ指向アプリケーションデザイン", "Kleppmann,Martin", "オーム社", "201907")
	if err != nil {
		t.Fatal(err)
	}
	cover, err := book.NewCover("https://cover.openbd.jp/9784873118703.jpg", book.CoverSourceOpenBD)
	if err != nil {
		t.Fatal(err)
	}
	comment, err := book.NewComment("良書")
	if err != nil {
		t.Fatal(err)
	}
	rating, err := book.NewRating(5)
	if err != nil {
		t.Fatal(err)
	}
	b := book.New(isbn, bib, &cover, comment, rating)
	cover = book.Cover{} // 呼び出し側の変数を変えても本には影響しないこと
	return b
}

func TestNew_SetsFieldsCopiesCoverAndLeavesIdentityUnassigned(t *testing.T) {
	t.Parallel()
	b := mustBook(t)

	if b.ID != 0 || b.Version != 0 {
		t.Fatalf("ID/Version = %d/%d, want zero values before saving", b.ID, b.Version)
	}
	if b.ISBN.String() != "9784873118703" || b.Bibliography.Title() != "データ指向アプリケーションデザイン" ||
		b.Comment.String() != "良書" || b.Rating.Int() != 5 {
		t.Fatalf("got %+v", b)
	}
	if b.Cover == nil || b.Cover.URL() != "https://cover.openbd.jp/9784873118703.jpg" {
		t.Fatalf("Cover = %+v, want a copy of the given cover", b.Cover)
	}
}

func TestBook_ChangeReview(t *testing.T) {
	t.Parallel()
	b := mustBook(t)
	comment, _ := book.NewComment("読み返して評価が変わった")
	rating, _ := book.NewRating(4)

	b.ChangeReview(comment, rating, 3)

	if b.Comment != comment || b.Rating != rating || b.Version != 3 {
		t.Fatalf("got comment=%q rating=%d version=%d", b.Comment.String(), b.Rating.Int(), b.Version)
	}
	if b.Bibliography.Title() != "データ指向アプリケーションデザイン" || b.Cover == nil {
		t.Fatal("changing the review must not touch the catalog data")
	}
}

func TestBook_RefreshCatalog(t *testing.T) {
	t.Parallel()
	b := mustBook(t)
	bib, _ := book.NewBibliography("データ指向アプリケーションデザイン 第2版", "Kleppmann,Martin", "オライリー・ジャパン", "202601")

	b.RefreshCatalog(bib, nil)

	if b.Bibliography != bib || b.Cover != nil {
		t.Fatalf("got bibliography=%+v cover=%v, want the refreshed bibliography and no cover", b.Bibliography, b.Cover)
	}
	if b.Comment.String() != "良書" || b.Rating.Int() != 5 {
		t.Fatal("refreshing the catalog must not touch the review")
	}
}

// Read Model は書き込み側の VO・Entity に依存しない（docs/rules/testing.md）。
func TestReadModels_UsePlainFieldTypes(t *testing.T) {
	t.Parallel()
	writeSide := map[reflect.Type]bool{
		reflect.TypeOf(book.ISBN{}):          true,
		reflect.TypeOf(book.Bibliography{}):  true,
		reflect.TypeOf(book.Cover{}):         true,
		reflect.TypeOf(book.CoverSource("")): true,
		reflect.TypeOf(book.Comment{}):       true,
		reflect.TypeOf(book.Rating{}):        true,
		reflect.TypeOf(book.Book{}):          true,
	}
	for _, typ := range []reflect.Type{reflect.TypeOf(book.BookDetail{}), reflect.TypeOf(book.BookListItem{})} {
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
