## Purpose

Defines how the proxy expands caveman-speak backend responses back to fluent English before returning to the caller. Decompression is delegated to the `caveman` binary subprocess; the expanded form is ephemeral and never persisted.

## Requirements

### Requirement: Backend responses expanded to fluent English via caveman binary
The proxy SHALL invoke `caveman --decode` on the `choices[0].message.content` field of the upstream response and return the modified response to the caller. The binary handles caching internally. The user never sees caveman-speak.

#### Scenario: Caveman response decoded before returning to caller
- **WHEN** the upstream returns a response with caveman-speak content
- **THEN** the proxy invokes `caveman --decode`, replaces the content with the expanded English, and returns the response

#### Scenario: Decompressed form is not persisted to history
- **WHEN** a response has been decoded for the caller
- **THEN** the compressed (caveman) form is what is stored in conversation history; the expanded form is ephemeral
