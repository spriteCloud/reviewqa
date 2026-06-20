package serve

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const fxFeatureForRegen = `Feature: spritecloud

  Scenario: visit contact
    Given I open the page "/contact"
    Then the main heading reads "Contact"
`

const fxStepsForRegen = `import { expect } from '@playwright/test'
import { createBdd } from 'playwright-bdd'
import { steps } from '../lib/steps'

const { Given, When, Then } = createBdd()

Given(/^I open the page "([^"]+)"$/, async ({ page }, path: string) => {
  await steps.visit(page, path)
})

Then(/^the main heading reads "([^"]+)"$/, async ({ page }, text: string) => {
  await steps.expectH1(page, text)
})
`

func setupRegenFixture(t *testing.T) (workdir, featurePath, history string) {
	t.Helper()
	tmp := t.TempDir()
	featuresDir := filepath.Join(tmp, "tests", "e2e", "features")
	stepsDir := filepath.Join(tmp, "tests", "e2e", "steps")
	for _, d := range []string{featuresDir, stepsDir} {
		if err := os.MkdirAll(d, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	fp := filepath.Join(featuresDir, "x.feature")
	if err := os.WriteFile(fp, []byte(fxFeatureForRegen), 0o644); err != nil {
		t.Fatal(err)
	}
	sp := filepath.Join(stepsDir, "quail.steps.ts")
	if err := os.WriteFile(sp, []byte(fxStepsForRegen), 0o644); err != nil {
		t.Fatal(err)
	}
	return tmp, fp, filepath.Join(tmp, ".quail-history")
}

func TestReplaceScenarioWithStepRegen_StepsMatchExisting(t *testing.T) {
	workdir, fp, hist := setupRegenFixture(t)
	// Same patterns, different concrete values — should fall through
	// to plain ReplaceScenario without LLM contact.
	newBlock := `  Scenario: visit about
    Given I open the page "/about"
    Then the main heading reads "About"
`
	res, err := ReplaceScenarioWithStepRegen(
		context.Background(), nil, workdir, fp, "visit contact", newBlock, hist,
	)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if res.StepsUpdated {
		t.Errorf("expected StepsUpdated=false when steps match existing patterns")
	}
	if !strings.Contains(res.Note, "no regen needed") {
		t.Errorf("unexpected note: %q", res.Note)
	}
	updated, _ := os.ReadFile(fp)
	if !strings.Contains(string(updated), `visit about`) {
		t.Errorf("feature not updated")
	}
}

func TestReplaceScenarioWithStepRegen_LLMDisabled_NewPhrasing(t *testing.T) {
	workdir, fp, hist := setupRegenFixture(t)
	// "I visit the X page" is NOT one of the registered patterns.
	// With nil LLM client we should fall back to plain ReplaceScenario
	// and surface a warning Note.
	newBlock := `  Scenario: visit contact reworded
    Given I visit the "/contact" page
    Then the main heading reads "Contact"
`
	res, err := ReplaceScenarioWithStepRegen(
		context.Background(), nil, workdir, fp, "visit contact", newBlock, hist,
	)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if res.StepsUpdated {
		t.Errorf("expected StepsUpdated=false when LLM is disabled")
	}
	if !strings.Contains(res.Note, "LLM is disabled") {
		t.Errorf("expected warning Note about disabled LLM; got %q", res.Note)
	}
}

func TestReplaceScenarioWithStepRegen_MissingStepsFile(t *testing.T) {
	workdir, fp, hist := setupRegenFixture(t)
	// Remove steps.ts; should fall back to legacy ReplaceScenario.
	if err := os.Remove(filepath.Join(workdir, "tests", "e2e", "steps", "quail.steps.ts")); err != nil {
		t.Fatal(err)
	}
	newBlock := `  Scenario: visit about
    Given I open the page "/about"
    Then the main heading reads "About"
`
	res, err := ReplaceScenarioWithStepRegen(
		context.Background(), nil, workdir, fp, "visit contact", newBlock, hist,
	)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if res.StepsUpdated {
		t.Errorf("expected StepsUpdated=false when steps.ts missing")
	}
	if !strings.Contains(res.Note, "not found") {
		t.Errorf("expected note about missing steps; got %q", res.Note)
	}
}
