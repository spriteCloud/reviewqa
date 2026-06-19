package serve

import (
	"strings"
	"testing"
)

// findEnv returns the value for `key` in env, or "<missing>" if
// the key isn't set. Empty value means explicitly set-to-empty.
func findEnv(env []string, key string) string {
	prefix := key + "="
	for _, e := range env {
		if strings.HasPrefix(e, prefix) {
			return strings.TrimPrefix(e, prefix)
		}
	}
	return "<missing>"
}

// When LLM is enabled in Settings with an endpoint + model, the
// subprocess env reflects them so probe's applyLLMOverride does
// the right thing.
func TestProbeSubprocessEnv_EnabledForwardsSettings(t *testing.T) {
	useTempSettings(t)
	_ = SaveSettings(Settings{LLM: LLMSettings{
		Enabled:    true,
		Endpoint:   "http://100.82.34.115:11434",
		Model:      "qwen3-coder-next:latest",
		APIKey:     "",
		TimeoutSec: 90,
	}})
	env := probeSubprocessEnv()

	if got := findEnv(env, "QUAIL_LLM"); got != "http://100.82.34.115:11434" {
		t.Errorf("QUAIL_LLM = %q, want endpoint", got)
	}
	if got := findEnv(env, "QUAIL_MODEL"); got != "qwen3-coder-next:latest" {
		t.Errorf("QUAIL_MODEL = %q, want model", got)
	}
	if got := findEnv(env, "OPENAI_API_KEY"); got != "ollama" {
		t.Errorf("OPENAI_API_KEY = %q, want ollama (keyless local fallback)", got)
	}
	if got := findEnv(env, "QUAIL_LLM_TIMEOUT"); got != "90s" {
		t.Errorf("QUAIL_LLM_TIMEOUT = %q, want 90s", got)
	}
}

func TestProbeSubprocessEnv_DisabledStripsKey(t *testing.T) {
	useTempSettings(t)
	_ = SaveSettings(Settings{LLM: LLMSettings{
		Enabled:  false,
		Endpoint: "http://x",
		APIKey:   "would-be-real-key",
	}})
	t.Setenv("OPENAI_API_KEY", "real-key-from-host")
	env := probeSubprocessEnv()
	if got := findEnv(env, "OPENAI_API_KEY"); got != "" {
		// The trailing entry from the settings layer should override
		// the inherited one. Empty string means "disabled" downstream.
		// Note: findEnv returns the LAST matching value (since we
		// append) — but our test searches first-match. Find LAST.
		lastValue := ""
		seen := false
		for _, e := range env {
			if strings.HasPrefix(e, "OPENAI_API_KEY=") {
				lastValue = strings.TrimPrefix(e, "OPENAI_API_KEY=")
				seen = true
			}
		}
		if !seen || lastValue != "" {
			t.Errorf("OPENAI_API_KEY (last) = %q, want empty (disabled)", lastValue)
		}
	}
}

func TestProbeSubprocessEnv_ExplicitKey(t *testing.T) {
	useTempSettings(t)
	_ = SaveSettings(Settings{LLM: LLMSettings{
		Enabled:  true,
		Endpoint: "https://api.openai.com",
		Model:    "gpt-4o-mini",
		APIKey:   "sk-test",
	}})
	env := probeSubprocessEnv()
	// Last OPENAI_API_KEY entry should be sk-test (settings appended
	// after os.Environ so the user's saved key wins).
	lastValue := ""
	for _, e := range env {
		if strings.HasPrefix(e, "OPENAI_API_KEY=") {
			lastValue = strings.TrimPrefix(e, "OPENAI_API_KEY=")
		}
	}
	if lastValue != "sk-test" {
		t.Errorf("OPENAI_API_KEY (last) = %q, want sk-test", lastValue)
	}
}
