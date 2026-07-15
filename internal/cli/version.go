package cli

import (
	"fmt"
	"runtime"
	"runtime/debug"
	"strings"

	"github.com/spf13/cobra"
)

// version is overridden at release time via
// -ldflags "-X github.com/gleicon/devskills/internal/cli.version=<tag>"
// (goreleaser). Plain `go build`/`go install` leave it "dev", where
// resolveVersion falls back to the module's build info.
var version = "dev"

// resolveVersion picks the version string: an ldflags-injected tag wins;
// otherwise the module version from build info; otherwise "dev".
func resolveVersion(ldflags string, info *debug.BuildInfo) string {
	if ldflags != "dev" {
		return ldflags
	}
	if info != nil && info.Main.Version != "" && info.Main.Version != "(devel)" {
		return info.Main.Version
	}
	return ldflags
}

func buildVersion() string {
	info, _ := debug.ReadBuildInfo()
	return resolveVersion(version, info)
}

// versionInfo renders the version plus any VCS metadata the build recorded.
func versionInfo() string {
	var b strings.Builder
	fmt.Fprintf(&b, "devskills %s\n", buildVersion())
	fmt.Fprintf(&b, "  go: %s\n", runtime.Version())
	if info, ok := debug.ReadBuildInfo(); ok {
		for _, s := range info.Settings {
			switch s.Key {
			case "vcs.revision":
				fmt.Fprintf(&b, "  commit: %s\n", s.Value)
			case "vcs.time":
				fmt.Fprintf(&b, "  built: %s\n", s.Value)
			case "vcs.modified":
				if s.Value == "true" {
					b.WriteString("  dirty: true\n")
				}
			}
		}
	}
	return strings.TrimRight(b.String(), "\n")
}

func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print version and build info",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			cmd.Println(versionInfo())
			return nil
		},
	}
}
