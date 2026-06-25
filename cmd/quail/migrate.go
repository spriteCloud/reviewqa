package main

import (
	"bytes"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/spf13/cobra"
)

// newMigrateCmd is `quail migrate <workdir>` — a regex-only rewrite of
// .feature title and step strings that the v0.10.6 / v0.10.7 / v0.10.8
// quail-core releases changed format on. Re-probing would regenerate
// everything correctly, but it's slow (browser + LLM). This subcommand
// covers the cheap case: just update the strings in place.
//
// Targets only the well-defined transition points; other content stays
// untouched. --dry-run prints the diff without writing.
func newMigrateCmd() *cobra.Command {
	var dryRun bool
	var paths string
	cmd := &cobra.Command{
		Use:     "migrate [workdir]",
		Aliases: []string{"rewrite-titles"},
		Short:   "Rewrite .feature titles + Outline strings to the current quail-core format (v0.10.8).",
		Long: `Migrate old .feature files to the current format without re-probing.

Targets four well-defined string transitions:

  1. Feature: <X> — <kind> journey
     → Feature: <X> — <kind> · <URLPath>
     (URL path lifted from the "As a visitor of <URL>" line below)

  2. Scenario Outline: <selector> — <field> accepts <variant> [values]
     → Scenario Outline: fill the "<field>" field with <example>

  3. Inside @kind:component blocks:
     When I enter "<value>" into the "<field>" field
     → When I enter "<example>" into the "<field>" field

  4. Inside @kind:component blocks:
     | variant | value |
     → | case | example |

Manual edits elsewhere in the file are not touched.`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			workdir := "."
			if len(args) > 0 {
				workdir = args[0]
			}
			featuresDir := filepath.Join(workdir, paths)
			matches, err := filepath.Glob(filepath.Join(featuresDir, "*.feature"))
			if err != nil {
				return fmt.Errorf("glob %s: %w", featuresDir, err)
			}
			if len(matches) == 0 {
				return fmt.Errorf("no .feature files under %s", featuresDir)
			}
			changed, scanned := 0, 0
			for _, p := range matches {
				scanned++
				raw, err := os.ReadFile(p)
				if err != nil {
					return fmt.Errorf("read %s: %w", p, err)
				}
				out := migrateFeatureContent(raw)
				if bytes.Equal(out, raw) {
					continue
				}
				changed++
				if dryRun {
					fmt.Fprintf(cmd.OutOrStdout(), "--- would rewrite %s\n", p)
					continue
				}
				if err := os.WriteFile(p, out, 0o644); err != nil {
					return fmt.Errorf("write %s: %w", p, err)
				}
				fmt.Fprintf(cmd.OutOrStdout(), "rewrote %s\n", p)
			}
			summary := "no rewrites needed"
			if changed > 0 {
				summary = fmt.Sprintf("%d/%d file(s) rewritten", changed, scanned)
				if dryRun {
					summary = fmt.Sprintf("%d/%d file(s) would be rewritten (dry-run)", changed, scanned)
				}
			}
			fmt.Fprintln(cmd.OutOrStdout(), summary)
			return nil
		},
	}
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Print which files would change without writing")
	cmd.Flags().StringVar(&paths, "paths", "tests/e2e/features", "Subdir under workdir to search")
	return cmd
}

// Regex pre-compiles. featureTitleRe matches `Feature: X — kind journey`
// at start of line.
var (
	featureTitleRe = regexp.MustCompile(`(?m)^Feature: (.+?) — (\S+) journey$`)
	asVisitorRe    = regexp.MustCompile(`(?m)^\s*As a visitor of (\S+)\s*$`)
	outlineTitleRe = regexp.MustCompile(`(?m)^(\s*)Scenario Outline: \S+ — (\S+) accepts <variant>(?: values)?\s*$`)
	stepValueRe    = regexp.MustCompile(`(?m)^(\s*)When I enter "<value>" into the "([^"]+)" field\s*$`)
	exHeaderRe     = regexp.MustCompile(`(?m)^(\s*)\| variant \| value \|\s*$`)
)

func migrateFeatureContent(src []byte) []byte {
	// 1. Feature: line — needs the URL path from the "As a visitor of" line.
	src = featureTitleRe.ReplaceAllFunc(src, func(match []byte) []byte {
		m := featureTitleRe.FindSubmatch(match)
		if len(m) < 3 {
			return match
		}
		name, kind := string(m[1]), string(m[2])
		path := "/"
		if visitor := asVisitorRe.FindSubmatch(src); len(visitor) >= 2 {
			if u, err := url.Parse(string(visitor[1])); err == nil && u.Path != "" {
				path = u.Path
			}
		}
		return []byte(fmt.Sprintf("Feature: %s — %s · %s", name, kind, path))
	})

	// 2-4 happen only inside @kind:component blocks. Walk line-by-line so we
	// know whether the next Scenario Outline is component-tagged.
	lines := bytes.Split(src, []byte("\n"))
	for i, line := range lines {
		s := string(line)
		// Scenario Outline title rewrite — only inside @kind:component
		// blocks. Legacy @kind:param Outlines also use the same
		// "X accepts <variant> values" shape but their template
		// emission hasn't changed; touching them would create a
		// drift between live template output and migrated files.
		if m := outlineTitleRe.FindStringSubmatch(s); m != nil {
			if isWithinComponentOutline(lines, i) {
				indent, field := m[1], m[2]
				lines[i] = []byte(fmt.Sprintf(`%sScenario Outline: fill the "%s" field with <example>`, indent, field))
			}
			continue
		}
		// Step text — only swap "<value>" → "<example>" when the
		// surrounding Outline was component-tagged. Scan backwards
		// from this line for the nearest @kind:* tag.
		if m := stepValueRe.FindStringSubmatch(s); m != nil {
			if isWithinComponentOutline(lines, i) {
				indent, field := m[1], m[2]
				lines[i] = []byte(fmt.Sprintf(`%sWhen I enter "<example>" into the "%s" field`, indent, field))
				continue
			}
		}
		// Examples header.
		if m := exHeaderRe.FindStringSubmatch(s); m != nil {
			if isWithinComponentOutline(lines, i) {
				lines[i] = []byte(fmt.Sprintf("%s| case | example |", m[1]))
				continue
			}
		}
	}
	return bytes.Join(lines, []byte("\n"))
}

// isWithinComponentOutline scans backwards from idx looking for the
// nearest scenario boundary (a `@journey:` tag line OR a blank line
// followed by another tag block). Returns true when the nearest tag
// line contains `@kind:component`.
func isWithinComponentOutline(lines [][]byte, idx int) bool {
	for j := idx - 1; j >= 0; j-- {
		s := strings.TrimSpace(string(lines[j]))
		if strings.HasPrefix(s, "@journey:") || strings.HasPrefix(s, "@kind:") {
			return strings.Contains(s, "@kind:component")
		}
		// Don't walk past a Feature: line.
		if strings.HasPrefix(s, "Feature:") {
			return false
		}
	}
	return false
}
