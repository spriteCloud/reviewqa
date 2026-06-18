package serve

import (
	"context"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// Scratch mode: Run() must NOT error when the workdir doesn't
// exist. We start it, then immediately cancel.
func TestRun_ScratchMode_TolerantOfMissingWorkdir(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	missing := filepath.Join(t.TempDir(), "does-not-exist")

	errCh := make(chan error, 1)
	go func() {
		errCh <- Run(ctx, Options{
			Workdir:   missing,
			Addr:      "127.0.0.1:0", // ephemeral port
			NoBrowser: true,
			Logf:      func(string, ...any) {},
		})
	}()
	// Give it a moment to start, then cancel.
	time.Sleep(50 * time.Millisecond)
	cancel()
	select {
	case err := <-errCh:
		if err != nil {
			t.Fatalf("Run errored in scratch mode: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Run did not return after ctx cancel")
	}
}

// loadProject must mark missing workdirs as scratch.
func TestLoadProject_FlagsScratchForMissingDir(t *testing.T) {
	missing := filepath.Join(t.TempDir(), "missing")
	p := loadProject(missing)
	if !p.Scratch {
		t.Errorf("expected Scratch true for missing dir, got false")
	}
}

func TestLoadProject_FlagsScratchForEmptyPath(t *testing.T) {
	p := loadProject("")
	if !p.Scratch {
		t.Errorf("expected Scratch true for empty path, got false")
	}
}

// pickProbeDestination should land in ~/reviewqa-projects/<brand>
// when the workdir is empty (scratch mode).
func TestPickProbeDestination_ScratchUsesHomeProjects(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	u, _ := url.Parse("https://www.example.com/path")
	got := pickProbeDestination("", u)
	want := filepath.Join(os.Getenv("HOME"), "reviewqa-projects", "example")
	if got != want {
		t.Errorf("pickProbeDestination(scratch) = %q, want %q", got, want)
	}
}

// /api/project returns Scratch=true under scratch conditions.
func TestProjectEndpoint_ReturnsScratchFlag(t *testing.T) {
	missing := filepath.Join(t.TempDir(), "does-not-exist")
	srv := httptest.NewServer(Handler(missing))
	defer srv.Close()
	resp, err := srv.Client().Get(srv.URL + "/api/project")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	// We don't need to decode the JSON — the body just has to be
	// non-empty and Scratch=true is what we test via loadProject.
	var buf [4096]byte
	n, _ := resp.Body.Read(buf[:])
	body := string(buf[:n])
	if !contains(body, `"scratch":true`) {
		t.Errorf("expected scratch:true in body, got %s", body)
	}
}

func contains(haystack, needle string) bool {
	return len(haystack) >= len(needle) && (haystack == needle || stringContains(haystack, needle))
}

func stringContains(s, sub string) bool {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
