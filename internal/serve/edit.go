package serve

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// ScenarioRange locates one Scenario block inside a .feature file's
// source lines. start is the first line of the block (the first @tag
// line, or the Scenario line if there are no tags). end is exclusive —
// the line index of the next block (next Scenario / Background) or
// len(lines) if it's the last block in the file.
type ScenarioRange struct {
	Name  string
	Start int
	End   int
}

// findScenarioRange returns the line range that contains the Scenario
// whose name matches `name`. Tag lines that immediately precede the
// Scenario keyword (separated only by blank lines / comments) count as
// part of the block.
func findScenarioRange(lines []string, name string) (ScenarioRange, bool) {
	starts := scenarioStarts(lines)
	for i, s := range starts {
		if s.Name == name {
			end := len(lines)
			if i+1 < len(starts) {
				end = starts[i+1].Start
			}
			s.End = end
			return s, true
		}
	}
	return ScenarioRange{}, false
}

func scenarioStarts(lines []string) []ScenarioRange {
	var out []ScenarioRange
	tagAnchor := -1
	for i, raw := range lines {
		trimmed := strings.TrimSpace(raw)
		if strings.HasPrefix(trimmed, "@") {
			if tagAnchor == -1 {
				tagAnchor = i
			}
			continue
		}
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}
		if strings.HasPrefix(trimmed, "Scenario:") || strings.HasPrefix(trimmed, "Scenario Outline:") {
			start := i
			if tagAnchor != -1 {
				start = tagAnchor
			}
			name := strings.TrimSpace(strings.TrimPrefix(trimmed, "Scenario:"))
			name = strings.TrimSpace(strings.TrimPrefix(name, "Scenario Outline:"))
			out = append(out, ScenarioRange{Name: name, Start: start})
			tagAnchor = -1
			continue
		}
		// Any other content (Feature header lines, narrative, Background,
		// step lines) resets the pending tag anchor — tags only belong
		// to the Scenario keyword they immediately precede.
		if strings.HasPrefix(trimmed, "Background:") || strings.HasPrefix(trimmed, "Feature:") {
			tagAnchor = -1
		}
	}
	return out
}

// readLines reads a file as a slice of lines (without trailing
// newlines). The caller is responsible for re-joining with "\n" when
// writing back.
func readLines(path string) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var lines []string
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 0, 64<<10), 1<<20)
	for sc.Scan() {
		lines = append(lines, sc.Text())
	}
	return lines, sc.Err()
}

// writeLinesAtomic writes the lines to path via a tmpfile + rename in
// the same directory. Trailing newline preserved. Before write, the
// existing file (if any) is copied into the .reviewqa-history backup
// tree so the user has an undo trail.
//
// historyRoot lets callers point the backup tree somewhere other than
// the default (tests use this to keep tmpdir output tidy).
func writeLinesAtomic(path string, lines []string, historyRoot string) error {
	if err := backupFile(path, historyRoot); err != nil {
		return err
	}
	dir := filepath.Dir(path)
	tmp, err := os.CreateTemp(dir, ".reviewqa-edit-*")
	if err != nil {
		return err
	}
	tmpPath := tmp.Name()
	for _, line := range lines {
		if _, err := io.WriteString(tmp, line); err != nil {
			tmp.Close()
			os.Remove(tmpPath)
			return err
		}
		if _, err := tmp.Write([]byte("\n")); err != nil {
			tmp.Close()
			os.Remove(tmpPath)
			return err
		}
	}
	if err := tmp.Close(); err != nil {
		os.Remove(tmpPath)
		return err
	}
	return os.Rename(tmpPath, path)
}

func backupFile(path, historyRoot string) error {
	src, err := os.Open(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}
	defer src.Close()
	ts := backupTimestamp()
	target := filepath.Join(historyRoot, ts, filepath.Base(path))
	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		return err
	}
	dst, err := os.Create(target)
	if err != nil {
		return err
	}
	defer dst.Close()
	if _, err := io.Copy(dst, src); err != nil {
		return err
	}
	return nil
}

// backupTimestamp is a package-level variable so tests can swap it for
// deterministic output. Defaults to a runtime UTC stamp.
var backupTimestamp = func() string {
	return defaultTimestamp()
}

// DeleteScenario removes the Scenario named `scenarioName` from the
// .feature file at featurePath. historyRoot receives the pre-edit
// backup. Returns the count of scenarios deleted (always 0 or 1).
func DeleteScenario(featurePath, scenarioName, historyRoot string) (int, error) {
	lines, err := readLines(featurePath)
	if err != nil {
		return 0, err
	}
	rng, ok := findScenarioRange(lines, scenarioName)
	if !ok {
		return 0, fmt.Errorf("scenario %q not found", scenarioName)
	}
	// Trim trailing blank lines from the block so we don't leave a
	// stray gap behind.
	newLines := append([]string{}, lines[:rng.Start]...)
	newLines = append(newLines, lines[rng.End:]...)
	newLines = trimTrailingBlanks(newLines)
	if err := writeLinesAtomic(featurePath, newLines, historyRoot); err != nil {
		return 0, err
	}
	return 1, nil
}

// ReplaceScenario replaces the Scenario named `scenarioName` with the
// new Gherkin block (must parse and contain exactly one Scenario). If
// the replacement renames the Scenario, the new name is returned.
func ReplaceScenario(featurePath, scenarioName, newBlock, historyRoot string) (string, error) {
	if err := validateScenarioBlock(newBlock); err != nil {
		return "", err
	}
	lines, err := readLines(featurePath)
	if err != nil {
		return "", err
	}
	rng, ok := findScenarioRange(lines, scenarioName)
	if !ok {
		return "", fmt.Errorf("scenario %q not found", scenarioName)
	}
	blockLines := splitLines(newBlock)
	newLines := append([]string{}, lines[:rng.Start]...)
	newLines = append(newLines, blockLines...)
	newLines = append(newLines, lines[rng.End:]...)
	if err := writeLinesAtomic(featurePath, newLines, historyRoot); err != nil {
		return "", err
	}
	newName := scenarioBlockName(newBlock)
	return newName, nil
}

// AppendScenario adds a new Scenario block at the end of the feature
// file, separated by a single blank line from the existing content.
// Returns the new Scenario's name.
func AppendScenario(featurePath, newBlock, historyRoot string) (string, error) {
	if err := validateScenarioBlock(newBlock); err != nil {
		return "", err
	}
	lines, err := readLines(featurePath)
	if err != nil {
		return "", err
	}
	lines = trimTrailingBlanks(lines)
	lines = append(lines, "")
	lines = append(lines, splitLines(newBlock)...)
	if err := writeLinesAtomic(featurePath, lines, historyRoot); err != nil {
		return "", err
	}
	return scenarioBlockName(newBlock), nil
}

// validateScenarioBlock parses `block` as a fragment of a .feature
// file and accepts it only if it contains exactly one Scenario with
// at least one step. Anything else (multiple Scenarios, an embedded
// Feature header, no steps) is rejected to keep edits well-formed.
func validateScenarioBlock(block string) error {
	wrapper := "Feature: __reviewqa_edit_validate__\n\n" + block + "\n"
	feat, err := ParseFeatureBytes([]byte(wrapper))
	if err != nil {
		return fmt.Errorf("invalid Gherkin: %w", err)
	}
	if len(feat.Scenarios) != 1 {
		return fmt.Errorf("expected exactly one Scenario block, got %d", len(feat.Scenarios))
	}
	if len(feat.Scenarios[0].Steps) == 0 {
		return fmt.Errorf("Scenario must contain at least one step")
	}
	return nil
}

func scenarioBlockName(block string) string {
	for _, line := range splitLines(block) {
		t := strings.TrimSpace(line)
		if strings.HasPrefix(t, "Scenario:") {
			return strings.TrimSpace(strings.TrimPrefix(t, "Scenario:"))
		}
		if strings.HasPrefix(t, "Scenario Outline:") {
			return strings.TrimSpace(strings.TrimPrefix(t, "Scenario Outline:"))
		}
	}
	return ""
}

func splitLines(s string) []string {
	s = strings.ReplaceAll(s, "\r\n", "\n")
	out := strings.Split(s, "\n")
	for len(out) > 0 && out[len(out)-1] == "" {
		out = out[:len(out)-1]
	}
	return out
}

func trimTrailingBlanks(lines []string) []string {
	for len(lines) > 0 && strings.TrimSpace(lines[len(lines)-1]) == "" {
		lines = lines[:len(lines)-1]
	}
	return lines
}

func historyRootFor(workdir string) string {
	return filepath.Join(workdir, "tests", "e2e", ".reviewqa-history")
}
