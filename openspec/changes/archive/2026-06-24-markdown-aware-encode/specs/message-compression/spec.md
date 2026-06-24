## MODIFIED Requirements

### Requirement: Whole message compressed via caveman binary
The proxy SHALL compress each message as a single unit by invoking `caveman --encode` as a subprocess. The entire message content is passed via stdin; the compressed result is read from stdout. The binary handles caching internally.

#### Scenario: Full message passed to binary
- **WHEN** a message is ready for compression
- **THEN** the complete message content string is written to `caveman --encode` stdin; no pre-processing or segmentation is performed by the proxy

#### Scenario: Binary failure falls back to original
- **WHEN** `caveman --encode` exits non-zero or produces no output
- **THEN** the original message content is used verbatim; no error is surfaced to the user

## REMOVED Requirements

### Requirement: Encode handles code fences per compressor LLM instruction
**Reason**: Replaced by unified rule: all fenced blocks are verbatim. The language-tag distinction (compress comments in full-tag blocks) is removed.
**Migration**: No caller change needed. The binary interface (`--encode` stdin/stdout) is unchanged. Callers that relied on comment compression in tagged blocks will now receive those blocks verbatim.

### Requirement: Encode passes through untagged code blocks
**Reason**: Superseded by the unified verbatim-all-fenced-blocks rule, which covers tagged and untagged blocks identically. Separate scenario no longer needed.
**Migration**: Behaviour is preserved (untagged blocks still pass through verbatim); only the language-tagged block behaviour changes.

## ADDED Requirements

### Requirement: Encode preserves all markdown structural invariants
The `caveman --encode` command SHALL produce output that passes structural validation: heading count unchanged, all fenced code blocks identical to input, and all URLs present.

#### Scenario: All fenced code blocks verbatim regardless of language tag
- **WHEN** input contains a fenced code block with any info string (e.g., ` ```python `, ` ```js `, ` ``` `)
- **THEN** the block appears verbatim in the encoded output; no line within the block is modified

#### Scenario: Inline backtick code preserved verbatim
- **WHEN** input contains inline backtick code (e.g., `` `funcName` ``)
- **THEN** the inline code appears verbatim in the encoded output

#### Scenario: Headings preserved verbatim
- **WHEN** input contains ATX headings (e.g., `## Section Title`)
- **THEN** heading lines appear verbatim in the encoded output

#### Scenario: URLs preserved verbatim
- **WHEN** input contains http or https URLs
- **THEN** all URLs appear verbatim in the encoded output

#### Scenario: Connective fluff removed
- **WHEN** input contains connective filler words (however, furthermore, additionally, in addition)
- **THEN** those words are absent from the encoded output

#### Scenario: Redundant phrasing contracted
- **WHEN** input contains verbose phrases (e.g., "in order to", "make sure to", "the reason is because")
- **THEN** shorter equivalents appear in the encoded output ("to", "ensure", "because")
