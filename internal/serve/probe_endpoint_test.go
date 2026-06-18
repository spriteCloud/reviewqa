package serve

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestProbeEndpoint_RejectsMissingURL(t *testing.T) {
	root := fixtureProject(t)
	workdir := filepath.Join(root, "tests", "e2e")
	srv := httptest.NewServer(Handler(workdir))
	defer srv.Close()

	resp, err := http.Post(srv.URL+"/api/probe", "application/json", strings.NewReader(`{}`))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status: got %d want 400", resp.StatusCode)
	}
}

func TestProbeEndpoint_RejectsInvalidURL(t *testing.T) {
	root := fixtureProject(t)
	workdir := filepath.Join(root, "tests", "e2e")
	srv := httptest.NewServer(Handler(workdir))
	defer srv.Close()

	cases := []string{
		`{"url":"not-a-url"}`,
		`{"url":"ftp://example.com"}`,
		`{"url":"http://"}`,
	}
	for _, body := range cases {
		resp, err := http.Post(srv.URL+"/api/probe", "application/json", strings.NewReader(body))
		if err != nil {
			t.Fatalf("post %s: %v", body, err)
		}
		resp.Body.Close()
		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("body=%s: status %d, want 400", body, resp.StatusCode)
		}
	}
}

func TestProbeEndpoint_RejectsGET(t *testing.T) {
	root := fixtureProject(t)
	workdir := filepath.Join(root, "tests", "e2e")
	srv := httptest.NewServer(Handler(workdir))
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/api/probe")
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusMethodNotAllowed {
		t.Fatalf("status: got %d want 405", resp.StatusCode)
	}
}

func TestProbeCwd_StepsUpFromTestsE2E(t *testing.T) {
	root := t.TempDir()
	workdir := filepath.Join(root, "tests", "e2e")
	got := probeCwd(workdir)
	if got != root {
		t.Errorf("probeCwd(%q) = %q, want %q", workdir, got, root)
	}
}

func TestProbeCwd_UsesWorkdirForOtherLayouts(t *testing.T) {
	root := t.TempDir()
	workdir := filepath.Join(root, "custom-layout")
	got := probeCwd(workdir)
	if got != workdir {
		t.Errorf("probeCwd(%q) = %q, want workdir unchanged", workdir, got)
	}
}

// Regression (v0.87.2): the probe subprocess must outlive a
// request-context cancel. Earlier versions used exec.CommandContext;
// when curl --max-time fired or the user closed the UI, the ctx
// cancel SIGKILL'd the probe mid-pipeline, leaving 167 static
// specs on disk but no .feature.
//
// We swap newProbeCmd to redirect spawns at a tiny sh script that
// sleeps, writes a sentinel, then exits. ProbeStream is invoked
// with a context cancelled immediately. The sentinel proves the
// subprocess ran to completion.
func TestProbeStream_SubprocessOutlivesRequestCancel(t *testing.T) {
	root := fixtureProject(t)
	workdir := filepath.Join(root, "tests", "e2e")

	sentinel := filepath.Join(t.TempDir(), "probe-finished")
	prev := newProbeCmd
	t.Cleanup(func() { newProbeCmd = prev })
	newProbeCmd = func(name string, args ...string) *exec.Cmd {
		return exec.Command("sh", "-c", fmt.Sprintf("sleep 0.3 && echo probe done && touch %q", sentinel))
	}

	rec := &flushRecorder{ResponseRecorder: httptest.NewRecorder()}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if _, err := ProbeStream(ctx, rec, workdir, ProbeRequest{URL: "https://example.com/"}); err != nil {
		t.Fatalf("ProbeStream: %v", err)
	}

	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if _, err := os.Stat(sentinel); err == nil {
			return
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Fatalf("sentinel %q missing — subprocess was killed by ctx cancel", sentinel)
}

type flushRecorder struct {
	*httptest.ResponseRecorder
}

func (f *flushRecorder) Flush() {}
