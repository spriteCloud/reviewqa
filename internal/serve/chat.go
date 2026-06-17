package serve

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/reviewqa/reviewqa/internal/llm"
)

// ChatMessage is one turn in the conversation maintained client-side.
type ChatMessage struct {
	Role    string `json:"role"`    // "user" | "assistant"
	Content string `json:"content"`
}

// ChatInput is the payload the UI sends to /api/scenario-chat. The
// backend is stateless — the full History is replayed every turn so
// the conversation survives page reloads (stored in localStorage on
// the client).
type ChatInput struct {
	Scenario string        `json:"scenario"`        // current Gherkin block
	URL      string        `json:"url,omitempty"`   // destination URL for DOM grounding
	History  []ChatMessage `json:"history,omitempty"`
	User     string        `json:"user"`            // the latest user message
}

// ChatResult is what the UI renders.
type ChatResult struct {
	Assistant string `json:"assistant"`         // conversational reply
	Proposed  string `json:"proposed,omitempty"` // updated Gherkin block (or empty)
	Valid     bool   `json:"valid"`             // proposed block parses cleanly
	Model     string `json:"model,omitempty"`
	Notes     string `json:"notes,omitempty"`
}

// chatSentinel separates the conversational reply from the proposed
// Gherkin block in the LLM's response. The system prompt instructs
// the model to emit this delimiter when (and only when) it wants the
// caller to apply a change.
const chatSentinel = "---SCENARIO---"

const chatSystemPrompt = `You are a Gherkin Scenario maintenance assistant. The user owns ONE specific Playwright + playwright-bdd Scenario and wants to refine it through conversation. Each turn you receive:

1. The current Scenario block (verbatim).
2. The destination page's DOM landmarks (when available) — title, headings, links, forms.
3. The full conversation history.
4. The user's new message.

Your job: read what the user wants in plain language, propose a concrete change, and (when the user is asking for an edit) emit the FULL updated Scenario block.

Output format — STRICT:
1. First, a 1-3 sentence plain-text reply explaining what you changed (or asking for clarification).
2. THEN, only when you have a concrete updated Scenario to propose, a single line containing exactly "---SCENARIO---" and the full updated block below it.

If the user is just asking a question (no edit needed), reply in plain text only — DO NOT emit the sentinel.

Hard rules for the updated block:
- Exactly one Scenario.
- All step keywords are Given / When / Then / And / But.
- Each step text MUST match one of the registered patterns below (substitute angle-bracketed placeholders with concrete values from the DOM when applicable).
- Tag the Scenario with @ai-refined.

Registered step patterns:
Given I open the landing page
Given I am on the landing page
Given I open the page "<path>"
When I click the link to "<href>"
When I navigate directly to "<href>"
When I enter "<value>" into the "<field>" field
When I submit the form
Then the page title contains "<title>"
Then the main heading reads "<text>"
Then I see the heading "<text>"
Then the URL contains "<fragment>"
Then no error message is shown in the form region
Then the page has at least <n> items
When I scroll to the bottom of the page
When I reload the page
When I submit the form twice in rapid succession
Then the form is not double-submitted
`

// Chat takes a scenario + history + user message and asks the
// configured LLM for a conversational reply and (optionally) a
// proposed updated block. Stateless — the client owns the history.
func Chat(ctx context.Context, in ChatInput) (*ChatResult, error) {
	if strings.TrimSpace(in.User) == "" {
		return nil, errors.New("user message is empty")
	}
	cfg := llmConfigFromEnv()
	if cfg.OpenAIAPIKey == "" {
		return nil, errors.New("LLM not configured; set REVIEWQA_LLM to a chat-completions endpoint")
	}
	client := llm.New(cfg)
	if !client.Enabled() {
		return nil, errors.New("LLM client disabled")
	}

	var lm *DOMLandmarks
	if strings.TrimSpace(in.URL) != "" {
		resolved, err := resolveTarget(in.URL, "")
		if err == nil {
			// DOM is best-effort — chat works without it.
			lm, _ = FetchAndParseDOM(ctx, resolved)
		}
	}

	user := buildChatUserPrompt(in, lm)
	raw, err := client.Chat(ctx, chatSystemPrompt, user)
	if err != nil {
		return nil, fmt.Errorf("llm: %w", err)
	}
	res := parseChatResponse(raw)
	res.Model = cfg.Model
	if res.Proposed != "" {
		if err := validateScenarioBlock(res.Proposed); err != nil {
			res.Valid = false
			res.Notes = "proposed block failed validation: " + err.Error()
		} else {
			res.Valid = true
		}
	}
	return res, nil
}

func buildChatUserPrompt(in ChatInput, lm *DOMLandmarks) string {
	var b strings.Builder
	b.WriteString("Current Scenario:\n")
	b.WriteString(in.Scenario)
	b.WriteString("\n\n")
	if lm != nil {
		fmt.Fprintf(&b, "Destination URL: %s\n", lm.URL)
		if lm.Title != "" {
			fmt.Fprintf(&b, "Title: %s\n", lm.Title)
		}
		if len(lm.Headings) > 0 {
			parts := make([]string, 0, len(lm.Headings))
			for _, h := range lm.Headings {
				parts = append(parts, fmt.Sprintf("h%d:%q", h.Level, h.Text))
			}
			fmt.Fprintf(&b, "Headings: %s\n", strings.Join(parts, "; "))
		}
		if len(lm.Forms) > 0 {
			fmt.Fprintf(&b, "Forms: %s\n", summariseForms(lm.Forms))
		}
		if len(lm.Buttons) > 0 {
			btns := make([]string, 0, len(lm.Buttons))
			for _, b2 := range lm.Buttons {
				n := b2.Text
				if n == "" {
					n = b2.AriaLabel
				}
				btns = append(btns, fmt.Sprintf("%q", n))
			}
			fmt.Fprintf(&b, "Buttons: %s\n", strings.Join(btns, ", "))
		}
		b.WriteString("\n")
	}
	if len(in.History) > 0 {
		b.WriteString("Conversation so far:\n")
		// Bound history to ~20 turns to keep token use sane.
		start := 0
		if len(in.History) > 20 {
			start = len(in.History) - 20
		}
		for _, m := range in.History[start:] {
			fmt.Fprintf(&b, "%s: %s\n", m.Role, m.Content)
		}
		b.WriteString("\n")
	}
	fmt.Fprintf(&b, "user: %s\n", in.User)
	return b.String()
}

func parseChatResponse(raw string) *ChatResult {
	idx := strings.Index(raw, chatSentinel)
	if idx == -1 {
		return &ChatResult{Assistant: strings.TrimSpace(raw)}
	}
	assistant := strings.TrimSpace(raw[:idx])
	rest := strings.TrimSpace(raw[idx+len(chatSentinel):])
	// Some models echo a trailing closing of their own back-tick block —
	// strip a leading ```gherkin / ``` and trailing ```.
	rest = strings.TrimPrefix(rest, "```gherkin")
	rest = strings.TrimPrefix(rest, "```")
	rest = strings.TrimSuffix(rest, "```")
	rest = strings.TrimSpace(rest)
	rest = extractScenarioBlock(rest)
	return &ChatResult{Assistant: assistant, Proposed: rest}
}
