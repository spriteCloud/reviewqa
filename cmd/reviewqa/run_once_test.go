package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRunOnce_NoE2EDir_Errors(t *testing.T) {
	tmp := t.TempDir()
	err := runOnce(tmp, false, filepath.Join(tmp, "report.json"), "")
	if err == nil {
		t.Fatal("runOnce should error when tests/e2e/ does not exist")
	}
}

func TestMergeReportIntoLedger_AppendsFindings(t *testing.T) {
	tmp := t.TempDir()
	if err := os.MkdirAll(filepath.Join(tmp, "tests", "e2e", "docs"), 0o755); err != nil {
		t.Fatal(err)
	}
	report := `{
  "suites": [
    {
      "title": "demo.spec.ts",
      "file": "demo.spec.ts",
      "specs": [
        {"title": "broken bit", "file": "demo.spec.ts", "tests": [
          {"projectName": "chromium", "results": [{"status": "failed", "errors": [{"message": "expected blah"}]}]}
        ]}
      ]
    }
  ]
}`
	reportPath := filepath.Join(tmp, "report.json")
	if err := os.WriteFile(reportPath, []byte(report), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := mergeReportIntoLedger(tmp, reportPath); err != nil {
		t.Fatalf("mergeReportIntoLedger: %v", err)
	}
	ledger, err := os.ReadFile(filepath.Join(tmp, "tests", "e2e", "docs", "findings.md"))
	if err != nil {
		t.Fatal(err)
	}
	if len(ledger) == 0 {
		t.Error("expected ledger to be populated after merge")
	}
}

func TestNewRunOnceCmd_Wired(t *testing.T) {
	cmd := newRunOnceCmd()
	if cmd.Use != "run-once" {
		t.Errorf("Use = %q; want run-once", cmd.Use)
	}
	// Flag presence check.
	for _, flag := range []string{"workdir", "record", "report", "grep"} {
		if cmd.Flags().Lookup(flag) == nil {
			t.Errorf("--%s flag missing", flag)
		}
	}
}
