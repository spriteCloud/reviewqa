package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type HealMode string

const (
	HealOnFailure HealMode = "on-failure"
	HealProactive HealMode = "proactive"
	HealOff       HealMode = "off"
)

type Config struct {
	GitHubToken      string
	Repo             string
	PRNumber         int
	OpenAIBaseURL    string
	OpenAIAPIKey     string
	Model            string
	LLMTimeout       time.Duration
	LLMTokenCap      int
	HealMode         HealMode
	AllowDiffToLLM   bool
	BranchPrefix     string
	DryRun           bool
	PlaywrightReport string
	WorkDir          string
}

func FromEnv() Config {
	c := Config{
		GitHubToken:      first(os.Getenv("REVIEWQA_GITHUB_TOKEN"), os.Getenv("GITHUB_TOKEN")),
		Repo:             os.Getenv("GITHUB_REPOSITORY"),
		OpenAIBaseURL:    firstNonEmpty(os.Getenv("OPENAI_BASE_URL"), "https://api.openai.com/v1"),
		OpenAIAPIKey:     os.Getenv("OPENAI_API_KEY"),
		Model:            firstNonEmpty(os.Getenv("REVIEWQA_MODEL"), "gpt-4o-mini"),
		LLMTimeout:       envDuration("REVIEWQA_LLM_TIMEOUT", 20*time.Second),
		LLMTokenCap:      envInt("REVIEWQA_LLM_TOKEN_CAP", 600),
		HealMode:         HealMode(firstNonEmpty(os.Getenv("REVIEWQA_HEAL_MODE"), string(HealOnFailure))),
		AllowDiffToLLM:   os.Getenv("REVIEWQA_ALLOW_DIFF_TO_LLM") == "1",
		BranchPrefix:     firstNonEmpty(os.Getenv("REVIEWQA_BRANCH_PREFIX"), "reviewqa"),
		PlaywrightReport: os.Getenv("REVIEWQA_PLAYWRIGHT_REPORT"),
		WorkDir:          firstNonEmpty(os.Getenv("REVIEWQA_WORKDIR"), "."),
	}
	if n := os.Getenv("REVIEWQA_PR"); n != "" {
		if v, err := strconv.Atoi(n); err == nil {
			c.PRNumber = v
		}
	}
	return c
}

func (c Config) Validate() error {
	switch c.HealMode {
	case HealOnFailure, HealProactive, HealOff:
	default:
		return fmt.Errorf("invalid REVIEWQA_HEAL_MODE %q; want on-failure|proactive|off", c.HealMode)
	}
	return nil
}

func (c Config) SplitRepo() (owner, name string, ok bool) {
	parts := strings.SplitN(c.Repo, "/", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", false
	}
	return parts[0], parts[1], true
}

func first(vals ...string) string {
	for _, v := range vals {
		if v != "" {
			return v
		}
	}
	return ""
}

func firstNonEmpty(a, b string) string {
	if a != "" {
		return a
	}
	return b
}

func envInt(k string, def int) int {
	if v := os.Getenv(k); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return def
}

func envDuration(k string, def time.Duration) time.Duration {
	if v := os.Getenv(k); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
	}
	return def
}
