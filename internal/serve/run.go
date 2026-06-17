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
	"strings"
	"sync"
	"time"
)

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
func streamCommand(ctx context.Context, w http.ResponseWriter, flusher http.Flusher, cmd *exec.Cmd) (int, error) {
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return -1, fmt.Errorf("stdout pipe: %w", err)
	}
	cmd.Stderr = cmd.Stdout
	if err := cmd.Start(); err != nil {
		return -1, fmt.Errorf("start: %w", err)
	}
	scanner := bufio.NewScanner(stdout)
	scanner.Buffer(make([]byte, 0, 64<<10), 1<<20)
	for scanner.Scan() {
		writeEvent(w, flusher, "line", map[string]any{"text": scanner.Text()})
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

	writeEvent(w, flusher, "line", map[string]any{"text": "# playwright test --grep " + scenarioName})
	pw := exec.CommandContext(ctx, pwBin, playwrightArgs...)
	pw.Dir = workdir
	pw.Env = os.Environ()
	exitCode, err := streamCommand(ctx, w, flusher, pw)
	if err != nil {
		writeEvent(w, flusher, "done", map[string]any{"exitCode": -1, "passed": false, "at": time.Now().UTC().Format(time.RFC3339)})
		return -1, err
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

