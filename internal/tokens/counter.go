package tokens

import "context"

// Counter counts tokens in a text string for a specific model.
type Counter interface {
	Count(ctx context.Context, text string) (int, error)
}

// Result holds the token count outcome for one model.
type Result struct {
	Provider   string
	Model      string
	Tokens     int
	Approx     bool
	MarginPct  int
	SkipReason string
}

// Entry pairs a model's metadata with its Counter.
type Entry struct {
	Provider  string
	Model     string
	Approx    bool
	MarginPct int
	counter   Counter
}

// Count delegates to the entry's Counter.
func (e *Entry) Count(ctx context.Context, text string) Result {
	if e.counter == nil {
		return Result{
			Provider:   e.Provider,
			Model:      e.Model,
			SkipReason: "no counter configured",
		}
	}
	n, err := e.counter.Count(ctx, text)
	if err != nil {
		return Result{
			Provider:   e.Provider,
			Model:      e.Model,
			SkipReason: err.Error(),
		}
	}
	return Result{
		Provider:  e.Provider,
		Model:     e.Model,
		Tokens:    n,
		Approx:    e.Approx,
		MarginPct: e.MarginPct,
	}
}
