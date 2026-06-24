## Purpose

Structural validation utilities for markdown text within the `internal/compress` package. Provides extraction functions for headings, fenced code blocks, and URLs, plus a `Validate` function that checks whether a compressed output preserves the structural invariants of its original.

## Requirements

### Requirement: Extract headings from markdown text
The `internal/compress` package SHALL expose an `ExtractHeadings(text string) []string` function that returns all ATX heading lines (lines matching `^#{1,6}\s+`) in document order.

#### Scenario: Returns all headings in order
- **WHEN** text contains multiple ATX headings at various levels
- **THEN** all heading lines are returned in document order

#### Scenario: Returns empty slice for text with no headings
- **WHEN** text contains no ATX heading lines
- **THEN** an empty (non-nil) slice is returned

### Requirement: Extract fenced code blocks from markdown text
The `internal/compress` package SHALL expose an `ExtractCodeBlocks(text string) []string` function that returns all fenced code block contents (including fence lines) using CommonMark rules: the closing fence must use the same character and be at least as long as the opening fence; unclosed fences are ignored.

#### Scenario: Returns exact block content including fence lines
- **WHEN** text contains a properly closed fenced code block
- **THEN** the returned string includes the opening fence line, all interior lines, and the closing fence line exactly as written

#### Scenario: Ignores unclosed fences
- **WHEN** text contains an opening fence with no matching closing fence
- **THEN** no block is returned for that unclosed fence

#### Scenario: Backtick and tilde fences are both recognised
- **WHEN** text contains both ` ``` ` and `~~~` fenced blocks
- **THEN** both blocks are extracted

#### Scenario: Closing fence must use same character as opening
- **WHEN** an opening ` ``` ` fence is followed by a `~~~` line
- **THEN** the `~~~` line is not treated as a closing fence

### Requirement: Extract URLs from markdown text
The `internal/compress` package SHALL expose an `ExtractURLs(text string) []string` function that returns all `http://` and `https://` URLs found in the text.

#### Scenario: Returns all http and https URLs
- **WHEN** text contains multiple http and https URLs
- **THEN** all URLs are returned

#### Scenario: Returns empty slice when no URLs present
- **WHEN** text contains no http or https URLs
- **THEN** an empty (non-nil) slice is returned

### Requirement: Validate compressed markdown preserves structural invariants
The `internal/compress` package SHALL expose a `Validate(original, compressed string) ValidationResult` function. `ValidationResult` SHALL have `IsValid bool` and `Errors []string` fields. Validation SHALL check: (1) heading count matches, (2) code blocks are identical, (3) URL sets are equal. Any failed check appends to `Errors` and sets `IsValid` to false.

#### Scenario: Returns valid when all checks pass
- **WHEN** compressed text has the same heading count, identical code blocks, and the same URL set as original
- **THEN** `IsValid` is true and `Errors` is empty

#### Scenario: Heading count mismatch sets IsValid false
- **WHEN** compressed text has a different number of headings than original
- **THEN** `IsValid` is false and `Errors` contains a heading-count error message

#### Scenario: Code block mismatch sets IsValid false
- **WHEN** a code block in the original is missing or altered in the compressed text
- **THEN** `IsValid` is false and `Errors` contains a code-block error message

#### Scenario: URL set mismatch sets IsValid false
- **WHEN** a URL present in the original is absent from the compressed text
- **THEN** `IsValid` is false and `Errors` contains a URL error message

#### Scenario: Multiple failures accumulate in Errors
- **WHEN** both heading count and URL set fail
- **THEN** `Errors` contains entries for both failures
