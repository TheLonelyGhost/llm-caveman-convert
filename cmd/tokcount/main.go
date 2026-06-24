package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/TheLonelyGhost/llm-caveman-convert/internal/tokens"
)

func main() {
	jsonOut := flag.Bool("json", false, "output results as JSON")
	flag.Parse()

	text, err := readStdin()
	if err != nil {
		fmt.Fprintf(os.Stderr, "tokcount: read stdin: %v\n", err)
		os.Exit(1)
	}

	results := countAll(context.Background(), text)

	if *jsonOut {
		if err := tokens.WriteJSON(os.Stdout, results); err != nil {
			fmt.Fprintf(os.Stderr, "tokcount: write json: %v\n", err)
			os.Exit(1)
		}
		return
	}
	tokens.WriteReport(os.Stdout, results)
}

func readStdin() (string, error) {
	b, err := io.ReadAll(os.Stdin)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func countAll(ctx context.Context, text string) []tokens.Result {
	results := make([]tokens.Result, 0, len(tokens.Registry))
	for i := range tokens.Registry {
		results = append(results, tokens.Registry[i].Count(ctx, text))
	}
	return results
}
