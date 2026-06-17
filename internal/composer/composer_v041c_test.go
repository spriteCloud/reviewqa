package composer

import (
	"strings"
	"testing"
)

func TestRegisteredPatterns_RaceAndReloadVocab(t *testing.T) {
	cases := []struct {
		text   string
		expect bool
	}{
		{"I submit the form twice in rapid succession", true},
		{"the form is not double-submitted", true},
		{"I reload the page", true},
		{"I submit the form THREE times", false},
		{"the form is not triple-submitted", false},
	}
	for _, c := range cases {
		got := matchesRegisteredPattern(c.text)
		if got != c.expect {
			t.Errorf("matchesRegisteredPattern(%q) = %v; want %v", c.text, got, c.expect)
		}
	}
}

func TestSystemPrompt_MentionsRaceAndReloadFamilies(t *testing.T) {
	// The system prompt has to advertise the new vocabulary so the LLM
	// knows it can use it — otherwise we add patterns the model will
	// never propose.
	for _, sub := range []string{"rapid succession", "double-submitted", "reload"} {
		if !strings.Contains(systemPrompt, sub) {
			t.Errorf("systemPrompt missing %q vocabulary cue", sub)
		}
	}
}

func TestValidate_DoubleSubmitScenario(t *testing.T) {
	in := []ExtraScenario{
		{
			Name: "Double-submit is debounced",
			Steps: []Step{
				{Keyword: "Given", Text: "I open the landing page"},
				{Keyword: "When", Text: "I submit the form twice in rapid succession"},
				{Keyword: "Then", Text: "the form is not double-submitted"},
			},
		},
	}
	out := Validate(in)
	if len(out) != 1 {
		t.Fatalf("validate dropped a valid race-scenario; got %d", len(out))
	}
}
