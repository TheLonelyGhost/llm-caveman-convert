## 1. Project Setup

- [x] 1.1 Initialize project structure: two entry points — `caveman` binary and proxy component
- [x] 1.2 Add compressor LLM client dependency (for `caveman` binary)
- [x] 1.3 Add backend LLM client dependency (OpenAI-compatible, for proxy)
- [x] 1.4 Define configuration schema: binary (compressor model/endpoint, cache location); proxy (backend model, binary path)
- [x] 1.5 Set up test framework and write first passing test

## 2. caveman Binary — Encode

- [x] 2.1 Implement compression cache: persistent key-value store keyed on SHA-256(input_text), encode namespace
- [x] 2.2 Implement `--encode` flag: check cache first; on miss, call compressor LLM with caveman-speak system prompt (code fence rules, prose rules), write stdout, store in cache
- [x] 2.3 Implement exit-zero on success, exit-non-zero on LLM failure with no stdout output and no cache write
- [x] 2.4 Write tests: cache hit skips LLM, cache miss populates cache, encodes prose, preserves identifiers, tagged code block comments compressed, untagged block unchanged, LLM failure exits non-zero

## 3. caveman Binary — Decode

- [x] 3.1 Implement decompression cache: persistent key-value store keyed on SHA-256(input_text), decode namespace
- [x] 3.2 Implement `--decode` flag: check cache first; on miss, call compressor LLM with expand-to-English system prompt, write stdout, store in cache
- [x] 3.3 Implement exit-zero on success, exit-non-zero on LLM failure with no stdout output and no cache write
- [x] 3.4 Write tests: cache hit skips LLM, cache miss populates cache, decodes caveman-speak to fluent English, encode and decode caches independent, LLM failure exits non-zero

## 4. caveman Binary — Configuration

- [x] 4.1 Implement binary configuration loading (compressor model, endpoint)
- [x] 4.2 Implement configurable cache location in binary configuration (default path + override)
- [x] 4.3 Implement auto-creation of parent directories on first cache write
- [x] 4.4 Write tests: binary routes LLM calls to configured model; cache written to configured path; missing parent dirs created automatically

## 5. HTTP Message Extraction

- [x] 5.1 Implement `/v1/chat/completions` route; return 404 for all other paths
- [x] 5.2 Implement client header forwarding: copy all request headers to upstream request; rewrite `Host` to upstream host
- [x] 5.3 Implement OpenAI-compatible request body parser: extract `messages[].content` per entry
- [x] 5.4 Implement subprocess invocation: pipe content to `caveman --encode`, read stdout result
- [x] 5.5 Implement fallback: on non-zero exit from binary, use original content verbatim
- [x] 5.6 Implement request reassembly: replace content fields, preserve all other fields unchanged
- [x] 5.7 Implement proxy startup validation: fail fast if `caveman` binary not found or not executable
- [x] 5.8 Implement response body parser: extract `choices[0].message.content` from complete upstream response
- [x] 5.9 Implement subprocess invocation for decode: pipe response content to `caveman --decode`, read stdout result
- [x] 5.10 Implement decode failure fallback: on non-zero exit, return response with original caveman content unchanged
- [x] 5.11 Write tests: route 404 on unknown path; headers forwarded, Host rewritten; content extracted and re-encoded; non-content fields unchanged; encode binary failure fallback; response content decoded; decode binary failure returns caveman content

## 6. System Prompt Augmentation

- [x] 6.1 Implement augmentation: append caveman-response instruction to extracted system message content before encode
- [x] 6.2 Handle no-system-message case: inject system message with instruction only
- [x] 6.3 Write tests: augmented system content contains instruction, original content preserved, no-system-message case

## 7. Conversation History

- [x] 7.1 Implement history store: append compressed user turn after encode
- [x] 7.2 Implement history store: append compressed assistant turn (caveman form, not decoded)
- [x] 7.3 Ensure history entries are immutable after storage (no re-encoding)
- [x] 7.4 Implement history serialization for backend request: send stored compressed bytes unchanged
- [x] 7.5 Write tests: history contains compressed forms, bytes stable across turns, display shows original/decoded

## 8. Integration

- [x] 8.1 Wire full request pipeline: extract messages → invoke `caveman --encode` → augment system prompt → reassemble → forward with client headers to backend
- [x] 8.2 Wire full response pipeline: receive complete response → invoke `caveman --decode` → return to caller → store compressed to history
- [x] 8.3 Write end-to-end test: full 2-turn conversation; verify backend never sees raw English, caller never sees caveman-speak
- [x] 8.4 Validate token savings: measure token counts on representative prompts against baseline
