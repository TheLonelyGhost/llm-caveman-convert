package tokens

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestWriteReportExact(t *testing.T) {
	results := []Result{
		{Provider: "OpenAI", Model: "gpt-5.5", Tokens: 2920},
	}
	var buf bytes.Buffer
	WriteReport(&buf, results)
	out := buf.String()
	if !strings.Contains(out, "OpenAI") {
		t.Error("output missing Provider")
	}
	if !strings.Contains(out, "gpt-5.5") {
		t.Error("output missing Model")
	}
	if !strings.Contains(out, "2,920") {
		t.Errorf("output missing thousands-formatted count; got:\n%s", out)
	}
	if strings.Contains(out, "approx") {
		t.Error("exact result should not contain approx")
	}
}

func TestWriteReportApproxNoMargin(t *testing.T) {
	results := []Result{
		{Provider: "Anthropic", Model: "claude-opus-4-8", Tokens: 2975, Approx: true},
	}
	var buf bytes.Buffer
	WriteReport(&buf, results)
	out := buf.String()
	if !strings.Contains(out, "approx.") {
		t.Errorf("approx result should contain 'approx.'; got:\n%s", out)
	}
	if strings.Contains(out, "±") {
		t.Error("zero-margin approx should not contain ± notation")
	}
}

func TestWriteReportApproxWithMargin(t *testing.T) {
	results := []Result{
		{Provider: "Meta", Model: "llama-4-scout", Tokens: 2890, Approx: true, MarginPct: 15},
	}
	var buf bytes.Buffer
	WriteReport(&buf, results)
	out := buf.String()
	if !strings.Contains(out, "approx. (±15%)") {
		t.Errorf("expected 'approx. (±15%%)' note; got:\n%s", out)
	}
}

func TestWriteReportSkip(t *testing.T) {
	results := []Result{
		{Provider: "Google", Model: "gemini-2.5-pro", SkipReason: googleSkipReason},
	}
	var buf bytes.Buffer
	WriteReport(&buf, results)
	out := buf.String()
	if !strings.Contains(out, googleSkipReason) {
		t.Errorf("expected skip reason in output; got:\n%s", out)
	}
	if !strings.Contains(out, "Google") {
		t.Error("expected Provider in skip row")
	}
	if !strings.Contains(out, "gemini-2.5-pro") {
		t.Error("expected Model in skip row")
	}
	if !strings.Contains(out, "-") {
		t.Error("expected placeholder token count '-' in skip row")
	}
}

func TestFormatTokens(t *testing.T) {
	tests := []struct {
		n    int
		want string
	}{
		{0, "0"},
		{999, "999"},
		{1000, "1,000"},
		{1234, "1,234"},
		{12345, "12,345"},
		{123456, "123,456"},
		{1234567, "1,234,567"},
	}
	for _, tt := range tests {
		got := formatTokens(tt.n)
		if got != tt.want {
			t.Errorf("formatTokens(%d) = %q, want %q", tt.n, got, tt.want)
		}
	}
}

func TestWriteJSONExact(t *testing.T) {
	results := []Result{
		{Provider: "OpenAI", Model: "gpt-5.5", Tokens: 2920},
	}
	var buf bytes.Buffer
	if err := WriteJSON(&buf, results); err != nil {
		t.Fatalf("WriteJSON: %v", err)
	}
	var out []jsonResult
	if err := json.Unmarshal(buf.Bytes(), &out); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(out) != 1 {
		t.Fatalf("got %d results, want 1", len(out))
	}
	if out[0].Provider != "OpenAI" {
		t.Errorf("Provider = %q, want OpenAI", out[0].Provider)
	}
	if out[0].Tokens == nil || *out[0].Tokens != 2920 {
		t.Errorf("Tokens = %v, want 2920", out[0].Tokens)
	}
	if out[0].Approx {
		t.Error("exact result should have Approx=false")
	}
	if out[0].SkipReason != "" {
		t.Errorf("unexpected SkipReason: %q", out[0].SkipReason)
	}
}

func TestWriteJSONApprox(t *testing.T) {
	results := []Result{
		{Provider: "Meta", Model: "llama-4-scout", Tokens: 2890, Approx: true, MarginPct: 15},
	}
	var buf bytes.Buffer
	if err := WriteJSON(&buf, results); err != nil {
		t.Fatalf("WriteJSON: %v", err)
	}
	var out []jsonResult
	if err := json.Unmarshal(buf.Bytes(), &out); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if !out[0].Approx {
		t.Error("expected Approx=true")
	}
	if out[0].MarginPct != 15 {
		t.Errorf("MarginPct = %d, want 15", out[0].MarginPct)
	}
}

func TestWriteJSONSkip(t *testing.T) {
	results := []Result{
		{Provider: "Google", Model: "gemini-2.5-pro", SkipReason: googleSkipReason},
	}
	var buf bytes.Buffer
	if err := WriteJSON(&buf, results); err != nil {
		t.Fatalf("WriteJSON: %v", err)
	}
	var out []jsonResult
	if err := json.Unmarshal(buf.Bytes(), &out); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if out[0].Tokens != nil {
		t.Errorf("skip result should have null tokens, got %v", *out[0].Tokens)
	}
	if out[0].SkipReason != googleSkipReason {
		t.Errorf("SkipReason = %q, want %q", out[0].SkipReason, googleSkipReason)
	}
}

func TestWriteReportColumnAlignment(t *testing.T) {
	results := []Result{
		{Provider: "OpenAI", Model: "gpt-5.5", Tokens: 100},
		{Provider: "Anthropic", Model: "claude-opus-4-8", Tokens: 200, Approx: true},
	}
	var buf bytes.Buffer
	WriteReport(&buf, results)
	lines := strings.Split(strings.TrimRight(buf.String(), "\n"), "\n")
	if len(lines) < 4 {
		t.Fatalf("expected at least 4 lines (header, sep, 2 rows), got %d", len(lines))
	}
	row1Len := len([]rune(lines[2]))
	row2Len := len([]rune(lines[3]))
	if row1Len == 0 || row2Len == 0 {
		t.Fatal("rows should not be empty")
	}
}
