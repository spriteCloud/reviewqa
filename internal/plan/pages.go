package plan

import (
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/reviewqa/reviewqa/internal/ast"
	"github.com/reviewqa/reviewqa/internal/diff"
)

// pageRoot is a candidate page file: either a static HTML page or a TSX/JSX
// entry that mounts React components.
type pageRoot struct {
	Path     string         // workdir-relative path
	Stem     string         // basename without extension
	URL      string         // derived page URL (e.g. "/", "/home")
	IsHTML   bool           // true for .html / .htm
	Mounted  []string       // component names referenced by JSX usage (for tsx/jsx roots)
	Anchors  []ast.LocatorAnchor // direct anchors found in the page source (HTML or JSX)
}

var (
	reJSXComponentUsage = regexp.MustCompile(`<([A-Z][\w$]*)\b`)
	rePageHTMLTestID    = regexp.MustCompile(`data-testid\s*=\s*['"]([^'"]+)['"]`)
	rePageHTMLAria      = regexp.MustCompile(`aria-label\s*=\s*['"]([^'"]+)['"]`)
	rePageHTMLRole      = regexp.MustCompile(`role\s*=\s*['"]([^'"]+)['"]`)
	rePageHTMLTag       = regexp.MustCompile(`<\s*([a-zA-Z][\w-]*)`)
)

// findPageRoots walks workDir for candidate page entry files and returns the
// detected roots with their mounted component names and anchors prepopulated.
// Walks are bounded — depth ≤ 6, common artifact dirs skipped.
func findPageRoots(workDir string) []pageRoot {
	if workDir == "" {
		workDir = "."
	}
	abs, err := filepath.Abs(workDir)
	if err != nil {
		abs = workDir
	}
	skip := map[string]bool{
		"node_modules": true, "dist": true, "build": true, ".next": true,
		"vendor": true, ".git": true, "target": true, "out": true,
	}
	var roots []pageRoot
	_ = filepath.WalkDir(abs, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return nil
		}
		rel, _ := filepath.Rel(abs, path)
		depth := strings.Count(filepath.ToSlash(rel), "/")
		if depth > 6 {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if d.IsDir() {
			if skip[d.Name()] {
				return filepath.SkipDir
			}
			return nil
		}
		if !isPageFile(rel, d.Name()) {
			return nil
		}
		content, err := os.ReadFile(path)
		if err != nil {
			return nil
		}
		stem := strings.TrimSuffix(d.Name(), filepath.Ext(d.Name()))
		root := pageRoot{
			Path:   filepath.ToSlash(rel),
			Stem:   stem,
			URL:    deriveURL(filepath.ToSlash(rel), stem),
			IsHTML: isHTMLExt(d.Name()),
		}
		if root.IsHTML {
			root.Anchors = extractHTMLAnchors(root.Path, content)
		} else {
			root.Mounted = extractJSXComponentNames(content)
		}
		roots = append(roots, root)
		return nil
	})
	sort.Slice(roots, func(i, j int) bool {
		return len(roots[i].Path) < len(roots[j].Path)
	})
	return roots
}

func isPageFile(rel, name string) bool {
	slash := "/" + filepath.ToSlash(rel) // leading slash so /pages/ matches even at the root
	lower := strings.ToLower(name)
	if lower == "index.html" || lower == "index.htm" {
		return true
	}
	switch name {
	case "index.tsx", "index.jsx", "App.tsx", "App.jsx", "routes.tsx", "routes.jsx":
		return true
	}
	if strings.Contains(slash, "/pages/") && (strings.HasSuffix(lower, ".tsx") || strings.HasSuffix(lower, ".jsx")) {
		return true
	}
	if strings.Contains(slash, "/app/") && (lower == "page.tsx" || lower == "page.jsx") {
		return true
	}
	return false
}

func isHTMLExt(name string) bool {
	ext := strings.ToLower(filepath.Ext(name))
	return ext == ".html" || ext == ".htm"
}

// deriveURL maps a page file path to its canonical relative URL.
//
//	index.html              -> "/"
//	pages/index.tsx         -> "/"
//	pages/About.tsx         -> "/about"
//	app/dashboard/page.tsx  -> "/dashboard"
func deriveURL(slashPath, stem string) string {
	lower := "/" + strings.ToLower(slashPath) // leading slash for substring matches
	switch {
	case strings.HasSuffix(lower, "/index.html"),
		strings.HasSuffix(lower, "/pages/index.tsx"),
		strings.HasSuffix(lower, "/pages/index.jsx"):
		return "/"
	case strings.HasSuffix(lower, "/app.tsx"), strings.HasSuffix(lower, "/app.jsx"),
		strings.HasSuffix(lower, "/routes.tsx"), strings.HasSuffix(lower, "/routes.jsx"):
		return "/"
	}
	if i := strings.Index(lower, "/app/"); i != -1 && (strings.HasSuffix(lower, "/page.tsx") || strings.HasSuffix(lower, "/page.jsx")) {
		dir := lower[i+len("/app/"):]
		dir = strings.TrimSuffix(dir, "/page.tsx")
		dir = strings.TrimSuffix(dir, "/page.jsx")
		if dir == "" {
			return "/"
		}
		return "/" + dir
	}
	if strings.Contains(lower, "/pages/") {
		return "/" + strings.ToLower(stem)
	}
	return "/"
}

func extractJSXComponentNames(content []byte) []string {
	seen := map[string]bool{}
	var out []string
	for _, m := range reJSXComponentUsage.FindAllSubmatch(content, -1) {
		name := string(m[1])
		if !seen[name] {
			seen[name] = true
			out = append(out, name)
		}
	}
	return out
}

func capitalizeFirst(s string) string {
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

func extractHTMLAnchors(file string, content []byte) []ast.LocatorAnchor {
	var anchors []ast.LocatorAnchor
	lines := strings.Split(string(content), "\n")
	for i, line := range lines {
		tag := ""
		if m := rePageHTMLTag.FindStringSubmatch(line); m != nil {
			tag = strings.ToLower(m[1])
		}
		if m := rePageHTMLTestID.FindStringSubmatch(line); m != nil {
			anchors = append(anchors, ast.LocatorAnchor{TestID: m[1], File: file, Line: i + 1, Tag: tag})
		}
		if m := rePageHTMLAria.FindStringSubmatch(line); m != nil {
			anchors = append(anchors, ast.LocatorAnchor{Aria: m[1], File: file, Line: i + 1, Tag: tag})
		}
		if m := rePageHTMLRole.FindStringSubmatch(line); m != nil {
			anchors = append(anchors, ast.LocatorAnchor{Role: m[1], File: file, Line: i + 1, Tag: tag})
		}
	}
	return anchors
}

// groupByPage post-processes a build's items. It folds component E2E items
// (TmplPlaywrightE2E) into page-scoped happy-flow items where the components
// are mounted on the same page root, and emits a happy-flow item for each
// static HTML page touched by the diff.
func groupByPage(items []Item, files []diff.File, layout Layout) []Item {
	style := os.Getenv("REVIEWQA_E2E_STYLE")
	roots := findPageRoots(layout.WorkDir)
	if len(roots) == 0 && style != "page-flow" {
		return items
	}

	// 1. Group TSX components by which page root mounts them.
	type group struct {
		root    pageRoot
		symbols []ast.Symbol
	}
	groupedByRoot := map[string]*group{}
	var nonE2E []Item
	var ungroupedE2E []Item
	for _, it := range items {
		if it.Template != TmplPlaywrightE2E {
			nonE2E = append(nonE2E, it)
			continue
		}
		var matched *pageRoot
		for i := range roots {
			r := &roots[i]
			if r.IsHTML {
				continue
			}
			for _, n := range r.Mounted {
				if n == it.Symbol.Name {
					matched = r
					break
				}
			}
			if matched != nil {
				break
			}
		}
		if matched == nil {
			ungroupedE2E = append(ungroupedE2E, it)
			continue
		}
		g, ok := groupedByRoot[matched.Path]
		if !ok {
			g = &group{root: *matched}
			groupedByRoot[matched.Path] = g
		}
		g.symbols = append(g.symbols, it.Symbol)
	}

	// 2. Emit one happy-flow item per group. Auto-mode requires ≥2 symbols;
	// page-flow forces grouping for any size; per-component would have been
	// short-circuited upstream.
	var out []Item
	out = append(out, nonE2E...)
	for _, g := range groupedByRoot {
		if len(g.symbols) < 2 && style != "page-flow" {
			for _, s := range g.symbols {
				ungroupedE2E = append(ungroupedE2E, Item{
					Symbol:   s,
					Template: TmplPlaywrightE2E,
					OutPath:  filepath.ToSlash(filepath.Join("tests", "e2e", s.Name+".spec.ts")),
				})
			}
			continue
		}
		sort.Slice(g.symbols, func(i, j int) bool { return g.symbols[i].Line < g.symbols[j].Line })
		out = append(out, Item{
			Symbol:   g.symbols[0],
			Symbols:  g.symbols,
			PageURL:  g.root.URL,
			Template: TmplPlaywrightHappyFlow,
			OutPath:  filepath.ToSlash(filepath.Join("tests", "e2e", g.root.Stem+".spec.ts")),
		})
	}
	out = append(out, ungroupedE2E...)

	// 3. Emit happy-flow items for static HTML pages touched by the diff.
	htmlInDiff := map[string]bool{}
	for _, f := range files {
		if f.Status == "removed" {
			continue
		}
		if isHTMLExt(f.Path) {
			htmlInDiff[filepath.ToSlash(f.Path)] = true
		}
	}
	for _, r := range roots {
		if !r.IsHTML || !htmlInDiff[r.Path] || len(r.Anchors) == 0 {
			continue
		}
		synthetic := ast.Symbol{
			Name:     capitalizeFirst(r.Stem),
			Kind:     ast.KindComponent,
			File:     r.Path,
			Language: "ts",
			Anchors:  r.Anchors,
		}
		out = append(out, Item{
			Symbol:   synthetic,
			Symbols:  []ast.Symbol{synthetic},
			PageURL:  r.URL,
			Template: TmplPlaywrightHappyFlow,
			OutPath:  filepath.ToSlash(filepath.Join("tests", "e2e", r.Stem+".spec.ts")),
		})
	}
	return out
}
