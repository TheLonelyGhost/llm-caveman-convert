package caveman

import (
	"context"

	"github.com/TheLonelyGhost/llm-caveman-convert/internal/cache"
	"github.com/TheLonelyGhost/llm-caveman-convert/internal/compress"
	"github.com/TheLonelyGhost/llm-caveman-convert/internal/llm"
)

const maxEncodeRetries = 2

// Encode compresses input to caveman-speak using the LLM, with caching.
// Output is validated for structural invariants (headings, code blocks, URLs).
// On failure the LLM is asked to fix the result up to maxEncodeRetries times.
// The cache is written only when validation passes. On exhausted retries the
// last result is returned uncached.
func Encode(ctx context.Context, c *cache.Cache, client *llm.Client, input string) (string, error) {
	if v, ok := c.GetEncode(input); ok {
		return v, nil
	}

	out, err := client.Encode(ctx, input)
	if err != nil {
		return "", err
	}

	for attempt := 0; attempt < maxEncodeRetries; attempt++ {
		result := compress.Validate(input, out)
		if result.IsValid {
			_ = c.SetEncode(input, out)
			return out, nil
		}
		fixed, fixErr := client.Fix(ctx, input, out, result.Errors)
		if fixErr != nil {
			break
		}
		out = fixed
	}

	if result := compress.Validate(input, out); result.IsValid {
		_ = c.SetEncode(input, out)
	}

	return out, nil
}

// Decode expands caveman-speak to plain English using the LLM, with caching.
func Decode(ctx context.Context, c *cache.Cache, client *llm.Client, input string) (string, error) {
	if v, ok := c.GetDecode(input); ok {
		return v, nil
	}
	out, err := client.Decode(ctx, input)
	if err != nil {
		return "", err
	}
	_ = c.SetDecode(input, out)
	return out, nil
}
