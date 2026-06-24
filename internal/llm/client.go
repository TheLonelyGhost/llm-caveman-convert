package llm

import (
	"context"
	"fmt"
	"strings"

	openai "github.com/sashabaranov/go-openai"
)

const encodeSystemPrompt = `Markdown compressor. Convert to caveman-speak.

PRESERVE EXACTLY — do not modify:
- Fenced code blocks (` + "```" + ` or ~~~, any info string): copy every line verbatim
- Inline backtick code (` + "`" + `like this` + "`" + `): copy verbatim
- ATX headings (# ## ### etc.): copy heading line verbatim
- URLs (http:// https://): copy verbatim
- File paths, commands, env vars ($HOME, NODE_ENV): copy verbatim
- Technical terms, library names, API names, proper nouns
- Dates, version numbers, numeric values
- Table structure: compress cell text, keep | separators and alignment rows

COMPRESS natural language outside the above:
- Drop articles (a, an, the), filler (just, really, basically, actually, simply), hedging, pleasantries
- Drop connective fluff: however, furthermore, additionally, in addition
- Contract redundant phrases: "in order to"→"to", "make sure to"→"ensure", "the reason is because"→"because", "utilize"→"use"
- Drop "you should", "make sure to", "remember to" — state the action directly
- Use short synonyms: "fix" not "implement a solution for", "use" not "utilize"
- Fragments OK
- Merge redundant bullets that say the same thing differently

Output compressed text only. Do not wrap output in a code fence.`

const decodeSystemPrompt = `Text decompressor. Expand caveman-speak to fluent English. Output expanded text only.`

const fixSystemPrompt = `You are fixing a caveman-compressed markdown file. Specific validation errors were found.

CRITICAL RULES:
- Do NOT recompress or rephrase the file
- ONLY fix the listed errors — leave everything else exactly as-is
- The ORIGINAL is provided as reference only (to restore missing content)
- Preserve caveman style in all untouched sections

HOW TO FIX:
- Missing URL: find it in ORIGINAL, restore it exactly where it belongs in COMPRESSED
- Code block mismatch: find the exact code block in ORIGINAL, restore it in COMPRESSED
- Heading mismatch: restore the exact heading text from ORIGINAL into COMPRESSED
- Do not touch any section not mentioned in the errors

Return ONLY the fixed compressed file. No explanation.`

// Client wraps an OpenAI-compatible API for encode/decode completions.
type Client struct {
	client *openai.Client
	model  string
}

// New creates a Client targeting the given base URL, API key, and model.
func New(baseURL, apiKey, model string) *Client {
	cfg := openai.DefaultConfig(apiKey)
	if baseURL != "" {
		cfg.BaseURL = baseURL
	}
	return &Client{
		client: openai.NewClientWithConfig(cfg),
		model:  model,
	}
}

// Encode compresses input to caveman-speak.
func (c *Client) Encode(ctx context.Context, input string) (string, error) {
	return c.complete(ctx, encodeSystemPrompt, input)
}

// Decode expands caveman-speak to plain English.
func (c *Client) Decode(ctx context.Context, input string) (string, error) {
	return c.complete(ctx, decodeSystemPrompt, input)
}

// Fix asks the LLM to restore only the listed validation errors in compressed,
// using original as reference. All other content is left unchanged.
func (c *Client) Fix(ctx context.Context, original, compressed string, errors []string) (string, error) {
	errList := strings.Join(errors, "\n")
	user := fmt.Sprintf("ERRORS TO FIX:\n%s\n\nORIGINAL (reference only):\n%s\n\nCOMPRESSED (fix this):\n%s", errList, original, compressed)
	return c.complete(ctx, fixSystemPrompt, user)
}

func (c *Client) complete(ctx context.Context, system, user string) (string, error) {
	resp, err := c.client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model: c.model,
		Messages: []openai.ChatCompletionMessage{
			{Role: openai.ChatMessageRoleSystem, Content: system},
			{Role: openai.ChatMessageRoleUser, Content: user},
		},
	})
	if err != nil {
		return "", err
	}
	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("no choices in response")
	}
	return resp.Choices[0].Message.Content, nil
}
