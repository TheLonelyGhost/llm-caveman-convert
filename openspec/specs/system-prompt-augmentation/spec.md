## Purpose

Governs how the proxy modifies the system prompt before forwarding to the backend LLM. A fixed caveman-response instruction is appended to direct the backend to respond in caveman-speak with exact preservation rules for identifiers, code, numbers, and URLs.

## Requirements

### Requirement: System prompt augmented with caveman-response instruction
Before sending to the backend LLM, the system SHALL append a fixed caveman-response instruction to the user's system prompt. If no system prompt exists, the instruction alone is used as the system prompt.

#### Scenario: Existing system prompt augmented
- **WHEN** the user has provided a system prompt
- **THEN** the caveman-response instruction is appended to it before sending to the backend; the user's original system prompt is not altered

#### Scenario: No system prompt — instruction used alone
- **WHEN** no system prompt exists
- **THEN** the caveman-response instruction is sent as the system prompt

#### Scenario: Augmentation instruction is fixed
- **WHEN** the system prompt augmentation is applied
- **THEN** the same fixed instruction text is appended on every call; because the instruction is constant, the binary's cache will hit after the first invocation

### Requirement: Caveman-response instruction specifies exact preservation rules
The augmentation instruction SHALL direct the backend to omit articles, filler, and pleasantries while keeping identifiers, code, numbers, and URLs verbatim.

#### Scenario: Backend preserves identifiers in response
- **WHEN** the backend responds to a prompt about a function named `calculate_discount`
- **THEN** the response contains `calculate_discount` verbatim, not paraphrased
