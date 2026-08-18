package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"runtime"
	"runtime/debug"
	"slices"
	"strconv"
	"strings"
	"time"

	"charm.land/lipgloss/v2"
	"github.com/spf13/cobra"
)

// version is overridden at release time via
// -ldflags "-X github.com/gleicon/devskills/internal/cli.version=<tag>"
// (goreleaser). Plain `go build`/`go install` leave it "dev", where
// resolveVersion falls back to the module's build info.
var version = "dev"

// latestReleaseURL is the GitHub API endpoint `version --check` queries —
// the only network call devskills itself makes (subprocesses like `doctor
// --fix` installers do their own).
const latestReleaseURL = "https://api.github.com/repos/gleicon/devskills/releases/latest"

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
	lines := []string{
		fmt.Sprintf("devskills %s", buildVersion()),
		fmt.Sprintf("  go: %s", runtime.Version()),
	}
	if info, ok := debug.ReadBuildInfo(); ok {
		for _, s := range info.Settings {
			switch s.Key {
			case "vcs.revision":
				lines = append(lines, fmt.Sprintf("  commit: %s", s.Value))
			case "vcs.time":
				lines = append(lines, fmt.Sprintf("  built: %s", s.Value))
			case "vcs.modified":
				if s.Value == "true" {
					lines = append(lines, "  dirty: true")
				}
			}
		}
	}
	return strings.Join(lines, "\n")
}

// releaseNums parses "v1.2.3" or "1.2.3" into its numeric fields. Dev builds
// and module pseudo-versions don't parse; callers treat that as "not comparable".
func releaseNums(s string) ([3]int, bool) {
	var n [3]int
	parts := strings.Split(strings.TrimPrefix(s, "v"), ".")
	if len(parts) != len(n) {
		return n, false
	}
	for i, p := range parts {
		v, err := strconv.Atoi(p)
		if err != nil || v < 0 {
			return [3]int{}, false
		}
		n[i] = v
	}
	return n, true
}

// fetchLatestTag asks the GitHub releases API for the newest release tag.
func fetchLatestTag(ctx context.Context, client *http.Client, url string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("github api: %s", resp.Status)
	}
	var release struct {
		TagName string `json:"tag_name"`
	}
	if err := json.NewDecoder(io.LimitReader(resp.Body, 1<<20)).Decode(&release); err != nil {
		return "", fmt.Errorf("decode github response: %w", err)
	}
	if _, ok := releaseNums(release.TagName); !ok {
		return "", fmt.Errorf("github api: unexpected release tag %q", release.TagName)
	}
	return release.TagName, nil
}

// releaseStatus renders the verdict of the current build vs the latest release tag.
func releaseStatus(current, latest string) string {
	latestV := "v" + strings.TrimPrefix(latest, "v")
	cur, ok := releaseNums(current)
	if !ok {
		return fmt.Sprintf("latest release: %s (current build %q is not a release — no comparison)", latestV, current)
	}
	currentV := "v" + strings.TrimPrefix(current, "v")
	lat, _ := releaseNums(latest)
	switch slices.Compare(cur[:], lat[:]) {
	case 0:
		return fmt.Sprintf("up to date: %s is the latest release", latestV)
	case -1:
		return fmt.Sprintf("newer release available: %s (current %s)\nupdate: go install github.com/gleicon/devskills@latest — or https://github.com/gleicon/devskills/releases/latest", latestV, currentV)
	default:
		return fmt.Sprintf("current %s is ahead of the latest release %s", currentV, latestV)
	}
}

func newVersionCmd() *cobra.Command {
	var check bool
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Print version and build info",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			lipgloss.Fprintln(cmd.OutOrStdout(), versionInfo())
			if !check {
				return nil
			}
			client := &http.Client{Timeout: 10 * time.Second}
			tag, err := fetchLatestTag(cmd.Context(), client, latestReleaseURL)
			if err != nil {
				return fmt.Errorf("check latest release: %w", err)
			}
			lipgloss.Fprintln(cmd.OutOrStdout(), releaseStatus(buildVersion(), tag))
			return nil
		},
	}
	cmd.Flags().BoolVar(&check, "check", false, "check GitHub for a newer release (network call)")
	return cmd
}
