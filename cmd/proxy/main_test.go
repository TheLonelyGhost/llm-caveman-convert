package main

import (
	"encoding/json"
	"os"
	"testing"
)

func TestEnvOrDefaultUsesEnv(t *testing.T) {
	t.Setenv("TEST_KEY_XYZ", "from-env")
	if got := envOrDefault("TEST_KEY_XYZ", "default"); got != "from-env" {
		t.Errorf("got %q, want %q", got, "from-env")
	}
}

func TestEnvOrDefaultFallsBack(t *testing.T) {
	t.Setenv("TEST_KEY_XYZ", "")
	if got := envOrDefault("TEST_KEY_XYZ", "fallback"); got != "fallback" {
		t.Errorf("got %q, want %q", got, "fallback")
	}
}

func TestLoadConfigFromEnv(t *testing.T) {
	t.Setenv("PROXY_CONFIG", "")
	t.Setenv("PROXY_BACKEND_URL", "https://backend.example.com")
	t.Setenv("PROXY_CAVEMAN_BIN", "/usr/bin/caveman")
	t.Setenv("PROXY_LISTEN_ADDR", ":9090")

	cfg, err := loadConfig("")
	if err != nil {
		t.Fatal(err)
	}
	if cfg.BackendBaseURL != "https://backend.example.com" {
		t.Errorf("BackendBaseURL: got %q", cfg.BackendBaseURL)
	}
	if cfg.CavemanBin != "/usr/bin/caveman" {
		t.Errorf("CavemanBin: got %q", cfg.CavemanBin)
	}
	if cfg.ListenAddr != ":9090" {
		t.Errorf("ListenAddr: got %q", cfg.ListenAddr)
	}
}

func TestLoadConfigFromFile(t *testing.T) {
	data := map[string]string{
		"backend_base_url": "https://file-backend.example.com",
		"caveman_bin":      "/bin/caveman",
		"listen_addr":      ":7070",
	}
	b, _ := json.Marshal(data)
	f, err := os.CreateTemp(t.TempDir(), "proxy-cfg-*.json")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := f.Write(b); err != nil {
		t.Fatal(err)
	}
	f.Close()

	cfg, err := loadConfig(f.Name())
	if err != nil {
		t.Fatal(err)
	}
	if cfg.BackendBaseURL != "https://file-backend.example.com" {
		t.Errorf("BackendBaseURL: got %q", cfg.BackendBaseURL)
	}
}

func TestLoadConfigFromEnvConfigPath(t *testing.T) {
	data := map[string]string{
		"backend_base_url": "https://env-path.example.com",
	}
	b, _ := json.Marshal(data)
	f, err := os.CreateTemp(t.TempDir(), "proxy-cfg-*.json")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := f.Write(b); err != nil {
		t.Fatal(err)
	}
	f.Close()

	t.Setenv("PROXY_CONFIG", f.Name())
	cfg, err := loadConfig("")
	if err != nil {
		t.Fatal(err)
	}
	if cfg.BackendBaseURL != "https://env-path.example.com" {
		t.Errorf("BackendBaseURL: got %q", cfg.BackendBaseURL)
	}
}
