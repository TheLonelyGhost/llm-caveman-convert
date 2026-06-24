// Package tokens provides token counting for major frontier LLM providers.
// It exposes a Counter interface and a Registry of pre-configured model entries
// covering OpenAI, Anthropic, xAI, Google, and Meta Llama models.
// Local BPE counting is performed via tiktoken-go; Google Gemini counting
// uses the Google Gen AI API when GOOGLE_API_KEY is present.
package tokens
