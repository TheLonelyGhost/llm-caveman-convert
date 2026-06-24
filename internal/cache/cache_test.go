package cache_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/TheLonelyGhost/llm-caveman-convert/internal/cache"
)

func TestEncodeCacheRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "sub", "cache.db")
	c, err := cache.New(path)
	if err != nil {
		t.Fatal(err)
	}

	if _, ok := c.GetEncode("hello world"); ok {
		t.Fatal("expected cache miss")
	}

	if err := c.SetEncode("hello world", "hello world cave"); err != nil {
		t.Fatal(err)
	}

	v, ok := c.GetEncode("hello world")
	if !ok {
		t.Fatal("expected cache hit after set")
	}
	if v != "hello world cave" {
		t.Fatalf("got %q", v)
	}
}

func TestDecodeCacheRoundTrip(t *testing.T) {
	c, err := cache.New(filepath.Join(t.TempDir(), "cache.db"))
	if err != nil {
		t.Fatal(err)
	}

	if err := c.SetDecode("cave speak", "fluent English"); err != nil {
		t.Fatal(err)
	}
	v, ok := c.GetDecode("cave speak")
	if !ok || v != "fluent English" {
		t.Fatalf("got %q, ok=%v", v, ok)
	}
}

func TestEncodeDecodeNamespacesIndependent(t *testing.T) {
	c, err := cache.New(filepath.Join(t.TempDir(), "cache.db"))
	if err != nil {
		t.Fatal(err)
	}

	_ = c.SetEncode("text", "encoded")
	if _, ok := c.GetDecode("text"); ok {
		t.Fatal("encode entry should not appear in decode namespace")
	}

	_ = c.SetDecode("text", "decoded")
	v, _ := c.GetEncode("text")
	if v != "encoded" {
		t.Fatalf("decode set should not overwrite encode namespace, got %q", v)
	}
}

func TestCachePersistsAcrossInstances(t *testing.T) {
	path := filepath.Join(t.TempDir(), "cache.db")

	c1, err := cache.New(path)
	if err != nil {
		t.Fatal(err)
	}
	_ = c1.SetEncode("input", "output")

	c2, err := cache.New(path)
	if err != nil {
		t.Fatal(err)
	}
	v, ok := c2.GetEncode("input")
	if !ok || v != "output" {
		t.Fatalf("expected persisted cache hit, got %q ok=%v", v, ok)
	}
}

func TestCacheAutoCreatesDirs(t *testing.T) {
	path := filepath.Join(t.TempDir(), "a", "b", "c", "cache.db")
	c, err := cache.New(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := c.SetEncode("x", "y"); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("cache file not created: %v", err)
	}
}

func TestCacheKeyIsStable(t *testing.T) {
	dir := t.TempDir()
	for range 3 {
		c, err := cache.New(filepath.Join(dir, "cache.db"))
		if err != nil {
			t.Fatal(err)
		}
		_ = c.SetEncode("stable input", "stable output")
	}
	c, err := cache.New(filepath.Join(dir, "cache.db"))
	if err != nil {
		t.Fatal(err)
	}
	v, ok := c.GetEncode("stable input")
	if !ok || v != "stable output" {
		t.Fatalf("unstable key: got %q ok=%v", v, ok)
	}
}

func TestNewCacheFailsOnBadLoadData(t *testing.T) {
	path := filepath.Join(t.TempDir(), "cache.db")
	if err := os.WriteFile(path, []byte("not valid json"), 0o600); err != nil {
		t.Fatal(err)
	}
	_, err := cache.New(path)
	if err == nil {
		t.Fatal("expected error loading corrupt cache file")
	}
}

func TestSetEncodeFailsOnReadOnlyDir(t *testing.T) {
	if os.Getuid() == 0 {
		t.Skip("root bypasses permissions")
	}
	dir := t.TempDir()
	path := filepath.Join(dir, "cache.db")
	c, err := cache.New(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(dir, 0o555); err != nil {
		t.Fatal(err)
	}
	defer os.Chmod(dir, 0o755)
	if err := c.SetEncode("k", "v"); err == nil {
		t.Fatal("expected error writing to read-only dir")
	}
}
