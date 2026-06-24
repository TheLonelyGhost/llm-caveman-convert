package main

import (
	"context"
	"strings"
	"testing"

	"github.com/TheLonelyGhost/llm-caveman-convert/internal/tokens"
)

func TestReadStdin(t *testing.T) {
	text, err := readStdin()
	if err != nil {
		t.Skip("stdin not available in test environment")
	}
	_ = text
}

func TestCountAll(t *testing.T) {
	results := countAll(context.Background(), "hello world")
	if len(results) != len(tokens.Registry) {
		t.Errorf("got %d results, want %d", len(results), len(tokens.Registry))
	}
	for _, r := range results {
		if r.Provider == "" {
			t.Errorf("result missing Provider: %+v", r)
		}
		if r.Model == "" {
			t.Errorf("result missing Model: %+v", r)
		}
	}
}

func TestCountAllEmpty(t *testing.T) {
	results := countAll(context.Background(), "")
	for _, r := range results {
		if r.SkipReason != "" {
			continue
		}
		if r.Tokens != 0 {
			t.Errorf("model %s/%s: expected 0 tokens for empty input, got %d", r.Provider, r.Model, r.Tokens)
		}
	}
}

func TestCountAllNonEmpty(t *testing.T) {
	results := countAll(context.Background(), strings.Repeat("hello world ", 100))
	for _, r := range results {
		if r.SkipReason != "" {
			continue
		}
		if r.Tokens == 0 {
			t.Errorf("model %s/%s: expected non-zero tokens, got 0", r.Provider, r.Model)
		}
	}
}
