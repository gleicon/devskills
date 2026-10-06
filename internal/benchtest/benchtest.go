// Package benchtest keeps tests that drive assistant CLIs off the real ones.
package benchtest

import (
	"os"
	"path/filepath"
)

// ShimAssistants puts stand-ins for claude, codex and opencode first on PATH.
// Each fails loudly, so a test that forgets its fake fails instead of running,
// and billing, the real assistant. A fake a test installs goes in front of
// them. The caller removes the returned dir when its tests finish.
func ShimAssistants() (string, error) {
	dir, err := os.MkdirTemp("", "devskills-shim-*")
	if err != nil {
		return "", err
	}
	for _, name := range []string{"claude", "codex", "opencode"} {
		script := "#!/bin/sh\necho \"a test reached the " + name + " shim: install a fake\" >&2\nexit 97\n"
		if err := os.WriteFile(filepath.Join(dir, name), []byte(script), 0o755); err != nil {
			return dir, err
		}
	}
	return dir, os.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))
}
