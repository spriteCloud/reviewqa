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

// RunScenarioStream executes the named Scenario via npx playwright
// test scoped with --grep and streams stdout as Server-Sent Events
// to the response writer. Returns the exit code (so the caller can
// emit a final "done" event with the verdict).
//
// The function blocks until the playwright process exits or the
// request context is cancelled (browser closed / Stop button).
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

	args := []string{"playwright", "test", "--reporter=line"}
	if featureRel != "" {
		args = append(args, featureRel)
	}
	// Regex-escape the scenario name so grep matches it literally.
	grep := "^.*" + regexp.QuoteMeta(scenarioName) + "$"
	args = append(args, "--grep", grep)

	cmd := exec.CommandContext(ctx, "npx", args...)
	cmd.Dir = workdir
	cmd.Env = os.Environ()
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return -1, fmt.Errorf("stdout pipe: %w", err)
	}
	cmd.Stderr = cmd.Stdout

	if err := cmd.Start(); err != nil {
		return -1, fmt.Errorf("start: %w", err)
	}

	writeEvent(w, flusher, "start", map[string]any{
		"workdir":  workdir,
		"feature":  featureRel,
		"scenario": scenarioName,
		"command":  "npx " + strings.Join(args, " "),
		"at":       time.Now().UTC().Format(time.RFC3339),
	})

	scanner := bufio.NewScanner(stdout)
	scanner.Buffer(make([]byte, 0, 64<<10), 1<<20)
	for scanner.Scan() {
		writeEvent(w, flusher, "line", map[string]any{"text": scanner.Text()})
	}
	// Drain any remaining read errors silently — exit code is what
	// matters for the verdict.
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

