package compress

import (
	"fmt"
	"regexp"
	"strings"
)

var (
	headingRe = regexp.MustCompile(`(?m)^#{1,6}\s+.*`)
	urlRe     = regexp.MustCompile(`https?://[^\s)]+`)
	fenceRe   = regexp.MustCompile(`^(\s{0,3})(\x60{3,}|~{3,})(.*)$`)
)

// ValidationResult holds the outcome of a structural validation check.
type ValidationResult struct {
	IsValid bool
	Errors  []string
}

func (r *ValidationResult) addError(msg string) {
	r.IsValid = false
	r.Errors = append(r.Errors, msg)
}

// ExtractHeadings returns all ATX heading lines in document order.
// Always returns a non-nil slice.
func ExtractHeadings(text string) []string {
	out := headingRe.FindAllString(text, -1)
	if out == nil {
		return make([]string, 0)
	}
	return out
}

// ExtractCodeBlocks returns all fenced code blocks (including fence lines)
// using CommonMark rules: closing fence must use the same character and be at
// least as long as the opening fence; unclosed fences are silently dropped.
func ExtractCodeBlocks(text string) []string {
	blocks := make([]string, 0)
	lines := strings.Split(text, "\n")
	n := len(lines)
	i := 0
	for i < n {
		m := fenceRe.FindStringSubmatch(lines[i])
		if m == nil {
			i++
			continue
		}
		fenceChar := rune(m[2][0])
		fenceLen := len(m[2])
		block := []string{lines[i]}
		i++
		closed := false
		for i < n {
			cm := fenceRe.FindStringSubmatch(lines[i])
			if cm != nil && rune(cm[2][0]) == fenceChar && len(cm[2]) >= fenceLen && strings.TrimSpace(cm[3]) == "" {
				block = append(block, lines[i])
				closed = true
				i++
				break
			}
			block = append(block, lines[i])
			i++
		}
		if closed {
			blocks = append(blocks, strings.Join(block, "\n"))
		}
	}
	return blocks
}

// ExtractURLs returns all http and https URLs found in text.
// Trailing sentence punctuation (., ,, )) is stripped from each match.
func ExtractURLs(text string) []string {
	raw := urlRe.FindAllString(text, -1)
	out := make([]string, 0, len(raw))
	for _, u := range raw {
		out = append(out, strings.TrimRight(u, ".,)"))
	}
	return out
}

func toSet(ss []string) map[string]struct{} {
	m := make(map[string]struct{}, len(ss))
	for _, s := range ss {
		m[s] = struct{}{}
	}
	return m
}

func setDiff(a, b map[string]struct{}) []string {
	var out []string
	for k := range a {
		if _, ok := b[k]; !ok {
			out = append(out, k)
		}
	}
	return out
}

// Validate checks that compressed preserves the structural invariants of
// original: heading count, exact code blocks, and URL set.
func Validate(original, compressed string) ValidationResult {
	result := ValidationResult{IsValid: true}

	origHeadings := ExtractHeadings(original)
	compHeadings := ExtractHeadings(compressed)
	if len(origHeadings) != len(compHeadings) {
		result.addError(fmt.Sprintf("heading count mismatch: original %d, compressed %d", len(origHeadings), len(compHeadings)))
	}

	origBlocks := ExtractCodeBlocks(original)
	compBlocks := ExtractCodeBlocks(compressed)
	equal := len(origBlocks) == len(compBlocks)
	if equal {
		for i := range origBlocks {
			if origBlocks[i] != compBlocks[i] {
				equal = false
				break
			}
		}
	}
	if !equal {
		result.addError("code blocks not preserved exactly")
	}

	origURLs := toSet(ExtractURLs(original))
	compURLs := toSet(ExtractURLs(compressed))
	lost := setDiff(origURLs, compURLs)
	added := setDiff(compURLs, origURLs)
	if len(lost) > 0 || len(added) > 0 {
		result.addError(fmt.Sprintf("URL mismatch: lost=%v added=%v", lost, added))
	}

	return result
}
