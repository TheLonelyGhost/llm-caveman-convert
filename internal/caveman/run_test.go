package caveman_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"sync/atomic"
	"testing"

	openai "github.com/sashabaranov/go-openai"

	"github.com/TheLonelyGhost/llm-caveman-convert/internal/cache"
	"github.com/TheLonelyGhost/llm-caveman-convert/internal/caveman"
	"github.com/TheLonelyGhost/llm-caveman-convert/internal/llm"
)

func fakeSuccess(t *testing.T, response string) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		resp := openai.ChatCompletionResponse{
			Choices: []openai.ChatCompletionChoice{
				{Message: openai.ChatCompletionMessage{Content: response}},
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
}

func fakeFailure(t *testing.T) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "internal error", http.StatusInternalServerError)
	}))
}

func newCache(t *testing.T) *cache.Cache {
	t.Helper()
	c, err := cache.New(filepath.Join(t.TempDir(), "cache.db"))
	if err != nil {
		t.Fatal(err)
	}
	return c
}

func TestEncodeFailedLLMNotCached(t *testing.T) {
	srv := fakeFailure(t)
	defer srv.Close()

	c := newCache(t)
	client := llm.New(srv.URL, "key", "model")

	_, err := caveman.Encode(context.Background(), c, client, "hello world")
	if err == nil {
		t.Fatal("expected error from failing LLM")
	}

	if _, ok := c.GetEncode("hello world"); ok {
		t.Error("failed encode should not write to cache")
	}
}

func TestDecodeFailedLLMNotCached(t *testing.T) {
	srv := fakeFailure(t)
	defer srv.Close()

	c := newCache(t)
	client := llm.New(srv.URL, "key", "model")

	_, err := caveman.Decode(context.Background(), c, client, "cave speak")
	if err == nil {
		t.Fatal("expected error from failing LLM")
	}

	if _, ok := c.GetDecode("cave speak"); ok {
		t.Error("failed decode should not write to cache")
	}
}

func TestEncodeCacheHitSkipsLLM(t *testing.T) {
	called := false
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.Header().Set("Content-Type", "application/json")
		resp := openai.ChatCompletionResponse{
			Choices: []openai.ChatCompletionChoice{
				{Message: openai.ChatCompletionMessage{Content: "llm result"}},
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	c := newCache(t)
	_ = c.SetEncode("hello", "cached result")
	client := llm.New(srv.URL, "key", "model")

	out, err := caveman.Encode(context.Background(), c, client, "hello")
	if err != nil {
		t.Fatal(err)
	}
	if out != "cached result" {
		t.Errorf("expected cached result, got %q", out)
	}
	if called {
		t.Error("LLM should not be called on cache hit")
	}
}

func TestDecodeCacheHitSkipsLLM(t *testing.T) {
	called := false
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.Header().Set("Content-Type", "application/json")
		resp := openai.ChatCompletionResponse{
			Choices: []openai.ChatCompletionChoice{
				{Message: openai.ChatCompletionMessage{Content: "llm result"}},
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	c := newCache(t)
	_ = c.SetDecode("cave", "cached expansion")
	client := llm.New(srv.URL, "key", "model")

	out, err := caveman.Decode(context.Background(), c, client, "cave")
	if err != nil {
		t.Fatal(err)
	}
	if out != "cached expansion" {
		t.Errorf("expected cached expansion, got %q", out)
	}
	if called {
		t.Error("LLM should not be called on cache hit")
	}
}

func TestEncodeCacheMissPopulatesCache(t *testing.T) {
	srv := fakeSuccess(t, "compressed")
	defer srv.Close()

	c := newCache(t)
	client := llm.New(srv.URL, "key", "model")

	out, err := caveman.Encode(context.Background(), c, client, "input text")
	if err != nil {
		t.Fatal(err)
	}
	if out != "compressed" {
		t.Errorf("got %q", out)
	}
	v, ok := c.GetEncode("input text")
	if !ok || v != "compressed" {
		t.Errorf("cache not populated: ok=%v val=%q", ok, v)
	}
}

func TestDecodeCacheMissPopulatesCache(t *testing.T) {
	srv := fakeSuccess(t, "expanded")
	defer srv.Close()

	c := newCache(t)
	client := llm.New(srv.URL, "key", "model")

	out, err := caveman.Decode(context.Background(), c, client, "cave text")
	if err != nil {
		t.Fatal(err)
	}
	if out != "expanded" {
		t.Errorf("got %q", out)
	}
	v, ok := c.GetDecode("cave text")
	if !ok || v != "expanded" {
		t.Errorf("cache not populated: ok=%v val=%q", ok, v)
	}
}

// fakeSequence returns a server that replies with successive responses on each
// request. Once the slice is exhausted it returns a 500.
func fakeSequence(t *testing.T, responses []string) *httptest.Server {
	t.Helper()
	var call atomic.Int32
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		idx := int(call.Add(1)) - 1
		if idx >= len(responses) {
			http.Error(w, "no more responses", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		resp := openai.ChatCompletionResponse{
			Choices: []openai.ChatCompletionChoice{
				{Message: openai.ChatCompletionMessage{Content: responses[idx]}},
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
}

// inputWithURL is plain prose + one URL — Validate will fail if the URL is
// dropped from the output.
const inputWithURL = "See https://example.com for details."

func TestEncodeValidationFailureTriggersFix(t *testing.T) {
	// First call (Encode): drops the URL → validation fails.
	// Second call (Fix): restores the URL → validation passes → cached.
	badOut := "See details."
	goodOut := "See https://example.com details."
	srv := fakeSequence(t, []string{badOut, goodOut})
	defer srv.Close()

	c := newCache(t)
	client := llm.New(srv.URL, "key", "model")

	out, err := caveman.Encode(context.Background(), c, client, inputWithURL)
	if err != nil {
		t.Fatal(err)
	}
	if out != goodOut {
		t.Errorf("got %q, want %q", out, goodOut)
	}
	v, ok := c.GetEncode(inputWithURL)
	if !ok || v != goodOut {
		t.Errorf("cache should hold fixed result: ok=%v val=%q", ok, v)
	}
}

func TestEncodeCacheNotWrittenOnExhaustedRetries(t *testing.T) {
	// Both LLM calls (Encode + Fix) drop the URL → validation never passes → not cached.
	// Only 2 responses are consumed: call 1 = Encode, call 2 = Fix (attempt 0).
	// The post-loop Validate is purely in-memory; no third LLM call is made.
	badOut := "See details."
	srv := fakeSequence(t, []string{badOut, badOut})
	defer srv.Close()

	c := newCache(t)
	client := llm.New(srv.URL, "key", "model")

	out, err := caveman.Encode(context.Background(), c, client, inputWithURL)
	if err != nil {
		t.Fatal(err)
	}
	if out != badOut {
		t.Errorf("expected best-effort result %q, got %q", badOut, out)
	}
	if _, ok := c.GetEncode(inputWithURL); ok {
		t.Error("cache must not be written when validation never passes")
	}
}

func TestEncodeCacheWrittenOnlyAfterValidationPass(t *testing.T) {
	// First call: drops URL.  Second call: restores URL.
	// Cache must contain the second (valid) result, not the first.
	badOut := "See details."
	goodOut := "See https://example.com details."
	srv := fakeSequence(t, []string{badOut, goodOut})
	defer srv.Close()

	c := newCache(t)
	client := llm.New(srv.URL, "key", "model")

	_, err := caveman.Encode(context.Background(), c, client, inputWithURL)
	if err != nil {
		t.Fatal(err)
	}
	v, ok := c.GetEncode(inputWithURL)
	if !ok {
		t.Fatal("expected cache entry")
	}
	if v == badOut {
		t.Errorf("cache must not hold the invalid intermediate result")
	}
	if v != goodOut {
		t.Errorf("cache should hold validated result %q, got %q", goodOut, v)
	}
}
