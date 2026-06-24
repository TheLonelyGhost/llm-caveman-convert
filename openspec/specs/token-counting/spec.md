# Token Counting

## Purpose

Provides token counting for all models in the registry via the `tokcount` CLI and a reusable `internal/tokens` package.

## Requirements

### Requirement: Count tokens from stdin
The system SHALL read raw text from stdin and count tokens for each model in the registry, writing a human-readable report to stdout.

#### Scenario: Single file count
- **WHEN** text is piped to `tokcount` via stdin
- **THEN** the tool prints a table with one row per model showing provider, model name, token count, and approximation note

#### Scenario: Empty input
- **WHEN** stdin is empty
- **THEN** the tool prints a table with all models showing 0 tokens

### Requirement: Local BPE tokenization for OpenAI models
The system SHALL count tokens for OpenAI models using the `o200k_base` tiktoken encoding locally, with no API call required.

#### Scenario: OpenAI model count
- **WHEN** text is provided
- **THEN** `gpt-5.5`, `gpt-5.4`, and `gpt-5.4-mini` each show a token count with no approximation marker

### Requirement: Approximate tokenization for Anthropic models
The system SHALL count tokens for Anthropic models using the `cl100k_base` tiktoken encoding as an approximation, and SHALL mark the result as approximate.

#### Scenario: Anthropic approximate count
- **WHEN** text is provided
- **THEN** `claude-opus-4-8`, `claude-sonnet-4-6`, and `claude-haiku-4-5` each show a token count marked `~`

### Requirement: Approximate tokenization for xAI Grok models
The system SHALL count tokens for xAI Grok models using the `o200k_base` tiktoken encoding as an approximation, and SHALL mark the result as approximate.

#### Scenario: Grok approximate count
- **WHEN** text is provided
- **THEN** `grok-4.3`, `grok-4.20-0309`, and `grok-4.20-multi-agent-0309` each show a token count marked `~`

### Requirement: API-based tokenization for Google Gemini models
The system SHALL count tokens for Google Gemini models via the Google Gen AI API when `GOOGLE_API_KEY` is set, and SHALL silently emit a placeholder row when the key is absent.

#### Scenario: Gemini count with key present
- **WHEN** `GOOGLE_API_KEY` is set and text is provided
- **THEN** `gemini-2.5-pro`, `gemini-2.5-flash`, and `gemini-2.0-flash` each show an exact token count with no approximation marker

#### Scenario: Gemini count with key absent
- **WHEN** `GOOGLE_API_KEY` is not set
- **THEN** each Gemini model row displays `(requires GOOGLE_API_KEY)` in place of a count

### Requirement: Approximate tokenization for Meta Llama models with margin warning
The system SHALL count tokens for Meta Llama models using BPE approximation and SHALL mark results with an explicit margin-of-error label.

#### Scenario: Llama 4 approximate count
- **WHEN** text is provided
- **THEN** `llama-4-scout` and `llama-4-maverick` each show a token count marked `~ (±15%)`

#### Scenario: Llama 3.3 approximate count
- **WHEN** text is provided
- **THEN** `llama-3.3-70b` shows a token count marked `~ (±10%)`

### Requirement: Aligned table output
The system SHALL format output as a left-aligned table with fixed-width columns for provider, model, tokens, and note, written to stdout.

#### Scenario: Consistent column alignment
- **WHEN** the report is printed
- **THEN** provider and model columns are left-aligned and padded so all rows line up vertically

#### Scenario: Thousands separator in token counts
- **WHEN** a token count is 1000 or greater
- **THEN** the count is formatted with comma thousands separators (e.g., `2,920`)

### Requirement: Counter interface reusability
The `internal/tokens` package SHALL expose a `Counter` interface that other packages in the module can import and use independently of the CLI.

#### Scenario: Interface contract
- **WHEN** a caller holds a value satisfying `Counter`
- **THEN** it can call `Count(ctx, text)` and receive `(int, error)` without knowing the underlying strategy

### Requirement: Tiktoken BPE encoder initialized once per process
The system SHALL initialize each tiktoken encoder at startup (not per-call) to avoid repeated BPE vocabulary loading.

#### Scenario: Repeated counts
- **WHEN** the registry is iterated to count tokens for multiple models sharing an encoding
- **THEN** the BPE vocabulary file is loaded only once per encoding across the process lifetime
