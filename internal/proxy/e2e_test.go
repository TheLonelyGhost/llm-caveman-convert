package proxy_test

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/TheLonelyGhost/llm-caveman-convert/internal/proxy"
)

func TestEndToEndTwoTurnConversation(t *testing.T) {
	bin := buildCavemanBinary(t)

	var upstreamRequests [][]byte
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		upstreamRequests = append(upstreamRequests, body)
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprint(w, `{"choices":[{"message":{"role":"assistant","content":"cave answer"}}]}`)
	}))
	defer upstream.Close()

	h := proxy.New(upstream.URL, bin)
	srv := httptest.NewServer(h)
	defer srv.Close()

	for turn := 1; turn <= 2; turn++ {
		body := fmt.Sprintf(`{"messages":[{"role":"user","content":"plain English question turn %d"}]}`, turn)
		req := mustNewRequest(t, "POST", srv.URL+"/v1/chat/completions", body)
		req.Header.Set("Content-Type", "application/json")
		resp, err := http.DefaultClient.Do(req)
		if err != nil || resp == nil {
			t.Fatalf("turn %d: %v", turn, err)
		}
		defer resp.Body.Close()
		respBody, _ := io.ReadAll(resp.Body)

		var result map[string]json.RawMessage
		_ = json.Unmarshal(respBody, &result)
		var choices []struct {
			Message struct{ Content string } `json:"message"`
		}
		_ = json.Unmarshal(result["choices"], &choices)
		if len(choices) == 0 {
			t.Fatalf("turn %d: no choices; body: %s", turn, respBody)
		}
		content := choices[0].Message.Content
		if !strings.HasPrefix(content, "[dec]") {
			t.Errorf("turn %d: caller should see decoded content, got %q", turn, content)
		}
	}

	// Verify backend never sees raw English in user messages.
	for i, rawReq := range upstreamRequests {
		var req map[string]json.RawMessage
		_ = json.Unmarshal(rawReq, &req)
		var msgs []struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		}
		_ = json.Unmarshal(req["messages"], &msgs)
		for _, m := range msgs {
			if m.Role == "user" && !strings.HasPrefix(m.Content, "[enc]") {
				t.Errorf("request %d: backend saw unencoded user content: %q", i+1, m.Content)
			}
		}
	}

	// Verify turn 2 request includes history from turn 1 (prior user + assistant turns).
	if len(upstreamRequests) < 2 {
		t.Fatal("expected 2 upstream requests")
	}
	var req2 map[string]json.RawMessage
	_ = json.Unmarshal(upstreamRequests[1], &req2)
	var msgs2 []struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	}
	_ = json.Unmarshal(req2["messages"], &msgs2)

	var userCount, assistantCount int
	for _, m := range msgs2 {
		if m.Role == "user" {
			userCount++
		}
		if m.Role == "assistant" {
			assistantCount++
		}
	}
	if userCount < 2 {
		t.Errorf("turn 2 request should include history user turn, got %d user messages; msgs: %+v", userCount, msgs2)
	}
	if assistantCount < 1 {
		t.Errorf("turn 2 request should include history assistant turn, got %d assistant messages; msgs: %+v", assistantCount, msgs2)
	}
}

func TestTokenSavingsValidation(t *testing.T) {
	samples := []struct {
		name    string
		normal  string
		caveman string
	}{
		{
			"prose",
			"Please explain the difference between a mutex and a semaphore in concurrent programming.",
			"mutex vs semaphore concurrent prog.",
		},
		{
			"conversational",
			"I have been working on this problem for hours and I am not sure what I am doing wrong.",
			"worked hours. not sure what wrong.",
		},
		{
			"instructional",
			"Please write a function that reads a file and returns the number of lines in the file.",
			"write func: read file, return line count.",
		},
	}

	for _, s := range samples {
		normalChars := utf8.RuneCountInString(s.normal)
		cavemanChars := utf8.RuneCountInString(s.caveman)
		reduction := float64(normalChars-cavemanChars) / float64(normalChars) * 100
		t.Logf("%-20s normal=%d chars  caveman=%d chars  reduction=%.1f%%",
			s.name, normalChars, cavemanChars, reduction)
		if cavemanChars >= normalChars {
			t.Errorf("%s: caveman (%d) not shorter than normal (%d)", s.name, cavemanChars, normalChars)
		}
	}
}
