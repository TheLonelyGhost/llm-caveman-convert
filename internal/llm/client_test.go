package llm_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	openai "github.com/sashabaranov/go-openai"

	"github.com/TheLonelyGhost/llm-caveman-convert/internal/llm"
)

func fakeLLMServer(t *testing.T, response string) *httptest.Server {
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

func TestEncodeCallsLLM(t *testing.T) {
	srv := fakeLLMServer(t, "cave speak result")
	defer srv.Close()

	c := llm.New(srv.URL, "test-key", "test-model")
	out, err := c.Encode(context.Background(), "please explain this to me")
	if err != nil {
		t.Fatal(err)
	}
	if out != "cave speak result" {
		t.Fatalf("got %q", out)
	}
}

func TestDecodeCallsLLM(t *testing.T) {
	srv := fakeLLMServer(t, "expanded fluent English")
	defer srv.Close()

	c := llm.New(srv.URL, "test-key", "test-model")
	out, err := c.Decode(context.Background(), "explain this")
	if err != nil {
		t.Fatal(err)
	}
	if out != "expanded fluent English" {
		t.Fatalf("got %q", out)
	}
}

func TestLLMFailureReturnsError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "internal error", http.StatusInternalServerError)
	}))
	defer srv.Close()

	c := llm.New(srv.URL, "test-key", "test-model")
	_, err := c.Encode(context.Background(), "input")
	if err == nil {
		t.Fatal("expected error from failing LLM")
	}
}

func TestLLMUsesConfiguredModel(t *testing.T) {
	var gotModel string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req openai.ChatCompletionRequest
		_ = json.NewDecoder(r.Body).Decode(&req)
		gotModel = req.Model
		w.Header().Set("Content-Type", "application/json")
		resp := openai.ChatCompletionResponse{
			Choices: []openai.ChatCompletionChoice{
				{Message: openai.ChatCompletionMessage{Content: "ok"}},
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	c := llm.New(srv.URL, "test-key", "my-special-model")
	_, _ = c.Encode(context.Background(), "text")
	if gotModel != "my-special-model" {
		t.Fatalf("expected model my-special-model, got %q", gotModel)
	}
}

func TestFixCallsLLM(t *testing.T) {
	var gotBody string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req openai.ChatCompletionRequest
		_ = json.NewDecoder(r.Body).Decode(&req)
		for _, m := range req.Messages {
			if m.Role == openai.ChatMessageRoleUser {
				gotBody = m.Content
			}
		}
		w.Header().Set("Content-Type", "application/json")
		resp := openai.ChatCompletionResponse{
			Choices: []openai.ChatCompletionChoice{
				{Message: openai.ChatCompletionMessage{Content: "fixed output"}},
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	c := llm.New(srv.URL, "test-key", "test-model")
	out, err := c.Fix(context.Background(), "original text", "compressed text", []string{"heading count mismatch: original 2, compressed 1"})
	if err != nil {
		t.Fatal(err)
	}
	if out != "fixed output" {
		t.Fatalf("got %q", out)
	}
	if !strings.Contains(gotBody, "heading count mismatch") {
		t.Errorf("expected errors in user message, got: %q", gotBody)
	}
	if !strings.Contains(gotBody, "original text") {
		t.Errorf("expected original in user message, got: %q", gotBody)
	}
	if !strings.Contains(gotBody, "compressed text") {
		t.Errorf("expected compressed in user message, got: %q", gotBody)
	}
}
