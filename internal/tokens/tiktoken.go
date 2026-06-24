package tokens

import (
	"context"
	"fmt"

	tiktoken "github.com/pkoukk/tiktoken-go"
)

type tiktokenCounter struct {
	enc *tiktoken.Tiktoken
}

// newTiktokenCounter initializes a BPE encoder for the given encoding name
// (e.g. "o200k_base", "cl100k_base") and returns a Counter that reuses it.
func newTiktokenCounter(encoding string) (Counter, error) {
	enc, err := tiktoken.GetEncoding(encoding)
	if err != nil {
		return nil, fmt.Errorf("tiktoken encoding %q: %w", encoding, err)
	}
	return &tiktokenCounter{enc: enc}, nil
}

// Count returns the number of BPE tokens in text.
func (c *tiktokenCounter) Count(_ context.Context, text string) (int, error) {
	return len(c.enc.EncodeOrdinary(text)), nil
}
