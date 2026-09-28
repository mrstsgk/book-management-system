package main

import (
	"testing"
)

func TestSampleBooks_AreValidAndDistinct(t *testing.T) {
	if len(sampleBooks) != 5 {
		t.Fatalf("len(sampleBooks) = %d, want 5", len(sampleBooks))
	}
	seen := map[string]bool{}
	for _, s := range sampleBooks {
		b, err := s.toBook()
		if err != nil {
			t.Fatalf("%s: %v", s.isbn, err)
		}
		if seen[b.ISBN.String()] {
			t.Fatalf("duplicate ISBN %s", b.ISBN.String())
		}
		seen[b.ISBN.String()] = true
		if b.Cover != nil {
			t.Fatalf("%s: sample data must not carry a cover (楽天由来の情報を持たない)", s.isbn)
		}
	}
}
