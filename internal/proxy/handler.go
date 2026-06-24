package proxy

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os/exec"
	"strings"
)

const augmentInstruction = "Respond in compressed caveman-speak: omit articles, filler, pleasantries. Fragments ok. Keep identifiers, code, numbers exact."

type message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// Handler is an HTTP handler that proxies chat completion requests,
// compressing user messages and expanding assistant responses via caveman.
type Handler struct {
	backendURL string
	cavemanBin string
	httpClient *http.Client
	history    *History
}

// New creates a Handler that forwards to backendURL and invokes cavemanBin for encode/decode.
func New(backendURL, cavemanBin string) *Handler {
	return &Handler{
		backendURL: strings.TrimRight(backendURL, "/"),
		cavemanBin: cavemanBin,
		httpClient: &http.Client{},
		history:    NewHistory(),
	}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/v1/chat/completions" {
		http.NotFound(w, r)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "read body", http.StatusBadRequest)
		return
	}

	var raw map[string]json.RawMessage
	if err := json.Unmarshal(body, &raw); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	var msgs []message
	if err := json.Unmarshal(raw["messages"], &msgs); err != nil {
		http.Error(w, "invalid messages", http.StatusBadRequest)
		return
	}

	msgs = h.augmentSystemPrompt(msgs)
	compressed := make([]message, len(msgs))
	for i, m := range msgs {
		compressed[i] = m
		if m.Role == "user" {
			if out, err := h.caveman("--encode", m.Content); err == nil {
				compressed[i].Content = out
			}
			h.history.AppendUser(compressed[i].Content)
		}
	}

	// Build outgoing: system messages first, then history, then current non-system messages.
	var systemMsgs, otherMsgs []message
	for _, m := range compressed {
		if m.Role == "system" {
			systemMsgs = append(systemMsgs, m)
		} else {
			otherMsgs = append(otherMsgs, m)
		}
	}
	// History contains prior turns (user+assistant); current turn's user msg is already appended.
	// Exclude the just-appended current user from history to avoid duplication: history has it,
	// and otherMsgs also has it. Use all but last entry from history (prior turns only).
	priorHistory := h.history.Prior()
	outgoing := systemMsgs
	outgoing = append(outgoing, priorHistory...)
	outgoing = append(outgoing, otherMsgs...)
	msgsJSON, _ := json.Marshal(outgoing)
	raw["messages"] = msgsJSON
	modifiedBody, _ := json.Marshal(raw)

	upstreamURL := h.backendURL + "/v1/chat/completions"
	upstreamReq, err := http.NewRequestWithContext(r.Context(), r.Method, upstreamURL, bytes.NewReader(modifiedBody))
	if err != nil {
		http.Error(w, "upstream request", http.StatusInternalServerError)
		return
	}

	for k, vs := range r.Header {
		if strings.EqualFold(k, "Host") {
			continue
		}
		for _, v := range vs {
			upstreamReq.Header.Add(k, v)
		}
	}
	upstreamReq.Header.Set("Content-Type", "application/json")

	resp, err := h.httpClient.Do(upstreamReq) //nolint:gosec // SSRF is intentional: proxy forwards to operator-configured backend URL
	if err != nil || resp == nil {
		http.Error(w, fmt.Sprintf("upstream: %v", err), http.StatusBadGateway)
		return
	}
	defer func() { _ = resp.Body.Close() }()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, "read upstream response", http.StatusBadGateway)
		return
	}

	var rawResp map[string]json.RawMessage
	if err := json.Unmarshal(respBody, &rawResp); err != nil {
		for k, vs := range resp.Header {
			for _, v := range vs {
				w.Header().Add(k, v)
			}
		}
		w.WriteHeader(resp.StatusCode)
		_, _ = w.Write(respBody)
		return
	}

	var choices []json.RawMessage
	if err := json.Unmarshal(rawResp["choices"], &choices); err == nil && len(choices) > 0 {
		var rawChoice map[string]json.RawMessage
		if err := json.Unmarshal(choices[0], &rawChoice); err == nil {
			var msg message
			if err := json.Unmarshal(rawChoice["message"], &msg); err == nil {
				caveman := msg.Content
				h.history.AppendAssistant(caveman)
				if decoded, err := h.caveman("--decode", caveman); err == nil {
					msg.Content = decoded
				}
				rawChoice["message"], _ = json.Marshal(msg)
				choices[0], _ = json.Marshal(rawChoice)
				rawResp["choices"], _ = json.Marshal(choices)
				respBody, _ = json.Marshal(rawResp)
			}
		}
	}

	for k, vs := range resp.Header {
		if strings.EqualFold(k, "Content-Length") {
			continue
		}
		for _, v := range vs {
			w.Header().Add(k, v)
		}
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(resp.StatusCode)
	_, _ = w.Write(respBody)
}

func (h *Handler) augmentSystemPrompt(msgs []message) []message {
	for i, m := range msgs {
		if m.Role == "system" {
			msgs[i].Content = m.Content + "\n" + augmentInstruction
			return msgs
		}
	}
	return append([]message{{Role: "system", Content: augmentInstruction}}, msgs...)
}

func (h *Handler) caveman(flag, input string) (string, error) {
	cmd := exec.Command(h.cavemanBin, flag) //nolint:gosec // G204: cavemanBin is operator-configured binary path, not user input
	cmd.Stdin = strings.NewReader(input)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("caveman %s: %w", flag, err)
	}
	return stdout.String(), nil
}
