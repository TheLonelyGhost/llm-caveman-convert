package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/TheLonelyGhost/llm-caveman-convert/internal/cache"
	"github.com/TheLonelyGhost/llm-caveman-convert/internal/caveman"
	"github.com/TheLonelyGhost/llm-caveman-convert/internal/config"
	"github.com/TheLonelyGhost/llm-caveman-convert/internal/llm"
)

func main() {
	encode := flag.Bool("encode", false, "compress stdin to caveman-speak")
	decode := flag.Bool("decode", false, "expand caveman-speak to plain English")
	cfgPath := flag.String("config", "", "path to config file (JSON)")
	flag.Parse()

	if *encode == *decode {
		fmt.Fprintln(os.Stderr, "caveman: exactly one of --encode or --decode required")
		os.Exit(1)
	}

	cfg, err := loadConfig(*cfgPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "caveman: config: %v\n", err)
		os.Exit(1)
	}

	input, err := io.ReadAll(os.Stdin)
	if err != nil {
		fmt.Fprintf(os.Stderr, "caveman: read stdin: %v\n", err)
		os.Exit(1)
	}
	text := string(input)

	c, err := cache.New(cfg.CachePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "caveman: cache: %v\n", err)
		os.Exit(1)
	}

	client := llm.New(cfg.BaseURL, cfg.APIKey, cfg.Model)
	ctx := context.Background()

	var result string
	if *encode {
		result, err = caveman.Encode(ctx, c, client, text)
	} else {
		result, err = caveman.Decode(ctx, c, client, text)
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "caveman: %v\n", err)
		os.Exit(1)
	}

	fmt.Print(result)
}

func loadConfig(path string) (*config.BinaryConfig, error) {
	if path == "" {
		path = os.Getenv("CAVEMAN_CONFIG")
	}
	if path == "" {
		return config.FromEnv(), nil
	}
	return config.LoadBinaryConfig(path)
}
