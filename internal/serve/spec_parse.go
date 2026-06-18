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
//
// Describe is the most recent test.describe() title that wraps
// this test in source order — empty when the test is top-level.
// The UI groups tests by Describe so the Tests view reads as
// describe → tests, the same shape as feature → scenarios.
//
// LastRun is the last-run verdict joined in from the workdir's
// LastRunIndex on load — same shape Scenarios use.
type SpecTest struct {
	Name     string         `json:"name"`
	Kind     string         `json:"kind,omitempty"` // "test" | "describe"
	Describe string         `json:"describe,omitempty"`
	LastRun  *LastRunRecord `json:"lastRun,omitempty"`
}

// testRe matches Playwright/Jest test calls — `test(…)`, `it(…)`,
// `test.describe(…)`, plus `.only`, `.skip`, `.fixme` variants. Go's
// regexp engine lacks backreferences, so we alternate over the three
// quote shapes and strip the wrappers at match time.
var testRe = regexp.MustCompile(`(?m)\b(test|it)(?:\.(only|skip|fixme|describe))?\s*\(\s*(?:'([^']+)'|"([^"]+)"|` + "`([^`]+)`" + `)`)

// parseSpecFile reads a Playwright spec file and extracts every
// test() / it() / describe() call. Returns the test list. Errors
// from the read (missing file, permission) surface to the caller.
// describeRange tracks one test.describe() block's source span.
type describeRange struct {
	title string
	start int // byte offset of the describe() call
	end   int // byte offset just past the matching `})`
}

// findBlockEnd returns the byte offset just past the closing `}`
// of the block opened by the first `{` at or after `from`. Returns
// len(src) when the source is unbalanced.
func findBlockEnd(src string, from int) int {
	depth := 0
	started := false
	for i := from; i < len(src); i++ {
		switch src[i] {
		case '{':
			depth++
			started = true
		case '}':
			depth--
			if started && depth == 0 {
				return i + 1
			}
		}
	}
	return len(src)
}

func parseSpecFile(path string) ([]SpecTest, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	src := string(b)
	matchPositions := testRe.FindAllStringSubmatchIndex(src, -1)

	// First pass: pick up every describe with its source range so
	// we can ask "which describe contains this test offset?" later.
	// Each entry in matchPositions has the same shape as the
	// matchgroups in testRe; pairs at indices 4..9 are the three
	// quoted-title groups (start, end).
	var describes []describeRange
	for _, mp := range matchPositions {
		if mp[4] < 0 && mp[6] < 0 && mp[8] < 0 {
			continue
		}
		// Modifier group is m[2] at indices [4..5].
		modStart, modEnd := mp[4], mp[5]
		_ = modStart
		_ = modEnd
		// Actually m[2] is at indices 4,5; titles begin at 6.
		// matchPositions indices for our regex:
		//   0,1 = full match
		//   2,3 = m[1] "test"|"it"
		//   4,5 = m[2] modifier
		//   6,7 = m[3] single-quoted title
		//   8,9 = m[4] double-quoted title
		//   10,11 = m[5] backtick title
		modifier := ""
		if mp[4] >= 0 {
			modifier = src[mp[4]:mp[5]]
		}
		title := pickTitleFromPositions(src, mp)
		if title == "" || modifier != "describe" {
			continue
		}
		describes = append(describes, describeRange{
			title: title,
			start: mp[0],
			end:   findBlockEnd(src, mp[1]),
		})
	}

	// Second pass: emit each non-describe test with its innermost
	// containing describe (last-match-wins for nested describes).
	seen := make(map[string]bool, len(matchPositions))
	out := make([]SpecTest, 0, len(matchPositions))
	for _, mp := range matchPositions {
		modifier := ""
		if mp[4] >= 0 {
			modifier = src[mp[4]:mp[5]]
		}
		if modifier == "describe" {
			continue
		}
		title := pickTitleFromPositions(src, mp)
		if title == "" || seen[title] {
			continue
		}
		seen[title] = true
		// Find innermost describe containing this match.
		describe := ""
		for _, d := range describes {
			if mp[0] >= d.start && mp[0] < d.end {
				describe = d.title
			}
		}
		out = append(out, SpecTest{Name: title, Kind: "test", Describe: describe})
	}
	return out, nil
}

// pickTitleFromPositions returns the trimmed title text from
// whichever of the three quoted-group captures matched.
func pickTitleFromPositions(src string, mp []int) string {
	for _, pair := range [][2]int{{6, 7}, {8, 9}, {10, 11}} {
		if mp[pair[0]] >= 0 && mp[pair[1]] >= 0 {
			return strings.TrimSpace(src[mp[pair[0]]:mp[pair[1]]])
		}
	}
	return ""
}

// loadSpecs walks the workdir's Playwright spec roots (tests/, e2e/,
// playwright/, spec/, __tests__/), parses every *.spec.* and
// *.test.* file, and returns them sorted by path.
//
// Quiet on errors — a spec dir might not exist, or a single file
// might be unreadable; we skip and keep going.
//
// v0.84: returns nil for reviewqa-generated projects. The
// `tests/e2e/*.spec.ts` files under such a project are layer
// artifacts (a11y / mobile / contract / security / …) emitted
// alongside the .feature files; they're already covered by Run
// from the feature view and surfacing them as a separate "Tests"
// sidebar section just clutters the UI.
func loadSpecs(workdir string) []SpecRef {
	if isReviewqaProject(workdir) {
		return nil
	}
	idx := LoadLastRunIndex(workdir)
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
			// Join in the last-run verdict per test (keyed by name —
			// same key Playwright uses for --grep). Mirrors how
			// handleFeature populates Scenario.LastRun.
			for i := range tests {
				if rec, ok := idx[tests[i].Name]; ok {
					rec := rec
					tests[i].LastRun = &rec
				}
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

// isReviewqaProject returns true ONLY for reviewqa-generated
// workdirs — those with a .feature file or the reviewqa steps
// module on disk. Distinct from `looksLikeReviewqaProject` (which
// is broader and ALSO accepts vanilla Playwright projects), this
// one says "yes, reviewqa emitted into this tree." Used by
// loadSpecs to skip the Tests sidebar section for native projects.
func isReviewqaProject(workdir string) bool {
	if _, err := os.Stat(filepath.Join(workdir, "tests", "e2e", "steps", "reviewqa.steps.ts")); err == nil {
		return true
	}
	if matches, _ := filepath.Glob(filepath.Join(workdir, "tests", "e2e", "features", "*.feature")); len(matches) > 0 {
		return true
	}
	return false
}

func isSpecFile(name string) bool {
	for _, suf := range []string{".spec.ts", ".spec.js", ".spec.mts", ".spec.mjs", ".test.ts", ".test.js"} {
		if strings.HasSuffix(name, suf) {
			return true
		}
	}
	return false
}
