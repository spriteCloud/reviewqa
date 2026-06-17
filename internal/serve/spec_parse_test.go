package serve

import (
	"os"
	"path/filepath"
	"testing"
)

const sampleSpecTS = `
import { test, expect } from '@playwright/test'

test.describe('signup flow', () => {
  test('happy path', async ({ page }) => {
    await page.goto('https://example.com')
    await expect(page).toHaveTitle(/example/i)
  })

  test.only('focused', async ({ page }) => {})
  test.skip('flaky for now', async ({ page }) => {})
  test.fixme('broken', async ({ page }) => {})
})

test('top-level test', async ({ page }) => {})

// it() alias from playwright-extra etc.
it('legacy mocha-style', async ({ page }) => {})
`

func TestParseSpecFile_ExtractsTitles(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "example.spec.ts")
	if err := os.WriteFile(path, []byte(sampleSpecTS), 0o644); err != nil {
		t.Fatal(err)
	}
	tests, err := parseSpecFile(path)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	wantNames := []string{"signup flow", "happy path", "focused", "flaky for now", "broken", "top-level test", "legacy mocha-style"}
	if len(tests) != len(wantNames) {
		t.Fatalf("got %d tests, want %d: %+v", len(tests), len(wantNames), tests)
	}
	for i, w := range wantNames {
		if tests[i].Name != w {
			t.Errorf("[%d] name = %q, want %q", i, tests[i].Name, w)
		}
	}
}

func TestLoadSpecs_FindsFiles(t *testing.T) {
	root := t.TempDir()
	specDir := filepath.Join(root, "tests")
	if err := os.MkdirAll(specDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(specDir, "demo.spec.ts"), []byte(sampleSpecTS), 0o644); err != nil {
		t.Fatal(err)
	}
	specs := loadSpecs(root)
	if len(specs) != 1 {
		t.Fatalf("specs = %d, want 1: %+v", len(specs), specs)
	}
	if specs[0].Path != "tests/demo.spec.ts" {
		t.Errorf("path = %q, want tests/demo.spec.ts", specs[0].Path)
	}
	if len(specs[0].Tests) == 0 {
		t.Errorf("no tests extracted")
	}
}

func TestLooksLikeReviewqaProject_VanillaPlaywright(t *testing.T) {
	dir := t.TempDir()
	// Just a playwright.config.ts is enough.
	if err := os.WriteFile(filepath.Join(dir, "playwright.config.ts"), []byte("export default { testDir: './tests' }"), 0o644); err != nil {
		t.Fatal(err)
	}
	if !looksLikeReviewqaProject(dir) {
		t.Errorf("vanilla playwright project should be accepted")
	}
}

func TestLooksLikeReviewqaProject_SpecRootOnly(t *testing.T) {
	dir := t.TempDir()
	specDir := filepath.Join(dir, "e2e")
	if err := os.MkdirAll(specDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(specDir, "a.spec.ts"), []byte(`test('x', () => {})`), 0o644); err != nil {
		t.Fatal(err)
	}
	if !looksLikeReviewqaProject(dir) {
		t.Errorf("dir with e2e/*.spec.ts should be accepted")
	}
}
