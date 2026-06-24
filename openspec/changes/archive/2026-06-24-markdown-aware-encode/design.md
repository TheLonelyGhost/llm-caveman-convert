## Context

`caveman --encode` calls a single LLM round-trip using a system prompt that handles fenced blocks inconsistently (compresses comments in language-tagged blocks; passes untagged blocks verbatim). No structural validation exists. The Python caveman-compress skill has a complete ruleset and a validate-then-fix loop, but it invokes Python and the Anthropic SDK directly. The goal is to bring parity into the existing Go binary, translating all Python logic into Go.

Current encode path:
```
stdin → caveman.Encode() → cache hit? → stdout
                         → miss → client.Encode() → LLM → cache.Set → stdout
```

## Goals / Non-Goals

**Goals:**
- Replace `encodeSystemPrompt` with a unified prompt that treats ALL fenced blocks as verbatim and applies the full caveman-compress ruleset
- Add `internal/compress` package: Go port of the skill's Python validators (headings, code blocks, URLs)
- Add `client.Fix()` to `llm.Client` for targeted correction of a failed-validation attempt
- Extend `caveman.Encode()` with a validate-and-retry loop (max 2 retries); write cache only on validation pass

**Non-Goals:**
- `--decode` path unchanged
- Proxy unchanged
- Cache storage format unchanged
- Sensitive-path detection (skill's `is_sensitive_path`) — irrelevant for stdin/stdout
- Warning-only validators (file paths, bullet count) — do not trigger retries; silently dropped
- Python runtime not required at all

## Decisions

### D1: All fenced blocks verbatim (no language-tag distinction)

**Decision:** Remove the "full lang tag → compress comments" rule. All fenced blocks (` ``` ` or `~~~`, any info string) are copied exactly.

**Rationale:** The skill's CRITICAL RULE is unambiguous: anything inside ` ``` ... ``` ` is read-only. The current distinction adds complexity for marginal gain and is the primary source of mis-compression. Simpler rule = fewer LLM errors = fewer retries.

**Alternative considered:** Keep comment compression for tagged blocks, extend validation to check per-line. Rejected: unprovable via structural validation (we can't diff "compressed comments" reliably).

---

### D2: New `internal/compress` package for validation

**Decision:** Create `internal/compress/validate.go` as a pure-Go port of the skill's `validate.py` extractors and validators. Only the error-producing checks are ported (headings count, code blocks exact-match, URL set equality). Warning-only checks (paths, bullets) are omitted.

**Rationale:** Keeping validation in a dedicated package makes it independently testable and decoupled from the LLM client. Warnings don't drive retries so adding them adds noise with no benefit.

**Extractors to port:**
- `ExtractHeadings(text string) []string` — regex `^#{1,6}\s+(.*)` multiline
- `ExtractCodeBlocks(text string) []string` — line-based CommonMark fence parser (same char, same or longer length closing fence)
- `ExtractURLs(text string) []string` — `https?://[^\s)]+`

**Validators:**
- `ValidateHeadings(orig, comp)` — error if len differs
- `ValidateCodeBlocks(orig, comp)` — error if slice not equal
- `ValidateURLs(orig, comp)` — error if sets differ

**`Validate(orig, comp string) ValidationResult`** returns `IsValid bool` + `Errors []string`.

---

### D3: `Fix()` lives on `llm.Client`

**Decision:** Add `func (c *Client) Fix(ctx, original, compressed string, errors []string) (string, error)` to `llm.Client`.

**Rationale:** `Fix()` is another LLM call with a different prompt — same pattern as `Encode()`/`Decode()`. Keeping prompt construction at the client layer maintains the existing separation: `internal/llm` owns all prompts; `internal/caveman` orchestrates.

**Alternative considered:** Build the fix prompt in `caveman.Encode()`, call a generic `client.Complete()`. Rejected: leaks prompt strings out of the `llm` package.

---

### D4: Cache deferred until validation passes

**Decision:** `caveman.Encode()` does not call `cache.SetEncode()` until `Validate()` returns `IsValid: true`. If all retries are exhausted and validation still fails, return the last LLM result anyway (best-effort) but do not cache it.

**Rationale:** Caching an invalid result would permanently serve broken output for that input. Returning best-effort on exhaustion preserves the original fallback behaviour (proxy falls back to original message on non-zero exit — but here we exit zero with the last attempt, consistent with current single-shot behaviour).

**Alternative considered:** Return error on exhausted retries. Rejected: breaks callers (proxy falls back gracefully on error, but that means the original uncompressed message is used — worse UX than a slightly-imperfect compression).

---

### D5: Max retries = 2 (matching the skill)

**Decision:** `const maxEncodeRetries = 2` in `internal/caveman/run.go`.

**Rationale:** Direct parity with `MAX_RETRIES = 2` in `compress.py`. Two attempts means up to 3 LLM calls total (1 compress + 2 fix). Beyond that, diminishing returns.

## Risks / Trade-offs

- **Latency spike on retry**: Up to 3 LLM calls vs 1 today. Mitigation: retries only trigger on structural validation failure, which should be rare with the improved prompt. Cache hit path is unchanged.
- **Fix prompt may not converge**: The fix prompt asks the LLM to restore only specific missing elements. If the LLM recompresses entirely, the fixed version may fail the same checks. Mitigation: best-effort return after exhaustion avoids hanging; prompt instructs "ONLY fix listed errors, leave everything else as-is".
- **Regex extractors diverge from CommonMark spec**: The Go port uses the same regexes as the Python skill. Both are approximate. Mitigation: edge cases (nested fences, ATX headings with trailing `#`) are handled by the line-based fence parser (same algorithm as validate.py).

## Migration Plan

- No schema or cache format changes; existing cache entries remain valid
- Existing cached encode results may be structurally invalid under the new rules — they will be served from cache as-is (cache key is unchanged: SHA-256 of raw input). This is acceptable: stale cache entries expire naturally as inputs change.
- No deployment steps; binary is replaced in-place
