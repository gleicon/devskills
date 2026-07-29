// Package doctor knows the external tools a few devskills skills shell out to and
// how to install each. It is data plus resolution logic; detection and execution
// are injected by the caller, so the command stays testable without touching the
// system.
package doctor

// Tool is an external binary a skill depends on, with the ways to install it.
type Tool struct {
	Name       string      // binary name (also the PATH lookup)
	Skill      string      // the /ds-* skill that needs it
	DocURL     string      // manual-install fallback
	Installers []Installer // tried in order; first applicable wins
}

// Installer is one way to install a Tool: a command that applies when its
// prerequisite binary is present and the platform matches.
type Installer struct {
	Requires string   // prerequisite binary that must be on PATH (brew, go, npm)
	GOOS     string   // "" for any platform, else the runtime.GOOS it applies to
	Command  []string // argv to run
}

// Tools is the external-tool catalog, carried over from the retired
// external-tools.sh: brew on macOS, otherwise the language toolchain.
var Tools = []Tool{
	{
		Name:   "osv-scanner",
		Skill:  "ds-osv",
		DocURL: "https://github.com/google/osv-scanner/releases",
		Installers: []Installer{
			{Requires: "brew", GOOS: "darwin", Command: []string{"brew", "install", "osv-scanner"}},
			{Requires: "go", Command: []string{"go", "install", "github.com/google/osv-scanner/cmd/osv-scanner@latest"}},
		},
	},
	{
		Name:   "ast-grep",
		Skill:  "ds-security-review",
		DocURL: "https://github.com/ast-grep/ast-grep/releases",
		Installers: []Installer{
			{Requires: "brew", GOOS: "darwin", Command: []string{"brew", "install", "ast-grep"}},
			{Requires: "npm", Command: []string{"npm", "install", "-g", "@ast-grep/cli"}},
		},
	},
	{
		Name:   "semgrep",
		Skill:  "ds-semgrep",
		DocURL: "https://semgrep.dev/docs/getting-started/",
		Installers: []Installer{
			{Requires: "brew", GOOS: "darwin", Command: []string{"brew", "install", "semgrep"}},
			{Requires: "pip", Command: []string{"pip", "install", "semgrep"}},
		},
	},
	{
		Name:   "tldt",
		Skill:  "ds-tldt",
		DocURL: "https://github.com/gleicon/tldt",
		Installers: []Installer{
			{Requires: "go", Command: []string{"go", "install", "github.com/gleicon/tldt/cmd/tldt@latest"}},
		},
	},
}

// Installer returns the first installer that applies on goos and whose
// prerequisite binary is present; ok is false when none fits (no package manager
// or toolchain for this platform).
func (t Tool) Installer(goos string, hasBinary func(string) bool) (Installer, bool) {
	for _, in := range t.Installers {
		if in.GOOS != "" && in.GOOS != goos {
			continue
		}
		if !hasBinary(in.Requires) {
			continue
		}
		return in, true
	}
	return Installer{}, false
}
