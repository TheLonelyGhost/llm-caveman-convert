## Why

LLM API costs scale with token usage. Compressing prompts and responses using caveman-speak (dropping articles, filler, hedging) yields 25–60% token savings on prose, without semantic loss. The savings compound over multi-turn conversations as history grows. The user should experience no compression — input and output appear as normal English.

## What Changes

- New standalone binary (`caveman`) that reads text from stdin and writes to stdout: `--encode` compresses to caveman-speak, `--decode` expands to plain English; the binary owns the compressor LLM configuration and is the sole interface for compression/decompression
- New proxy component that listens on `/v1/chat/completions`, forwards all client HTTP headers to the upstream backend (rewriting `Host`), intercepts OpenAI-compatible request bodies to extract and compress message text via the `caveman` binary, waits for the complete upstream response, and decodes `choices[0].message.content` through `caveman --decode` before returning to the caller
- System prompt is augmented to instruct the backend to respond in caveman-speak
- Backend responses (caveman) are passed through the `caveman --decode` binary before display
- Conversation history is stored in compressed form for token efficiency and provider prefix-cache stability
- Compression and decompression results are cached inside the `caveman` binary (hash of input) to avoid redundant LLM calls and preserve exact bytes for provider caching

## Capabilities

### New Capabilities

- `caveman-cli`: Standalone binary with `--encode` (plain English → caveman-speak) and `--decode` (caveman-speak → plain English) flags; reads from stdin, writes to stdout; owns the compressor LLM configuration and persistent cache
- `http-message-extraction`: Proxy component that parses OpenAI-compatible HTTP request/response bodies, extracts message text per turn, invokes `caveman --encode` or `caveman --decode` as a subprocess, and reassembles the body
- `message-compression`: Whole-message compression contract between the proxy and the `caveman` binary; cache by hash of full message; fallback to original on binary failure
- `response-decompression`: Expand caveman-speak LLM responses to fluent English via `caveman --decode`; cache by hash of compressed input
- `compression-cache`: Persistent cache inside the `caveman` binary mapping `hash(original_message)` → `compressed_message`; checked before every encode LLM call; written on success only
- `decompression-cache`: Persistent cache inside the `caveman` binary mapping `hash(compressed_response)` → `expanded_response`; checked before every decode LLM call; written on success only
- `system-prompt-augmentation`: Append caveman-response instruction to user's system prompt before sending to backend
- `conversation-history`: Store all turns in compressed form; decompress only for display; never re-compress stored history

### Modified Capabilities

## Impact

- New dependency: small/cheap LLM endpoint for compression and decompression (configured in the `caveman` binary, not the proxy)
- `caveman` binary is independently testable and composable — any tool that can pipe text can use it; binary manages its own cache transparently
- Proxy component invokes `caveman` as a subprocess; adds one process-spawn latency on cache miss
- Provider prefix-cache compatibility requires history bytes to be stable across turns — achieved by storing compressed form only
- No changes to backend LLM API contract; proxy is transparent to the backend
- Proxy waits for full upstream response before decoding and returning; SSE/streaming not supported
