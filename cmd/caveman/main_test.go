package main

import (
	"encoding/json"
	"os"
	"testing"
)

func TestLoadConfigFromEnv(t *testing.T) {
	t.Setenv("CAVEMAN_CONFIG", "")
	t.Setenv("CAVEMAN_MODEL", "gpt-4o")
	t.Setenv("CAVEMAN_BASE_URL", "https://api.example.com")
	t.Setenv("CAVEMAN_API_KEY", "sk-test")

	cfg, err := loadConfig("")
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Model != "gpt-4o" {
		t.Errorf("Model: got %q", cfg.Model)
	}
	if cfg.BaseURL != "https://api.example.com" {
		t.Errorf("BaseURL: got %q", cfg.BaseURL)
	}
	if cfg.APIKey != "sk-test" {
		t.Errorf("APIKey: got %q", cfg.APIKey)
	}
	if cfg.CachePath == "" {
		t.Error("CachePath should default")
	}
}

func TestLoadConfigFromFile(t *testing.T) {
	data := map[string]string{
		"model":      "gpt-4o-mini",
		"base_url":   "https://file.example.com",
		"api_key":    "sk-file",
		"cache_path": "/tmp/test-cache.db",
	}
	b, _ := json.Marshal(data)
	f, err := os.CreateTemp(t.TempDir(), "caveman-cfg-*.json")
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
	if cfg.Model != "gpt-4o-mini" {
		t.Errorf("Model: got %q", cfg.Model)
	}
}

func TestLoadConfigFromEnvConfigPath(t *testing.T) {
	data := map[string]string{
		"model": "env-path-model",
	}
	b, _ := json.Marshal(data)
	f, err := os.CreateTemp(t.TempDir(), "caveman-cfg-*.json")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := f.Write(b); err != nil {
		t.Fatal(err)
	}
	f.Close()

	t.Setenv("CAVEMAN_CONFIG", f.Name())
	cfg, err := loadConfig("")
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Model != "env-path-model" {
		t.Errorf("Model: got %q", cfg.Model)
	}
}
