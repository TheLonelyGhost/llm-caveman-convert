## 1. New `internal/compress` package

- [x] 1.1 Create `internal/compress/validate.go` with `ExtractHeadings`, `ExtractCodeBlocks`, `ExtractURLs`, and `Validate` (port of validate.py error-producing validators only)
- [x] 1.2 Create `internal/compress/validate_test.go` covering all scenarios in `specs/markdown-validation/spec.md`

## 2. Update `internal/llm/client.go`

- [x] 2.1 Replace `encodeSystemPrompt` with unified markdown-aware prompt: all fenced blocks verbatim, inline backticks verbatim, headings verbatim, URLs/paths/env-vars verbatim, tables structure preserved, redundant-phrasing contractions, connective-fluff removal, fragment style
- [x] 2.2 Add `Fix(ctx context.Context, original, compressed string, errors []string) (string, error)`
- [x] 2.3 Update `TestEncodeCallsLLM` and add `TestFixCallsLLM` in `internal/llm/client_test.go`

## 3. Update `internal/caveman/run.go`

- [x] 3.1 Add `const maxEncodeRetries = 2`
- [x] 3.2 Rewrite `Encode()` to: (a) check cache, (b) call `client.Encode()`, (c) call `compress.Validate()`, (d) on failure loop up to `maxEncodeRetries` calling `client.Fix()` + re-validate, (e) call `cache.SetEncode()` only on validation pass, (f) return last result regardless
- [x] 3.3 Update `internal/caveman/run_test.go`: add tests for retry-on-validation-failure, cache-not-written-on-exhausted-retries, cache-written-only-after-validation-pass

## 4. Update `caveman-cli` spec

- [x] 4.1 Remove obsolete scenarios "Encode handles code fences per compressor LLM instruction" and "Encode passes through untagged code blocks" from `openspec/specs/caveman-cli/spec.md`
- [x] 4.2 Add new scenarios to `openspec/specs/caveman-cli/spec.md` covering all-fenced-blocks-verbatim, inline-backtick-verbatim, headings-verbatim, URLs-verbatim, connective-fluff-removed, redundant-phrasing-contracted

## 5. Quality gates

- [x] 5.1 Run `task fmt && task vet && task lint && task test:cover && task vuln` — all pass, coverage ≥ 80%
