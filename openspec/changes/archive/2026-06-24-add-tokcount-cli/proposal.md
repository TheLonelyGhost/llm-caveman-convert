## Why

The project compresses LLM message content to reduce token spend, but has no way to measure the actual savings. A dedicated token-counting CLI fills this gap, giving users a concrete report of how many tokens a file consumes across the major frontier model families.

## What Changes

- New binary `cmd/tokcount` — reads stdin, writes a per-model token count report to stdout
- New internal package `internal/tokens` — `Counter` interface, model registry, tiktoken-based local counters, Google Gemini API counter, and report formatter
- New dependency: `github.com/pkoukk/tiktoken-go` for local BPE tokenization (o200k_base and cl100k_base)
- New dependency: `google.golang.org/genai` for Gemini token counting via API

## Capabilities

### New Capabilities

- `token-counting`: Count tokens for a given text against a fixed registry of frontier models (OpenAI, Anthropic, xAI, Google, Meta), with per-model accuracy metadata, and a human-readable table report

### Modified Capabilities

_(none)_

## Impact

- `go.mod` / `go.sum`: two new direct dependencies
- `internal/tokens/`: new package, no coupling to existing packages at creation time; designed for future use by `internal/proxy` or `cmd/caveman`
- `cmd/tokcount/`: new binary, standalone
- No changes to existing `caveman` CLI, proxy, or cache behaviour
