## Purpose

Handles all HTTP-level concerns for the proxy: listening on the correct endpoint, forwarding client headers to the upstream, parsing OpenAI-compatible request/response bodies, and invoking the `caveman` binary as a subprocess for each message.

## Requirements

### Requirement: Proxy listens on /v1/chat/completions
The proxy SHALL accept requests on the `/v1/chat/completions` endpoint only. Requests to other paths SHALL return an appropriate HTTP error.

#### Scenario: Request to /v1/chat/completions accepted
- **WHEN** a client sends a POST request to `/v1/chat/completions`
- **THEN** the proxy processes the request and forwards it to the backend

#### Scenario: Request to unknown path rejected
- **WHEN** a client sends a request to any path other than `/v1/chat/completions`
- **THEN** the proxy returns an HTTP 404 response

### Requirement: Proxy forwards all client HTTP headers to upstream
The proxy SHALL forward all HTTP headers received from the client to the upstream backend unchanged, with the exception of the `Host` header which SHALL be rewritten to the upstream host.

#### Scenario: Authorization header forwarded
- **WHEN** the client sends an `Authorization` header
- **THEN** the same header and value is present in the request forwarded to the upstream

#### Scenario: Custom headers forwarded
- **WHEN** the client sends any additional headers (e.g., `X-Request-Id`, `OpenAI-Organization`)
- **THEN** those headers are forwarded to the upstream unchanged

#### Scenario: Host header rewritten
- **WHEN** the proxy forwards a request to the upstream
- **THEN** the `Host` header reflects the upstream host, not the proxy's host

### Requirement: Proxy extracts message content from OpenAI-compatible HTTP request bodies
The proxy component SHALL parse incoming HTTP request bodies in OpenAI chat completions format, extract the `content` field of each entry in the `messages` array, and present each content string as a discrete unit for compression.

#### Scenario: User message content extracted
- **WHEN** an HTTP request contains a `messages` array with a user-role entry
- **THEN** the proxy extracts the `content` string from that entry for compression

#### Scenario: System message content extracted for augmentation
- **WHEN** an HTTP request contains a `messages` array with a system-role entry
- **THEN** the proxy extracts the `content` string and passes it to the system-prompt-augmentation component

#### Scenario: Multiple messages extracted in order
- **WHEN** a request contains multiple messages
- **THEN** each message's content is extracted and processed in order; the original array structure is preserved on reassembly

#### Scenario: Non-content fields untouched
- **WHEN** a message entry contains fields other than `content` (e.g., `role`, `name`)
- **THEN** those fields are passed through unchanged in the reassembled request

### Requirement: Proxy invokes caveman binary as subprocess for each message
The proxy SHALL invoke `caveman --encode` as a subprocess for each extracted user message content, passing the content via stdin and reading the result from stdout. On failure (non-zero exit), the proxy SHALL use the original content verbatim.

#### Scenario: Successful encode subprocess invocation
- **WHEN** the proxy invokes `caveman --encode` with message content on stdin
- **THEN** the stdout result replaces the original content in the reassembled request

#### Scenario: Binary exits non-zero — fallback to original
- **WHEN** `caveman --encode` exits non-zero
- **THEN** the proxy uses the original message content verbatim; no error is surfaced to the caller

#### Scenario: Binary not found — proxy fails at startup
- **WHEN** the `caveman` binary is not present or not executable at the configured path
- **THEN** the proxy fails to start with a clear error message

### Requirement: Proxy extracts and decodes content from OpenAI-compatible HTTP response bodies
The proxy SHALL wait for the complete HTTP response from the upstream backend, extract the `content` field from `choices[0].message.content`, invoke `caveman --decode` as a subprocess, replace the content with the decoded result, and return the modified response to the caller.

#### Scenario: Assistant response content decoded
- **WHEN** the upstream returns a complete HTTP response with an assistant message containing caveman-speak
- **THEN** the proxy invokes `caveman --decode`, replaces the content with the decoded result, and returns the modified response

#### Scenario: Decode failure — fallback to caveman content
- **WHEN** `caveman --decode` exits non-zero
- **THEN** the proxy returns the response with the original caveman-speak content; no error is surfaced to the caller
