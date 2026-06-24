# Agent Instructions

Before change complete, run all quality gates in order:

```bash
task fmt
task vet
task lint
task test:cover
task vuln
```

All must pass zero errors. Fix failures before finishing.

Test coverage must remain at or above 80%. `task test:cover` enforces this threshold and will fail if coverage drops below it.

## Lint notes

`task lint` runs `golangci-lint` then `nilaway` as a second pass. Both must pass.

**nilaway**: does not treat `t.Fatal`/`t.Fatalf` as terminating. After any `http.Client.Do` (or `http.Get`/`http.Post`), guard with `err != nil || resp == nil` before dereferencing `resp`.

**errcheck**: always check or explicitly discard (`_ =`) return values including `defer f.Close()` — use `defer func() { _ = f.Close() }()`.

**bodyclose**: every `http.Response` must have its `Body` drained and closed, including fire-and-forget requests.

**gosec**: intentional SSRF (G704) and subprocess (G204) are suppressed with `//nolint:gosec // <reason>`. Use `http.Server{ReadTimeout: ..., WriteTimeout: ..., IdleTimeout: ...}` instead of `http.ListenAndServe` (G114). Directory permissions must be `0o750` or less (G301).

**revive**: all exported symbols need doc comments.

## Testing notes

`cmd/` packages (`package main`) can be tested with a `main_test.go` using `package main` — this gives access to unexported functions like `loadConfig` and `envOrDefault`.

`nilaway` false positives on production nil-checks-before-deref can be suppressed by adding `|| ptr == nil` to the existing error guard.
