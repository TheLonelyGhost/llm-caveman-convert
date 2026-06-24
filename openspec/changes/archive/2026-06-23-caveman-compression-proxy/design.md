## Context

LLM API costs scale directly with token usage. English prose is token-inefficient — articles, filler, hedging, and pleasantries consume tokens without adding semantic value. Caveman-speak (stripping these) yields 25–60% token savings on prose segments, validated empirically across cl100k and o200k tokenizers. Savings compound over multi-turn conversations as context history grows.

This design covers a transparent compression proxy: user writes normal English, sees normal English, but the backend LLM only ever sees compressed text. The system has two components: a standalone `caveman` binary (owns LLM interaction, stdin/stdout interface) and a proxy component (owns HTTP parsing, cache, and subprocess invocation).

Constraints from exploration:
- Provider prefix-cache requires byte-identical history across turns → history must be stored compressed, never re-compressed
- Wenyan compression is tokenizer-dependent (breaks on cl100k); caveman-speak is tokenizer-agnostic → caveman is the primary strategy
- Historical messages are unlikely to change once stored; re-compression instability is not a practical concern for history
- System prompt augmentation overhead (~29 tokens) breaks even at ~29 tokens of conversation content; all realistic usage is profitable

## Goals / Non-Goals

**Goals:**
- Transparent compression: user sees only normal English at input and output
- Whole-message compression: entire message sent to `caveman` binary as a unit
- Standalone binary: `caveman --encode` / `caveman --decode` readable from any tool via stdin/stdout; independently testable
- Clean component boundary: proxy handles HTTP and caching; binary handles LLM interaction
- Cached: compression and decompression results persisted by message hash in the binary to avoid redundant LLM calls
- Conversation-history efficient: history stored compressed; decompressed only for display
- Provider prefix-cache compatible: history bytes stable across turns
- Configurable: compressor LLM configured in the binary, independently of the backend LLM

**Non-Goals:**
- Wenyan or other language-based compression (tokenizer-dependent, not universally safe)
- Lossless compression (caveman is lossy on style, not on semantics — acceptable trade-off)
- Client-side structural parsing of messages before compression (code fence handling is the compressor LLM's responsibility, expressed in binary's system prompt)
- Real-time streaming / SSE (not supported; request/response only)
- Cache inside the proxy (cache is the binary's responsibility)
- Multi-user shared compression cache (cache is per-binary-deployment)

## Decisions

### Decision: Caveman-speak as sole compression strategy

**Chosen**: Caveman-speak only (drop articles, filler, hedging; keep identifiers, code, numbers exact).

**Alternatives considered**:
- Wenyan: ~23% savings on o200k but costs more tokens on cl100k. Tokenizer-dependent, unpredictable across providers. Rejected.
- Abbreviation shorthand: ~10–20% savings, rule-based, no LLM needed. Too shallow; doesn't handle complex prose.
- Combined: better ratio but compounding complexity and failure modes.

**Rationale**: Caveman-speak is tokenizer-agnostic, yields 25–60% savings, and is reversible by a small LLM decompressor with high reliability.

---

### Decision: Whole-message compression unit

**Chosen**: Each message is compressed as a single unit by the compressor LLM. No structural parsing or segmentation is performed by the system before passing the message to the compressor.

**Alternatives considered**:
- Segment-level compression: splits message into prose/code/url segments, compresses only prose. More granular cache hits but adds structural parsing complexity and splits messages the LLM would handle better as a whole.
- Paragraph-level: natural boundary but harder to detect reliably; still requires parsing.

**Rationale**: Historical messages are unlikely to change once stored, so cache instability from whole-message keying is not a practical problem. Whole-message compression is simpler, lets the LLM reason about context across the full message, and avoids a parsing layer that can misclassify edge cases. Cache key is `hash(whole_message)`.

---

### Decision: Compressor/decompressor exposed as standalone binary with stdin/stdout interface

**Chosen**: A standalone `caveman` binary with two modes: `caveman --encode` (plain English → caveman-speak) and `caveman --decode` (caveman-speak → plain English). Reads from stdin, writes to stdout. Owns compressor LLM configuration and persistent cache.

**Alternatives considered**:
- Library/module: tighter coupling between proxy and LLM logic; harder to test the compression in isolation; not composable with non-proxy callers.
- Embedded in proxy: same issues as library; also harder to replace the compressor strategy without touching the proxy.
- HTTP microservice: more composable than embedded but adds network hop, service lifecycle, and deployment complexity for what is fundamentally a text transform.

**Rationale**: A stdin/stdout binary is the Unix primitive for composable text transformation. It is independently testable (`echo "text" | caveman --encode`), replaceable (swap the binary without changing the proxy), and requires no network or service management. The proxy invokes it as a subprocess, keeping the component boundary clean.

---

### Decision: Cache lives in the binary, not the proxy

**Chosen**: The `caveman` binary maintains its own persistent cache for both encode and decode results, keyed on `hash(input_text)` in independent namespaces. The cache location is configurable; if the configured path does not exist, the binary creates all necessary parent directories on first write. The proxy invokes the binary unconditionally; cache hits are transparent to the proxy.

**Alternatives considered**:
- Cache in the proxy (check before subprocess invocation): proxy avoids process-spawn cost on hit, but now owns cache lifecycle, eviction policy, and storage — concerns that belong with the component that owns the LLM interaction.
- Shared external cache: adds infrastructure dependency; overkill for a single binary.

**Rationale**: The binary owns the LLM interaction and knows best what is safe to cache (successful results only, never failures). Keeping the cache in the binary preserves a clean interface: proxy always invokes `caveman`, binary always returns the right answer — cached or fresh. The proxy stays simple and stateless with respect to compression.

---

### Decision: Proxy handles /v1/chat/completions only, request/response only (no streaming)

**Chosen**: The proxy exposes a single endpoint (`/v1/chat/completions`). It waits for the complete upstream HTTP response, decodes `choices[0].message.content` via `caveman --decode`, and returns the modified response. SSE/streaming is not supported. All client headers are forwarded to upstream; `Host` is rewritten to the upstream host.

**Alternatives considered**:
- SSE streaming: accumulate `delta.content` fragments, decode at sentence boundary, re-emit as `data:` events. Adds significant complexity — SSE parser, accumulation buffer, sentence-boundary detection, re-serialization of events. Out of scope.
- Pass SSE through raw without decoding: client would see caveman-speak tokens. Unacceptable.

**Rationale**: Request/response is sufficient for a first implementation. The proxy's value is in token savings, not streaming latency. Streaming can be added later without changing the compression contract.

---

### Decision: Proxy component handles HTTP extraction and subprocess invocation

**Chosen**: The proxy parses OpenAI-compatible HTTP request/response bodies, extracts message content fields, invokes `caveman --encode` or `caveman --decode` as a subprocess per message, and reassembles the body.

**Alternatives considered**:
- Generic middleware approach: parse at a higher level (e.g., entire request body as text). Too coarse — would corrupt JSON structure.
- Rewrite at transport level: intercept raw bytes. Requires TLS termination, more complex.

**Rationale**: OpenAI-compatible request format is well-specified. Extracting `messages[].content` fields is targeted and safe. Reassembly preserves all other fields unchanged. The proxy is the right place to own the HTTP concern; the binary has no knowledge of HTTP.

---

### Decision: Code fence handling is a compressor LLM instruction concern

**Chosen**: The compressor LLM system prompt instructs it to handle code fences appropriately: preserve code logic verbatim, compress comments inside explicitly-tagged blocks (e.g., ` ```python `), and pass opaque or untagged blocks through unchanged. No regex-based segmentation or language allowlist is maintained by the system.

**Alternatives considered**:
- System-side language allowlist + regex comment extraction: deterministic but brittle; requires maintaining COMMENT_LANGS/OPAQUE_LANGS lists; misclassifies shorthand tags and unknown languages.
- Full code block pass-through regardless: safe but leaves comment token savings on the table.

**Rationale**: The compressor LLM can recognize code fence language tags and comment syntax more reliably than a regex heuristic. Delegating this to the LLM instruction eliminates a structural parsing layer while preserving the same safety properties: code logic untouched, comments compressed where safe, opaque blocks unchanged.

---

### Decision: System prompt augmented with caveman-response instruction

**Chosen**: Append a fixed instruction to the user's system prompt before sending to backend:
> "Respond in compressed caveman-speak: omit articles, filler, pleasantries. Fragments ok. Keep identifiers, code, numbers exact."

**Alternatives considered**:
- Compress only input, accept normal response: simpler but history grows at full response size, eliminating compounding savings.
- Per-request instruction in user turn: costs tokens on every turn; less reliable than system prompt.

**Rationale**: Backend responding in caveman-speak is essential for history compression to compound. The 29-token overhead breaks even at 29 tokens of conversation — negligible.

---

### Decision: History stored compressed, decompressed only for display

**Chosen**: After decompression for display, the compressed form is what gets appended to history. Decompressed text is ephemeral — never persisted.

**Alternatives considered**:
- Store decompressed, compress on send: re-compressing non-deterministic output would break provider prefix-cache; costs compressor LLM on every call.
- Store both forms: doubles storage; complex to keep in sync.

**Rationale**: Compressed history is byte-stable across turns. Provider prefix-cache hits require exact bytes. Storing only compressed form is the only design that satisfies both constraints simultaneously.

---

### Decision: Two independent caches

**Chosen**:
- **Compression cache**: `hash(raw_message)` → `compressed_message`; managed by the binary; checked before encode LLM call
- **Decompression cache**: `hash(compressed_response)` → `expanded_response`; managed by the binary; checked before decode LLM call

**Rationale**: Independent namespaces avoid key collisions. Both caches live inside the binary alongside the LLM calls they gate. The proxy has no cache responsibility.

---

### Decision: Comment compression failure is silent whole-message fallback

**Chosen**: If the `caveman --encode` binary exits non-zero, produces empty output, or produces output longer than the input by a significant margin, the proxy uses the original message verbatim. No error surfaced to user.

**Rationale**: Compression is best-effort. An uncompressed message is semantically identical to the original. Surfacing an error would be confusing — the user's prompt was not corrupted.

## Risks / Trade-offs

**Binary subprocess latency on cache miss** → Mitigation: process-spawn cost is low (~ms); LLM call dominates latency; cache warms quickly for repeated messages.

**Caveman response degrades on highly technical backend responses** → Mitigation: `caveman --decode` has full context to reconstruct; tested on technical content (JWT, auth, architecture). Acceptable loss.

**Non-deterministic binary output breaks message cache on cold start** → Mitigation: cache is checked before binary is invoked. Once cached, output is deterministic. First call per message is the only non-deterministic moment.

**System prompt augmentation visible to backend** → Accepted. Backend needs this instruction to respond in caveman. Not a confidentiality concern for the compression system itself.

**Binary not found or not executable** → Mitigation: proxy validates binary presence at startup; fails fast with a clear error rather than silently passing uncompressed messages.

**Compressor LLM may mishandle unknown code fence languages** → Accepted. Binary's compressor instruction is best-effort on untagged or exotic blocks; proxy fallback to original message preserves correctness.

## Open Questions

- **Binary system prompt**: Exact LLM system prompt inside `caveman` binary is unspecified. Needs tuning — must instruct correct code fence handling, caveman prose rules, and appropriate behavior for untagged blocks.
- **Cache eviction policy**: TTL? LRU? Max size? Depends on deployment context. Should be configurable in the binary.
- **Compressor fallback threshold**: What constitutes a "bad" encode result? Length ratio? Exit code only? Threshold unspecified.
- **Binary distribution**: How is the `caveman` binary delivered alongside the proxy? Same package, separate install, or PATH-resolved? Unspecified.
