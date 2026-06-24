package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/TheLonelyGhost/llm-caveman-convert/internal/config"
	"github.com/TheLonelyGhost/llm-caveman-convert/internal/proxy"
)

func main() {
	cfgPath := flag.String("config", "", "path to proxy config file (JSON)")
	flag.Parse()

	cfg, err := loadConfig(*cfgPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "proxy: config: %v\n", err)
		os.Exit(1)
	}

	if err := proxy.ValidateBinary(cfg.CavemanBin); err != nil {
		fmt.Fprintf(os.Stderr, "proxy: %v\n", err)
		os.Exit(1)
	}

	h := proxy.New(cfg.BackendBaseURL, cfg.CavemanBin)
	fmt.Fprintf(os.Stderr, "proxy: listening on %s\n", cfg.ListenAddr)
	srv := &http.Server{
		Addr:         cfg.ListenAddr,
		Handler:      h,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 120 * time.Second,
		IdleTimeout:  120 * time.Second,
	}
	if err := srv.ListenAndServe(); err != nil {
		fmt.Fprintf(os.Stderr, "proxy: %v\n", err)
		os.Exit(1)
	}
}

func loadConfig(path string) (*config.ProxyConfig, error) {
	if path == "" {
		path = os.Getenv("PROXY_CONFIG")
	}
	if path == "" {
		return &config.ProxyConfig{
			BackendBaseURL: os.Getenv("PROXY_BACKEND_URL"),
			CavemanBin:     envOrDefault("PROXY_CAVEMAN_BIN", "caveman"),
			ListenAddr:     envOrDefault("PROXY_LISTEN_ADDR", ":8080"),
		}, nil
	}
	return config.LoadProxyConfig(path)
}

func envOrDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
