package cache

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
)

type namespace string

// Namespace constants for encode and decode cache buckets.
const (
	EncodeNS namespace = "encode"
	DecodeNS namespace = "decode"
)

type store struct {
	mu       sync.RWMutex
	path     string
	lockPath string
	entries  map[string]string
}

func newStore(path string) (*store, error) {
	s := &store{
		path:     path,
		lockPath: path + ".lock",
		entries:  make(map[string]string),
	}
	if err := s.load(); err != nil && !os.IsNotExist(err) {
		return nil, err
	}
	return s, nil
}

func (s *store) key(ns namespace, input string) string {
	h := sha256.Sum256([]byte(input))
	return string(ns) + ":" + hex.EncodeToString(h[:])
}

func (s *store) Get(ns namespace, input string) (string, bool) {
	s.mu.RLock()
	k := s.key(ns, input)
	v, ok := s.entries[k]
	s.mu.RUnlock()

	if ok {
		// Fast path: found in memory
		return v, true
	}

	// Slow path: not in memory, reload from disk to check for writes from other processes
	s.mu.Lock()
	defer s.mu.Unlock()

	// Double-check after acquiring write lock (another goroutine may have loaded it)
	if v, ok := s.entries[k]; ok {
		return v, true
	}

	// Reload from disk
	if err := s.loadUnlocked(); err != nil && !os.IsNotExist(err) {
		// On error, return miss
		return "", false
	}

	// Check again after reload
	v, ok = s.entries[k]
	return v, ok
}

func (s *store) Set(ns namespace, input, output string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.entries[s.key(ns, input)] = output
	return s.saveUnlocked()
}



func (s *store) load() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.loadUnlocked()
}

func (s *store) loadUnlocked() error {
	if err := os.MkdirAll(filepath.Dir(s.lockPath), 0o750); err != nil {
		return err
	}
	// Acquire shared lock for reading
	lockFile, err := s.acquireLock(lockShared)
	if err != nil {
		return err
	}
	defer func() {
		_ = s.releaseLock(lockFile)
	}()

	f, err := os.Open(s.path)
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()
	return json.NewDecoder(f).Decode(&s.entries)
}

func (s *store) saveUnlocked() error {
	// Acquire exclusive lock for writing
	lockFile, err := s.acquireLock(lockExclusive)
	if err != nil {
		return err
	}
	defer func() {
		_ = s.releaseLock(lockFile)
	}()

	// Reload from disk to merge any concurrent writes from other processes
	diskEntries := make(map[string]string)
	if f, err := os.Open(s.path); err == nil {
		_ = json.NewDecoder(f).Decode(&diskEntries)
		_ = f.Close()
		// Merge disk entries with our in-memory entries (our entries take precedence)
		for k, v := range diskEntries {
			if _, exists := s.entries[k]; !exists {
				s.entries[k] = v
			}
		}
	}

	if err := os.MkdirAll(filepath.Dir(s.path), 0o750); err != nil {
		return err
	}
	tmp := s.path + ".tmp"
	f, err := os.Create(tmp)
	if err != nil {
		return err
	}
	if err := json.NewEncoder(f).Encode(s.entries); err != nil {
		_ = f.Close()
		_ = os.Remove(tmp)
		return err
	}
	if err := f.Close(); err != nil {
		_ = os.Remove(tmp)
		return err
	}
	return os.Rename(tmp, s.path)
}

// Cache is a persistent disk-backed key-value store for encode/decode results.
type Cache struct {
	s *store
}

// New creates or loads a Cache at the given file path.
func New(path string) (*Cache, error) {
	s, err := newStore(path)
	if err != nil {
		return nil, err
	}
	return &Cache{s: s}, nil
}

// GetEncode returns a cached encode result for input, if present.
func (c *Cache) GetEncode(input string) (string, bool) { return c.s.Get(EncodeNS, input) }

// GetDecode returns a cached decode result for input, if present.
func (c *Cache) GetDecode(input string) (string, bool) { return c.s.Get(DecodeNS, input) }

// SetEncode stores an encode result for input.
func (c *Cache) SetEncode(input, output string) error { return c.s.Set(EncodeNS, input, output) }

// SetDecode stores a decode result for input.
func (c *Cache) SetDecode(input, output string) error { return c.s.Set(DecodeNS, input, output) }
