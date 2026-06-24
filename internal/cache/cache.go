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
	mu      sync.RWMutex
	path    string
	entries map[string]string
}

func newStore(path string) (*store, error) {
	s := &store{path: path, entries: make(map[string]string)}
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
	defer s.mu.RUnlock()
	v, ok := s.entries[s.key(ns, input)]
	return v, ok
}

func (s *store) Set(ns namespace, input, output string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.entries[s.key(ns, input)] = output
	return s.save()
}

func (s *store) load() error {
	f, err := os.Open(s.path)
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()
	return json.NewDecoder(f).Decode(&s.entries)
}

func (s *store) save() error {
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
