package caveman

import (
	"context"

	"github.com/TheLonelyGhost/llm-caveman-convert/internal/cache"
	"github.com/TheLonelyGhost/llm-caveman-convert/internal/compress"
	"github.com/TheLonelyGhost/llm-caveman-convert/internal/llm"
)

// Encode compresses input to caveman-speak using the LLM, with caching.
// Output is validated for structural invariants (headings, code blocks, URLs).
// If validation fails, a single fix-up call is made. The cache is written only
// when validation passes. Invalid output is returned uncached.
func Encode(ctx context.Context, c *cache.Cache, client *llm.Client, input string) (string, error) {
	if v, ok := c.GetEncode(input); ok {
		return v, nil
	}

	out, err := client.Encode(ctx, input)
	if err != nil {
		return "", err
	}

	if result := compress.Validate(input, out); result.IsValid {
		_ = c.SetEncode(input, out)
		return out, nil
	} else {
		fixed, fixErr := client.Fix(ctx, input, out, result.Errors)
		if fixErr == nil {
			out = fixed
		}
	}

	if compress.Validate(input, out).IsValid {
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
