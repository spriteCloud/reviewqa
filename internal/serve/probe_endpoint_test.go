package serve

import (
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
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
