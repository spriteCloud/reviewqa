package main

import (
	"strings"
	"testing"
)

const fixture = `time=2026-06-16T23:13:27.021+02:00 level=INFO msg="llm humanization disabled"
--- tests/e2e/example-com-convert.spec.ts ---
test.describe('Example', () => {
  test('convert', async ({ page }) => {
    // Step 1 — visit https://example.com (intent: form)
    await page.goto(TARGET)
    await expect(page).toHaveTitle(/Example/i)
    await page.getByPlaceholder('Email').first().fill('test@example.com')
  })
})
--- tests/e2e/example-com-browse-blog.spec.ts ---
test.describe('Example', () => {
  test('browse', async ({ page }) => {
    // Step 1 — visit https://example.com (intent: nav)
    await page.goto(TARGET)
    await expect(page).toHaveTitle(/Example/i)
    await expect(page.getByRole('heading', { level: 1, name: /Home/i }).first()).toBeVisible()
    // Step 2 — click "/blog" and land on the next page (intent: nav)
    {
      const link = page.locator('a[href="/blog"]').first()
      await link.click()
    }
    await expect(page).toHaveTitle(/Blog/i)
    await expect(page.getByRole('heading', { level: 1, name: /Blog/i }).first()).toBeVisible()
    // Step 3 — outbound click "/contact"
    {
      const link = page.locator('a[href="/contact"]').first()
      await link.click()
    }
    await expect(page.getByRole('banner').first()).toBeVisible()
  })
})
--- tests/e2e/example-com-explore-contact.spec.ts ---
test.describe('Example', () => {
  test('explore', async ({ page }) => {
    // Step 1 — visit https://example.com (intent: nav)
    await page.goto(TARGET)
    // Step 2 — click "/contact" and land on the next page (intent: nav)
    {
      const link = page.locator('a[href="/contact"]').first()
      await link.click()
    }
  })
})
`

func TestParseSpecs_Shape(t *testing.T) {
	specs := parseSpecs(strings.NewReader(fixture))
	if len(specs) != 3 {
		t.Fatalf("expected 3 specs; got %d", len(specs))
	}

	convert := specs[0]
	if convert.kind != "convert" {
		t.Errorf("expected kind=convert; got %q", convert.kind)
	}
	if !convert.hasFillCall {
		t.Error("expected hasFillCall=true on convert spec")
	}

	browse := specs[1]
	if browse.kind != "browse" {
		t.Errorf("expected kind=browse; got %q", browse.kind)
	}
	if !browse.hasH1Assertion {
		t.Error("expected hasH1Assertion=true on browse spec")
	}
	if browse.hasFillCall {
		t.Error("did not expect hasFillCall on browse spec")
	}
	// Step 2 has h1 (non-banner); Step 3 is banner-only.
	if len(browse.chainedSteps) != 2 {
		t.Fatalf("expected 2 chained steps; got %d", len(browse.chainedSteps))
	}
	if browse.chainedSteps[0].bannerOnly {
		t.Error("Step 2 should not be banner-only (it has an h1)")
	}
	if !browse.chainedSteps[1].bannerOnly {
		t.Error("Step 3 should be banner-only")
	}

	explore := specs[2]
	if len(explore.chainedSteps) != 1 {
		t.Fatalf("expected 1 chained step; got %d", len(explore.chainedSteps))
	}
	if !explore.chainedSteps[0].empty {
		t.Error("Step 2 of the explore spec has no assertions — should be empty=true")
	}
}

func TestRenderReport_HighlightsLeakageAndCoverage(t *testing.T) {
	specs := parseSpecs(strings.NewReader(fixture))
	report := renderReport("test.example", specs)
	if !strings.Contains(report, "## site: test.example") {
		t.Error("report missing site header")
	}
	if !strings.Contains(report, "| 3 | ") {
		t.Errorf("report missing total spec count of 3; got: %q", report)
	}
	// Only the browse spec in the fixture has an h1 assertion.
	if !strings.Contains(report, "1/3") {
		t.Errorf("report missing h1 coverage 1/3; got: %q", report)
	}
}

func TestRenderReport_EmptyInput(t *testing.T) {
	report := renderReport("nothing", nil)
	if !strings.Contains(report, "No specs emitted") {
		t.Errorf("expected empty-input message; got: %q", report)
	}
}

func TestKindFromFile_HandlesMultiTokenHost(t *testing.T) {
	tests := map[string]string{
		"spritecloud-com-convert.spec.ts":                      "convert",
		"books-toscrape-com-browse-books-category.spec.ts":     "browse",
		"es-wikipedia-org-explore-madrid.spec.ts":              "explore",
		"playwright-dev-read-docs-installation.spec.ts":        "read",
		"something-weird-without-a-kind.spec.ts":               "unknown",
	}
	for file, want := range tests {
		got := kindFromFile(file)
		if got != want {
			t.Errorf("kindFromFile(%q) = %q; want %q", file, got, want)
		}
	}
}
