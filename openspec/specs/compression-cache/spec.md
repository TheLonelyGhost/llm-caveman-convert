## Purpose

Persistent cache for compression (encode) results inside the `caveman` binary. Maps a deterministic hash of raw input text to its compressed caveman-speak output, avoiding redundant compressor LLM calls.

## Requirements

### Requirement: Compression results cached by input hash inside the binary
The `caveman` binary SHALL maintain a persistent compression cache mapping `hash(raw_message_text)` to `compressed_message_text`. The cache SHALL be checked before any compressor LLM call. The cache SHALL only be written after the compressed output passes structural validation; intermediate retry attempts that fail validation SHALL NOT be written to the cache.

#### Scenario: Cache hit skips compressor LLM
- **WHEN** a message's hash exists in the compression cache
- **THEN** the cached compressed text is returned immediately and no compressor LLM call is made

#### Scenario: Cache miss populates cache after compression and validation pass
- **WHEN** a message's hash is not in the compression cache and compression succeeds and validation passes
- **THEN** the result is stored in the cache before being written to stdout

#### Scenario: Failed compression not cached
- **WHEN** the compressor LLM fails
- **THEN** no entry is written to the cache; the binary exits non-zero

#### Scenario: Validation failure on first attempt triggers retry without caching
- **WHEN** the first LLM response fails structural validation
- **THEN** a fix LLM call is made; the failed intermediate result is not written to the cache

#### Scenario: Validation failure after all retries exhausted — best-effort result returned uncached
- **WHEN** all retry attempts are exhausted and the final result still fails validation
- **THEN** the last LLM result is written to stdout; no cache entry is written; the binary exits zero

### Requirement: Cache location is configurable and auto-created
The binary SHALL accept a configurable path for the compression cache. If the path or any of its parent directories do not exist at the time of first write, the binary SHALL create them automatically.

#### Scenario: Custom cache path used when configured
- **WHEN** a non-default compression cache path is specified in configuration
- **THEN** the binary reads and writes the compression cache at that path

#### Scenario: Parent directories created on first write
- **WHEN** the configured compression cache path does not exist
- **THEN** the binary creates all necessary parent directories before writing the first cache entry

### Requirement: Cache keys are stable across invocations
The binary SHALL use a deterministic hash (e.g., SHA-256) of the exact input text as the cache key, ensuring identical input always produces the same cache key across separate process executions.

#### Scenario: Same message text always hits cache
- **WHEN** the same raw message text is passed to `caveman --encode` on two separate invocations
- **THEN** both invocations return the same cached compressed form after the first
