package cache_test

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
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

func TestConcurrentAccessSameProcess(t *testing.T) {
	path := filepath.Join(t.TempDir(), "cache.db")
	c, err := cache.New(path)
	if err != nil {
		t.Fatal(err)
	}

	const numGoroutines = 10
	const numOps = 100

	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	for i := range numGoroutines {
		go func(id int) {
			defer wg.Done()
			for j := range numOps {
				key := fmt.Sprintf("key-%d-%d", id, j)
				val := fmt.Sprintf("val-%d-%d", id, j)
				if err := c.SetEncode(key, val); err != nil {
					t.Errorf("goroutine %d: SetEncode failed: %v", id, err)
					return
				}
				if v, ok := c.GetEncode(key); !ok || v != val {
					t.Errorf("goroutine %d: expected %q, got %q ok=%v", id, val, v, ok)
					return
				}
			}
		}(i)
	}

	wg.Wait()
}

func TestConcurrentAccessMultipleProcesses(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping multi-process test in short mode")
	}

	path := filepath.Join(t.TempDir(), "cache.db")
	c, err := cache.New(path)
	if err != nil {
		t.Fatal(err)
	}

	// Pre-populate with some data
	for i := range 10 {
		key := fmt.Sprintf("init-key-%d", i)
		val := fmt.Sprintf("init-val-%d", i)
		if err := c.SetEncode(key, val); err != nil {
			t.Fatal(err)
		}
	}

	// Spawn multiple test processes that will concurrently access the cache
	const numProcesses = 5
	var wg sync.WaitGroup
	wg.Add(numProcesses)

	for i := range numProcesses {
		go func(id int) {
			defer wg.Done()
			cmd := exec.Command(os.Args[0], "-test.run=TestCacheWorker")
			cmd.Env = append(os.Environ(),
				fmt.Sprintf("CACHE_PATH=%s", path),
				fmt.Sprintf("WORKER_ID=%d", id),
			)
			if output, err := cmd.CombinedOutput(); err != nil {
				t.Errorf("worker %d failed: %v\nOutput: %s", id, err, output)
			}
		}(i)
	}

	wg.Wait()

	// Reload cache from disk to see worker writes
	c, err = cache.New(path)
	if err != nil {
		t.Fatal(err)
	}

	// Verify all workers wrote their data
	for i := range numProcesses {
		for j := range 50 {
			key := fmt.Sprintf("worker-%d-key-%d", i, j)
			val := fmt.Sprintf("worker-%d-val-%d", i, j)
			if v, ok := c.GetEncode(key); !ok || v != val {
				t.Errorf("missing or incorrect data for worker %d op %d: got %q ok=%v", i, j, v, ok)
			}
		}
	}
}

// TestCacheWorker is a helper test that runs as a separate process
func TestCacheWorker(t *testing.T) {
	path := os.Getenv("CACHE_PATH")
	if path == "" {
		t.Skip("not running as cache worker")
	}

	workerID := os.Getenv("WORKER_ID")
	c, err := cache.New(path)
	if err != nil {
		t.Fatal(err)
	}

	// Each worker writes 50 entries
	for i := range 50 {
		key := fmt.Sprintf("worker-%s-key-%d", workerID, i)
		val := fmt.Sprintf("worker-%s-val-%d", workerID, i)
		if err := c.SetEncode(key, val); err != nil {
			t.Fatalf("worker %s: SetEncode failed: %v", workerID, err)
		}

		// Verify we can read it back immediately
		if v, ok := c.GetEncode(key); !ok || v != val {
			t.Fatalf("worker %s: expected %q, got %q ok=%v", workerID, val, v, ok)
		}
	}

	// Verify we can still read the initial data
	for i := range 10 {
		key := fmt.Sprintf("init-key-%d", i)
		val := fmt.Sprintf("init-val-%d", i)
		if v, ok := c.GetEncode(key); !ok || v != val {
			t.Fatalf("worker %s: init data missing: expected %q, got %q ok=%v", workerID, val, v, ok)
		}
	}
}
