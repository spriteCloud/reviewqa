package serve

import (
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

// SpecRef is a vanilla Playwright spec file surfaced to the sidebar
// alongside the reviewqa-native FeatureRef. The shape mirrors
// FeatureRef so the frontend can render both with the same code.
type SpecRef struct {
	Path  string     `json:"path"`
	Name  string     `json:"name"`
	Tests []SpecTest `json:"tests,omitempty"`
}

// SpecTest is one test('…') or test.describe('…') call extracted
// from a vanilla Playwright spec file. The Name is the literal
// title argument; the UI uses it for the `--grep` value when Run
// is hit.
type SpecTest struct {
	Name string `json:"name"`
	Kind string `json:"kind,omitempty"` // "test" | "describe"
}

// testRe matches Playwright/Jest test calls — `test(…)`, `it(…)`,
// `test.describe(…)`, plus `.only`, `.skip`, `.fixme` variants. Go's
// regexp engine lacks backreferences, so we alternate over the three
// quote shapes and strip the wrappers at match time.
var testRe = regexp.MustCompile(`(?m)\b(test|it)(?:\.(only|skip|fixme|describe))?\s*\(\s*(?:'([^']+)'|"([^"]+)"|` + "`([^`]+)`" + `)`)

// parseSpecFile reads a Playwright spec file and extracts every
// test() / it() / describe() call. Returns the test list. Errors
// from the read (missing file, permission) surface to the caller.
func parseSpecFile(path string) ([]SpecTest, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	matches := testRe.FindAllStringSubmatch(string(b), -1)
	seen := make(map[string]bool, len(matches))
	out := make([]SpecTest, 0, len(matches))
	for _, m := range matches {
		// m[1] = "test"|"it"; m[2] = modifier (only|skip|fixme|describe) or "";
		// title is whichever of m[3..5] (the three quote groups) is non-empty.
		title := ""
		for _, g := range m[3:] {
			if g != "" {
				title = strings.TrimSpace(g)
				break
			}
		}
		if title == "" || seen[title] {
			continue
		}
		seen[title] = true
		kind := "test"
		if m[2] == "describe" {
			kind = "describe"
		}
		out = append(out, SpecTest{Name: title, Kind: kind})
	}
	return out, nil
}

// loadSpecs walks the workdir's Playwright spec roots (tests/, e2e/,
// playwright/, spec/, __tests__/), parses every *.spec.* and
// *.test.* file, and returns them sorted by path.
//
// Quiet on errors — a spec dir might not exist, or a single file
// might be unreadable; we skip and keep going.
func loadSpecs(workdir string) []SpecRef {
	var out []SpecRef
	roots := findSpecRoots(workdir)
	for _, root := range roots {
		_ = filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
			if err != nil || info.IsDir() {
				if info != nil && info.IsDir() {
					name := info.Name()
					if name == "node_modules" || name == ".git" || name == "dist" || name == "build" {
						return filepath.SkipDir
					}
				}
				return nil
			}
			if !isSpecFile(info.Name()) {
				return nil
			}
			tests, perr := parseSpecFile(path)
			if perr != nil || len(tests) == 0 {
				return nil
			}
			rel, _ := filepath.Rel(workdir, path)
			out = append(out, SpecRef{
				Path:  filepath.ToSlash(rel),
				Name:  filepath.Base(path),
				Tests: tests,
			})
			return nil
		})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Path < out[j].Path })
	return out
}

func isSpecFile(name string) bool {
	for _, suf := range []string{".spec.ts", ".spec.js", ".spec.mts", ".spec.mjs", ".test.ts", ".test.js"} {
		if strings.HasSuffix(name, suf) {
			return true
		}
	}
	return false
}
