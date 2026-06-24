package tokens

import (
	"context"
	"errors"
	"testing"
)

type errCounter struct{}

func (e *errCounter) Count(_ context.Context, _ string) (int, error) {
	return 0, errors.New("forced error")
}

func TestEntryCountNilCounter(t *testing.T) {
	e := Entry{Provider: "Test", Model: "test-model", counter: nil}
	r := e.Count(context.Background(), "hello")
	if r.SkipReason == "" {
		t.Error("expected SkipReason for nil counter")
	}
	if r.Tokens != 0 {
		t.Errorf("expected 0 tokens, got %d", r.Tokens)
	}
}

func TestEntryCountError(t *testing.T) {
	e := Entry{Provider: "Test", Model: "test-model", counter: &errCounter{}}
	r := e.Count(context.Background(), "hello")
	if r.SkipReason == "" {
		t.Error("expected SkipReason for erroring counter")
	}
}

func TestEntryCountSuccess(t *testing.T) {
	c, err := newTiktokenCounter(encO200k)
	if err != nil {
		t.Fatalf("newTiktokenCounter: %v", err)
	}
	e := Entry{Provider: "Test", Model: "test-model", Approx: true, MarginPct: 10, counter: c}
	r := e.Count(context.Background(), "hello world")
	if r.Tokens == 0 {
		t.Error("expected non-zero tokens")
	}
	if !r.Approx {
		t.Error("expected Approx=true")
	}
	if r.MarginPct != 10 {
		t.Errorf("MarginPct = %d, want 10", r.MarginPct)
	}
	if r.SkipReason != "" {
		t.Errorf("unexpected SkipReason: %q", r.SkipReason)
	}
}
