## ADDED Requirements

### Requirement: Decompression results cached by input hash inside the binary
The `caveman` binary SHALL maintain a persistent decompression cache mapping `hash(compressed_response_text)` to `expanded_response_text`. The cache SHALL be checked before any decompressor LLM call and written on success.

#### Scenario: Cache hit skips decompressor LLM
- **WHEN** a compressed response's hash exists in the decompression cache
- **THEN** the cached expanded text is returned immediately and no decompressor LLM call is made

#### Scenario: Cache miss populates cache after decompression
- **WHEN** a compressed response's hash is not in the decompression cache and decompression succeeds
- **THEN** the result is stored in the cache before being written to stdout

#### Scenario: Compression and decompression caches are independent
- **WHEN** the same text appears as input to both `--encode` and `--decode`
- **THEN** lookups in each cache namespace are independent and do not interfere

### Requirement: Cache location is configurable and auto-created
The binary SHALL accept a configurable path for the decompression cache. If the path or any of its parent directories do not exist at the time of first write, the binary SHALL create them automatically.

#### Scenario: Custom cache path used when configured
- **WHEN** a non-default decompression cache path is specified in configuration
- **THEN** the binary reads and writes the decompression cache at that path

#### Scenario: Parent directories created on first write
- **WHEN** the configured decompression cache path does not exist
- **THEN** the binary creates all necessary parent directories before writing the first cache entry

### Requirement: Decompression cache key is the compressed form
The cache key SHALL be a deterministic hash of the compressed input text, not the expanded form, ensuring the cache is keyed on exactly what was passed to `caveman --decode`.

#### Scenario: Same caveman input always hits decompression cache
- **WHEN** identical caveman-speak text is passed to `caveman --decode` on two separate invocations
- **THEN** both invocations return the same cached expanded English form after the first
