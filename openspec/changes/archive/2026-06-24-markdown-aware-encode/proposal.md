## Why

The `--encode` compressor treats all fenced code blocks inconsistently and lacks validation: it compresses comments in language-tagged blocks but ignores markdown structural invariants (headings, URLs, inline code, tables). Since all inputs are markdown, the compressor should apply the full caveman-compress ruleset and verify output correctness before caching.

## What Changes

- Replace `encodeSystemPrompt` in `internal/llm/client.go` with a unified markdown-aware prompt aligned to the caveman-compress skill rules
- Add `Fix()` method to `llm.Client` for targeted error correction on failed validation
- Add `internal/compress/validate.go`: Go port of the skill's Python validators (headings, code blocks, URLs — the error-producing checks)
- Update `caveman.Encode()` in `internal/caveman/run.go` to run validate-and-retry loop (up to 2 retries); cache only on validation pass

## Capabilities

### New Capabilities
- `markdown-validation`: Structural validation of compressed markdown output (headings count, code block exact-match, URL set preservation); drives retry logic in the encode path

### Modified Capabilities
- `message-compression`: Encode prompt now treats ALL fenced blocks as verbatim, preserves headings/inline-code/URLs/paths/env-vars exactly, and applies full caveman-compress compression rules (redundant phrasing, connective fluff, fragment style)
- `compression-cache`: Cache write for encode is deferred until validation passes; a failed-then-retried encode that ultimately passes is cached on final success; exhausted retries produce no cache entry

## Impact

- `internal/llm/client.go`: prompt replacement + new `Fix()` method
- `internal/caveman/run.go`: retry loop wraps existing `Encode()` flow
- New package `internal/compress/` with `validate.go` and tests
- Existing `caveman-cli` spec scenarios for fenced-block handling become obsolete and will be replaced
- No changes to `--decode`, proxy, or cache storage format
