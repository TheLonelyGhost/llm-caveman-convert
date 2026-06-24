## 1. Dependencies

- [x] 1.1 Add `github.com/pkoukk/tiktoken-go` to `go.mod` and run `go mod tidy`
- [x] 1.2 Add `google.golang.org/genai` to `go.mod` and run `go mod tidy`

## 2. Counter Interface and Types

- [x] 2.1 Create `internal/tokens/counter.go` — define `Counter` interface (`Count(ctx, text) (int, error)`), `Result` struct (`Provider`, `Model`, `Tokens`, `Approx bool`, `MarginPct int`, `SkipReason string`), and `Entry` struct tying a model to its `Counter`
- [x] 2.2 Create `internal/tokens/doc.go` — package doc comment

## 3. Local BPE Counter

- [x] 3.1 Create `internal/tokens/tiktoken.go` — implement a `tiktokenCounter` that holds a pre-initialized `*tiktoken.Tiktoken` and implements `Counter`
- [x] 3.2 Add constructor `newTiktokenCounter(encoding string) (Counter, error)` that initializes the encoder once

## 4. Google Gemini Counter

- [x] 4.1 Create `internal/tokens/google.go` — implement a `geminiCounter` that calls `google.golang.org/genai` `CountTokens`; returns a sentinel `Result` with `SkipReason = "requires GOOGLE_API_KEY"` when key absent

## 5. Model Registry

- [x] 5.1 Create `internal/tokens/registry.go` — define package-level `Registry []Entry` initialized in `func init()` with all 15 model entries:
  - OpenAI: `gpt-5.5`, `gpt-5.4`, `gpt-5.4-mini` (o200k_base, exact)
  - Anthropic: `claude-opus-4-8`, `claude-sonnet-4-6`, `claude-haiku-4-5` (cl100k_base, approx, MarginPct=0)
  - xAI: `grok-4.3`, `grok-4.20-0309`, `grok-4.20-multi-agent-0309` (o200k_base, approx, MarginPct=0)
  - Google: `gemini-2.5-pro`, `gemini-2.5-flash`, `gemini-2.0-flash` (API or skip)
  - Meta: `llama-4-scout`, `llama-4-maverick` (o200k_base, approx, MarginPct=15); `llama-3.3-70b` (cl100k_base, approx, MarginPct=10)

## 6. Report Formatter

- [x] 6.1 Create `internal/tokens/report.go` — implement `WriteReport(w io.Writer, results []Result)` that prints the aligned table with header, separator line, and one row per result; format counts with thousands commas; render note column as `~`, `~ (±N%)`, or skip-reason string

## 7. CLI Binary

- [x] 7.1 Create `cmd/tokcount/main.go` — read all stdin, iterate `tokens.Registry`, call each `Counter`, collect `[]Result`, call `tokens.WriteReport(os.Stdout, results)`
- [x] 7.2 Create `cmd/tokcount/main_test.go` — test `loadInput` (or equivalent helper) and any unexported functions

## 8. Tests

- [x] 8.1 Add `internal/tokens/tiktoken_test.go` — test `tiktokenCounter.Count` for both `o200k_base` and `cl100k_base` with known inputs; assert counts match Python tiktoken reference values within 0%
- [x] 8.2 Add `internal/tokens/google_test.go` — test skip-reason sentinel path (no key); mock or skip API path
- [x] 8.3 Add `internal/tokens/report_test.go` — test `WriteReport` output format: column alignment, thousands separator, note rendering for exact/approx/skip cases
- [x] 8.4 Add `internal/tokens/registry_test.go` — assert Registry has exactly 15 entries; assert each entry has a non-nil Counter and non-empty Provider/Model

## 9. Build Integration

- [x] 9.1 Add `tokcount` binary to `Taskfile.yml` build task alongside `caveman` and `proxy`
- [x] 9.2 Verify `task fmt && task vet && task lint && task test:cover && task vuln` all pass with coverage ≥ 80%
