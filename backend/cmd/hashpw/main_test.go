// 非公開の run を検証するため、このテストは in-package（package main）に置く。
package main

import (
	"bytes"
	"strings"
	"testing"

	"golang.org/x/crypto/bcrypt"
)

func TestRun(t *testing.T) {
	t.Parallel()
	var out bytes.Buffer
	if err := run(strings.NewReader("secret\n"), &out); err != nil {
		t.Fatal(err)
	}
	hash := strings.TrimSpace(out.String())
	// 末尾の改行はパスワードに含めない
	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte("secret")); err != nil {
		t.Fatalf("hash does not match the input: %v", err)
	}
}

func TestRun_EmptyIsError(t *testing.T) {
	t.Parallel()
	if err := run(strings.NewReader("\n"), &bytes.Buffer{}); err == nil {
		t.Fatal("expected error for empty password")
	}
}

func TestRun_NoTrailingNewline(t *testing.T) {
	t.Parallel()
	var out bytes.Buffer
	if err := run(strings.NewReader("secret"), &out); err != nil {
		t.Fatal(err)
	}
	if err := bcrypt.CompareHashAndPassword([]byte(strings.TrimSpace(out.String())), []byte("secret")); err != nil {
		t.Fatalf("hash does not match the input: %v", err)
	}
}
