package tokens

import "log"

const (
	encO200k  = "o200k_base"
	encCL100k = "cl100k_base"
)

// Registry is the package-level list of all supported model entries.
// It is initialized once at program startup.
var Registry []Entry //nolint:gochecknoglobals

func init() {
	o200k, err := newTiktokenCounter(encO200k)
	if err != nil {
		log.Fatalf("tokens: init o200k_base: %v", err)
	}
	cl100k, err := newTiktokenCounter(encCL100k)
	if err != nil {
		log.Fatalf("tokens: init cl100k_base: %v", err)
	}

	Registry = []Entry{
		{Provider: "OpenAI", Model: "gpt-5.5", counter: o200k},
		{Provider: "OpenAI", Model: "gpt-5.4", counter: o200k},
		{Provider: "OpenAI", Model: "gpt-5.4-mini", counter: o200k},

		{Provider: "Anthropic", Model: "claude-opus-4-8", Approx: true, MarginPct: 2, counter: cl100k},
		{Provider: "Anthropic", Model: "claude-sonnet-4-6", Approx: true, MarginPct: 2, counter: cl100k},
		{Provider: "Anthropic", Model: "claude-haiku-4-5", Approx: true, MarginPct: 2, counter: cl100k},

		{Provider: "xAI", Model: "grok-4.3", Approx: true, MarginPct: 2, counter: o200k},
		{Provider: "xAI", Model: "grok-4.20-0309", Approx: true, MarginPct: 2, counter: o200k},
		{Provider: "xAI", Model: "grok-4.20-multi-agent-0309", Approx: true, MarginPct: 2, counter: o200k},

		{Provider: "Google", Model: "gemini-2.5-pro", counter: newGeminiCounter("gemini-2.5-pro")},
		{Provider: "Google", Model: "gemini-2.5-flash", counter: newGeminiCounter("gemini-2.5-flash")},
		{Provider: "Google", Model: "gemini-2.0-flash", counter: newGeminiCounter("gemini-2.0-flash")},

		{Provider: "Meta", Model: "llama-4-scout", Approx: true, MarginPct: 15, counter: o200k},
		{Provider: "Meta", Model: "llama-4-maverick", Approx: true, MarginPct: 15, counter: o200k},
		{Provider: "Meta", Model: "llama-3.3-70b", Approx: true, MarginPct: 10, counter: cl100k},
	}
}
