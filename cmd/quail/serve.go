package main

import (
	"log"

	"github.com/spf13/cobra"

	"github.com/spriteCloud/quail/internal/serve"
)

// newServeCmd is the `quail serve` subcommand introduced in v0.65.
// It starts a localhost HTTP server that loads an existing
// quail-generated project and renders a read-only browser UI for
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
		Short: "Open a local browser UI for tailoring an existing quail-generated project.",
		Long: `Start a localhost HTTP server that loads a quail-generated
project (or a vanilla Playwright project) and renders the UI for
browsing, running, chat-editing and probing.

Scratch mode (v0.85): pass --workdir to a non-existent path (or
just run from any empty directory) and the server boots with no
project loaded. HOME shows the Probe + Import cards; once you
probe a URL, the result lands in ~/quail-projects/<brand>/ and
the UI auto-switches into it.`,
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
	cmd.Flags().StringVar(&workdir, "workdir", ".", "Root of the quail-generated project to load")
	cmd.Flags().StringVar(&addr, "addr", "127.0.0.1:8765", "Address to listen on (loopback only by default)")
	cmd.Flags().BoolVar(&noBrowser, "no-browser", false, "Skip auto-opening the default browser")
	return cmd
}
