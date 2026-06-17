package main

import (
	"context"

	"github.com/reviewqa/reviewqa/internal/ast"
	"github.com/reviewqa/reviewqa/internal/composer"
	"github.com/reviewqa/reviewqa/internal/config"
	"github.com/reviewqa/reviewqa/internal/llm"
	rlog "github.com/reviewqa/reviewqa/internal/log"
	"github.com/reviewqa/reviewqa/internal/plan"
)

// composeScenarios walks the post-probe item list and, when the LLM
// is enabled, asks the composer for additional Scenarios per Gherkin
// feature item. Returns the items list with `ExtraScenarios` populated
// in place. No-ops when the LLM is disabled (default).
func composeScenarios(ctx context.Context, cfg config.Config, items []plan.Item) []plan.Item {
	client := llm.New(cfg)
	if !client.Enabled() {
		return items
	}
	rlog.Info("composer: requesting LLM scenarios", "model", cfg.Model, "endpoint", cfg.OpenAIBaseURL)
	for i := range items {
		if items[i].Template != plan.TmplPlaywrightFeature {
			continue
		}
		j := buildJourneyForComposer(items[i])
		extras, err := composer.Propose(ctx, client, j, 3)
		if err != nil {
			rlog.Warn("composer: skipped journey", "kind", j.Kind, "err", err)
			continue
		}
		if len(extras) == 0 {
			continue
		}
		items[i].ExtraScenarios = toExtraScenarios(extras)
		items[i].LLMModel = cfg.Model
		rlog.Info("composer: added scenarios", "journey", j.Kind, "count", len(extras))
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
