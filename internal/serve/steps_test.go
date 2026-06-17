package serve

import "testing"

const sampleSteps = `import { expect } from '@playwright/test'
import { createBdd } from 'playwright-bdd'

const { Given, When, Then } = createBdd()

Given('I open the landing page', async ({ page }) => {
  await steps.visit(page, '/')
})

Given(/^I open the page "([^"]+)"$/, async ({ page }, path: string) => {
  await steps.visit(page, path)
})

When(/^I click the link to "([^"]+)"$/, async ({ page }, href: string) => {
  await page.goto(href)
})

Then('the form is not double-submitted', async ({ page }) => {})
`

func TestParseStepsBytes_ExtractsLiteralsAndRegex(t *testing.T) {
	out := ParseStepsBytes([]byte(sampleSteps))
	if len(out) != 4 {
		t.Fatalf("expected 4 patterns, got %d: %+v", len(out), out)
	}

	wantPattern := []string{
		"I open the landing page",
		`^I open the page "([^"]+)"$`,
		`^I click the link to "([^"]+)"$`,
		"the form is not double-submitted",
	}
	wantKW := []string{"Given", "Given", "When", "Then"}
	wantRegex := []bool{false, true, true, false}

	for i, p := range out {
		if p.Pattern != wantPattern[i] {
			t.Errorf("pattern[%d]: got %q want %q", i, p.Pattern, wantPattern[i])
		}
		if p.Keyword != wantKW[i] {
			t.Errorf("keyword[%d]: got %q want %q", i, p.Keyword, wantKW[i])
		}
		if p.IsRegex != wantRegex[i] {
			t.Errorf("isRegex[%d]: got %v want %v", i, p.IsRegex, wantRegex[i])
		}
	}
}

func TestParseStepsBytes_IgnoresNonTopLevel(t *testing.T) {
	// Indented calls (inside a function body etc.) should not be picked
	// up — playwright-bdd registers patterns at module level.
	out := ParseStepsBytes([]byte(`function noop() { Given('inner', () => {}) }`))
	if len(out) != 0 {
		t.Fatalf("expected indented Given to be ignored, got %d: %+v", len(out), out)
	}
}
