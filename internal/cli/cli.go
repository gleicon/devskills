// Package cli builds and runs the devskills command tree.
package cli

import (
	"context"
	"io/fs"
	"os"
	"syscall"

	"github.com/charmbracelet/fang"
	"github.com/spf13/cobra"
)

// NewRootCmd builds the devskills root command and its subcommands. catalog is
// the embedded skill catalog (skills live under "skills/"), passed down from
// main to the commands that install from it.
func NewRootCmd(catalog fs.FS) *cobra.Command {
	root := &cobra.Command{
		Use:          "devskills",
		Short:        "Skills and project scaffolding for AI coding assistants",
		SilenceUsage: true,
	}
	root.AddCommand(newVersionCmd())
	root.AddCommand(newInstallCmd(catalog))
	root.AddCommand(newInitCmd(catalog))
	root.AddCommand(newConfigCmd(catalog))
	root.AddCommand(newDoctorCmd())
	root.AddCommand(newBenchCmd())
	return root
}

// Execute runs the CLI. AnsiColorScheme uses the 16 ANSI colors, so help
// inherits the user's terminal theme instead of fang's hardcoded palette.
func Execute(ctx context.Context, catalog fs.FS) error {
	return fang.Execute(ctx, NewRootCmd(catalog),
		fang.WithVersion(buildVersion()),
		fang.WithColorSchemeFunc(fang.AnsiColorScheme),
		// Cancel the context on SIGINT/SIGTERM instead of dying mid-run, so
		// long commands (bench) abort cleanly and their deferred sandbox
		// cleanup still runs.
		fang.WithNotifySignal(os.Interrupt, syscall.SIGTERM),
	)
}
