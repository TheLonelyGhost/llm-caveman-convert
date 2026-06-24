package proxy_test

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/TheLonelyGhost/llm-caveman-convert/internal/proxy"
)

func mustNewRequest(t *testing.T, method, url, body string) *http.Request {
	t.Helper()
	req, err := http.NewRequest(method, url, strings.NewReader(body))
	if err != nil {
		t.Fatalf("NewRequest: %v", err)
	}
	return req
}

// buildCavemanBinary compiles a fake caveman binary that echoes stdin prefixed by mode.
func buildCavemanBinary(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	src := filepath.Join(dir, "main.go")
	bin := filepath.Join(dir, "caveman")
	if runtime.GOOS == "windows" {
		bin += ".exe"
	}
	code := `package main
import("fmt";"io";"os")
func main(){
	b,_:=io.ReadAll(os.Stdin)
	for _,a:=range os.Args[1:]{
		if a=="--encode"{fmt.Print("[enc]"+string(b));return}
		if a=="--decode"{fmt.Print("[dec]"+string(b));return}
	}
	os.Exit(1)
}`
	if err := os.WriteFile(src, []byte(code), 0o644); err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command("go", "build", "-o", bin, src)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("build fake caveman: %v\n%s", err, out)
	}
	return bin
}

func fakeUpstream(t *testing.T, response string) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		t.Logf("upstream received: %s", body)
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprint(w, response)
	}))
}

func TestRouteNotFound(t *testing.T) {
	bin := buildCavemanBinary(t)
	h := proxy.New("http://localhost", bin)
	srv := httptest.NewServer(h)
	defer srv.Close()

	req := mustNewRequest(t, "GET", srv.URL+"/unknown", "")
	resp, err := http.DefaultClient.Do(req)
	if err != nil || resp == nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", resp.StatusCode)
	}
}

func TestRouteChatCompletionsAccepted(t *testing.T) {
	bin := buildCavemanBinary(t)
	upstreamResp := `{"choices":[{"message":{"role":"assistant","content":"[dec]cave resp"}}]}`
	upstream := fakeUpstream(t, upstreamResp)
	defer upstream.Close()

	h := proxy.New(upstream.URL, bin)
	srv := httptest.NewServer(h)
	defer srv.Close()

	body := `{"messages":[{"role":"user","content":"hello"}]}`
	req := mustNewRequest(t, "POST", srv.URL+"/v1/chat/completions", body)
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil || resp == nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
}

func TestHeadersForwarded(t *testing.T) {
	bin := buildCavemanBinary(t)
	var gotAuth, gotCustom, gotHost string
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		gotCustom = r.Header.Get("X-Custom")
		gotHost = r.Host
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"choices":[{"message":{"role":"assistant","content":"ok"}}]}`)
	}))
	defer upstream.Close()

	h := proxy.New(upstream.URL, bin)
	srv := httptest.NewServer(h)
	defer srv.Close()

	req := mustNewRequest(t, "POST", srv.URL+"/v1/chat/completions", `{"messages":[{"role":"user","content":"hi"}]}`)
	req.Header.Set("Authorization", "Bearer secret")
	req.Header.Set("X-Custom", "custom-value")
	req.Header.Set("Content-Type", "application/json")
	hresp, _ := http.DefaultClient.Do(req)
	if hresp != nil {
		_, _ = io.Copy(io.Discard, hresp.Body)
		hresp.Body.Close()
	}

	if gotAuth != "Bearer secret" {
		t.Errorf("Authorization not forwarded: %q", gotAuth)
	}
	if gotCustom != "custom-value" {
		t.Errorf("X-Custom not forwarded: %q", gotCustom)
	}
	// Host should be upstream host, not proxy host
	upstreamHost := strings.TrimPrefix(upstream.URL, "http://")
	if gotHost != upstreamHost {
		t.Errorf("Host not rewritten: got %q, want %q", gotHost, upstreamHost)
	}
}

func TestContentExtractedAndEncoded(t *testing.T) {
	bin := buildCavemanBinary(t)
	var gotBody []byte
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotBody, _ = io.ReadAll(r.Body)
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"choices":[{"message":{"role":"assistant","content":"ok"}}]}`)
	}))
	defer upstream.Close()

	h := proxy.New(upstream.URL, bin)
	srv := httptest.NewServer(h)
	defer srv.Close()

	body := `{"messages":[{"role":"user","content":"please explain this"}]}`
	req := mustNewRequest(t, "POST", srv.URL+"/v1/chat/completions", body)
	req.Header.Set("Content-Type", "application/json")
	r1, err := http.DefaultClient.Do(req)
	if err != nil || r1 == nil {
		t.Fatal(err)
	}
	_, _ = io.Copy(io.Discard, r1.Body)
	r1.Body.Close()

	var req2 map[string]json.RawMessage
	_ = json.Unmarshal(gotBody, &req2)
	var msgs []struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	}
	_ = json.Unmarshal(req2["messages"], &msgs)

	found := false
	for _, m := range msgs {
		if m.Role == "user" && strings.HasPrefix(m.Content, "[enc]") {
			found = true
		}
	}
	if !found {
		t.Errorf("user content not encoded; messages: %s", gotBody)
	}
}

func TestNonContentFieldsUnchanged(t *testing.T) {
	bin := buildCavemanBinary(t)
	var gotBody []byte
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotBody, _ = io.ReadAll(r.Body)
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprint(w, `{"choices":[{"message":{"role":"assistant","content":"ok"}}]}`)
	}))
	defer upstream.Close()

	h := proxy.New(upstream.URL, bin)
	srv := httptest.NewServer(h)
	defer srv.Close()

	body := `{"model":"gpt-4","messages":[{"role":"user","content":"hi","name":"alice"}]}`
	req := mustNewRequest(t, "POST", srv.URL+"/v1/chat/completions", body)
	req.Header.Set("Content-Type", "application/json")
	r2, err := http.DefaultClient.Do(req)
	if err != nil || r2 == nil {
		t.Fatal(err)
	}
	_, _ = io.Copy(io.Discard, r2.Body)
	r2.Body.Close()

	var raw map[string]json.RawMessage
	_ = json.Unmarshal(gotBody, &raw)
	var model string
	_ = json.Unmarshal(raw["model"], &model)
	if model != "gpt-4" {
		t.Errorf("model field changed: %q", model)
	}
}

func TestEncodeBinaryFailureFallback(t *testing.T) {
	dir := t.TempDir()
	bin := filepath.Join(dir, "caveman-fail")
	failCode := `package main
import "os"
func main(){ os.Exit(1) }`
	src := filepath.Join(dir, "main.go")
	if err := os.WriteFile(src, []byte(failCode), 0o644); err != nil {
		t.Fatal(err)
	}
	if out, err := exec.Command("go", "build", "-o", bin, src).CombinedOutput(); err != nil {
		t.Fatalf("build fail binary: %v\n%s", err, out)
	}

	var gotBody []byte
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotBody, _ = io.ReadAll(r.Body)
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprint(w, `{"choices":[{"message":{"role":"assistant","content":"ok"}}]}`)
	}))
	defer upstream.Close()

	h := proxy.New(upstream.URL, bin)
	srv := httptest.NewServer(h)
	defer srv.Close()

	originalContent := "please explain this to me"
	body := fmt.Sprintf(`{"messages":[{"role":"user","content":%q}]}`, originalContent)
	req := mustNewRequest(t, "POST", srv.URL+"/v1/chat/completions", body)
	req.Header.Set("Content-Type", "application/json")
	r3, err := http.DefaultClient.Do(req)
	if err != nil || r3 == nil {
		t.Fatal(err)
	}
	_, _ = io.Copy(io.Discard, r3.Body)
	r3.Body.Close()

	var raw map[string]json.RawMessage
	_ = json.Unmarshal(gotBody, &raw)
	var msgs []struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	}
	_ = json.Unmarshal(raw["messages"], &msgs)
	for _, m := range msgs {
		if m.Role == "user" && m.Content != originalContent {
			t.Errorf("expected fallback to original %q, got %q", originalContent, m.Content)
		}
	}
}

func TestResponseContentDecoded(t *testing.T) {
	bin := buildCavemanBinary(t)
	upstream := fakeUpstream(t, `{"choices":[{"message":{"role":"assistant","content":"cave resp"}}]}`)
	defer upstream.Close()

	h := proxy.New(upstream.URL, bin)
	srv := httptest.NewServer(h)
	defer srv.Close()

	req := mustNewRequest(t, "POST", srv.URL+"/v1/chat/completions", `{"messages":[{"role":"user","content":"hi"}]}`)
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil || resp == nil {
		t.Fatal(err)
	}

	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	var result map[string]json.RawMessage
	if err := json.Unmarshal(body, &result); err != nil {
		t.Fatalf("response not valid JSON: %v\nbody: %s", err, body)
	}
	var choices []struct {
		Message struct{ Content string } `json:"message"`
	}
	_ = json.Unmarshal(result["choices"], &choices)
	if len(choices) == 0 {
		t.Fatalf("no choices in response; body: %s", body)
	}
	if !strings.HasPrefix(choices[0].Message.Content, "[dec]") {
		t.Errorf("expected decoded content, got %q", choices[0].Message.Content)
	}
}

func TestDecodeBinaryFailureReturnsCavemanContent(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "main.go")
	bin := filepath.Join(dir, "caveman")
	code := `package main
import("fmt";"io";"os")
func main(){
	b,_:=io.ReadAll(os.Stdin)
	for _,a:=range os.Args[1:]{
		if a=="--encode"{fmt.Print("[enc]"+string(b));return}
		if a=="--decode"{os.Exit(1)}
	}
	os.Exit(1)
}`
	if err := os.WriteFile(src, []byte(code), 0o644); err != nil {
		t.Fatal(err)
	}
	if out, err := exec.Command("go", "build", "-o", bin, src).CombinedOutput(); err != nil {
		t.Fatalf("build decode-fail binary: %v\n%s", err, out)
	}

	upstream := fakeUpstream(t, `{"choices":[{"message":{"role":"assistant","content":"cave speak"}}]}`)
	defer upstream.Close()

	h := proxy.New(upstream.URL, bin)
	srv := httptest.NewServer(h)
	defer srv.Close()

	req := mustNewRequest(t, "POST", srv.URL+"/v1/chat/completions", `{"messages":[{"role":"user","content":"hi"}]}`)
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil || resp == nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	var result map[string]json.RawMessage
	_ = json.Unmarshal(body, &result)
	var choices []struct {
		Message struct{ Content string } `json:"message"`
	}
	_ = json.Unmarshal(result["choices"], &choices)
	if len(choices) == 0 {
		t.Fatalf("no choices; body: %s", body)
	}
	if choices[0].Message.Content != "cave speak" {
		t.Errorf("expected caveman fallback, got %q", choices[0].Message.Content)
	}
}

func TestInvalidJSONBodyReturns400(t *testing.T) {
	bin := buildCavemanBinary(t)
	h := proxy.New("http://localhost", bin)
	srv := httptest.NewServer(h)
	defer srv.Close()

	req := mustNewRequest(t, "POST", srv.URL+"/v1/chat/completions", "not json{{{")
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil || resp == nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}

func TestInvalidMessagesFieldReturns400(t *testing.T) {
	bin := buildCavemanBinary(t)
	h := proxy.New("http://localhost", bin)
	srv := httptest.NewServer(h)
	defer srv.Close()

	req := mustNewRequest(t, "POST", srv.URL+"/v1/chat/completions", `{"messages":"not-an-array"}`)
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil || resp == nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}

func TestUpstreamErrorReturns502(t *testing.T) {
	bin := buildCavemanBinary(t)
	h := proxy.New("http://127.0.0.1:1", bin)
	srv := httptest.NewServer(h)
	defer srv.Close()

	req := mustNewRequest(t, "POST", srv.URL+"/v1/chat/completions", `{"messages":[{"role":"user","content":"hi"}]}`)
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil || resp == nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusBadGateway {
		t.Fatalf("expected 502, got %d", resp.StatusCode)
	}
}

func TestNonJSONUpstreamResponsePassedThrough(t *testing.T) {
	bin := buildCavemanBinary(t)
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprint(w, "plain text response")
	}))
	defer upstream.Close()

	h := proxy.New(upstream.URL, bin)
	srv := httptest.NewServer(h)
	defer srv.Close()

	req := mustNewRequest(t, "POST", srv.URL+"/v1/chat/completions", `{"messages":[{"role":"user","content":"hi"}]}`)
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil || resp == nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if string(body) != "plain text response" {
		t.Errorf("expected passthrough body, got %q", body)
	}
}

func TestInvalidBackendURLReturns500(t *testing.T) {
	bin := buildCavemanBinary(t)
	h := proxy.New("http://\x00invalid", bin)
	srv := httptest.NewServer(h)
	defer srv.Close()

	req := mustNewRequest(t, "POST", srv.URL+"/v1/chat/completions", `{"messages":[{"role":"user","content":"hi"}]}`)
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil || resp == nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusInternalServerError && resp.StatusCode != http.StatusBadGateway {
		t.Fatalf("expected 500 or 502, got %d", resp.StatusCode)
	}
}
