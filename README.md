# llm-caveman-convert

> [!warning]
>
> This is an exploratory project where use of caveman-speak is tested as a compression algorithm for real-time token savings when using a Large Language Model (LLM). It is not intended for production use.

Transparent compression proxy for OpenAI-compatible LLM APIs. Compresses your messages to caveman-speak before sending, instructs the backend to respond in caveman-speak, and expands responses back to normal English before returning them. You write and read normal English; the backend only ever sees compressed text.

Token savings: 25–60% on prose. Savings compound as conversation history grows.

---

## How it works

```
you (normal English)
        ↓
  proxy: caveman --encode
        ↓
backend LLM (caveman-speak)
        ↓
  proxy: caveman --decode
        ↓
you (normal English)
```

- **`caveman`** — standalone CLI that compresses/decompresses text via a small compressor LLM. Maintains a persistent cache so repeated messages never hit the LLM twice.
- **`proxy`** — HTTP proxy that listens on `/v1/chat/completions`, intercepts OpenAI-compatible request/response bodies, invokes `caveman` as a subprocess, and forwards everything else unchanged.
- **`tokcount`** — utility CLI that counts tokens for a given text across major frontier model families (OpenAI, Anthropic, xAI, Google, Meta). Use it to measure compression savings.

Conversation history is stored compressed for token efficiency and provider prefix-cache stability. Decompressed text is ephemeral.

---

## Binaries

### `caveman`

Compresses plain English to caveman-speak, or expands it back.

```sh
echo "In order to make sure the system works correctly, you should check the logs." | caveman --encode
# → "ensure system works correctly, check logs"

echo "ensure system works correctly, check logs" | caveman --decode
# → "To ensure the system works correctly, you should check the logs."
```

What is preserved verbatim:
- All fenced code blocks (` ```python `, ` ```js `, ` ``` `, etc.)
- Inline backtick code (`` `funcName` ``)
- ATX headings (`## Section Title`)
- HTTP/HTTPS URLs
- Identifiers, numbers

What is removed or contracted:
- Articles, filler, pleasantries, hedging
- Connective words: however, furthermore, additionally, in addition
- Verbose phrases: "in order to" → "to", "make sure to" → "ensure", "the reason is because" → "because"

Results are cached by `hash(input)` in independent encode/decode namespaces. Cache is checked before every LLM call. Cache directory is created automatically if absent.

The `caveman` binary owns its own compressor LLM configuration (endpoint, model), separate from the backend LLM used by the proxy.

On LLM failure: exits non-zero, writes nothing to stdout. The proxy falls back to the original message verbatim.

### `proxy`

OpenAI-compatible proxy with transparent caveman compression.

Listens on `/v1/chat/completions` only. All other paths return 404. All client headers are forwarded to the upstream; `Host` is rewritten to the upstream host.

Request flow:
1. Parse `messages[]` from request body
2. Compress each user message via `caveman --encode`
3. Append caveman-response instruction to system prompt
4. Forward to backend
5. Decode `choices[0].message.content` via `caveman --decode`
6. Return modified response to caller

Streaming/SSE is not supported. The proxy waits for the full upstream response before decoding.

The `caveman` binary must be present and executable at startup; proxy fails fast with a clear error if not found.

### `tokcount`

Counts tokens for stdin across frontier models.

```sh
cat myfile.txt | tokcount
```

Output (example):

```
Provider    Model                         Tokens    Note
OpenAI      gpt-5.5                        2,920
OpenAI      gpt-5.4                        2,920
OpenAI      gpt-5.4-mini                   2,920
Anthropic   claude-opus-4-8                2,874    ~
Anthropic   claude-sonnet-4-6              2,874    ~
Anthropic   claude-haiku-4-5               2,874    ~
xAI         grok-4.3                       2,920    ~
Google      gemini-2.5-pro     (requires GOOGLE_API_KEY)
Meta        llama-4-scout                  2,891    ~ (±15%)
...
```

- OpenAI: exact, via `o200k_base` BPE locally
- Anthropic: approximate (`~`), via `cl100k_base` BPE
- xAI: approximate (`~`), via `o200k_base` BPE
- Google: exact when `GOOGLE_API_KEY` set; placeholder when absent
- Meta: approximate with explicit margin (`~ (±15%)` or `~ (±10%)`)

> **Note:** Google Gemini token counts require a valid `GOOGLE_API_KEY` environment variable. Without it, each Gemini model row displays `(requires GOOGLE_API_KEY)` instead of a count.

> **Note:** Meta Llama token counts are BPE approximations with a significant margin of error (`±15%` for Llama 4, `±10%` for Llama 3.3). Do not rely on these counts for precise context-window budgeting.

---

## Building

Requires [Go](https://golang.org) and [Task](https://taskfile.dev).

```sh
task build          # all three binaries, all platforms (linux/darwin/windows × amd64/arm64)
task build:caveman  # caveman only
task build:proxy    # proxy only
task build:tokcount # tokcount only
```

Binaries land in `out/<os>_<arch>/`. A symlink to the current platform's binary is placed in `bin/`.

---

## Development

```sh
task fmt          # gofmt -s
task vet          # go vet
task lint         # golangci-lint + nilaway
task test:cover   # tests with 80% coverage enforcement
task vuln         # govulncheck
```

Install developer tools:

```sh
task tools:install
```

---

## Configuration

The `caveman` binary and the proxy are configured independently.

### `caveman`

| Variable | Description | Default |
|---|---|---|
| `CAVEMAN_BASE_URL` | Base URL for the compressor LLM API | — |
| `CAVEMAN_MODEL` | Model name for the compressor LLM | — |
| `CAVEMAN_API_KEY` | API key for the compressor LLM (optional) | — |
| `CAVEMAN_CONFIG` | Path to a JSON config file; bypasses all individual vars when set | — |

Cache location is not configurable via environment variable. It defaults to `$XDG_CACHE_HOME/caveman/cache.db` (or OS user cache dir, falling back to the system temp dir). It can be set via the `cache_path` field in a JSON config file.

Config precedence: `-config` flag > `CAVEMAN_CONFIG` file > individual `CAVEMAN_*` vars.

### `proxy`

| Variable | Description | Default |
|---|---|---|
| `PROXY_BACKEND_URL` | Base URL of the upstream LLM API (**required**) | — |
| `PROXY_LISTEN_ADDR` | TCP listen address for the HTTP proxy server | `:8080` |
| `PROXY_CAVEMAN_BIN` | Path or name of the `caveman` binary | `caveman` |
| `PROXY_CONFIG` | Path to a JSON config file; bypasses all individual vars when set | — |

Config precedence: `-config` flag > `PROXY_CONFIG` file > individual `PROXY_*` vars.

### `tokcount`

| Variable | Description | Default |
|---|---|---|
| `GOOGLE_API_KEY` | Gemini API key for exact token counts; omit to skip Gemini rows | — |
