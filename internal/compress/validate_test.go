package compress_test

import (
	"strings"
	"testing"

	"github.com/TheLonelyGhost/llm-caveman-convert/internal/compress"
)

func TestExtractHeadingsInOrder(t *testing.T) {
	text := "# H1\n\nsome text\n\n## H2\n\n### H3"
	got := compress.ExtractHeadings(text)
	want := []string{"# H1", "## H2", "### H3"}
	if len(got) != len(want) {
		t.Fatalf("len: got %d want %d: %v", len(got), len(want), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("[%d] got %q want %q", i, got[i], want[i])
		}
	}
}

func TestExtractHeadingsNone(t *testing.T) {
	got := compress.ExtractHeadings("no headings here")
	if got == nil {
		t.Error("expected non-nil empty slice, got nil")
	}
	if len(got) != 0 {
		t.Errorf("expected empty, got %v", got)
	}
}

func TestExtractCodeBlocksIncludesFenceLines(t *testing.T) {
	text := "```python\nprint('hi')\n```"
	blocks := compress.ExtractCodeBlocks(text)
	if len(blocks) != 1 {
		t.Fatalf("expected 1 block, got %d", len(blocks))
	}
	if blocks[0] != text {
		t.Errorf("block content mismatch:\ngot:  %q\nwant: %q", blocks[0], text)
	}
}

func TestExtractCodeBlocksIgnoresUnclosed(t *testing.T) {
	text := "```\nunclosed"
	blocks := compress.ExtractCodeBlocks(text)
	if len(blocks) != 0 {
		t.Errorf("expected no blocks for unclosed fence, got %d", len(blocks))
	}
}

func TestExtractCodeBlocksBacktickAndTilde(t *testing.T) {
	text := "```\nfoo\n```\n\n~~~\nbar\n~~~"
	blocks := compress.ExtractCodeBlocks(text)
	if len(blocks) != 2 {
		t.Fatalf("expected 2 blocks, got %d: %v", len(blocks), blocks)
	}
}

func TestExtractCodeBlocksClosingMustMatchChar(t *testing.T) {
	text := "```\nfoo\n~~~\nstill open\n```"
	blocks := compress.ExtractCodeBlocks(text)
	if len(blocks) != 1 {
		t.Fatalf("expected 1 block (closed by matching ``), got %d: %v", len(blocks), blocks)
	}
	if !strings.Contains(blocks[0], "~~~") {
		t.Error("expected the tilde line to be interior content, not closing fence")
	}
}

func TestExtractURLsHTTPAndHTTPS(t *testing.T) {
	text := "see https://example.com and http://other.org for more"
	urls := compress.ExtractURLs(text)
	set := make(map[string]bool)
	for _, u := range urls {
		set[u] = true
	}
	if !set["https://example.com"] {
		t.Error("missing https://example.com")
	}
	if !set["http://other.org"] {
		t.Error("missing http://other.org")
	}
}

func TestExtractURLsNone(t *testing.T) {
	got := compress.ExtractURLs("no urls here")
	if len(got) != 0 {
		t.Errorf("expected empty, got %v", got)
	}
}

func TestValidateAllPass(t *testing.T) {
	orig := "## Title\n\nSome verbose text with https://example.com and more words.\n\n```go\nfmt.Println(\"hi\")\n```"
	comp := "## Title\n\nTerse text https://example.com.\n\n```go\nfmt.Println(\"hi\")\n```"
	result := compress.Validate(orig, comp)
	if !result.IsValid {
		t.Errorf("expected valid, got errors: %v", result.Errors)
	}
	if len(result.Errors) != 0 {
		t.Errorf("expected no errors, got: %v", result.Errors)
	}
}

func TestValidateHeadingCountMismatch(t *testing.T) {
	orig := "## A\n\n## B\n\ntext"
	comp := "## A\n\ntext"
	result := compress.Validate(orig, comp)
	if result.IsValid {
		t.Error("expected invalid due to heading count mismatch")
	}
	found := false
	for _, e := range result.Errors {
		if strings.Contains(e, "heading") {
			found = true
		}
	}
	if !found {
		t.Errorf("expected heading error in %v", result.Errors)
	}
}

func TestValidateCodeBlockMismatch(t *testing.T) {
	orig := "text\n\n```\noriginal code\n```"
	comp := "text\n\n```\nmodified code\n```"
	result := compress.Validate(orig, comp)
	if result.IsValid {
		t.Error("expected invalid due to code block mismatch")
	}
	found := false
	for _, e := range result.Errors {
		if strings.Contains(e, "code block") {
			found = true
		}
	}
	if !found {
		t.Errorf("expected code block error in %v", result.Errors)
	}
}

func TestValidateURLSetMismatch(t *testing.T) {
	orig := "see https://example.com for details"
	comp := "see https://other.com for details"
	result := compress.Validate(orig, comp)
	if result.IsValid {
		t.Error("expected invalid due to URL mismatch")
	}
	found := false
	for _, e := range result.Errors {
		if strings.Contains(e, "URL") {
			found = true
		}
	}
	if !found {
		t.Errorf("expected URL error in %v", result.Errors)
	}
}

func TestValidateMultipleFailuresAccumulate(t *testing.T) {
	orig := "## H\n\nhttps://example.com"
	comp := "text only"
	result := compress.Validate(orig, comp)
	if result.IsValid {
		t.Error("expected invalid")
	}
	if len(result.Errors) < 2 {
		t.Errorf("expected ≥2 errors, got %d: %v", len(result.Errors), result.Errors)
	}
}
