package serve

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

// minimalReport produces a tiny Playwright JSON report blob that
// ParsePlaywrightJSON can read; just enough for the timeline reader.
func writeMinimalReport(t *testing.T, dir, scenario, status, at string, durationMs int) {
	t.Helper()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	report := map[string]any{
		"suites": []any{
			map[string]any{
				"specs": []any{
					map[string]any{
						"title": scenario,
						"tests": []any{
							map[string]any{
								"results": []any{
									map[string]any{
										"status":     status,
										"startTime":  at,
										"duration":   durationMs,
										"steps":      []any{},
									},
								},
							},
						},
					},
				},
			},
		},
	}
	b, _ := json.Marshal(report)
	name := "run-" + at + ".json"
	if err := os.WriteFile(filepath.Join(dir, name), b, 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestLoadScenarioTimeline_ReadsRunFiles(t *testing.T) {
	root := t.TempDir()
	runsRoot := filepath.Join(root, "tests", "e2e", ".quail-runs")
	writeMinimalReport(t, runsRoot, "demo scenario", "passed", "20260618-100000", 1234)
	writeMinimalReport(t, runsRoot, "demo scenario", "failed", "20260618-110000", 2345)
	writeMinimalReport(t, runsRoot, "other", "passed", "20260618-120000", 999)

	timeline := LoadScenarioTimeline(root, "demo scenario")
	if len(timeline) != 2 {
		t.Fatalf("got %d runs, want 2: %+v", len(timeline), timeline)
	}
	// Sorted oldest→newest.
	if timeline[0].DurationMs != 1234 || timeline[1].DurationMs != 2345 {
		t.Errorf("order wrong: %+v", timeline)
	}
}

func TestScenarioRunsEndpoint_RequiresScenario(t *testing.T) {
	root := fixtureProject(t)
	workdir := filepath.Join(root, "tests", "e2e")
	srv := httptest.NewServer(Handler(workdir))
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/api/scenario-runs")
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("status %d, want 400", resp.StatusCode)
	}
}
