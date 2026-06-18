package serve

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/reviewqa/reviewqa/internal/probe"
)

// ProbeRequest is the JSON body accepted by POST /api/probe.
//
// `URL` is required; everything else is optional and maps to a probe
// CLI flag. The serve UI surfaces a single input box and an "Advanced"
// disclosure for the other fields.
type ProbeRequest struct {
	URL         string `json:"url"`
	Coverage    string `json:"coverage,omitempty"`
	JourneyKind string `json:"journeyKind,omitempty"`
	LLM         string `json:"llm,omitempty"`
}

// ProbeStream invokes the reviewqa binary's `probe` subcommand and
// streams its stdout/stderr as Server-Sent Events to the response
// writer. Returns the exit code so the handler can emit a final
// "done" event with the verdict.
//
// The probe binary writes into tests/e2e/<...> relative to its cwd,
// so we set cwd = parent-of-workdir when workdir ends in tests/e2e
// (the normal serve layout). For any other workdir we use the
// workdir itself — the user can still verify the output landed in
// the right place after the run.
func ProbeStream(ctx context.Context, w http.ResponseWriter, workdir string, req ProbeRequest) (int, error) {
	if strings.TrimSpace(req.URL) == "" {
		return -1, errors.New("url is required")
	}
	u, err := url.Parse(strings.TrimSpace(req.URL))
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
		return -1, errors.New("url must be a valid http(s) URL")
	}

	exe, err := os.Executable()
	if err != nil {
		if path, lerr := exec.LookPath("reviewqa"); lerr == nil {
			exe = path
		} else {
			return -1, fmt.Errorf("locate reviewqa binary: %w", err)
		}
	}

	cwd := pickProbeDestination(workdir, u)

	// --local: write rendered files into the workdir, do NOT try to
	// open a GitHub PR. The serve UI is a local control room, not a
	// CI runner; gh.New would fail without GITHUB_TOKEN.
	args := []string{"probe", "--url", u.String(), "--local"}
	if req.Coverage != "" {
		args = append(args, "--coverage", req.Coverage)
	}
	if req.JourneyKind != "" {
		// The probe CLI has no --journey-kind flag yet; preserve the
		// field for the future filter wiring. For now it lands in the
		// log line so the UI shows the user what they asked for.
		_ = req.JourneyKind
	}
	if req.LLM != "" {
		args = append(args, "--llm", req.LLM)
	}

	flusher, ok := w.(http.Flusher)
	if !ok {
		return -1, errors.New("response writer does not support flushing")
	}
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("X-Accel-Buffering", "no")

	writeEvent(w, flusher, "start", map[string]any{
		"workdir": cwd,
		"url":     u.String(),
		"command": exe + " " + strings.Join(args, " "),
		"at":      time.Now().UTC().Format(time.RFC3339),
	})

	cmd := exec.CommandContext(ctx, exe, args...)
	cmd.Dir = cwd
	cmd.Env = os.Environ()
	exitCode, err := streamCommand(ctx, w, flusher, cmd)
	if err != nil {
		writeEvent(w, flusher, "done", map[string]any{
			"exitCode": -1,
			"passed":   false,
			"error":    err.Error(),
			"at":       time.Now().UTC().Format(time.RFC3339),
		})
		return -1, err
	}

	writeEvent(w, flusher, "done", map[string]any{
		"exitCode": exitCode,
		"passed":   exitCode == 0,
		"at":       time.Now().UTC().Format(time.RFC3339),
	})
	return exitCode, nil
}

// probeCwd returns the directory the probe subprocess should run in
// for an *in-place* re-probe (no URL switch). When workdir ends in
// `tests/e2e` (the normal serve layout) we step up two parents so
// probe re-emits into the same project root. Otherwise the workdir
// itself is used.
func probeCwd(workdir string) string {
	abs := filepath.Clean(workdir)
	parent := filepath.Dir(abs)
	if filepath.Base(parent) == "tests" && filepath.Base(abs) == "e2e" {
		return filepath.Dir(parent)
	}
	return abs
}

// pickProbeDestination chooses where the probe subprocess should
// write. If the URL's brand matches the current workdir's name, we
// re-probe in place (probeCwd). Otherwise we land in a new sibling
// dir named after `BrandFromOrigin(url)`, suffixing `-1`, `-2`, etc.
// if a non-reviewqa dir already squats the slot.
//
// Scratch mode (v0.85): when the current workdir is empty or
// doesn't exist on disk, we have no parent dir to plant a sibling
// in. Fall back to a per-user projects root at
// `~/reviewqa-projects/<brand>/` — same XDG-ish location pattern
// as the settings file at `~/.config/reviewqa/serve.json`. Falls
// back to cwd if the user's home dir can't be resolved.
//
// Created dirs are mkdir'd here so the probe subprocess has a cwd to
// run in. The frontend reads the destination from the SSE `start`
// event so it can /api/switch-project to the new dir on `done`.
func pickProbeDestination(workdir string, u *url.URL) string {
	brand := probe.BrandFromHost(u.Host)
	// Scratch mode: no current project; pick the user-level root.
	if workdir == "" {
		if home, err := os.UserHomeDir(); err == nil {
			parent := filepath.Join(home, "reviewqa-projects")
			return reserveSiblingDir(parent, brand, workdir)
		}
		// Last-resort fallback: cwd.
		cwd, _ := os.Getwd()
		return reserveSiblingDir(cwd, brand, workdir)
	}
	if info, err := os.Stat(workdir); err != nil || !info.IsDir() {
		if home, err := os.UserHomeDir(); err == nil {
			parent := filepath.Join(home, "reviewqa-projects")
			return reserveSiblingDir(parent, brand, workdir)
		}
		cwd, _ := os.Getwd()
		return reserveSiblingDir(cwd, brand, workdir)
	}
	current := probeCwd(workdir)
	if brand == "" {
		return current
	}
	// If the current project's name matches the brand, re-probe in place.
	if strings.EqualFold(filepath.Base(current), brand) ||
		strings.HasPrefix(strings.ToLower(filepath.Base(current)), brand) {
		return current
	}
	return reserveSiblingDir(filepath.Dir(current), brand, current)
}

// reserveSiblingDir picks a destination subdir of `parent` named
// after `brand`. If `parent/brand` is free it's created and
// returned. If a reviewqa-looking project already exists there we
// re-probe in place. If a non-reviewqa dir (or non-dir file) is
// squatting the slot, we suffix `-1`, `-2`, … until we find a free
// or compatible slot. Returns `fallback` if anything fails (parent
// can't be created, 100 collisions, etc).
func reserveSiblingDir(parent, brand, fallback string) string {
	if brand == "" {
		return fallback
	}
	if err := os.MkdirAll(parent, 0o755); err != nil {
		return fallback
	}
	candidate := filepath.Join(parent, brand)
	for i := 1; i < 100; i++ {
		info, err := os.Stat(candidate)
		if err != nil {
			// Doesn't exist — create it and use it.
			if mkerr := os.MkdirAll(candidate, 0o755); mkerr == nil {
				return candidate
			}
			return fallback
		}
		if !info.IsDir() {
			candidate = filepath.Join(parent, brand+"-"+strconv.Itoa(i))
			continue
		}
		// Exists as a dir. If it already looks like a reviewqa /
		// Playwright project we re-probe in place; otherwise suffix.
		if looksLikeReviewqaProject(candidate) {
			return candidate
		}
		candidate = filepath.Join(parent, brand+"-"+strconv.Itoa(i))
	}
	return fallback
}

// handleProbe is the http handler. Registered by Run as
// POST /api/probe. Validates the JSON body and delegates the streaming
// to ProbeStream.
func handleProbe(workdir string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var req ProbeRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if _, err := ProbeStream(r.Context(), w, workdir, req); err != nil {
			// If the stream hasn't been initiated yet, surface as 400.
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}
}
