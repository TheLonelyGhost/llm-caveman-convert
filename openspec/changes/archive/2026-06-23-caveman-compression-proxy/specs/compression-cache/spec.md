## ADDED Requirements

### Requirement: Compression results cached by input hash inside the binary
The `caveman` binary SHALL maintain a persistent compression cache mapping `hash(raw_message_text)` to `compressed_message_text`. The cache SHALL be checked before any compressor LLM call and written on success.

#### Scenario: Cache hit skips compressor LLM
- **WHEN** a message's hash exists in the compression cache
- **THEN** the cached compressed text is returned immediately and no compressor LLM call is made

#### Scenario: Cache miss populates cache after compression
- **WHEN** a message's hash is not in the compression cache and compression succeeds
- **THEN** the result is stored in the cache before being written to stdout

#### Scenario: Failed compression not cached
- **WHEN** the compressor LLM fails
- **THEN** no entry is written to the cache; the binary exits non-zero

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
