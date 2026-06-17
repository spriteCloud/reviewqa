// Package composer asks an LLM for additional Gherkin scenarios
// beyond what the deterministic templates emit. It's the answer to
// "templates are fixed; can scenarios be composed from parts to find
// more user flows?" — the deterministic baseline stays, and the model
// is constrained to compose Scenarios using ONLY the Given/When/Then
// patterns already registered in reviewqa.steps.ts.
//
// The composer is strictly OPT-IN. Default is off; the consumer must
// pass --llm <url>. The LLM is consulted at generation time only —
// emitted .feature files are deterministic Gherkin text and CI runs
// them with zero LLM dependency.
package composer

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"
)

// Client is the minimal LLM surface the composer needs. internal/llm
// Client satisfies it via its Chat method; tests pass a fake.
type Client interface {
	Chat(ctx context.Context, system, user string) (string, error)
}

// Journey is the probe-side context the composer renders into the
// model prompt. All fields optional; the model handles partial data.
type Journey struct {
	URL       string
	Kind      string // convert | contact | authenticate | ... | exercise
	Priority  string // critical | standard | nice-to-have
	Title     string
	H1        string
	Links     []string // ranked, top-10
	Forms     []string // human-readable form summaries
	HasForm   bool
}

// ExtraScenario is one LLM-proposed Scenario in a normalized shape.
// Validated against the registered step patterns; invalid scenarios
// are dropped before the template renders.
type ExtraScenario struct {
	Name  string   `json:"name"`
	Tags  []string `json:"tags,omitempty"`
	Steps []Step   `json:"steps"`
}

// Step is one Gherkin step inside an ExtraScenario.
type Step struct {
	Keyword string `json:"keyword"` // Given | When | And | Then
	Text    string `json:"text"`    // verbatim Gherkin (placeholders substituted)
}

const systemPrompt = `You are a senior QA engineer composing additional Gherkin scenarios for a Playwright + playwright-bdd test suite.

You will receive a probe summary of a single user journey. Compose UP TO 3 additional Scenarios beyond the deterministic happy path.

Constraints:
- Each step.text MUST be a verbatim match (after placeholder substitution) of one of the registered patterns below.
- Each Scenario has 3-6 steps and starts with Given.
- Use Given for setup, When for actions, Then for assertions, And to chain.
- Keep names short and concrete (under 10 words). Tag with @kind:edge or @kind:variant if relevant.
- Output STRICT JSON ONLY — no markdown fences, no commentary. Shape: an array of {name, tags?, steps:[{keyword,text}]}.

Registered step patterns (substitute the angle-bracketed placeholders with concrete values):

Given I open the landing page
Given I am on the landing page
Given I open the page "<path>"
When I click the link to "<href>"
When I navigate directly to "<href>"
When I enter "<value>" into the "<field>" field
When I submit the form
When I submit the form without filling any required field
Then the page title contains "<title>"
Then the main heading reads "<text>"
Then I see the heading "<text>"
Then no error message is shown in the form region
Then I remain on the same page
Then no success message is shown
`

// Propose asks the LLM for up to n additional scenarios for the
// journey. Returns ([], nil) when the model declines or returns
// nothing useful — never an error.
func Propose(ctx context.Context, llm Client, j Journey, n int) ([]ExtraScenario, error) {
	if llm == nil {
		return nil, nil
	}
	if n <= 0 {
		n = 3
	}
	user := buildUserPrompt(j, n)
	raw, err := llm.Chat(ctx, systemPrompt, user)
	if err != nil {
		return nil, fmt.Errorf("composer: llm chat: %w", err)
	}
	scenarios, err := Parse(raw)
	if err != nil {
		return nil, fmt.Errorf("composer: parse response: %w", err)
	}
	scenarios = Validate(scenarios)
	if len(scenarios) > n {
		scenarios = scenarios[:n]
	}
	return scenarios, nil
}

func buildUserPrompt(j Journey, n int) string {
	var b strings.Builder
	fmt.Fprintf(&b, "URL: %s\n", j.URL)
	fmt.Fprintf(&b, "Journey kind: %s (priority: %s)\n", j.Kind, j.Priority)
	if j.Title != "" {
		fmt.Fprintf(&b, "Title: %s\n", j.Title)
	}
	if j.H1 != "" {
		fmt.Fprintf(&b, "H1: %s\n", j.H1)
	}
	if len(j.Links) > 0 {
		fmt.Fprintf(&b, "Top links: %s\n", strings.Join(j.Links, ", "))
	}
	if j.HasForm {
		fmt.Fprintf(&b, "Page has a form. Form summaries: %s\n", strings.Join(j.Forms, "; "))
	}
	fmt.Fprintf(&b, "\nPropose up to %d additional Scenarios for this journey. JSON only.\n", n)
	return b.String()
}

// Parse extracts the JSON array of ExtraScenario from the model's raw
// response. Tolerant of leading prose / fenced code blocks — we slice
// from the first `[` to the last `]`.
func Parse(raw string) ([]ExtraScenario, error) {
	start := strings.Index(raw, "[")
	end := strings.LastIndex(raw, "]")
	if start < 0 || end < start {
		return nil, errors.New("composer: no JSON array in response")
	}
	jsonText := raw[start : end+1]
	var out []ExtraScenario
	if err := json.Unmarshal([]byte(jsonText), &out); err != nil {
		return nil, err
	}
	return out, nil
}

// Validate drops scenarios that don't conform to the registered step
// pattern set or carry an empty name. Returns the valid subset.
func Validate(in []ExtraScenario) []ExtraScenario {
	var out []ExtraScenario
	for _, s := range in {
		if strings.TrimSpace(s.Name) == "" {
			continue
		}
		if len(s.Steps) < 1 {
			continue
		}
		if !allStepsRegistered(s.Steps) {
			continue
		}
		out = append(out, s)
	}
	return out
}

// allStepsRegistered returns true when every step in the scenario
// matches one of the registered patterns. Empty placeholder substitutions
// pass — the runtime step definitions handle them with sensible defaults.
func allStepsRegistered(steps []Step) bool {
	for _, s := range steps {
		if !matchesRegisteredPattern(s.Text) {
			return false
		}
	}
	return true
}

// registeredPatterns is the canonical list of Gherkin step texts the
// composer is allowed to produce. The "<...>" placeholders are
// substituted with concrete values; the regex normalises them out.
var registeredPatterns = []*regexp.Regexp{
	regexp.MustCompile(`^I open the landing page$`),
	regexp.MustCompile(`^I am on the landing page$`),
	regexp.MustCompile(`^I open the page "[^"]+"$`),
	regexp.MustCompile(`^I click the link to "[^"]+"$`),
	regexp.MustCompile(`^I navigate directly to "[^"]+"$`),
	regexp.MustCompile(`^I enter "[^"]*" into the "[^"]+" field$`),
	regexp.MustCompile(`^I submit the form$`),
	regexp.MustCompile(`^I submit the form without filling any required field$`),
	regexp.MustCompile(`^the page title contains "[^"]+"$`),
	regexp.MustCompile(`^the main heading reads "[^"]+"$`),
	regexp.MustCompile(`^I see the heading "[^"]+"$`),
	regexp.MustCompile(`^no error message is shown in the form region$`),
	regexp.MustCompile(`^I remain on the same page$`),
	regexp.MustCompile(`^no success message is shown$`),
}

func matchesRegisteredPattern(text string) bool {
	text = strings.TrimSpace(text)
	for _, re := range registeredPatterns {
		if re.MatchString(text) {
			return true
		}
	}
	return false
}
