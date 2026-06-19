package serve

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"
)

// StepResult is one Gherkin step's verdict, harvested from
// Playwright's JSON reporter after a run completes.
type StepResult struct {
	Title      string `json:"title"`
	Status     string `json:"status"`               // passed | failed | skipped | timedOut
	DurationMs int    `json:"durationMs"`
	Error      string `json:"error,omitempty"`
}

// LastRunRecord is the on-disk record per Scenario name. Persisted
// into tests/e2e/.quail-runs/last-run.json so the UI can render
// "last passed Xm ago" pills across reloads.
type LastRunRecord struct {
	Status     string       `json:"status"`     // passed | failed | skipped | mixed
	At         string       `json:"at"`         // RFC3339
	DurationMs int          `json:"durationMs"`
	Steps      []StepResult `json:"steps,omitempty"`
}

// LastRunIndex is the on-disk format: a flat map keyed by Scenario
// name. New runs overwrite the entry; we keep no history (yet).
type LastRunIndex map[string]LastRunRecord

func runsDir(workdir string) string {
	return filepath.Join(workdir, "tests", "e2e", ".quail-runs")
}

func lastRunPath(workdir string) string {
	return filepath.Join(runsDir(workdir), "last-run.json")
}

// LoadLastRunIndex reads the last-run.json file (if it exists) and
// returns the map. Missing file → empty map (not an error).
func LoadLastRunIndex(workdir string) LastRunIndex {
	b, err := os.ReadFile(lastRunPath(workdir))
	if err != nil {
		return LastRunIndex{}
	}
	var idx LastRunIndex
	if err := json.Unmarshal(b, &idx); err != nil {
		return LastRunIndex{}
	}
	return idx
}

// ScenarioRun is one entry in a scenario's multi-run timeline (v0.82).
// Smaller than LastRunRecord — no step detail.
type ScenarioRun struct {
	At         string `json:"at"`
	Status     string `json:"status"`
	DurationMs int    `json:"durationMs"`
}

// LoadScenarioTimeline walks the per-run JSON reports in
// tests/e2e/.quail-runs/ (written one per Run by the JSON
// reporter) and returns this scenario's runs sorted oldest→newest.
// Capped at the last 20 entries so the popover stays compact.
//
// The .quail-runs/ dir already exists from v0.75 — the JSON
// reports were retained but never aggregated. This reader closes
// that loop.
func LoadScenarioTimeline(workdir, scenarioName string) []ScenarioRun {
	if scenarioName == "" {
		return nil
	}
	entries, err := os.ReadDir(runsDir(workdir))
	if err != nil {
		return nil
	}
	var out []ScenarioRun
	for _, e := range entries {
		name := e.Name()
		if e.IsDir() || !strings.HasPrefix(name, "run-") || !strings.HasSuffix(name, ".json") {
			continue
		}
		b, err := os.ReadFile(filepath.Join(runsDir(workdir), name))
		if err != nil {
			continue
		}
		recs := ParsePlaywrightJSON(b)
		rec, ok := recs[scenarioName]
		if !ok {
			continue
		}
		at := rec.At
		if at == "" {
			// Fall back to the report file's mtime so older runs still sort.
			if info, err := os.Stat(filepath.Join(runsDir(workdir), name)); err == nil {
				at = info.ModTime().UTC().Format(time.RFC3339)
			}
		}
		out = append(out, ScenarioRun{At: at, Status: rec.Status, DurationMs: rec.DurationMs})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].At < out[j].At })
	if len(out) > 20 {
		out = out[len(out)-20:]
	}
	return out
}

func writeLastRunIndex(workdir string, idx LastRunIndex) error {
	if err := os.MkdirAll(runsDir(workdir), 0o755); err != nil {
		return err
	}
	b, err := json.MarshalIndent(idx, "", "  ")
	if err != nil {
		return err
	}
	tmp := lastRunPath(workdir) + ".tmp"
	if err := os.WriteFile(tmp, b, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, lastRunPath(workdir))
}

// pwJSONReport mirrors the Playwright JSON reporter shape. We only
// declare what we read.
type pwJSONReport struct {
	Suites []pwSuite `json:"suites"`
}

type pwSuite struct {
	Suites []pwSuite `json:"suites"`
	Specs  []pwSpec  `json:"specs"`
}

type pwSpec struct {
	Title string   `json:"title"`
	Tests []pwTest `json:"tests"`
}

type pwTest struct {
	Results []pwResult `json:"results"`
}

type pwResult struct {
	Status   string   `json:"status"`
	Duration int      `json:"duration"`
	Steps    []pwStep `json:"steps"`
	Error    *pwErr   `json:"error,omitempty"`
}

type pwStep struct {
	Title    string   `json:"title"`
	Category string   `json:"category"`
	Duration int      `json:"duration"`
	Error    *pwErr   `json:"error,omitempty"`
	Steps    []pwStep `json:"steps,omitempty"`
}

type pwErr struct {
	Message string `json:"message,omitempty"`
}

// ParsePlaywrightJSON walks the report and returns one
// LastRunRecord per Scenario title. Only the OUTER-MOST steps in
// each result are Gherkin (playwright-bdd emits `test.step()` per
// Gherkin line, then the Playwright internals nest inside).
func ParsePlaywrightJSON(report []byte) map[string]LastRunRecord {
	var doc pwJSONReport
	if err := json.Unmarshal(report, &doc); err != nil {
		return nil
	}
	out := map[string]LastRunRecord{}
	var walk func([]pwSuite)
	walk = func(suites []pwSuite) {
		for _, s := range suites {
			for _, spec := range s.Specs {
				name := scenarioNameFromSpec(spec.Title)
				for _, t := range spec.Tests {
					for _, r := range t.Results {
						steps := make([]StepResult, 0, len(r.Steps))
						for _, st := range r.Steps {
							if !looksLikeGherkinStep(st.Title) {
								continue
							}
							errMsg := ""
							if st.Error != nil {
								errMsg = st.Error.Message
							}
							steps = append(steps, StepResult{
								Title:      st.Title,
								Status:     stepStatus(st),
								DurationMs: st.Duration,
								Error:      errMsg,
							})
						}
						out[name] = LastRunRecord{
							Status:     r.Status,
							At:         time.Now().UTC().Format(time.RFC3339),
							DurationMs: r.Duration,
							Steps:      steps,
						}
					}
				}
			}
			walk(s.Suites)
		}
	}
	walk(doc.Suites)
	return out
}

// scenarioNameFromSpec strips off the playwright-bdd path/feature
// prefix that the JSON reporter prepends to each spec title. The
// reporter format is roughly "Feature Name > Scenario Name" or
// sometimes just "Scenario Name".
func scenarioNameFromSpec(title string) string {
	if i := strings.LastIndex(title, " > "); i >= 0 {
		return strings.TrimSpace(title[i+3:])
	}
	return strings.TrimSpace(title)
}

func looksLikeGherkinStep(title string) bool {
	for _, kw := range []string{"Given ", "When ", "Then ", "And ", "But "} {
		if strings.HasPrefix(title, kw) {
			return true
		}
	}
	return false
}

func stepStatus(s pwStep) string {
	if s.Error != nil && s.Error.Message != "" {
		return "failed"
	}
	return "passed"
}

// RunPreflight reports whether the workdir looks runnable —
// node_modules/.bin/playwright present and the feature path resolves.
// Cached lightly to avoid hammering the filesystem when the UI polls.
type RunPreflight struct {
	Ready       bool   `json:"ready"`
	Message     string `json:"message"`
	PlaywrightBin string `json:"playwrightBin,omitempty"`
}

func Preflight(workdir string) RunPreflight {
	pwBin := filepath.Join(workdir, "node_modules", ".bin", "playwright")
	if info, err := os.Stat(pwBin); err != nil || info.IsDir() {
		return RunPreflight{
			Ready:   false,
			Message: "Playwright not installed. Run `npm install && npx playwright install` in the project to enable the Run button.",
		}
	}
	return RunPreflight{
		Ready:         true,
		Message:       "ready",
		PlaywrightBin: pwBin,
	}
}

// runFlightMu serialises concurrent runs against the SAME workdir.
// Playwright stores reports in test-results/ — overlapping invocations
// would corrupt the report. Keyed by absolute workdir path.
var (
	runFlightMu sync.Mutex
	runFlight   = map[string]bool{}
)

func acquireRun(workdir string) bool {
	runFlightMu.Lock()
	defer runFlightMu.Unlock()
	if runFlight[workdir] {
		return false
	}
	runFlight[workdir] = true
	return true
}

func releaseRun(workdir string) {
	runFlightMu.Lock()
	defer runFlightMu.Unlock()
	delete(runFlight, workdir)
}

// RunRequest is the POST body for /api/run-scenario.
type RunRequest struct {
	Feature  string `json:"feature"`
	Scenario string `json:"scenario"`
}

// streamCommand spawns cmd, pipes stdout+stderr, and writes each line
// as a "line" SSE event. Returns the exit code; -1 on spawn errors.
//
// As of v0.86 the streamer ALSO captures the last 200 lines into a
// package-scoped ring buffer (`lastStdoutLines`) so the probe
// endpoint can parse them post-hoc for item counts / WAF
// signatures without re-piping.
func streamCommand(ctx context.Context, w http.ResponseWriter, flusher http.Flusher, cmd *exec.Cmd) (int, error) {
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return -1, fmt.Errorf("stdout pipe: %w", err)
	}
	cmd.Stderr = cmd.Stdout
	if err := cmd.Start(); err != nil {
		return -1, fmt.Errorf("start: %w", err)
	}
	resetCapture()
	scanner := bufio.NewScanner(stdout)
	scanner.Buffer(make([]byte, 0, 64<<10), 1<<20)
	for scanner.Scan() {
		line := scanner.Text()
		captureLine(line)
		writeEvent(w, flusher, "line", map[string]any{"text": line})
	}
	_ = scanner.Err()
	exitCode := 0
	if err := cmd.Wait(); err != nil {
		var ee *exec.ExitError
		if errors.As(err, &ee) {
			exitCode = ee.ExitCode()
		} else {
			exitCode = -1
		}
	}
	return exitCode, nil
}

// RunScenarioStream executes the named Scenario via npx playwright
// test scoped with --grep and streams stdout as Server-Sent Events
// to the response writer. Returns the exit code (so the caller can
// emit a final "done" event with the verdict).
//
// The function blocks until the playwright process exits or the
// request context is cancelled (browser closed / Stop button).
//
// playwright-bdd v9 requires `bddgen` to run BEFORE `playwright test`
// — that step parses .feature files and writes .features-gen/*.spec.js.
// Without it, `playwright test --grep <scenario>` finds no tests.
// We run bddgen first (when the binary exists) and stream its output
// too, then run playwright. Both phases share the SSE channel.
func RunScenarioStream(ctx context.Context, w http.ResponseWriter, workdir, featureRel, scenarioName string) (int, error) {
	if scenarioName == "" {
		return -1, errors.New("scenario name is empty")
	}
	if !acquireRun(workdir) {
		return -1, errors.New("a run is already in progress for this workdir")
	}
	defer releaseRun(workdir)

	pre := Preflight(workdir)
	if !pre.Ready {
		return -1, errors.New(pre.Message)
	}

	flusher, ok := w.(http.Flusher)
	if !ok {
		return -1, errors.New("response writer does not support flushing")
	}
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("X-Accel-Buffering", "no")

	// playwright-bdd v9 generates .spec files into .features-gen/ when
	// `bddgen` runs. Without bddgen Playwright sees no tests. Run it
	// first when the local binary exists, then run playwright. The
	// .feature path is informational only — Playwright discovers tests
	// from the generated specs, not from .feature paths.
	_ = featureRel
	grep := regexp.QuoteMeta(scenarioName)
	bddgenBin := filepath.Join(workdir, "node_modules", ".bin", "bddgen")
	pwBin := filepath.Join(workdir, "node_modules", ".bin", "playwright")
	playwrightArgs := []string{"test", "--reporter=line", "--grep", grep}

	writeEvent(w, flusher, "start", map[string]any{
		"workdir":  workdir,
		"feature":  featureRel,
		"scenario": scenarioName,
		"command":  "(bddgen) " + pwBin + " " + strings.Join(playwrightArgs, " "),
		"at":       time.Now().UTC().Format(time.RFC3339),
	})

	if info, err := os.Stat(bddgenBin); err == nil && !info.IsDir() {
		writeEvent(w, flusher, "line", map[string]any{"text": "# bddgen — generating .features-gen/ from .feature files"})
		bddgen := exec.CommandContext(ctx, bddgenBin)
		bddgen.Dir = workdir
		bddgen.Env = os.Environ()
		bgExit, err := streamCommand(ctx, w, flusher, bddgen)
		if err != nil {
			writeEvent(w, flusher, "done", map[string]any{"exitCode": -1, "passed": false, "at": time.Now().UTC().Format(time.RFC3339)})
			return -1, err
		}
		if bgExit != 0 {
			writeEvent(w, flusher, "line", map[string]any{"text": fmt.Sprintf("bddgen exited %d — aborting before playwright test.", bgExit)})
			writeEvent(w, flusher, "done", map[string]any{"exitCode": bgExit, "passed": false, "at": time.Now().UTC().Format(time.RFC3339)})
			return bgExit, nil
		}
	}

	// Write the JSON report into the per-workdir .quail-runs/
	// directory so we can parse per-step verdicts after the run.
	//
	// Playwright 1.61's CLI does NOT accept `--reporter=name:path`
	// (it tries to resolve "name:path" as a Node module). Pass the
	// JSON output path via PLAYWRIGHT_JSON_OUTPUT_NAME instead and
	// request both reporters with `--reporter=line,json`.
	ts := time.Now().UTC().Format("20060102-150405")
	if err := os.MkdirAll(runsDir(workdir), 0o755); err != nil {
		return -1, fmt.Errorf("create runs dir: %w", err)
	}
	jsonAbs := filepath.Join(runsDir(workdir), "run-"+ts+".json")
	pwArgsWithJSON := []string{"test", "--reporter=line,json", "--grep", grep}
	writeEvent(w, flusher, "line", map[string]any{"text": "# playwright test --grep " + scenarioName})
	pw := exec.CommandContext(ctx, pwBin, pwArgsWithJSON...)
	pw.Dir = workdir
	pw.Env = append(os.Environ(), "PLAYWRIGHT_JSON_OUTPUT_NAME="+jsonAbs)
	exitCode, err := streamCommand(ctx, w, flusher, pw)
	if err != nil {
		writeEvent(w, flusher, "done", map[string]any{"exitCode": -1, "passed": false, "at": time.Now().UTC().Format(time.RFC3339)})
		return -1, err
	}

	// Parse the JSON report (if it exists), emit per-step events,
	// update the on-disk last-run.json.
	if b, rerr := os.ReadFile(jsonAbs); rerr == nil {
		records := ParsePlaywrightJSON(b)
		// Emit one steps event per matching Scenario (typically just
		// one, since the UI runs --grep against a single Scenario name).
		for name, rec := range records {
			writeEvent(w, flusher, "steps", map[string]any{
				"scenario":   name,
				"status":     rec.Status,
				"durationMs": rec.DurationMs,
				"steps":      rec.Steps,
			})
		}
		// Merge into the persistent index.
		idx := LoadLastRunIndex(workdir)
		for name, rec := range records {
			idx[name] = rec
		}
		_ = writeLastRunIndex(workdir, idx)
	}

	writeEvent(w, flusher, "done", map[string]any{
		"exitCode": exitCode,
		"passed":   exitCode == 0,
		"at":       time.Now().UTC().Format(time.RFC3339),
	})
	return exitCode, nil
}

func writeEvent(w http.ResponseWriter, flusher http.Flusher, event string, payload any) {
	data, err := json.Marshal(payload)
	if err != nil {
		data = []byte(`{"error":"marshal"}`)
	}
	fmt.Fprintf(w, "event: %s\n", event)
	fmt.Fprintf(w, "data: %s\n\n", data)
	flusher.Flush()
}

// ParseRunSummary scans a slice of line-reporter output lines for
// playwright's terminal summary and returns the structured verdict.
// Used by tests (the live SSE stream is verified via integration).
type RunSummary struct {
	Passed  int
	Failed  int
	Skipped int
	HasLine bool
}

var summaryRe = regexp.MustCompile(`(?i)(\d+)\s+passed|(\d+)\s+failed|(\d+)\s+skipped|(\d+)\s+did not run`)

func ParseRunSummary(lines []string) RunSummary {
	var s RunSummary
	for _, line := range lines {
		matches := summaryRe.FindAllStringSubmatch(line, -1)
		for _, m := range matches {
			s.HasLine = true
			switch {
			case m[1] != "":
				s.Passed = atoiSafe(m[1])
			case m[2] != "":
				s.Failed = atoiSafe(m[2])
			case m[3] != "":
				s.Skipped = atoiSafe(m[3])
			}
		}
	}
	return s
}

func atoiSafe(s string) int {
	n := 0
	for _, c := range s {
		if c < '0' || c > '9' {
			return n
		}
		n = n*10 + int(c-'0')
	}
	return n
}

