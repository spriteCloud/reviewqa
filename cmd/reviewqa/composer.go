package main

import (
	"context"
	"os"
	"strings"

	"github.com/reviewqa/reviewqa/internal/ast"
	"github.com/reviewqa/reviewqa/internal/composer"
	"github.com/reviewqa/reviewqa/internal/config"
	"github.com/reviewqa/reviewqa/internal/llm"
	rlog "github.com/reviewqa/reviewqa/internal/log"
	"github.com/reviewqa/reviewqa/internal/plan"
)

// buildLadder constructs the LLM model ladder. The primary rung is
// always the configured cfg.Model (the default qwen3-coder-next when
// --llm is set). Additional fallback rungs come from
// REVIEWQA_LLM_LADDER (comma-separated model ids) — each is tried in
// sequence on parse failure of the previous.
func buildLadder(cfg config.Config, primary *llm.Client) composer.Ladder {
	ladder := composer.Ladder{
		Rungs: []composer.Rung{{Model: cfg.Model, Client: primary}},
	}
	extra := strings.TrimSpace(os.Getenv("REVIEWQA_LLM_LADDER"))
	if extra == "" {
		return ladder
	}
	for _, m := range strings.Split(extra, ",") {
		m = strings.TrimSpace(m)
		if m == "" || m == cfg.Model {
			continue
		}
		// Clone the config with the alternate model; the same
		// endpoint + api key are reused.
		altCfg := cfg
		altCfg.Model = m
		ladder.Rungs = append(ladder.Rungs, composer.Rung{Model: m, Client: llm.New(altCfg)})
	}
	return ladder
}

// composeScenarios walks the post-probe item list and, when the LLM
// is enabled, asks the composer for additional Scenarios per Gherkin
// feature item. Returns the items list with `ExtraScenarios` populated
// in place. No-ops when the LLM is disabled (default).
//
// v0.34: composer now supports a model ladder. When REVIEWQA_LLM_LADDER
// is set (comma-separated model ids), each is tried in order and the
// first to return parseable scenarios wins. The chosen model is
// embedded as `@model:<id>` on each emitted scenario.
func composeScenarios(ctx context.Context, cfg config.Config, items []plan.Item) []plan.Item {
	client := llm.New(cfg)
	if !client.Enabled() {
		return items
	}
	feedback := composer.LoadFeedback(cfg.WorkDir)
	if len(feedback.FailedTitles) > 0 {
		rlog.Info("composer: feeding ledger findings to LLM", "failed_titles", len(feedback.FailedTitles))
	}
	ladder := buildLadder(cfg, client)
	rlog.Info("composer: requesting LLM scenarios",
		"primary_model", ladder.First().Model,
		"endpoint", cfg.OpenAIBaseURL,
		"ladder_rungs", len(ladder.Rungs))
	// v0.31 cross-journey dedup. Track every step-sequence we've
	// accepted across the whole suite — duplicate scenarios from
	// different journeys (e.g. "no error and no success" repeated
	// across three journeys in the spritecloud.com run) get dropped
	// after the first.
	seenKeys := map[string]bool{}
	for i := range items {
		if items[i].Template != plan.TmplPlaywrightFeature {
			continue
		}
		j := buildJourneyForComposer(items[i])
		extras, winningModel, err := composer.ProposeWithLadder(ctx, ladder, j, 3, feedback)
		if err != nil {
			rlog.Warn("composer: skipped journey", "kind", j.Kind, "err", err)
			continue
		}
		extras = composer.Dedup(extras)
		fresh := extras[:0]
		for _, s := range extras {
			k := composer.ScenarioKey(s)
			if seenKeys[k] {
				continue
			}
			seenKeys[k] = true
			fresh = append(fresh, s)
		}
		if len(fresh) == 0 {
			continue
		}
		items[i].ExtraScenarios = toExtraScenarios(fresh)
		items[i].LLMModel = winningModel
		rlog.Info("composer: added scenarios", "journey", j.Kind, "count", len(fresh), "model", winningModel)
	}
	return items
}

func buildJourneyForComposer(it plan.Item) composer.Journey {
	j := composer.Journey{
		URL:      it.PageURL,
		Kind:     it.JourneyKind,
		Priority: priorityForKind(it.JourneyKind),
	}
	if len(it.Symbols) > 0 {
		first := it.Symbols[0]
		j.Title = first.PageTitle
		j.H1 = firstH1Text(first.Contents)
		j.HasForm = first.HasForm
		for _, l := range first.Links {
			if l.Aria != "" {
				j.Links = append(j.Links, l.Aria)
			}
			if len(j.Links) >= 10 {
				break
			}
		}
		if first.HasForm {
			j.Forms = append(j.Forms, "form with inputs")
		}
	}
	return j
}

func toExtraScenarios(in []composer.ExtraScenario) []plan.ExtraScenario {
	out := make([]plan.ExtraScenario, 0, len(in))
	for _, e := range in {
		steps := make([]plan.ExtraScenarioStep, 0, len(e.Steps))
		for _, s := range e.Steps {
			steps = append(steps, plan.ExtraScenarioStep{Keyword: s.Keyword, Text: s.Text})
		}
		out = append(out, plan.ExtraScenario{Name: e.Name, Tags: e.Tags, Steps: steps})
	}
	return out
}

func priorityForKind(kind string) string {
	switch kind {
	case "convert", "contact", "authenticate":
		return "critical"
	case "explore", "read":
		return "nice-to-have"
	}
	return "standard"
}

func firstH1Text(contents []ast.ContentAnchor) string {
	for _, c := range contents {
		if c.Tag == "h1" {
			return c.Text
		}
	}
	return ""
}
