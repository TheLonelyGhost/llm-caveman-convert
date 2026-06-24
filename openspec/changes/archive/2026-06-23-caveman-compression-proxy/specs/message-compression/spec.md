## ADDED Requirements

### Requirement: Whole message compressed via caveman binary
The proxy SHALL compress each message as a single unit by invoking `caveman --encode` as a subprocess. The entire message content is passed via stdin; the compressed result is read from stdout. The binary handles caching internally.

#### Scenario: Full message passed to binary
- **WHEN** a message is ready for compression
- **THEN** the complete message content string is written to `caveman --encode` stdin; no pre-processing or segmentation is performed by the proxy

#### Scenario: Binary failure falls back to original
- **WHEN** `caveman --encode` exits non-zero or produces no output
- **THEN** the original message content is used verbatim; no error is surfaced to the user
