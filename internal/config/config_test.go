package config_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/TheLonelyGhost/llm-caveman-convert/internal/config"
)

func writeTempJSON(t *testing.T, v any) string {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	f, err := os.CreateTemp(t.TempDir(), "cfg-*.json")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := f.Write(b); err != nil {
		t.Fatal(err)
	}
	f.Close()
	return f.Name()
}

func TestFromEnv(t *testing.T) {
	t.Setenv("CAVEMAN_MODEL", "gpt-4o")
	t.Setenv("CAVEMAN_BASE_URL", "https://api.example.com")
	t.Setenv("CAVEMAN_API_KEY", "sk-test")

	cfg := config.FromEnv()
	if cfg.Model != "gpt-4o" {
		t.Errorf("model: got %q", cfg.Model)
	}
	if cfg.BaseURL != "https://api.example.com" {
		t.Errorf("base_url: got %q", cfg.BaseURL)
	}
	if cfg.APIKey != "sk-test" {
		t.Errorf("api_key: got %q", cfg.APIKey)
	}
	if cfg.CachePath == "" {
		t.Error("CachePath should default when not set")
	}
}

func TestFromEnvOptionalAPIKey(t *testing.T) {
	t.Setenv("CAVEMAN_MODEL", "gpt-4o")
	t.Setenv("CAVEMAN_BASE_URL", "https://api.example.com")
	t.Setenv("CAVEMAN_API_KEY", "")

	cfg := config.FromEnv()
	if cfg.APIKey != "" {
		t.Errorf("api_key: expected empty, got %q", cfg.APIKey)
	}
}

func TestLoadBinaryConfigBadJSON(t *testing.T) {
	f, err := os.CreateTemp(t.TempDir(), "cfg-*.json")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := f.Write([]byte("not json")); err != nil {
		t.Fatal(err)
	}
	f.Close()
	_, err = config.LoadBinaryConfig(f.Name())
	if err == nil {
		t.Fatal("expected error for bad JSON")
	}
}

func TestLoadProxyConfigBadJSON(t *testing.T) {
	f, err := os.CreateTemp(t.TempDir(), "cfg-*.json")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := f.Write([]byte("not json")); err != nil {
		t.Fatal(err)
	}
	f.Close()
	_, err = config.LoadProxyConfig(f.Name())
	if err == nil {
		t.Fatal("expected error for bad JSON")
	}
}

func TestLoadBinaryConfigRoundTrip(t *testing.T) {
	path := writeTempJSON(t, map[string]string{
		"model":      "gpt-4o",
		"base_url":   "https://api.example.com",
		"api_key":    "sk-test",
		"cache_path": "/tmp/cache.db",
	})

	cfg, err := config.LoadBinaryConfig(path)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Model != "gpt-4o" {
		t.Errorf("model: got %q", cfg.Model)
	}
	if cfg.BaseURL != "https://api.example.com" {
		t.Errorf("base_url: got %q", cfg.BaseURL)
	}
	if cfg.APIKey != "sk-test" {
		t.Errorf("api_key: got %q", cfg.APIKey)
	}
	if cfg.CachePath != "/tmp/cache.db" {
		t.Errorf("cache_path: got %q", cfg.CachePath)
	}
}

func TestLoadBinaryConfigDefaultCachePath(t *testing.T) {
	path := writeTempJSON(t, map[string]string{
		"model": "gpt-4o",
	})

	cfg, err := config.LoadBinaryConfig(path)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.CachePath == "" {
		t.Error("CachePath should default when not set")
	}
	defaultPath := config.DefaultCachePath()
	if cfg.CachePath != defaultPath {
		t.Errorf("CachePath: got %q, want %q", cfg.CachePath, defaultPath)
	}
}

func TestLoadBinaryConfigMissingFile(t *testing.T) {
	_, err := config.LoadBinaryConfig(filepath.Join(t.TempDir(), "nonexistent.json"))
	if err == nil {
		t.Error("expected error for missing file")
	}
}

func TestLoadProxyConfigRoundTrip(t *testing.T) {
	path := writeTempJSON(t, map[string]string{
		"backend_base_url": "https://backend.example.com",
		"caveman_bin":      "/usr/local/bin/caveman",
		"listen_addr":      ":9090",
	})

	cfg, err := config.LoadProxyConfig(path)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.BackendBaseURL != "https://backend.example.com" {
		t.Errorf("backend_base_url: got %q", cfg.BackendBaseURL)
	}
	if cfg.CavemanBin != "/usr/local/bin/caveman" {
		t.Errorf("caveman_bin: got %q", cfg.CavemanBin)
	}
	if cfg.ListenAddr != ":9090" {
		t.Errorf("listen_addr: got %q", cfg.ListenAddr)
	}
}

func TestLoadProxyConfigDefaults(t *testing.T) {
	path := writeTempJSON(t, map[string]string{
		"backend_base_url": "https://backend.example.com",
	})

	cfg, err := config.LoadProxyConfig(path)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.ListenAddr != ":8080" {
		t.Errorf("listen_addr default: got %q", cfg.ListenAddr)
	}
	if cfg.CavemanBin != "caveman" {
		t.Errorf("caveman_bin default: got %q", cfg.CavemanBin)
	}
}

func TestLoadProxyConfigMissingFile(t *testing.T) {
	_, err := config.LoadProxyConfig(filepath.Join(t.TempDir(), "nonexistent.json"))
	if err == nil {
		t.Error("expected error for missing file")
	}
}
