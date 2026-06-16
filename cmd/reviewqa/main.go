package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/spf13/cobra"

	_ "github.com/reviewqa/reviewqa/internal/ast/golang"
	_ "github.com/reviewqa/reviewqa/internal/ast/java"
	_ "github.com/reviewqa/reviewqa/internal/ast/python"
	_ "github.com/reviewqa/reviewqa/internal/ast/ts"

	"github.com/reviewqa/reviewqa/internal/config"
	"github.com/reviewqa/reviewqa/internal/diff"
	"github.com/reviewqa/reviewqa/internal/gen"
	"github.com/reviewqa/reviewqa/internal/gh"
	"github.com/reviewqa/reviewqa/internal/heal"
	"github.com/reviewqa/reviewqa/internal/llm"
	rlog "github.com/reviewqa/reviewqa/internal/log"
	"github.com/reviewqa/reviewqa/internal/merge"
	"github.com/reviewqa/reviewqa/internal/plan"
)

var (
	version = "0.1.0"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()
	root := newRoot()
	if err := root.ExecuteContext(ctx); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func newRoot() *cobra.Command {
	root := &cobra.Command{
		Use:           "reviewqa",
		Short:         "Generate tests for a PR and heal broken Playwright locators.",
		SilenceUsage:  true,
		SilenceErrors: true,
		Version:       version,
	}
	root.AddCommand(newGenerateCmd(), newHealCmd(), newScanCmd())
	return root
}

func newGenerateCmd() *cobra.Command {
	var pr int
	var dryRun bool
	cmd := &cobra.Command{
		Use:   "generate",
		Short: "Scan a PR's diff, emit test scaffolds, and open a follow-up PR.",
		RunE: func(cmd *cobra.Command, _ []string) error {
			cfg := config.FromEnv()
			if err := cfg.Validate(); err != nil {
				return err
			}
			if pr == 0 {
				pr = cfg.PRNumber
			}
			if pr == 0 {
				pr = readPRFromEvent()
			}
			if pr == 0 {
				return fmt.Errorf("missing --pr; set $REVIEWQA_PR or run inside a pull_request event")
			}
			cfg.PRNumber = pr
			cfg.DryRun = dryRun
			return runGenerate(cmd.Context(), cfg)
		},
	}
	cmd.Flags().IntVar(&pr, "pr", 0, "PR number")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Print plan instead of opening a PR")
	return cmd
}

func newHealCmd() *cobra.Command {
	var pr int
	var report string
	var dryRun bool
	cmd := &cobra.Command{
		Use:   "heal",
		Short: "Repair broken Playwright locators (defaults to on-failure).",
		RunE: func(cmd *cobra.Command, _ []string) error {
			cfg := config.FromEnv()
			if err := cfg.Validate(); err != nil {
				return err
			}
			if pr == 0 {
				pr = cfg.PRNumber
			}
			if pr == 0 {
				pr = readPRFromEvent()
			}
			cfg.PRNumber = pr
			cfg.DryRun = dryRun
			if report != "" {
				cfg.PlaywrightReport = report
			}
			return runHeal(cmd.Context(), cfg)
		},
	}
	cmd.Flags().IntVar(&pr, "pr", 0, "PR number")
	cmd.Flags().StringVar(&report, "report", "", "Path to Playwright JSON report (on-failure mode)")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Print edits instead of opening a PR")
	return cmd
}

func newScanCmd() *cobra.Command {
	var pr int
	cmd := &cobra.Command{
		Use:   "scan",
		Short: "Dry-run: print what generate/heal would do without opening a PR.",
		RunE: func(cmd *cobra.Command, _ []string) error {
			cfg := config.FromEnv()
			if pr == 0 {
				pr = cfg.PRNumber
			}
			if pr == 0 {
				pr = readPRFromEvent()
			}
			cfg.PRNumber = pr
			cfg.DryRun = true
			return runGenerate(cmd.Context(), cfg)
		},
	}
	cmd.Flags().IntVar(&pr, "pr", 0, "PR number")
	return cmd
}

func runGenerate(ctx context.Context, cfg config.Config) error {
	client, err := gh.New(ctx, cfg)
	if err != nil && !cfg.DryRun {
		return err
	}
	files, prInfo, err := fetchPRFilesAndInfo(ctx, client, cfg.PRNumber)
	if err != nil {
		return err
	}
	layout := plan.Detect(cfg.WorkDir)
	items := plan.Build(files, layout)
	if len(items) == 0 {
		rlog.Info("no new symbols in PR that warrant generated tests")
		writeStepSummary("reviewqa: no new symbols in PR that warrant generated tests.\n")
		return nil
	}
	rendered, err := gen.Render(items, cfg.WorkDir)
	if err != nil {
		return fmt.Errorf("render: %w", err)
	}
	llmClient := llm.New(cfg)
	for i := range rendered {
		rendered[i].Content = llmClient.Humanize(ctx, rendered[i].Symbol.Language, rendered[i].Symbol.Name, rendered[i].Content)
	}
	if cfg.DryRun || client == nil {
		printRendered(rendered)
		return nil
	}
	branch := fmt.Sprintf("%s/tests-pr-%d-%s", cfg.BranchPrefix, cfg.PRNumber, shortSHA(prInfo.HeadSHA))
	body := genPRBody(prInfo, rendered)
	files2 := applyExistingFileMerge(ctx, client, rendered, prInfo.HeadSHA)
	url, err := client.OpenPR(ctx, gh.PROpts{
		BaseBranch: prInfo.HeadBranch, NewBranch: branch,
		Title: fmt.Sprintf("reviewqa: tests for PR #%d", cfg.PRNumber),
		Body:  body, Files: files2,
	})
	if err != nil {
		return fmt.Errorf("open pr: %w", err)
	}
	rlog.Info("opened test PR", "url", url)
	writeStepSummary(generateSummary(prInfo, rendered, url))
	return nil
}

// siblingPath inserts "_reviewqa" before the test suffix so the generated
// file lands next to the existing one instead of overwriting it.
//
//	internal/diff/diff_test.go   -> internal/diff/diff_reviewqa_test.go
//	src/foo.test.ts              -> src/foo.reviewqa.test.ts
//	tests/test_users.py          -> tests/test_users_reviewqa.py
//	src/test/java/x/YTest.java   -> src/test/java/x/YReviewqaTest.java
func siblingPath(p string) string {
	dir, base := filepath.Split(p)
	switch {
	case strings.HasSuffix(base, "_test.go"):
		return dir + strings.TrimSuffix(base, "_test.go") + "_reviewqa_test.go"
	case strings.HasSuffix(base, ".test.ts"):
		return dir + strings.TrimSuffix(base, ".test.ts") + ".reviewqa.test.ts"
	case strings.HasSuffix(base, ".test.js"):
		return dir + strings.TrimSuffix(base, ".test.js") + ".reviewqa.test.js"
	case strings.HasSuffix(base, ".spec.ts"):
		return dir + strings.TrimSuffix(base, ".spec.ts") + ".reviewqa.spec.ts"
	case strings.HasSuffix(base, ".py"):
		return dir + strings.TrimSuffix(base, ".py") + "_reviewqa.py"
	case strings.HasSuffix(base, "Test.java"):
		return dir + strings.TrimSuffix(base, "Test.java") + "ReviewqaTest.java"
	}
	ext := filepath.Ext(base)
	return dir + strings.TrimSuffix(base, ext) + "_reviewqa" + ext
}

// applyExistingFileMerge folds rendered scaffolds into the existing tree:
// append-where-possible, sibling-when-merge-unsupported, fresh otherwise.
// Mutates rendered[i].Content/Path when the file already exists.
func applyExistingFileMerge(ctx context.Context, client *gh.Client, rendered []gen.Rendered, headSHA string) map[string][]byte {
	out := map[string][]byte{}
	for i, r := range rendered {
		existing, found, err := client.ReadFile(ctx, r.Path, headSHA)
		if err != nil {
			rlog.Warn("could not check existing test file; will write fresh", "path", r.Path, "err", err)
		}
		if !found {
			out[r.Path] = r.Content
			continue
		}
		mergeOneOrSibling(&rendered[i], r, []byte(existing), out)
	}
	return out
}

func mergeOneOrSibling(slot *gen.Rendered, r gen.Rendered, existing []byte, out map[string][]byte) {
	if merged, ok := merge.Append(r.Symbol.Language, existing, r.Content); ok {
		rlog.Info("appending to existing test file", "path", r.Path, "symbol", r.Symbol.Name)
		slot.Content = merged
		out[r.Path] = merged
		return
	}
	alt := siblingPath(r.Path)
	rlog.Info("existing test file present; writing to sibling", "from", r.Path, "to", alt)
	slot.Path = alt
	out[alt] = r.Content
}

func runHeal(ctx context.Context, cfg config.Config) error {
	if cfg.HealMode == config.HealOff {
		rlog.Info("heal mode = off; skipping")
		return nil
	}
	llmClient := llm.New(cfg)
	client, err := gh.New(ctx, cfg)
	if err != nil && !cfg.DryRun {
		return err
	}
	files, prInfo, err := fetchPRFilesAndInfo(ctx, client, cfg.PRNumber)
	if err != nil {
		return err
	}
	report, err := loadReportIfNeeded(cfg)
	if err != nil {
		return err
	}
	edits, err := heal.Run(ctx, cfg, files, report, llmClient)
	if err != nil {
		return err
	}
	if handled := emitOrSkipHealOutput(edits, cfg, client, prInfo); handled {
		return nil
	}
	return openHealPR(ctx, client, cfg, files, prInfo, edits)
}

// emitOrSkipHealOutput handles the three terminal cases for runHeal: no
// edits to apply (logs + writes summary), or the dry-run / missing-PR path
// (prints edits to stdout). Returns true when the caller should return nil
// without opening a PR.
func emitOrSkipHealOutput(edits []heal.Edit, cfg config.Config, client *gh.Client, prInfo *prSummary) bool {
	if len(edits) == 0 {
		rlog.Info("no locator edits to apply")
		writeStepSummary("reviewqa: no locator edits to apply.\n")
		return true
	}
	if cfg.DryRun || client == nil || prInfo == nil {
		printEdits(edits)
		return true
	}
	return false
}

// fetchPRFilesAndInfo pulls the PR's diff + blobs through the GitHub client.
// Returns (nil, nil, nil) when no client/PR is available — caller is
// expected to handle that as "skip heal" or "dry-run".
func fetchPRFilesAndInfo(ctx context.Context, client *gh.Client, prNum int) ([]diff.File, *prSummary, error) {
	if client == nil || prNum == 0 {
		return nil, nil, nil
	}
	raw, prObj, err := client.FetchDiff(ctx, prNum)
	if err != nil {
		return nil, nil, fmt.Errorf("fetch diff: %w", err)
	}
	files := diff.Parse(raw)
	paths := make([]string, 0, len(files))
	for _, f := range files {
		paths = append(paths, f.Path)
	}
	newBlobs, oldBlobs, _ := client.FileBlobs(ctx, prObj, paths)
	for i := range files {
		if v, ok := newBlobs[files[i].Path]; ok {
			files[i].NewBlob = v
		}
		if v, ok := oldBlobs[files[i].Path]; ok {
			files[i].OldBlob = v
		}
	}
	return files, &prSummary{
		Number: prObj.GetNumber(), Title: prObj.GetTitle(),
		HeadBranch: prObj.GetHead().GetRef(),
		HeadSHA:    prObj.GetHead().GetSHA(),
		URL:        prObj.GetHTMLURL(),
	}, nil
}

// loadReportIfNeeded loads the Playwright JSON report when heal mode is
// `on-failure`. Returns (nil, nil) otherwise.
func loadReportIfNeeded(cfg config.Config) (*heal.PlaywrightReport, error) {
	if cfg.HealMode != config.HealOnFailure {
		return nil, nil
	}
	path := cfg.PlaywrightReport
	if path == "" {
		path = filepath.Join(cfg.WorkDir, "playwright-report.json")
	}
	report, err := heal.LoadReport(path)
	if err != nil {
		return nil, fmt.Errorf("load report (%s): %w; pass --report or set REVIEWQA_HEAL_MODE=proactive", path, err)
	}
	return report, nil
}

func openHealPR(ctx context.Context, client *gh.Client, cfg config.Config, files []diff.File, prInfo *prSummary, edits []heal.Edit) error {
	indexed := map[string]string{}
	for _, f := range files {
		if f.NewBlob != "" {
			indexed[f.Path] = f.NewBlob
		}
	}
	patched := applyEdits(cfg.WorkDir, indexed, edits)
	branch := fmt.Sprintf("%s/heal-pr-%d-%s", cfg.BranchPrefix, cfg.PRNumber, shortSHA(prInfo.HeadSHA))
	url, err := client.OpenPR(ctx, gh.PROpts{
		BaseBranch: prInfo.HeadBranch, NewBranch: branch,
		Title: fmt.Sprintf("reviewqa: heal Playwright locators for PR #%d", cfg.PRNumber),
		Body:  healPRBody(prInfo, edits),
		Files: patched,
	})
	if err != nil {
		return err
	}
	rlog.Info("opened heal PR", "url", url)
	writeStepSummary(healSummary(prInfo, edits, url))
	return nil
}

type prSummary struct {
	Number     int
	Title      string
	HeadBranch string
	HeadSHA    string
	URL        string
}

func printRendered(rs []gen.Rendered) {
	for _, r := range rs {
		fmt.Println("---", r.Path, "---")
		fmt.Println(string(r.Content))
	}
}

func printEdits(es []heal.Edit) {
	for _, e := range es {
		fmt.Printf("%s:%d\n  - %s\n  + %s\n  reason: %s\n\n", e.File, e.Line, e.Before, e.After, e.Reason)
	}
}

func applyEdits(workDir string, indexed map[string]string, edits []heal.Edit) map[string][]byte {
	out := map[string][]byte{}
	bucket := map[string][]heal.Edit{}
	for _, e := range edits {
		bucket[e.File] = append(bucket[e.File], e)
	}
	for path, es := range bucket {
		content, ok := indexed[path]
		if !ok {
			b, err := os.ReadFile(filepath.Join(workDir, path))
			if err != nil {
				continue
			}
			content = string(b)
		}
		lines := strings.Split(content, "\n")
		for _, e := range es {
			if e.Line-1 < 0 || e.Line-1 >= len(lines) {
				continue
			}
			lines[e.Line-1] = e.After
		}
		out[path] = []byte(strings.Join(lines, "\n"))
	}
	return out
}

func genPRBody(pr *prSummary, rs []gen.Rendered) string {
	var b strings.Builder
	if pr != nil {
		fmt.Fprintf(&b, "Generated by reviewqa for #%d.\n\n", pr.Number)
	}
	b.WriteString("Files:\n")
	for _, r := range rs {
		fmt.Fprintf(&b, "- `%s` — covers `%s` (%s)\n", r.Path, r.Symbol.Name, r.Symbol.Language)
	}
	b.WriteString("\nEach scaffold contains one or more deterministic happy-path scenarios per component or symbol. Review and extend with edge cases.\n")
	return b.String()
}

func healPRBody(pr *prSummary, es []heal.Edit) string {
	var b strings.Builder
	if pr != nil {
		fmt.Fprintf(&b, "Generated by reviewqa for #%d.\n\n", pr.Number)
	}
	b.WriteString("Locator edits:\n")
	for _, e := range es {
		fmt.Fprintf(&b, "- `%s:%d` — %s\n", e.File, e.Line, e.Reason)
	}
	return b.String()
}

func shortSHA(sha string) string {
	if len(sha) > 7 {
		return sha[:7]
	}
	if sha == "" {
		return strconv.FormatInt(time.Now().Unix(), 10)
	}
	return sha
}

// writeStepSummary appends markdown to $GITHUB_STEP_SUMMARY when present.
// No-op outside GitHub Actions.
func writeStepSummary(md string) {
	p := os.Getenv("GITHUB_STEP_SUMMARY")
	if p == "" {
		return
	}
	f, err := os.OpenFile(p, os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return
	}
	defer f.Close()
	_, _ = f.WriteString(md)
}

func generateSummary(pr *prSummary, rs []gen.Rendered, url string) string {
	var b strings.Builder
	b.WriteString("### reviewqa — generated tests\n\n")
	if pr != nil {
		fmt.Fprintf(&b, "Source PR: [#%d](%s)\n\n", pr.Number, pr.URL)
	}
	fmt.Fprintf(&b, "Follow-up PR: %s\n\n", url)
	b.WriteString("| Test file | Covers | Language |\n")
	b.WriteString("|---|---|---|\n")
	for _, r := range rs {
		fmt.Fprintf(&b, "| `%s` | `%s` | %s |\n", r.Path, r.Symbol.Name, r.Symbol.Language)
	}
	b.WriteString("\n")
	return b.String()
}

func healSummary(pr *prSummary, es []heal.Edit, url string) string {
	var b strings.Builder
	b.WriteString("### reviewqa — locator heal\n\n")
	if pr != nil {
		fmt.Fprintf(&b, "Source PR: [#%d](%s)\n\n", pr.Number, pr.URL)
	}
	fmt.Fprintf(&b, "Heal PR: %s\n\n", url)
	b.WriteString("| File | Line | Reason |\n")
	b.WriteString("|---|---|---|\n")
	for _, e := range es {
		fmt.Fprintf(&b, "| `%s` | %d | %s |\n", e.File, e.Line, e.Reason)
	}
	b.WriteString("\n")
	return b.String()
}

// readPRFromEvent extracts the PR number from $GITHUB_EVENT_PATH if present.
func readPRFromEvent() int {
	p := os.Getenv("GITHUB_EVENT_PATH")
	if p == "" {
		return 0
	}
	b, err := os.ReadFile(p)
	if err != nil {
		return 0
	}
	var event struct {
		PullRequest struct {
			Number int `json:"number"`
		} `json:"pull_request"`
		Number int `json:"number"`
	}
	if err := json.Unmarshal(b, &event); err != nil {
		return 0
	}
	if event.PullRequest.Number != 0 {
		return event.PullRequest.Number
	}
	return event.Number
}
