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
