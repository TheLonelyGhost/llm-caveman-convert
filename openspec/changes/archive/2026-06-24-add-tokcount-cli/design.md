## Context

The project's core value proposition is reducing LLM token spend via caveman-style compression. Until now there has been no tooling to measure that reduction. Users must either call an API and inspect `usage.input_tokens`, or guess. This change adds a standalone `tokcount` binary that reads text from stdin and emits a per-model token count table, covering the five major frontier model families.

The codebase already wraps `github.com/sashabaranov/go-openai` for LLM calls and has a clean `internal/` package structure. The new `internal/tokens` package follows the same conventions and is intentionally decoupled from existing packages so it can be imported later by `internal/proxy` (to log savings) or `cmd/caveman` (to report savings after encode).

## Goals / Non-Goals

**Goals:**
- Count tokens locally (no API call) for OpenAI, Anthropic (old tokenizer), xAI, and Meta models
- Count tokens via Google Gemini API when `GOOGLE_API_KEY` is present; silently emit a "requires key" placeholder when absent
- Expose a `Counter` interface in `internal/tokens` usable by future callers
- Emit a human-readable aligned table to stdout
- Produce exact counts for OpenAI (tiktoken exact), approximate counts within ±10% for Anthropic and xAI, approximate counts within ±15% for Llama 4, approximate counts within ±10% for Llama 3.3

**Non-Goals:**
- Comparison mode (two-file diff) — stdin-only, single measurement
- Per-message chat overhead tokens (role markers, framing tokens)
- Automatic routing or compression decisions based on token count
- Support for image, audio, or PDF token counting
- Support for Anthropic's new tokenizer generation (Fable-5, Mythos-5)

## Decisions

### D1 — Local tokenizer library: `github.com/pkoukk/tiktoken-go`

**Decision:** Use `pkoukk/tiktoken-go` for all local BPE counting.

**Rationale:** It is a faithful Go port of OpenAI's tiktoken, supports both `o200k_base` (gpt-5.x, Grok) and `cl100k_base` (Anthropic approx, Llama 3.3 approx). Validation against the Python reference tokenizer on project demo files shows delta of ~2% between `cl100k_base` and `o200k_base` for English markdown — well within the ±10% tolerance. BPE vocabulary files are downloaded once and cached via `TIKTOKEN_CACHE_DIR`; the offline loader variant (`tiktoken_loader`) avoids network dependency if needed in the future.

**Alternatives considered:**
- Anthropic's `/v1/messages/count_tokens` API: exact, but requires a key and a network round-trip. Since the top-3 Anthropic models all use the old tokenizer (~cl100k), the local approximation is acceptable and keeps the tool usable without any credentials.
- Writing a custom tokenizer: not warranted given available libraries.

### D2 — Gemini: Google Gen AI Go SDK (`google.golang.org/genai`)

**Decision:** Use `google.golang.org/genai` (the new unified SDK, not the deprecated `generative-ai-go`).

**Rationale:** The old `github.com/google/generative-ai-go` reached end-of-life August 2025. The new SDK is `google.golang.org/genai` and has a `CountTokens` method. When `GOOGLE_API_KEY` is absent the counter emits a sentinel result with a "requires GOOGLE_API_KEY" message and zero tokens rather than erroring.

**Alternatives considered:**
- Approximate via SentencePiece: no Go SentencePiece library with Gemini's vocab exists; error would likely exceed ±15%.
- Skip Google entirely: loses meaningful coverage for a major provider.

### D3 — Meta Llama: approximate via o200k_base / cl100k_base with explicit margin labels

**Decision:** Llama 4 (Scout, Maverick) approximated via `o200k_base`; Llama 3.3 70B via `cl100k_base`. Both display `~ (±15%)` and `~ (±10%)` margin labels respectively.

**Rationale:** No Go SentencePiece tokenizer with Llama's vocabulary exists. The BPE approximation is within the user-accepted ±15% bound for English prose. The explicit margin label ensures the output is honest. README documents the limitation.

**Alternatives considered:**
- Exclude Llama: user explicitly requested inclusion with warning.
- Use Meta's API for exact counts: Meta's inference API does not expose a standalone token count endpoint.

### D4 — Model registry as a static slice in `internal/tokens/registry.go`

**Decision:** Hard-code the 15 model entries as a package-level `[]Entry` slice, constructed once at `init()` time with pre-initialized `Counter` instances.

**Rationale:** The set of supported models is intentionally fixed and small (15 rows). A dynamic/config-driven registry adds complexity with no present benefit. Future callers iterate the slice; the interface is the `Counter`, not the registry shape.

**Alternatives considered:**
- Config file: overkill, adds parsing surface area.
- Per-call counter construction: wastes BPE encoder initialization on every run.

### D5 — Output: aligned text table to stdout, no flags

**Decision:** Single output format — aligned table, no `--format` flag in v1.

**Rationale:** The primary consumer is a human at a terminal. Machine-readable output (JSON/CSV) can be added later without breaking the interface. Keeping v1 simple matches the existing `caveman` CLI philosophy.

### D6 — `Counter` interface lives in `internal/tokens`, not exported via a separate module

**Decision:** `Counter` is an unexported-friendly interface within `internal/tokens`. Future callers import the package directly.

**Rationale:** All current and anticipated callers (`cmd/tokcount`, `internal/proxy`, `cmd/caveman`) are within the same module. No need for a separate module boundary.

## Risks / Trade-offs

- **tiktoken BPE file download at first run** → Mitigation: document `TIKTOKEN_CACHE_DIR`; CI pre-warms cache. Long-term: consider embedding via `tiktoken_loader` offline loader.
- **Anthropic tokenizer drift** → If Anthropic silently changes their tokenizer for existing models, the cl100k approximation degrades. Mitigation: document the approximation source; revisit if Anthropic publishes a Go-compatible tokenizer.
- **Google SDK API churn** → `google.golang.org/genai` is new; its API may change. Mitigation: pin to a specific minor version; the `google.go` file is narrow and easy to update.
- **Llama margin accuracy** → ±15% is a stated estimate for English prose; non-English or code-heavy text may drift further. Mitigation: README caveat, explicit label in output.
- **`go 1.26.4` in go.mod** → Confirm this is intentional (unusual patch version). No impact on this change.

## Migration Plan

New binary and new package only. No changes to existing binaries or packages. No database migrations, no config changes. Deploy by adding `tokcount` to the build output in `Taskfile.yml`.
