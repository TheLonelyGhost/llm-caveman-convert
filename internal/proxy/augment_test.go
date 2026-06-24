package proxy_test

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/TheLonelyGhost/llm-caveman-convert/internal/proxy"
)

func TestSystemPromptAugmented(t *testing.T) {
	bin := buildCavemanBinary(t)
	var gotMsgs []struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	}
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var raw map[string]json.RawMessage
		body, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(body, &raw)
		_ = json.Unmarshal(raw["messages"], &gotMsgs)
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprint(w, `{"choices":[{"message":{"role":"assistant","content":"ok"}}]}`)
	}))
	defer upstream.Close()

	h := proxy.New(upstream.URL, bin)
	srv := httptest.NewServer(h)
	defer srv.Close()

	body := `{"messages":[{"role":"system","content":"You are helpful."},{"role":"user","content":"hello"}]}`
	req := mustNewRequest(t, "POST", srv.URL+"/v1/chat/completions", body)
	req.Header.Set("Content-Type", "application/json")
	resp, _ := http.DefaultClient.Do(req)
	if resp != nil {
		_, _ = io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
	}

	var systemContent string
	for _, m := range gotMsgs {
		if m.Role == "system" {
			systemContent = m.Content
		}
	}
	if !strings.Contains(systemContent, "caveman-speak") {
		t.Errorf("system prompt not augmented: %q", systemContent)
	}
	if !strings.Contains(systemContent, "You are helpful") {
		t.Errorf("original system prompt content lost: %q", systemContent)
	}
}

func TestNoSystemMessageInjectsInstruction(t *testing.T) {
	bin := buildCavemanBinary(t)
	var gotMsgs []struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	}
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var raw map[string]json.RawMessage
		body, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(body, &raw)
		_ = json.Unmarshal(raw["messages"], &gotMsgs)
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprint(w, `{"choices":[{"message":{"role":"assistant","content":"ok"}}]}`)
	}))
	defer upstream.Close()

	h := proxy.New(upstream.URL, bin)
	srv := httptest.NewServer(h)
	defer srv.Close()

	body := `{"messages":[{"role":"user","content":"hello"}]}`
	req := mustNewRequest(t, "POST", srv.URL+"/v1/chat/completions", body)
	req.Header.Set("Content-Type", "application/json")
	resp, _ := http.DefaultClient.Do(req)
	if resp != nil {
		_, _ = io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
	}

	if len(gotMsgs) < 2 {
		t.Fatalf("expected injected system message, got %d messages", len(gotMsgs))
	}
	if gotMsgs[0].Role != "system" {
		t.Errorf("first message should be system, got %q", gotMsgs[0].Role)
	}
	if !strings.Contains(gotMsgs[0].Content, "caveman-speak") {
		t.Errorf("injected system message lacks instruction: %q", gotMsgs[0].Content)
	}
}
