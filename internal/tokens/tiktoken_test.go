package tokens

import (
	"context"
	"testing"
)

func TestTiktokenCounterO200k(t *testing.T) {
	c, err := newTiktokenCounter(encO200k)
	if err != nil {
		t.Fatalf("newTiktokenCounter o200k: %v", err)
	}

	tests := []struct {
		text string
		want int
	}{
		{"Hello, world!", 4},
		{"", 0},
		{"The quick brown fox jumps over the lazy dog", 9},
	}

	for _, tt := range tests {
		got, err := c.Count(context.Background(), tt.text)
		if err != nil {
			t.Fatalf("Count(%q): %v", tt.text, err)
		}
		if got != tt.want {
			t.Errorf("o200k Count(%q) = %d, want %d", tt.text, got, tt.want)
		}
	}
}

func TestTiktokenCounterCL100k(t *testing.T) {
	c, err := newTiktokenCounter(encCL100k)
	if err != nil {
		t.Fatalf("newTiktokenCounter cl100k: %v", err)
	}

	tests := []struct {
		text string
		want int
	}{
		{"Hello, world!", 4},
		{"", 0},
		{"The quick brown fox jumps over the lazy dog", 9},
	}

	for _, tt := range tests {
		got, err := c.Count(context.Background(), tt.text)
		if err != nil {
			t.Fatalf("Count(%q): %v", tt.text, err)
		}
		if got != tt.want {
			t.Errorf("cl100k Count(%q) = %d, want %d", tt.text, got, tt.want)
		}
	}
}

func TestTiktokenCounterInvalidEncoding(t *testing.T) {
	_, err := newTiktokenCounter("not_a_real_encoding")
	if err == nil {
		t.Error("expected error for invalid encoding, got nil")
	}
}
