package tokens

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

const (
	colProvider = 12
	colModel    = 30
	colTokens   = 8
)

// WriteReport writes an aligned token count table to w.
func WriteReport(w io.Writer, results []Result) {
	header := fmt.Sprintf("%-*s  %-*s  %*s  %s",
		colProvider, "Provider",
		colModel, "Model",
		colTokens, "Tokens",
		"Note",
	)
	sep := strings.Repeat("─", len(header))
	_, _ = fmt.Fprintln(w, header)
	_, _ = fmt.Fprintln(w, sep)
	for _, r := range results {
		_, _ = fmt.Fprintln(w, formatRow(r))
	}
}

func formatRow(r Result) string {
	if r.SkipReason != "" {
		return fmt.Sprintf("%-*s  %-*s  %*s  (%s)",
			colProvider, r.Provider,
			colModel, r.Model,
			colTokens, "-",
			r.SkipReason,
		)
	}
	return fmt.Sprintf("%-*s  %-*s  %*s  %s",
		colProvider, r.Provider,
		colModel, r.Model,
		colTokens, formatTokens(r.Tokens),
		noteString(r),
	)
}

func noteString(r Result) string {
	if !r.Approx {
		return ""
	}
	if r.MarginPct > 0 {
		return fmt.Sprintf("approx. (±%d%%)", r.MarginPct)
	}
	return "approx."
}

func formatTokens(n int) string {
	s := fmt.Sprintf("%d", n)
	out := make([]byte, 0, len(s)+len(s)/3)
	for i := 0; i < len(s); i++ {
		pos := len(s) - i
		if i > 0 && pos%3 == 0 {
			out = append(out, ',')
		}
		out = append(out, s[i])
	}
	return string(out)
}

// jsonResult is the JSON representation of a single model's token count.
type jsonResult struct {
	Provider   string `json:"provider"`
	Model      string `json:"model"`
	Tokens     *int   `json:"tokens"`
	Approx     bool   `json:"approx"`
	MarginPct  int    `json:"margin_pct,omitempty"`
	SkipReason string `json:"skip_reason,omitempty"`
}

// WriteJSON writes results as a JSON array to w.
func WriteJSON(w io.Writer, results []Result) error {
	out := make([]jsonResult, len(results))
	for i, r := range results {
		jr := jsonResult{
			Provider:   r.Provider,
			Model:      r.Model,
			Approx:     r.Approx,
			MarginPct:  r.MarginPct,
			SkipReason: r.SkipReason,
		}
		if r.SkipReason == "" {
			n := r.Tokens
			jr.Tokens = &n
		}
		out[i] = jr
	}
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(out)
}
