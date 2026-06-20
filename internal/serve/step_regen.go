package serve

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spriteCloud/quail/internal/composer"
	"github.com/spriteCloud/quail/internal/llm"
)

// stepsRelPath is the canonical step-definitions file emitted by the
// quail probe. ReplaceScenarioWithStepRegen reads it (and optionally
// writes it back) so a UI edit to a step phrase can update its
// matching pattern in lockstep.
const stepsRelPath = "tests/e2e/steps/quail.steps.ts"

// StepRegenResult is the per-edit outcome returned to the serve
// handler so the UI can toast the user about what changed.
type StepRegenResult struct {
	// NewName is the (possibly renamed) scenario after the edit.
	NewName string
	// StepsUpdated is true when the matching pattern in quail.steps.ts
	// was rewritten in lockstep with the .feature edit.
	StepsUpdated bool
	// Note is a human-readable summary of what happened (used by the
	// UI toast).
	Note string
}

// ReplaceScenarioWithStepRegen replaces a scenario in a .feature file,
// and — if the new Gherkin steps fall outside the registered patterns
// in tests/e2e/steps/quail.steps.ts — routes the change through the
// paired-humanize machinery so the .feature and .steps.ts stay in
// lockstep.
//
// Behavior:
//   - Steps all match existing patterns → behaves exactly like
//     ReplaceScenario (write .feature only).
//   - Some step doesn't match AND llmClient is enabled → call
//     HumanizeSuite. If it returns paired rewrites, atomically dual-
//     write .feature + .steps.ts (both backed up under historyRoot).
//   - HumanizeSuite fails the arity / binding guards (or LLM disabled)
//     → fall back to writing .feature only and surface a Note so the
//     UI can tell the user the test may be unrunnable until they edit
//     the step-defs.
//
// historyRoot is the parent for .quail-history backups (callers pass
// the same value they pass to ReplaceScenario).
//
// v0.96.0.
func ReplaceScenarioWithStepRegen(
	ctx context.Context,
	llmClient *llm.Client,
	workdir, featurePath, scenarioName, newBlock, historyRoot string,
) (StepRegenResult, error) {
	if err := validateScenarioBlock(newBlock); err != nil {
		return StepRegenResult{}, err
	}

	// 1) Read current steps.ts. If absent, fall back to the legacy
	// single-file path (no regen possible).
	stepsPath := filepath.Join(workdir, stepsRelPath)
	stepsContent, stepsErr := os.ReadFile(stepsPath)
	if stepsErr != nil {
		name, err := ReplaceScenario(featurePath, scenarioName, newBlock, historyRoot)
		return StepRegenResult{NewName: name, Note: "step-defs file not found; .feature updated only"}, err
	}
	patterns := composer.ExtractStepPatterns(stepsContent)

	// 2) Build the candidate new .feature in memory so we can run the
	// Gherkin-safe guard against the new step text.
	lines, err := readLines(featurePath)
	if err != nil {
		return StepRegenResult{}, err
	}
	rng, ok := findScenarioRange(lines, scenarioName)
	if !ok {
		return StepRegenResult{}, fmt.Errorf("scenario %q not found", scenarioName)
	}
	newFeatureLines := append([]string{}, lines[:rng.Start]...)
	newFeatureLines = append(newFeatureLines, splitLines(newBlock)...)
	newFeatureLines = append(newFeatureLines, lines[rng.End:]...)
	candidateFeature := []byte(strings.Join(newFeatureLines, "\n") + "\n")

	// 3) Steps all match existing patterns → write feature, done.
	if composer.IsGherkinSafeAgainst(candidateFeature, patterns) {
		name, err := ReplaceScenario(featurePath, scenarioName, newBlock, historyRoot)
		return StepRegenResult{NewName: name, Note: "steps match existing patterns; no regen needed"}, err
	}

	// 4) Some step doesn't match. If LLM is disabled, write feature
	// and surface a warning — user will need to update step-defs by
	// hand, or the test will be unrunnable.
	if llmClient == nil || !llmClient.Enabled() {
		name, err := ReplaceScenario(featurePath, scenarioName, newBlock, historyRoot)
		return StepRegenResult{
			NewName: name,
			Note:    "WARNING: new step phrasing doesn't match registered patterns and the LLM is disabled; the test may be unrunnable until you update tests/e2e/steps/quail.steps.ts manually",
		}, err
	}

	// 5) Pair-call HumanizeSuite to produce matching .feature + steps.ts.
	symbol := strings.TrimSuffix(filepath.Base(featurePath), filepath.Ext(featurePath))
	newFeatures, newSteps := llmClient.HumanizeSuite(ctx, symbol,
		[]llm.SuiteFile{{Path: featurePath, Body: candidateFeature}},
		llm.SuiteFile{Path: stepsPath, Body: stepsContent},
	)

	// 6) HumanizeSuite returns originals on any guard fail. Detect:
	// if newSteps.Body matches stepsContent byte-for-byte, the regen
	// didn't happen.
	stepsChanged := !bytesEqual(newSteps.Body, stepsContent)
	if !stepsChanged {
		name, err := ReplaceScenario(featurePath, scenarioName, newBlock, historyRoot)
		return StepRegenResult{
			NewName: name,
			Note:    "the LLM could not produce a safe step-def rewrite (guard tripped); .feature updated but test-defs may need a manual touch",
		}, err
	}

	// 7) Atomic dual-write. writeLinesAtomic backs up to .quail-history
	// before each write. If the second write fails we restore the first
	// from its just-made backup so the suite never lands half-edited.
	featureLinesOut := splitLines(string(newFeatures[0].Body))
	stepsLinesOut := splitLines(string(newSteps.Body))

	// .feature first (cheaper rollback).
	if err := writeLinesAtomic(featurePath, featureLinesOut, historyRoot); err != nil {
		return StepRegenResult{}, fmt.Errorf("write feature: %w", err)
	}
	if err := writeLinesAtomic(stepsPath, stepsLinesOut, historyRoot); err != nil {
		// Roll back the feature write from the backup we just made.
		if rerr := restoreFromHistory(featurePath, historyRoot); rerr != nil {
			return StepRegenResult{}, fmt.Errorf("write steps: %w (and restore failed: %v)", err, rerr)
		}
		return StepRegenResult{}, fmt.Errorf("write steps: %w (feature rolled back)", err)
	}
	newName := scenarioBlockName(string(newFeatures[0].Body))
	if newName == "" {
		newName = scenarioBlockName(newBlock)
	}
	return StepRegenResult{
		NewName:      newName,
		StepsUpdated: true,
		Note:         "step definitions regenerated in lockstep with feature edit",
	}, nil
}

func bytesEqual(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// restoreFromHistory copies the most recent .quail-history backup of
// path back over path. Used to roll back a successful first write when
// the second write of a paired edit fails.
func restoreFromHistory(path, historyRoot string) error {
	dir := historyRootDir(path, historyRoot)
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}
	// backupFile names use a timestamp prefix so lex-sort = chrono-sort.
	var latest string
	target := filepath.Base(path)
	for _, e := range entries {
		if strings.HasSuffix(e.Name(), "."+target) && e.Name() > latest {
			latest = e.Name()
		}
	}
	if latest == "" {
		return fmt.Errorf("no backup found in %s", dir)
	}
	raw, err := os.ReadFile(filepath.Join(dir, latest))
	if err != nil {
		return err
	}
	return os.WriteFile(path, raw, 0o644)
}

// historyRootDir mirrors backupFile's path computation. We can't call
// it directly because it's unexported and bundled with the file-write
// logic; the layout here ("workdir/.quail-history") is the canonical
// shape backupFile uses.
func historyRootDir(path, historyRoot string) string {
	if historyRoot == "" {
		historyRoot = filepath.Join(filepath.Dir(path), ".quail-history")
	}
	return historyRoot
}
