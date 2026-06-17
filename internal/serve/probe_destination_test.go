package serve

import (
	"net/url"
	"os"
	"path/filepath"
	"testing"
)

func mustParseURL(t *testing.T, s string) *url.URL {
	t.Helper()
	u, err := url.Parse(s)
	if err != nil {
		t.Fatal(err)
	}
	return u
}

// Re-probing the same URL as the current workdir's name should
// keep us in place (no sibling dir created).
func TestPickProbeDestination_RepProbeInPlace(t *testing.T) {
	parent := t.TempDir()
	current := filepath.Join(parent, "spritecloud")
	fixtureProjectAt(t, current)
	got := pickProbeDestination(current, mustParseURL(t, "https://www.spritecloud.com"))
	if got != current {
		t.Errorf("got %q, want re-probe in place %q", got, current)
	}
}

// Probing a fresh URL while serving an existing project should
// create a sibling dir named after the URL's brand.
func TestPickProbeDestination_NewBrandCreatesSibling(t *testing.T) {
	parent := t.TempDir()
	current := filepath.Join(parent, "spritecloud")
	fixtureProjectAt(t, current)
	got := pickProbeDestination(current, mustParseURL(t, "https://petstore3.swagger.io"))
	want := filepath.Join(parent, "petstore3.swagger")
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
	if info, err := os.Stat(got); err != nil || !info.IsDir() {
		t.Errorf("destination %q was not created as a dir", got)
	}
}

// A pre-existing NON-reviewqa squatter dir should NOT be probed
// into. We pick a -1 suffix instead.
func TestPickProbeDestination_CollidesWithNonReviewqaDir(t *testing.T) {
	parent := t.TempDir()
	current := filepath.Join(parent, "spritecloud")
	fixtureProjectAt(t, current)
	squatter := filepath.Join(parent, "petstore3.swagger")
	if err := os.MkdirAll(filepath.Join(squatter, "random-files"), 0o755); err != nil {
		t.Fatal(err)
	}
	got := pickProbeDestination(current, mustParseURL(t, "https://petstore3.swagger.io"))
	want := filepath.Join(parent, "petstore3.swagger-1")
	if got != want {
		t.Errorf("got %q, want %q (squatter present)", got, want)
	}
}

// A pre-existing reviewqa project for the same brand should
// re-probe in place (no -1 suffix).
func TestPickProbeDestination_CollidesWithReviewqaSibling(t *testing.T) {
	parent := t.TempDir()
	current := filepath.Join(parent, "spritecloud")
	fixtureProjectAt(t, current)
	sibling := filepath.Join(parent, "petstore3.swagger")
	fixtureProjectAt(t, sibling)
	got := pickProbeDestination(current, mustParseURL(t, "https://petstore3.swagger.io"))
	if got != sibling {
		t.Errorf("got %q, want %q (existing reviewqa sibling)", got, sibling)
	}
}
