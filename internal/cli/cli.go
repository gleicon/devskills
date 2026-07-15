// Package cli builds and runs the devskills command tree.
package cli

import (
	"context"

	"github.com/charmbracelet/fang"
	"github.com/spf13/cobra"
)

// NewRootCmd builds the devskills root command and its subcommands.
func NewRootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:          "devskills",
		Short:        "Skills and project scaffolding for AI coding assistants",
		SilenceUsage: true,
	}
	root.AddCommand(newVersionCmd())
	return root
}

// Execute runs the CLI. AnsiColorScheme uses the 16 ANSI colors, so help
// inherits the user's terminal theme instead of fang's hardcoded palette.
func Execute(ctx context.Context) error {
	return fang.Execute(ctx, NewRootCmd(),
		fang.WithVersion(buildVersion()),
		fang.WithColorSchemeFunc(fang.AnsiColorScheme),
	)
}
