package tokens

import (
	"context"
	"testing"
)

func TestGeminiCounterSkipWhenNoKey(t *testing.T) {
	t.Setenv(googleAPIKeyEnv, "")
	c := newGeminiCounter("gemini-2.5-pro")
	_, err := c.Count(context.Background(), "hello")
	if err == nil {
		t.Fatal("expected error when GOOGLE_API_KEY is absent, got nil")
	}
	if err.Error() != googleSkipReason {
		t.Errorf("error = %q, want %q", err.Error(), googleSkipReason)
	}
}

func TestNewGeminiCounterAlwaysReturnsGeminiCounter(t *testing.T) {
	t.Setenv(googleAPIKeyEnv, "")
	c := newGeminiCounter("gemini-2.5-pro")
	if _, ok := c.(*geminiCounter); !ok {
		t.Errorf("expected *geminiCounter regardless of key presence, got %T", c)
	}
}
