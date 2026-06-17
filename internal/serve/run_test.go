package serve

import (
	"os"
	"path/filepath"
	"testing"
)

func TestPreflight_ReadyWhenPlaywrightPresent(t *testing.T) {
	root := t.TempDir()
	pwBin := filepath.Join(root, "node_modules", ".bin", "playwright")
	if err := os.MkdirAll(filepath.Dir(pwBin), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(pwBin, []byte("#!/bin/sh\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	got := Preflight(root)
	if !got.Ready {
		t.Errorf("expected Ready=true; got %+v", got)
	}
	if got.PlaywrightBin == "" {
		t.Errorf("expected PlaywrightBin to be set")
	}
}

func TestPreflight_NotReadyWhenMissing(t *testing.T) {
	root := t.TempDir()
	got := Preflight(root)
	if got.Ready {
		t.Errorf("expected Ready=false; got %+v", got)
	}
	if got.Message == "" {
		t.Errorf("expected a helpful message")
	}
}

func TestParseRunSummary_AllPassed(t *testing.T) {
	lines := []string{
		"Running 12 tests using 6 workers",
		"  ✓  1 [chromium] › convert.feature:23 (1.2s)",
		"12 passed (8.2s)",
	}
	s := ParseRunSummary(lines)
	if !s.HasLine || s.Passed != 12 || s.Failed != 0 {
		t.Errorf("summary: %+v", s)
	}
}

func TestParseRunSummary_MixedFailures(t *testing.T) {
	lines := []string{
		"Running 5 tests using 3 workers",
		"  ✘  1 [chromium] › convert.feature:23 (1.2s)",
		"3 passed (8.2s)",
		"2 failed",
	}
	s := ParseRunSummary(lines)
	if s.Passed != 3 || s.Failed != 2 {
		t.Errorf("expected 3/2; got %+v", s)
	}
}

func TestParseRunSummary_NoOutput(t *testing.T) {
	if ParseRunSummary(nil).HasLine {
		t.Errorf("expected HasLine=false for empty input")
	}
}

const pwReportFixture = `{
  "suites": [{
    "suites": [{
      "specs": [{
        "title": "Feature Foo > visits the about page",
        "tests": [{
          "results": [{
            "status": "passed",
            "duration": 4321,
            "steps": [
              {"title": "Given I open the landing page", "duration": 120, "category": null, "steps": []},
              {"title": "When I navigate directly to \"/about\"", "duration": 340, "category": null, "steps": []},
              {"title": "Then I see the heading \"About\"", "duration": 17, "category": null, "steps": []},
              {"title": "Before Hooks", "duration": 80, "category": null, "steps": []}
            ]
          }]
        }]
      }],
      "specs2": []
    }]
  }]
}`

func TestParsePlaywrightJSON_ExtractsGherkinSteps(t *testing.T) {
	idx := ParsePlaywrightJSON([]byte(pwReportFixture))
	if len(idx) != 1 {
		t.Fatalf("expected 1 scenario, got %d", len(idx))
	}
	rec, ok := idx["visits the about page"]
	if !ok {
		t.Fatalf("scenario name lookup failed: %+v", idx)
	}
	if rec.Status != "passed" {
		t.Errorf("status: got %q", rec.Status)
	}
	if rec.DurationMs != 4321 {
		t.Errorf("duration: got %d", rec.DurationMs)
	}
	if len(rec.Steps) != 3 {
		t.Errorf("expected 3 Gherkin steps (non-Gherkin 'Before Hooks' filtered out); got %d: %+v", len(rec.Steps), rec.Steps)
	}
	if rec.Steps[0].Title != "Given I open the landing page" {
		t.Errorf("first step title: got %q", rec.Steps[0].Title)
	}
	if rec.Steps[0].Status != "passed" {
		t.Errorf("first step status: got %q", rec.Steps[0].Status)
	}
}

func TestLastRunIndex_RoundTrip(t *testing.T) {
	root := t.TempDir()
	idx := LastRunIndex{
		"convert journey terminal": {
			Status:     "passed",
			At:         "2026-06-17T20:32:44Z",
			DurationMs: 4321,
			Steps: []StepResult{
				{Title: "Given I open the landing page", Status: "passed", DurationMs: 120},
			},
		},
	}
	if err := writeLastRunIndex(root, idx); err != nil {
		t.Fatal(err)
	}
	loaded := LoadLastRunIndex(root)
	if got, want := loaded["convert journey terminal"].Status, "passed"; got != want {
		t.Errorf("status round-trip: got %q want %q", got, want)
	}
	if len(loaded["convert journey terminal"].Steps) != 1 {
		t.Errorf("steps lost in round-trip")
	}
}

func TestLastRunIndex_MissingFileReturnsEmpty(t *testing.T) {
	idx := LoadLastRunIndex(t.TempDir())
	if len(idx) != 0 {
		t.Errorf("expected empty index for missing file; got %d", len(idx))
	}
}

func TestAcquireRun_SerialisesPerWorkdir(t *testing.T) {
	wd := "/tmp/some-workdir"
	defer releaseRun(wd)
	if !acquireRun(wd) {
		t.Fatal("first acquire should succeed")
	}
	if acquireRun(wd) {
		t.Errorf("second acquire should fail while the first is in-flight")
	}
	releaseRun(wd)
	if !acquireRun(wd) {
		t.Errorf("acquire after release should succeed")
	}
}
