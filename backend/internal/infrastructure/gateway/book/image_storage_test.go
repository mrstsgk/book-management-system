package book_test

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	domainbook "github.com/mrstsgk/book-management-system/backend/internal/domain/book"
	gwbook "github.com/mrstsgk/book-management-system/backend/internal/infrastructure/gateway/book"
)

// connectTestStorage targets the LocalStack S3 in backend/docker-compose.yml. Like
// the Postgres contract tests, it skips rather than fails when that isn't running.
func connectTestStorage(t *testing.T) domainbook.ImageStorage {
	t.Helper()
	const endpoint = "http://localhost:4566"
	client := &http.Client{Timeout: 2 * time.Second}
	res, err := client.Get(endpoint + "/_localstack/health")
	if err != nil {
		t.Skipf("skipping: LocalStack not reachable (run `make db-up` first): %v", err)
	}
	_ = res.Body.Close()

	s, err := gwbook.NewImageStorage(context.Background(), gwbook.Config{
		Endpoint: endpoint, Region: "ap-northeast-1", Bucket: "book-images",
		AccessKeyID: "test", SecretAccessKey: "test",
	})
	if err != nil {
		t.Fatalf("NewImageStorage: %v", err)
	}
	return s
}

// onlyReader hides Seek so Put takes its buffering path.
type onlyReader struct{ io.Reader }

func TestImageStorage_PutURLDelete(t *testing.T) {
	storage := connectTestStorage(t)
	ctx := context.Background()
	body := []byte("\x89PNG\r\n\x1a\ncontract-test")
	image, err := domainbook.NewImage("image/png", int64(len(body)))
	if err != nil {
		t.Fatal(err)
	}

	bodies := []struct {
		name string
		body func() io.Reader
	}{
		{name: "シーク可能なボディ", body: func() io.Reader { return bytes.NewReader(body) }},
		{name: "シークできないボディ", body: func() io.Reader { return onlyReader{bytes.NewReader(body)} }},
	}
	for _, tt := range bodies {
		t.Run(tt.name, func(t *testing.T) {
			key, err := domainbook.NewImageKey("contract-test/" + strings.ReplaceAll(t.Name(), "/", "_") + ".png")
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { _ = storage.Delete(ctx, key) })

			if err := storage.Put(ctx, key, image, tt.body()); err != nil {
				t.Fatalf("Put: %v", err)
			}
			url, err := storage.URL(ctx, key)
			if err != nil {
				t.Fatalf("URL: %v", err)
			}
			if !strings.Contains(url, "X-Amz-Expires=900") {
				t.Errorf("URL %q is not a 15-minute presigned URL", url)
			}

			res, err := http.Get(url)
			if err != nil {
				t.Fatalf("GET presigned URL: %v", err)
			}
			got, _ := io.ReadAll(res.Body)
			_ = res.Body.Close()
			if res.StatusCode != http.StatusOK || !bytes.Equal(got, body) || res.Header.Get("Content-Type") != "image/png" {
				t.Fatalf("GET = %d %q (%s), want 200 with the stored bytes as image/png", res.StatusCode, got, res.Header.Get("Content-Type"))
			}

			if err := storage.Delete(ctx, key); err != nil {
				t.Fatalf("Delete: %v", err)
			}
			res, err = http.Get(url)
			if err != nil {
				t.Fatalf("GET after delete: %v", err)
			}
			_ = res.Body.Close()
			if res.StatusCode != http.StatusNotFound {
				t.Fatalf("GET after delete = %d, want 404", res.StatusCode)
			}
		})
	}
}
