package browser

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestRunnerDir_RespectsXDG(t *testing.T) {
	t.Setenv("XDG_CACHE_HOME", "/custom/cache")
	got := RunnerDir()
	want := "/custom/cache/reviewqa/playwright-runner"
	if got != want {
		t.Errorf("RunnerDir = %q, want %q", got, want)
	}
}

func TestRunnerDir_DefaultsToUserCache(t *testing.T) {
	t.Setenv("XDG_CACHE_HOME", "")
	home, err := os.UserHomeDir()
	if err != nil {
		t.Skip("no home dir on this platform")
	}
	got := RunnerDir()
	want := filepath.Join(home, ".cache", "reviewqa", "playwright-runner")
	if got != want {
		t.Errorf("RunnerDir = %q, want %q", got, want)
	}
}

// EnsureRunner must be a no-op when the sentinel exists and the
// package is on disk. We prove it by pointing XDG at a tempdir that
// we hand-populate, then running EnsureRunner with PATH cleared so
// any subprocess invocation would fail.
func TestEnsureRunner_NoopWhenReady(t *testing.T) {
	cache := t.TempDir()
	t.Setenv("XDG_CACHE_HOME", cache)

	runner := filepath.Join(cache, "reviewqa", "playwright-runner")
	if err := os.MkdirAll(filepath.Join(runner, "node_modules", "@playwright", "test"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(runner, "node_modules", "@playwright", "test", "package.json"), []byte(`{}`), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(runner, runnerSentinel), []byte("ok\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	// Cleared PATH guarantees any npm/npx exec call would explode.
	// If EnsureRunner is idempotent it never reaches that point.
	t.Setenv("PATH", "")

	got, err := EnsureRunner(context.Background())
	if err != nil {
		t.Fatalf("EnsureRunner on pre-populated runner: %v", err)
	}
	if got != runner {
		t.Errorf("EnsureRunner returned %q, want %q", got, runner)
	}
}

// When the sentinel is missing AND no npm is reachable, EnsureRunner
// must return ErrBrowserUnavailable (wrapped) so callers can
// distinguish "environment broken" from "browser ran but found
// nothing".
func TestEnsureRunner_ReturnsUnavailableWithoutNpm(t *testing.T) {
	cache := t.TempDir()
	t.Setenv("XDG_CACHE_HOME", cache)
	t.Setenv("PATH", "")

	_, err := EnsureRunner(context.Background())
	if err == nil {
		t.Fatal("expected error when npm is unreachable")
	}
	if !errorsIs(err, ErrBrowserUnavailable) {
		t.Errorf("expected wraps ErrBrowserUnavailable; got %v", err)
	}
}

// Tiny local wrapper so we don't have to import "errors" for one
// call. Mirrors errors.Is behavior on the wrapped sentinel.
func errorsIs(err, target error) bool {
	for err != nil {
		if err == target {
			return true
		}
		type unwrapper interface{ Unwrap() error }
		if u, ok := err.(unwrapper); ok {
			err = u.Unwrap()
			continue
		}
		return false
	}
	return false
}
