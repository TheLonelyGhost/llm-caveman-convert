package tokens

import (
	"testing"
)

func TestRegistryLength(t *testing.T) {
	if len(Registry) != 15 {
		t.Errorf("Registry has %d entries, want 15", len(Registry))
	}
}

func TestRegistryEntries(t *testing.T) {
	for i, e := range Registry {
		if e.Provider == "" {
			t.Errorf("entry %d: empty Provider", i)
		}
		if e.Model == "" {
			t.Errorf("entry %d: empty Model", i)
		}
		if e.counter == nil {
			t.Errorf("entry %d (%s/%s): nil counter", i, e.Provider, e.Model)
		}
	}
}

func TestRegistryProviders(t *testing.T) {
	want := map[string]int{
		"OpenAI":    3,
		"Anthropic": 3,
		"xAI":       3,
		"Google":    3,
		"Meta":      3,
	}
	got := make(map[string]int)
	for _, e := range Registry {
		got[e.Provider]++
	}
	for provider, wantCount := range want {
		if got[provider] != wantCount {
			t.Errorf("provider %q: got %d entries, want %d", provider, got[provider], wantCount)
		}
	}
}

func TestRegistryApproxFlags(t *testing.T) {
	for _, e := range Registry {
		if e.Provider == "OpenAI" && e.Approx {
			t.Errorf("OpenAI model %q should not be marked approximate", e.Model)
		}
		if e.Approx && e.MarginPct == 0 {
			t.Errorf("approx model %s/%s has no MarginPct set", e.Provider, e.Model)
		}
	}
}
