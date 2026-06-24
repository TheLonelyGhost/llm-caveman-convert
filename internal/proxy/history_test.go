package proxy_test

import (
	"testing"

	"github.com/TheLonelyGhost/llm-caveman-convert/internal/proxy"
)

func TestHistoryStoresCompressedForms(t *testing.T) {
	h := proxy.NewHistory()
	h.AppendUser("[enc]please explain")
	h.AppendAssistant("jwt = header.payload.sig")

	msgs := h.Messages()
	if len(msgs) != 2 {
		t.Fatalf("expected 2 turns, got %d", len(msgs))
	}
	if msgs[0].Role != "user" || msgs[0].Content != "[enc]please explain" {
		t.Errorf("unexpected user turn: %+v", msgs[0])
	}
	if msgs[1].Role != "assistant" || msgs[1].Content != "jwt = header.payload.sig" {
		t.Errorf("unexpected assistant turn: %+v", msgs[1])
	}
}

func TestHistoryByteStability(t *testing.T) {
	h := proxy.NewHistory()
	h.AppendUser("compressed user msg")
	h.AppendAssistant("compressed assistant msg")

	first := h.Messages()
	second := h.Messages()

	for i := range first {
		if first[i].Content != second[i].Content {
			t.Errorf("turn %d content changed between calls", i)
		}
	}
}

func TestHistoryImmutableAfterStore(t *testing.T) {
	h := proxy.NewHistory()
	h.AppendUser("original")
	msgs := h.Messages()
	msgs[0].Content = "mutated"

	stored := h.Messages()
	if stored[0].Content != "original" {
		t.Errorf("history mutated externally: %q", stored[0].Content)
	}
}

func TestPriorEmptyHistoryReturnsNil(t *testing.T) {
	h := proxy.NewHistory()
	if got := h.Prior(); got != nil {
		t.Errorf("expected nil, got %v", got)
	}
}

func TestPriorSingleTurnReturnsEmpty(t *testing.T) {
	h := proxy.NewHistory()
	h.AppendUser("only turn")
	prior := h.Prior()
	if len(prior) != 0 {
		t.Errorf("expected empty prior for single turn, got %v", prior)
	}
}
