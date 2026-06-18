package probe

import (
	"testing"

	"github.com/reviewqa/reviewqa/internal/ast"
	"github.com/reviewqa/reviewqa/internal/mindmap"
)

func TestSynthesiseFallbackJourneys_EmitsOneFromLanding(t *testing.T) {
	landingURL := "https://example.com/"
	subURL := "https://example.com/about"
	m := &mindmap.Map{
		Origin: landingURL,
		Order:  []string{landingURL, subURL},
		Pages: map[string]*mindmap.Page{
			landingURL: {
				URL:  landingURL,
				Tags: []string{mindmap.TagLanding},
				Links: []ast.LocatorAnchor{
					{Aria: subURL, Text: "About"},
					{Aria: "https://example.com/contact", Text: "Contact"},
					{Aria: "https://example.com/services", Text: "Services"},
				},
			},
			subURL: {URL: subURL},
		},
	}
	js := synthesiseFallbackJourneys(m, landingURL)
	if len(js) != 1 {
		t.Fatalf("got %d journeys, want 1: %+v", len(js), js)
	}
	if js[0].Kind != mindmap.JourneyDiscover {
		t.Errorf("kind = %q, want discover", js[0].Kind)
	}
	if len(js[0].Steps) != 2 {
		t.Fatalf("steps = %d, want 2", len(js[0].Steps))
	}
	if js[0].Steps[0].Page.URL != landingURL || js[0].Steps[1].Page.URL != subURL {
		t.Errorf("steps mis-routed: %+v", js[0].Steps)
	}
}

// With fewer than 3 landing links we deliberately bail — better
// to honestly report "no journeys" than emit a thin Scenario.
func TestSynthesiseFallbackJourneys_BailsBelowThreshold(t *testing.T) {
	m := &mindmap.Map{
		Origin: "https://example.com/",
		Order:  []string{"https://example.com/"},
		Pages: map[string]*mindmap.Page{
			"https://example.com/": {
				URL:   "https://example.com/",
				Tags:  []string{mindmap.TagLanding},
				Links: []ast.LocatorAnchor{{Aria: "/x"}, {Aria: "/y"}},
			},
		},
	}
	if js := synthesiseFallbackJourneys(m, "https://example.com/"); len(js) != 0 {
		t.Errorf("expected nil for thin landing, got %+v", js)
	}
}
