package serve

import (
	"context"
	"strings"
	"testing"
)

func TestComposeDeterministic_UsesH1FromDOM(t *testing.T) {
	lm := &DOMLandmarks{
		URL:      "https://example.com/about",
		Headings: []DOMHeading{{Level: 1, Text: "About us"}},
	}
	res := composeDeterministic("I want to confirm the about page", lm)
	if !strings.Contains(res.Gherkin, `Then I see the heading "About us"`) {
		t.Errorf("expected H1 assertion; got %s", res.Gherkin)
	}
	if !strings.Contains(res.Gherkin, `When I navigate directly to "/about"`) {
		t.Errorf("expected navigation step; got %s", res.Gherkin)
	}
	if err := validateScenarioBlock(res.Gherkin); err != nil {
		t.Errorf("deterministic output must parse cleanly: %v", err)
	}
}

func TestComposeDeterministic_FallsBackWhenNoHeading(t *testing.T) {
	lm := &DOMLandmarks{URL: "https://example.com/"}
	res := composeDeterministic("anything", lm)
	if !strings.Contains(res.Gherkin, "I remain on the same page") {
		t.Errorf("expected fallback assertion; got %s", res.Gherkin)
	}
}

func TestExtractScenarioBlock_StripsPreamble(t *testing.T) {
	raw := "Here is the scenario you asked for:\n\n  @ai-composed\n  Scenario: foo\n    Given I open the landing page\n    Then I remain on the same page\n\nLet me know if you need refinements!"
	got := extractScenarioBlock(raw)
	if !strings.HasPrefix(strings.TrimSpace(got), "@ai-composed") {
		t.Errorf("expected to start at @ai-composed; got %q", got)
	}
	if strings.Contains(got, "Let me know") {
		t.Errorf("trailing chatter not trimmed; got %q", got)
	}
}

func TestFirstLineOfFreeGherkin_StripsScenarioPrefix(t *testing.T) {
	got := firstLineOfFreeGherkin("Scenario: visit the about page\n  ...")
	if got != "visit the about page" {
		t.Errorf("got %q", got)
	}
}

func TestComposeSteps_RejectsEmptyInput(t *testing.T) {
	if _, err := ComposeSteps(context.Background(), ComposeInput{}); err == nil {
		t.Errorf("expected error for empty input")
	}
}
