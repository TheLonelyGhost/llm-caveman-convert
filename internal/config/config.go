package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// BinaryConfig holds settings for the caveman CLI binary.
type BinaryConfig struct {
	Model     string `json:"model"`
	BaseURL   string `json:"base_url"`
	APIKey    string `json:"api_key"`
	CachePath string `json:"cache_path"`
}

// ProxyConfig holds settings for the caveman proxy server.
type ProxyConfig struct {
	BackendBaseURL string `json:"backend_base_url"`
	CavemanBin     string `json:"caveman_bin"`
	ListenAddr     string `json:"listen_addr"`
}

// DefaultCachePath returns the platform-appropriate default cache file path.
func DefaultCachePath() string {
	dir, err := os.UserCacheDir()
	if err != nil {
		dir = os.TempDir()
	}
	return filepath.Join(dir, "caveman", "cache.db")
}

// FromEnv builds a BinaryConfig from environment variables.
func FromEnv() *BinaryConfig {
	cfg := &BinaryConfig{
		Model:   os.Getenv("CAVEMAN_MODEL"),
		BaseURL: os.Getenv("CAVEMAN_BASE_URL"),
		APIKey:  os.Getenv("CAVEMAN_API_KEY"),
	}
	cfg.CachePath = DefaultCachePath()
	return cfg
}

// LoadBinaryConfig reads a BinaryConfig from a JSON file.
func LoadBinaryConfig(path string) (*BinaryConfig, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }()
	var cfg BinaryConfig
	if err := json.NewDecoder(f).Decode(&cfg); err != nil {
		return nil, err
	}
	if cfg.CachePath == "" {
		cfg.CachePath = DefaultCachePath()
	}
	return &cfg, nil
}

// LoadProxyConfig reads a ProxyConfig from a JSON file.
func LoadProxyConfig(path string) (*ProxyConfig, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }()
	var cfg ProxyConfig
	if err := json.NewDecoder(f).Decode(&cfg); err != nil {
		return nil, err
	}
	if cfg.ListenAddr == "" {
		cfg.ListenAddr = ":8080"
	}
	if cfg.CavemanBin == "" {
		cfg.CavemanBin = "caveman"
	}
	return &cfg, nil
}
