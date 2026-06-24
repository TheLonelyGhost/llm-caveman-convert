## ADDED Requirements

### Requirement: Conversation history stored in compressed form
The system SHALL store all turns (user and assistant) in their compressed form. The decompressed form of any turn SHALL NOT be persisted.

#### Scenario: User turn stored compressed
- **WHEN** a user message is compressed and sent to the backend
- **THEN** the compressed form of the message is appended to history; the original raw text is not stored in history

#### Scenario: Assistant turn stored compressed
- **WHEN** the backend returns a caveman-speak response
- **THEN** the caveman-speak form is appended to history; the expanded English shown to the user is not stored in history

#### Scenario: History sent to backend is fully compressed
- **WHEN** a new request is made in a multi-turn conversation
- **THEN** all prior turns sent to the backend are in their stored compressed form — no re-compression occurs

### Requirement: Compressed history is byte-stable across turns
The system SHALL never re-compress stored history entries. Once a turn is stored, its compressed bytes are immutable.

#### Scenario: Prior turn bytes unchanged on subsequent calls
- **WHEN** a conversation has N turns and a new (N+1)th turn is sent
- **THEN** the bytes of turns 1 through N in the request are identical to the bytes sent in the Nth call, enabling provider prefix-cache hits

### Requirement: Display layer shows decompressed content
The proxy SHALL never return caveman-speak to the caller in response content. The caller receives fluent English responses. User messages are not echoed back by the proxy — the caller retains the original text they submitted.

#### Scenario: Assistant response returned as expanded English
- **WHEN** the backend response has been decompressed
- **THEN** the proxy returns the expanded English in `choices[0].message.content`; caveman-speak is never present in the proxy's response to the caller
