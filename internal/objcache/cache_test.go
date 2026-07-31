package objcache

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestLookupPutRoundTrip(t *testing.T) {
	dir := t.TempDir()
	s, err := New(dir, 1024, 10<<20)
	if err != nil {
		t.Fatal(err)
	}
	data := []byte("hello-image-bytes")
	path, err := s.Put("prof1", "bucket", "a/b.png", "etag-1", "2026-01-02T03:04:05Z", "image/png", int64(len(data)), bytes.NewReader(data))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatal(err)
	}
	// hit
	p, meta, ok := s.Lookup("prof1", "bucket", "a/b.png", "etag-1", "2026-01-02T03:04:05Z")
	if !ok {
		t.Fatal("expected cache hit")
	}
	if p != path {
		t.Fatalf("path mismatch %s vs %s", p, path)
	}
	if meta.Size != int64(len(data)) {
		t.Fatalf("size %d", meta.Size)
	}
	// etag mismatch miss
	if _, _, ok := s.Lookup("prof1", "bucket", "a/b.png", "etag-2", "2026-01-02T03:04:05Z"); ok {
		t.Fatal("expected miss on etag change")
	}
	// different profile miss
	if _, _, ok := s.Lookup("prof2", "bucket", "a/b.png", "etag-1", ""); ok {
		t.Fatal("expected miss on profile")
	}
	files, n, d := s.Stats()
	if files != 1 || n != int64(len(data)) || d != dir {
		t.Fatalf("stats %d %d %s", files, n, d)
	}
	if err := s.Clear(); err != nil {
		t.Fatal(err)
	}
	entries, _ := os.ReadDir(dir)
	for _, e := range entries {
		if filepath.Ext(e.Name()) == ".bin" {
			t.Fatal("bin left after clear")
		}
	}
}
