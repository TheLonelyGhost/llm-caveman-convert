## ADDED Requirements

### Requirement: Binary accepts --encode flag to compress stdin to caveman-speak
The `caveman` binary SHALL accept an `--encode` flag. When invoked with `--encode`, it SHALL read plain English text from stdin, compress it to caveman-speak using the compressor LLM, and write the result to stdout. The binary SHALL exit zero on success.

#### Scenario: Encode plain English
- **WHEN** `caveman --encode` is invoked with plain English text on stdin
- **THEN** caveman-speak is written to stdout and the process exits zero

#### Scenario: Encode preserves identifiers and code references
- **WHEN** input contains identifiers, inline code, numbers, or URLs
- **THEN** those tokens appear verbatim in the encoded output

#### Scenario: Encode handles code fences per compressor LLM instruction
- **WHEN** input contains a fenced code block with a recognized full language tag (e.g., ` ```python `)
- **THEN** comment lines within the block are compressed; code logic lines are unchanged

#### Scenario: Encode passes through untagged code blocks
- **WHEN** input contains a fenced code block with no language tag or a shorthand tag
- **THEN** the block content appears verbatim in the encoded output

#### Scenario: Encode cache hit — returns cached result without LLM call
- **WHEN** `caveman --encode` is invoked and the input hash exists in the cache
- **THEN** the cached compressed text is written to stdout; no LLM call is made

#### Scenario: Encode cache miss — result stored after successful LLM call
- **WHEN** `caveman --encode` is invoked, no cache entry exists, and the LLM call succeeds
- **THEN** the result is written to stdout and stored in the cache before the process exits

#### Scenario: Encode fails — exits non-zero and writes nothing to stdout
- **WHEN** the compressor LLM call fails
- **THEN** the binary exits non-zero and writes no output to stdout; no cache entry is written; the caller is responsible for fallback

### Requirement: Binary accepts --decode flag to expand caveman-speak to plain English
The `caveman` binary SHALL accept a `--decode` flag. When invoked with `--decode`, it SHALL read caveman-speak text from stdin, expand it to fluent plain English using the compressor LLM, and write the result to stdout. The binary SHALL exit zero on success.

#### Scenario: Decode caveman-speak
- **WHEN** `caveman --decode` is invoked with caveman-speak on stdin
- **THEN** fluent plain English is written to stdout and the process exits zero

#### Scenario: Decode cache hit — returns cached result without LLM call
- **WHEN** `caveman --decode` is invoked and the input hash exists in the cache
- **THEN** the cached expanded text is written to stdout; no LLM call is made

#### Scenario: Decode cache miss — result stored after successful LLM call
- **WHEN** `caveman --decode` is invoked, no cache entry exists, and the LLM call succeeds
- **THEN** the result is written to stdout and stored in the cache before the process exits

#### Scenario: Decode fails — exits non-zero and writes nothing to stdout
- **WHEN** the decompressor LLM call fails
- **THEN** the binary exits non-zero and writes no output to stdout; no cache entry is written

### Requirement: Binary maintains a persistent cache for encode and decode results
The `caveman` binary SHALL maintain a persistent cache mapping `hash(input_text)` to output text, shared across `--encode` and `--decode` invocations. The cache SHALL be checked before any LLM call. Encode and decode results SHALL be stored in independent namespaces to avoid key collisions. The cache location SHALL be configurable; if the configured path does not exist, the binary SHALL create all necessary parent directories before writing.

#### Scenario: Cache persists across invocations
- **WHEN** `caveman --encode` is invoked with identical stdin on two separate process executions
- **THEN** the second invocation returns the cached result without making an LLM call

#### Scenario: Encode and decode caches are independent
- **WHEN** the same text appears as input to both `--encode` and `--decode`
- **THEN** lookups in each namespace are independent and do not interfere

#### Scenario: Cache directory created automatically
- **WHEN** the configured cache location does not exist
- **THEN** the binary creates all necessary parent directories before writing the first cache entry

#### Scenario: Custom cache location used when configured
- **WHEN** a non-default cache path is specified in configuration
- **THEN** the binary reads and writes the cache at that path

### Requirement: Binary LLM configuration is independent of the proxy
The `caveman` binary SHALL be configured with its own compressor LLM endpoint and model, separate from any backend LLM configuration used by the proxy.

#### Scenario: Binary uses its own configured model
- **WHEN** the binary is invoked
- **THEN** LLM calls are routed to the model specified in the binary's configuration, not the backend model
