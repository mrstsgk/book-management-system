package book_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/textproto"
	"reflect"
	"strings"
	"testing"

	domainbook "github.com/mrstsgk/book-management-system/backend/internal/domain/book"
	domaincommon "github.com/mrstsgk/book-management-system/backend/internal/domain/common"
	httpbook "github.com/mrstsgk/book-management-system/backend/internal/presentation/http/book"
	"github.com/mrstsgk/book-management-system/backend/internal/presentation/http/common"
	bookcmd "github.com/mrstsgk/book-management-system/backend/internal/usecase/book/command"
)

type fakeCreate func(ctx context.Context, cmd bookcmd.CreateCommand) (*domainbook.BookDetail, error)

func (f fakeCreate) Execute(ctx context.Context, cmd bookcmd.CreateCommand) (*domainbook.BookDetail, error) {
	return f(ctx, cmd)
}

type fakeUpdate func(ctx context.Context, cmd bookcmd.UpdateCommand) (*domainbook.BookDetail, error)

func (f fakeUpdate) Execute(ctx context.Context, cmd bookcmd.UpdateCommand) (*domainbook.BookDetail, error) {
	return f(ctx, cmd)
}

// serve wires the handler the same way cmd/api/main.go does (NewEcho + Register).
func serve(t *testing.T, h *httpbook.Handler, method, path, body string) *httptest.ResponseRecorder {
	t.Helper()
	orig := slog.Default()
	slog.SetDefault(slog.New(slog.NewTextHandler(io.Discard, nil)))
	t.Cleanup(func() { slog.SetDefault(orig) })

	e := common.NewEcho()
	h.Register(e.Group("/api/books"))
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	return rec
}

func decode[T any](t *testing.T, rec *httptest.ResponseRecorder) T {
	t.Helper()
	var v T
	if err := json.Unmarshal(rec.Body.Bytes(), &v); err != nil {
		t.Fatalf("invalid JSON body %q: %v", rec.Body.String(), err)
	}
	return v
}

var detail = &domainbook.BookDetail{
	ID: 1, Title: "人間失格", Price: 1500, Status: 2, Version: 1,
	Authors: []domainbook.AuthorSummary{{ID: 1, Name: "太宰治", Version: 1}},
}

var detailResponse = httpbook.Response{
	ID: 1, Title: "人間失格", Price: 1500, Status: 2, Version: 1,
	Authors: []httpbook.AuthorResponse{{ID: 1, Name: "太宰治", Version: 1}},
}

func TestHandlerCreate(t *testing.T) {
	t.Run("作成して200で書籍詳細を返す", func(t *testing.T) {
		var got bookcmd.CreateCommand
		h := &httpbook.Handler{CreateUC: fakeCreate(func(_ context.Context, cmd bookcmd.CreateCommand) (*domainbook.BookDetail, error) {
			got = cmd
			return detail, nil
		})}
		rec := serve(t, h, http.MethodPost, "/api/books", `{"title":"人間失格","price":0,"authorIds":[1,2],"status":2}`)

		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200 (body=%s)", rec.Code, rec.Body.String())
		}
		want := bookcmd.CreateCommand{Title: "人間失格", Price: 0, AuthorIDs: []int64{1, 2}, Status: 2}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("usecase received %+v, want %+v", got, want)
		}
		if res := decode[httpbook.Response](t, rec); !reflect.DeepEqual(res, detailResponse) {
			t.Fatalf("body = %+v, want %+v", res, detailResponse)
		}
	})

	invalid := []struct {
		name string
		body string
		want []common.FieldError
	}{
		{name: "必須項目なしは400", body: `{}`, want: []common.FieldError{
			{Field: "title", Rule: "required"}, {Field: "price", Rule: "required"},
			{Field: "authorIds", Rule: "required"}, {Field: "status", Rule: "required"},
		}},
		{name: "タイトル256文字は400", body: `{"title":"` + strings.Repeat("あ", 256) + `","price":1,"authorIds":[1],"status":1}`, want: []common.FieldError{{Field: "title", Rule: "max"}}},
		{name: "価格が負は400", body: `{"title":"a","price":-1,"authorIds":[1],"status":1}`, want: []common.FieldError{{Field: "price", Rule: "min"}}},
		{name: "価格が上限超過は400", body: `{"title":"a","price":100000000,"authorIds":[1],"status":1}`, want: []common.FieldError{{Field: "price", Rule: "max"}}},
		{name: "著者0人は400", body: `{"title":"a","price":1,"authorIds":[],"status":1}`, want: []common.FieldError{{Field: "authorIds", Rule: "min"}}},
		{name: "著者IDが0以下は400", body: `{"title":"a","price":1,"authorIds":[0],"status":1}`, want: []common.FieldError{{Field: "authorIds[0]", Rule: "gt"}}},
		{name: "不正な出版状況は400", body: `{"title":"a","price":1,"authorIds":[1],"status":3}`, want: []common.FieldError{{Field: "status", Rule: "oneof"}}},
	}
	for _, tt := range invalid {
		t.Run(tt.name, func(t *testing.T) {
			h := &httpbook.Handler{CreateUC: fakeCreate(func(context.Context, bookcmd.CreateCommand) (*domainbook.BookDetail, error) {
				t.Error("usecase must not be called")
				return nil, nil
			})}
			rec := serve(t, h, http.MethodPost, "/api/books", tt.body)

			if rec.Code != http.StatusBadRequest {
				t.Fatalf("status = %d, want 400", rec.Code)
			}
			if got := decode[common.ErrorResponse](t, rec).Errors; !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("errors = %+v, want %+v", got, tt.want)
			}
		})
	}

	t.Run("存在しない著者などドメインルール違反は400", func(t *testing.T) {
		h := &httpbook.Handler{CreateUC: fakeCreate(func(context.Context, bookcmd.CreateCommand) (*domainbook.BookDetail, error) {
			return nil, domaincommon.ErrInvalid
		})}
		if rec := serve(t, h, http.MethodPost, "/api/books", `{"title":"a","price":1,"authorIds":[999],"status":1}`); rec.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want 400", rec.Code)
		}
	})
}

func TestHandlerUpdate(t *testing.T) {
	body := `{"title":"人間失格","price":1500,"authorIds":[1],"status":2,"version":1}`

	t.Run("パスのIDとバージョンをusecaseに渡し200で返す", func(t *testing.T) {
		var got bookcmd.UpdateCommand
		h := &httpbook.Handler{UpdateUC: fakeUpdate(func(_ context.Context, cmd bookcmd.UpdateCommand) (*domainbook.BookDetail, error) {
			got = cmd
			return detail, nil
		})}
		rec := serve(t, h, http.MethodPut, "/api/books/5", body)

		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200 (body=%s)", rec.Code, rec.Body.String())
		}
		want := bookcmd.UpdateCommand{ID: 5, Title: "人間失格", Price: 1500, AuthorIDs: []int64{1}, Status: 2, Version: 1}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("usecase received %+v, want %+v", got, want)
		}
	})

	statuses := []struct {
		name string
		path string
		body string
		err  error
		want int
	}{
		{name: "不正なIDは400", path: "/api/books/abc", body: body, want: http.StatusBadRequest},
		{name: "バージョンなしは400", path: "/api/books/5", body: `{"title":"a","price":1,"authorIds":[1],"status":1}`, want: http.StatusBadRequest},
		{name: "出版済みから未出版への変更は400", path: "/api/books/5", body: body, err: domaincommon.ErrInvalid, want: http.StatusBadRequest},
		{name: "存在しない書籍は404", path: "/api/books/5", body: body, err: domaincommon.ErrNotFound, want: http.StatusNotFound},
		{name: "楽観的ロックの競合は409", path: "/api/books/5", body: body, err: domaincommon.ErrConflict, want: http.StatusConflict},
	}
	for _, tt := range statuses {
		t.Run(tt.name, func(t *testing.T) {
			h := &httpbook.Handler{UpdateUC: fakeUpdate(func(context.Context, bookcmd.UpdateCommand) (*domainbook.BookDetail, error) {
				if tt.err == nil {
					t.Error("usecase must not be called")
				}
				return nil, tt.err
			})}
			if rec := serve(t, h, http.MethodPut, tt.path, tt.body); rec.Code != tt.want {
				t.Fatalf("status = %d, want %d (body=%s)", rec.Code, tt.want, rec.Body.String())
			}
		})
	}
}

type fakeGet func(ctx context.Context, id int64) (*domainbook.BookDetail, error)

func (f fakeGet) Execute(ctx context.Context, id int64) (*domainbook.BookDetail, error) {
	return f(ctx, id)
}

type fakeUpload func(ctx context.Context, cmd bookcmd.UploadImageCommand) (*domainbook.BookDetail, error)

func (f fakeUpload) Execute(ctx context.Context, cmd bookcmd.UploadImageCommand) (*domainbook.BookDetail, error) {
	return f(ctx, cmd)
}

func strPtr(s string) *string { return &s }

func TestHandlerGet(t *testing.T) {
	t.Run("URLと画像URLを含む書籍詳細を返す", func(t *testing.T) {
		var gotID int64
		d := *detail
		d.AmazonURL = strPtr("https://www.amazon.co.jp/dp/4101006059")
		d.ImageURL = strPtr("http://localhost:4566/book-images/books/1/a.png?X-Amz-Expires=900")
		h := &httpbook.Handler{GetUC: fakeGet(func(_ context.Context, id int64) (*domainbook.BookDetail, error) {
			gotID = id
			return &d, nil
		})}
		rec := serve(t, h, http.MethodGet, "/api/books/1", "")

		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200 (body=%s)", rec.Code, rec.Body.String())
		}
		want := detailResponse
		want.AmazonURL = d.AmazonURL
		want.ImageURL = d.ImageURL
		if got := decode[httpbook.Response](t, rec); gotID != 1 || !reflect.DeepEqual(got, want) {
			t.Fatalf("id=%d body = %+v, want %+v", gotID, got, want)
		}
	})

	t.Run("URLも画像もなければnullで返す", func(t *testing.T) {
		h := &httpbook.Handler{GetUC: fakeGet(func(context.Context, int64) (*domainbook.BookDetail, error) { return detail, nil })}
		rec := serve(t, h, http.MethodGet, "/api/books/1", "")

		if !strings.Contains(rec.Body.String(), `"amazonUrl":null`) || !strings.Contains(rec.Body.String(), `"imageUrl":null`) {
			t.Fatalf("body = %s, want amazonUrl and imageUrl null", rec.Body.String())
		}
	})

	statuses := []struct {
		name string
		path string
		err  error
		want int
	}{
		{name: "不正なIDは400", path: "/api/books/abc", want: http.StatusBadRequest},
		{name: "存在しない書籍は404", path: "/api/books/1", err: domaincommon.ErrNotFound, want: http.StatusNotFound},
	}
	for _, tt := range statuses {
		t.Run(tt.name, func(t *testing.T) {
			h := &httpbook.Handler{GetUC: fakeGet(func(context.Context, int64) (*domainbook.BookDetail, error) {
				if tt.err == nil {
					t.Error("usecase must not be called")
				}
				return nil, tt.err
			})}
			if rec := serve(t, h, http.MethodGet, tt.path, ""); rec.Code != tt.want {
				t.Fatalf("status = %d, want %d", rec.Code, tt.want)
			}
		})
	}
}

func TestHandlerCreateAndUpdate_PassAmazonURL(t *testing.T) {
	const u = "https://www.amazon.co.jp/dp/4101006059"
	var created, updated *string
	h := &httpbook.Handler{
		CreateUC: fakeCreate(func(_ context.Context, cmd bookcmd.CreateCommand) (*domainbook.BookDetail, error) {
			created = cmd.AmazonURL
			return detail, nil
		}),
		UpdateUC: fakeUpdate(func(_ context.Context, cmd bookcmd.UpdateCommand) (*domainbook.BookDetail, error) {
			updated = cmd.AmazonURL
			return detail, nil
		}),
	}

	serve(t, h, http.MethodPost, "/api/books", `{"title":"a","price":1,"authorIds":[1],"status":1,"amazonUrl":"`+u+`"}`)
	serve(t, h, http.MethodPut, "/api/books/1", `{"title":"a","price":1,"authorIds":[1],"status":1,"amazonUrl":"`+u+`","version":1}`)

	if created == nil || *created != u || updated == nil || *updated != u {
		t.Fatalf("create got %v, update got %v, want %s for both", created, updated, u)
	}

	rec := serve(t, h, http.MethodPost, "/api/books", `{"title":"a","price":1,"authorIds":[1],"status":1,"amazonUrl":"`+"https://www.amazon.co.jp/"+strings.Repeat("a", 2048)+`"}`)
	if rec.Code != http.StatusBadRequest || !reflect.DeepEqual(decode[common.ErrorResponse](t, rec).Errors, []common.FieldError{{Field: "amazonUrl", Rule: "max"}}) {
		t.Fatalf("status = %d body = %s, want 400 amazonUrl/max", rec.Code, rec.Body.String())
	}
}

// serveUpload posts content as the multipart field `field`, declaring declaredType for it.
func serveUpload(t *testing.T, h *httpbook.Handler, path, field, declaredType string, content []byte) *httptest.ResponseRecorder {
	t.Helper()
	var body bytes.Buffer
	w := multipart.NewWriter(&body)
	header := textproto.MIMEHeader{}
	header.Set("Content-Disposition", `form-data; name="`+field+`"; filename="cover.png"`)
	header.Set("Content-Type", declaredType)
	part, err := w.CreatePart(header)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := part.Write(content); err != nil {
		t.Fatal(err)
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}

	orig := slog.Default()
	slog.SetDefault(slog.New(slog.NewTextHandler(io.Discard, nil)))
	t.Cleanup(func() { slog.SetDefault(orig) })
	e := common.NewEcho()
	h.Register(e.Group("/api/books"))
	req := httptest.NewRequest(http.MethodPost, path, &body)
	req.Header.Set("Content-Type", w.FormDataContentType())
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	return rec
}

var pngBytes = []byte("\x89PNG\r\n\x1a\n\x00\x00\x00\rIHDR")

func TestHandlerUploadImage(t *testing.T) {
	t.Run("中身から判定した種別・サイズ・本文をusecaseに渡し200で返す", func(t *testing.T) {
		var got bookcmd.UploadImageCommand
		var gotBody []byte
		h := &httpbook.Handler{UploadImageUC: fakeUpload(func(_ context.Context, cmd bookcmd.UploadImageCommand) (*domainbook.BookDetail, error) {
			got = cmd
			gotBody, _ = io.ReadAll(cmd.Body)
			return detail, nil
		})}
		rec := serveUpload(t, h, "/api/books/1/image", "image", "image/png", pngBytes)

		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200 (body=%s)", rec.Code, rec.Body.String())
		}
		if got.BookID != 1 || got.ContentType != "image/png" || got.Size != int64(len(pngBytes)) || !bytes.Equal(gotBody, pngBytes) {
			t.Fatalf("usecase received %+v body=%q; the body must be rewound after sniffing", got, gotBody)
		}
	})

	t.Run("申告された種別ではなく中身の種別を渡す", func(t *testing.T) {
		var gotType string
		h := &httpbook.Handler{UploadImageUC: fakeUpload(func(_ context.Context, cmd bookcmd.UploadImageCommand) (*domainbook.BookDetail, error) {
			gotType = cmd.ContentType
			return nil, domaincommon.ErrInvalid
		})}
		rec := serveUpload(t, h, "/api/books/1/image", "image", "image/png", []byte("<script>alert(1)</script>"))

		if rec.Code != http.StatusBadRequest || strings.HasPrefix(gotType, "image/") {
			t.Fatalf("status = %d, usecase got type %q; a spoofed header must not reach the usecase as an image type", rec.Code, gotType)
		}
	})

	t.Run("imageフィールドがなければ400でusecaseを呼ばない", func(t *testing.T) {
		h := &httpbook.Handler{UploadImageUC: fakeUpload(func(context.Context, bookcmd.UploadImageCommand) (*domainbook.BookDetail, error) {
			t.Error("usecase must not be called")
			return nil, nil
		})}
		rec := serveUpload(t, h, "/api/books/1/image", "file", "image/png", pngBytes)

		if rec.Code != http.StatusBadRequest || !reflect.DeepEqual(decode[common.ErrorResponse](t, rec).Errors, []common.FieldError{{Field: "image", Rule: "required"}}) {
			t.Fatalf("status = %d body = %s, want 400 image/required", rec.Code, rec.Body.String())
		}
	})

	t.Run("本文が上限を超えると413でusecaseを呼ばない", func(t *testing.T) {
		h := &httpbook.Handler{UploadImageUC: fakeUpload(func(context.Context, bookcmd.UploadImageCommand) (*domainbook.BookDetail, error) {
			t.Error("usecase must not be called")
			return nil, nil
		})}
		rec := serveUpload(t, h, "/api/books/1/image", "image", "image/png", bytes.Repeat([]byte{0}, 6<<20+1))

		if rec.Code != http.StatusRequestEntityTooLarge {
			t.Fatalf("status = %d, want 413", rec.Code)
		}
	})

	statuses := []struct {
		name string
		path string
		err  error
		want int
	}{
		{name: "不正なIDは400", path: "/api/books/abc/image", want: http.StatusBadRequest},
		{name: "存在しない書籍は404", path: "/api/books/1/image", err: domaincommon.ErrNotFound, want: http.StatusNotFound},
		{name: "更新の競合は409", path: "/api/books/1/image", err: domaincommon.ErrConflict, want: http.StatusConflict},
	}
	for _, tt := range statuses {
		t.Run(tt.name, func(t *testing.T) {
			h := &httpbook.Handler{UploadImageUC: fakeUpload(func(context.Context, bookcmd.UploadImageCommand) (*domainbook.BookDetail, error) {
				if tt.err == nil {
					t.Error("usecase must not be called")
				}
				return nil, tt.err
			})}
			if rec := serveUpload(t, h, tt.path, "image", "image/png", pngBytes); rec.Code != tt.want {
				t.Fatalf("status = %d, want %d", rec.Code, tt.want)
			}
		})
	}
}
