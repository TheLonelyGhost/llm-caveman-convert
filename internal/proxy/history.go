package proxy

import "sync"

type turn struct {
	role    string
	content string
}

// History tracks compressed conversation turns for multi-turn context injection.
type History struct {
	mu    sync.Mutex
	turns []turn
}

// NewHistory creates an empty History.
func NewHistory() *History { return &History{} }

// AppendUser records a compressed user message.
func (h *History) AppendUser(compressed string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.turns = append(h.turns, turn{role: "user", content: compressed})
}

// AppendAssistant records a compressed assistant message.
func (h *History) AppendAssistant(compressed string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.turns = append(h.turns, turn{role: "assistant", content: compressed})
}

// Messages returns all recorded turns as chat messages.
func (h *History) Messages() []message {
	h.mu.Lock()
	defer h.mu.Unlock()
	out := make([]message, len(h.turns))
	for i, t := range h.turns {
		out[i] = message{Role: t.role, Content: t.content}
	}
	return out
}

// Prior returns all turns except the last, for injecting history before the current turn.
func (h *History) Prior() []message {
	h.mu.Lock()
	defer h.mu.Unlock()
	if len(h.turns) == 0 {
		return nil
	}
	prior := h.turns[:len(h.turns)-1]
	out := make([]message, len(prior))
	for i, t := range prior {
		out[i] = message{Role: t.role, Content: t.content}
	}
	return out
}
