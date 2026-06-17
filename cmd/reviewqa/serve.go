package main

import (
	"log"

	"github.com/spf13/cobra"

	"github.com/reviewqa/reviewqa/internal/serve"
)

// newServeCmd is the `reviewqa serve` subcommand introduced in v0.65.
// It starts a localhost HTTP server that loads an existing
// reviewqa-generated project and renders a read-only browser UI for
// its Features, Scenarios, and stakeholder docs.
func newServeCmd() *cobra.Command {
	// Surface the binary version to the serve package so the topbar
	// pill can always show the running binary, not a hard-coded string.
	serve.BinaryVersion = version

	var workdir string
	var addr string
	var noBrowser bool
	cmd := &cobra.Command{
		Use:   "serve",
		Short: "Open a local browser UI for tailoring an existing reviewqa-generated project.",
		Long: `Start a localhost HTTP server that loads a reviewqa-generated
project (the kind ` + "`reviewqa probe`" + ` produces) and renders a
read-only UI for browsing every Feature, Scenario, and step. Useful
when you want to review the generated suite without diving into the
IDE — sidebar lists each .feature file, click any to see its Scenarios
rendered with brand styling. Stakeholder docs (catalogue, summary,
findings) are surfaced alongside.

This is Phase A of the serve workflow. Editing, locator suggestion,
and AI-composed step bindings ship in follow-up releases.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			return serve.Run(ctx, serve.Options{
				Workdir:   workdir,
				Addr:      addr,
				NoBrowser: noBrowser,
				Logf:      log.Printf,
			})
		},
	}
	cmd.Flags().StringVar(&workdir, "workdir", ".", "Root of the reviewqa-generated project to load")
	cmd.Flags().StringVar(&addr, "addr", "127.0.0.1:8765", "Address to listen on (loopback only by default)")
	cmd.Flags().BoolVar(&noBrowser, "no-browser", false, "Skip auto-opening the default browser")
	return cmd
}
